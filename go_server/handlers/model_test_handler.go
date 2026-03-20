package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type ModelTestHandler struct {
	multiModelService *services.MultiModelService
}

func NewModelTestHandler(multiModelService *services.MultiModelService) *ModelTestHandler {
	return &ModelTestHandler{
		multiModelService: multiModelService,
	}
}

type ModelTestRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ModelTestResponse struct {
	Success      bool   `json:"success"`
	Model        string `json:"model"`
	Response     string `json:"response,omitempty"`
	Error        string `json:"error,omitempty"`
	ResponseTime int64  `json:"response_time_ms"`
}

func (h *ModelTestHandler) TestModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ModelTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	if req.Prompt == "" {
		req.Prompt = "Hello, please respond with 'Connection successful' in English."
	}

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.multiModelService.TestChat(ctx, req.Model, req.Prompt)
	responseTime := time.Since(startTime).Milliseconds()

	if err != nil {
		utils.RespondJSON(w, ModelTestResponse{
			Success:      false,
			Model:        req.Model,
			Error:        err.Error(),
			ResponseTime: responseTime,
		})
		return
	}

	utils.RespondJSON(w, ModelTestResponse{
		Success:      true,
		Model:        req.Model,
		Response:     result,
		ResponseTime: responseTime,
	})
}

func (h *ModelTestHandler) TestAllModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		req.Prompt = "Hello, please respond with 'Connection successful' in English."
	}

	modelsList := []string{
		models.ModelGemini,
		models.ModelGPT,
		models.ModelDeepSeek,
	}

	results := make([]ModelTestResponse, 0, len(modelsList))

	for _, model := range modelsList {
		startTime := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		result, err := h.multiModelService.TestChat(ctx, model, req.Prompt)
		responseTime := time.Since(startTime).Milliseconds()
		cancel()

		if err != nil {
			results = append(results, ModelTestResponse{
				Success:      false,
				Model:        model,
				Error:        err.Error(),
				ResponseTime: responseTime,
			})
		} else {
			results = append(results, ModelTestResponse{
				Success:      true,
				Model:        model,
				Response:     result,
				ResponseTime: responseTime,
			})
		}
	}

	utils.RespondJSON(w, map[string]interface{}{
		"results": results,
	})
}
