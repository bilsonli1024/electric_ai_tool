package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type CopywritingHandler struct {
	copywritingService *services.CopywritingService
	authService        *services.AuthService
	unifiedTaskService *services.UnifiedTaskService
}

func NewCopywritingHandler(copywritingService *services.CopywritingService, authService *services.AuthService, unifiedTaskService *services.UnifiedTaskService) *CopywritingHandler {
	return &CopywritingHandler{
		copywritingService: copywritingService,
		authService:        authService,
		unifiedTaskService: unifiedTaskService,
	}
}

func (h *CopywritingHandler) AnalyzeCompetitors(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		log.Printf("AnalyzeCompetitors error: user_id not found in context")
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		log.Printf("AnalyzeCompetitors error: user_id is not int64, got %T", userIDValue)
		utils.RespondError(w, fmt.Errorf("invalid user_id"), http.StatusUnauthorized)
		return
	}

	var req models.AnalyzeCompetitorsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		utils.RespondError(w, fmt.Errorf("URLs are required"), http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	if req.TaskName == "" {
		req.TaskName = fmt.Sprintf("文案任务_%d", time.Now().Unix())
	}

	// 获取用户名
	username := r.Header.Get("X-Username")
	if username == "" {
		user, err := h.authService.GetUserByID(userID)
		if err == nil && user != nil {
			username = user.Username
		}
	}

	// 创建旧格式任务（保持向后兼容）
	taskID, err := h.copywritingService.CreateTask(userID, req.URLs, req.Model, req.TaskName)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 创建统一任务
	configJSON, _ := json.Marshal(map[string]interface{}{
		"urls": req.URLs,
	})
	unifiedTask := &models.UnifiedTask{
		UserID:        userID,
		Username:      username,
		TaskName:      req.TaskName,
		TaskType:      "copywriting",
		Status:        0, // 分析中
		TaskConfig:    string(configJSON),
		AnalyzeModel:  req.Model,
		GenerateModel: req.Model,
	}
	unifiedTaskID, err := h.unifiedTaskService.CreateTask(unifiedTask)
	if err != nil {
		log.Printf("Failed to create unified task: %v", err)
		// 不影响主流程，继续执行
	} else {
		log.Printf("Created unified task: id=%d, type=copywriting, name=%s", unifiedTaskID, req.TaskName)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	analysis, err := h.copywritingService.AnalyzeCompetitors(ctx, req.URLs, req.Model)
	if err != nil {
		h.copywritingService.UpdateTaskStatus(taskID, models.CopyStatusAnalyzeFailed, err.Error())
		if unifiedTaskID > 0 {
			h.unifiedTaskService.UpdateTaskStatus(unifiedTaskID, 10, err.Error()) // 分析失败
		}
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	if err := h.copywritingService.SaveAnalysisResult(taskID, analysis); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 更新统一任务状态为分析完成
	if unifiedTaskID > 0 {
		analysisJSON, _ := json.Marshal(analysis)
		h.unifiedTaskService.UpdateTaskResult(unifiedTaskID, string(analysisJSON), "", 1) // 分析完成
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":    analysis,
		"task_id": taskID,
		"unified_task_id": unifiedTaskID,
	})
}

func (h *CopywritingHandler) GenerateCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TaskID                    int64                      `json:"task_id"`
		SelectedKeywords          []string                   `json:"selectedKeywords"`
		SelectedSellingPoints     []string                   `json:"selectedSellingPoints"`
		SelectedReviewInsights    []string                   `json:"selectedReviewInsights"`
		SelectedImageInsights     []string                   `json:"selectedImageInsights"`
		ProductDetails            models.ProductDetails      `json:"productDetails"`
		Model                     string                     `json:"model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.TaskID == 0 {
		utils.RespondError(w, fmt.Errorf("task_id is required"), http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	h.copywritingService.UpdateTaskStatus(req.TaskID, models.CopyStatusGenerating, "")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	copyReq := models.GenerateCopyRequest{
		SelectedKeywords:       req.SelectedKeywords,
		SelectedSellingPoints:  req.SelectedSellingPoints,
		SelectedReviewInsights: req.SelectedReviewInsights,
		SelectedImageInsights:  req.SelectedImageInsights,
		ProductDetails:         req.ProductDetails,
		Model:                  req.Model,
	}

	copy, err := h.copywritingService.GenerateCopy(ctx, copyReq)
	if err != nil {
		h.copywritingService.UpdateTaskStatus(req.TaskID, models.CopyStatusGenerateFailed, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	if err := h.copywritingService.SaveGeneratedCopy(req.TaskID, copy, &req.ProductDetails); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":    copy,
		"task_id": req.TaskID,
	})
}

func (h *CopywritingHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		log.Printf("GetTasks error: user_id not found in context")
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		log.Printf("GetTasks error: user_id is not int64, got %T", userIDValue)
		utils.RespondError(w, fmt.Errorf("invalid user_id"), http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	tasks, total, err := h.copywritingService.GetUserTasks(userID, limit, offset)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  tasks,
		"total": total,
	})
}

func (h *CopywritingHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		utils.RespondError(w, fmt.Errorf("task_id is required"), http.StatusBadRequest)
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	task, err := h.copywritingService.GetTaskByID(taskID)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": task,
	})
}

func (h *CopywritingHandler) SearchTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		log.Printf("SearchTasks error: user_id not found in context")
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}
	
	userID, ok := userIDValue.(int64)
	if !ok {
		log.Printf("SearchTasks error: user_id is not int64, got %T", userIDValue)
		utils.RespondError(w, fmt.Errorf("invalid user_id"), http.StatusUnauthorized)
		return
	}

	keyword := r.URL.Query().Get("keyword")
	limitStr := r.URL.Query().Get("limit")

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	tasks, err := h.copywritingService.SearchCompletedTasks(userID, keyword, limit)
	if err != nil {
		log.Printf("SearchTasks error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = []*models.CopywritingTask{}
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": tasks,
	})
}
