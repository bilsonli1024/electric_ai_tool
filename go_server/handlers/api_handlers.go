package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type Handler struct {
	aiService *services.AIService
}

func NewHandler(aiService *services.AIService) *Handler {
	return &Handler{aiService: aiService}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}
	utils.RespondJSON(w, models.HealthResponse{Status: "ok", Port: port})
}

func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	sellingPoints, err := h.aiService.AnalyzeSellingPoints(ctx, req)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.AnalyzeResponse{Data: sellingPoints})
}

func (h *Handler) GenerateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.GenerateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	dataURL, err := h.aiService.GenerateImage(ctx, req)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.ImageResponse{Data: dataURL})
}

func (h *Handler) EditImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.EditImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	dataURL, err := h.aiService.EditImage(ctx, req)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.ImageResponse{Data: dataURL})
}

func (h *Handler) APlusContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.APlusContentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	modules, err := h.aiService.GenerateAPlusContent(ctx, req)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.APlusContentResponse{Data: modules})
}
