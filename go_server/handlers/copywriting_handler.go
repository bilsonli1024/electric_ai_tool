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
	copywritingService     *services.CopywritingService
	authService            *services.AuthService
	unifiedTaskService     *services.UnifiedTaskService
	taskCenterService      *services.TaskCenterService
	copywritingTaskService *services.CopywritingTaskService
}

func NewCopywritingHandler(
	copywritingService *services.CopywritingService,
	authService *services.AuthService,
	unifiedTaskService *services.UnifiedTaskService,
	taskCenterService *services.TaskCenterService,
	copywritingTaskService *services.CopywritingTaskService,
) *CopywritingHandler {
	return &CopywritingHandler{
		copywritingService:     copywritingService,
		authService:            authService,
		unifiedTaskService:     unifiedTaskService,
		taskCenterService:      taskCenterService,
		copywritingTaskService: copywritingTaskService,
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

	// 获取用户邮箱
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to get user info"), http.StatusInternalServerError)
		return
	}
	operator := user.Email

	// 1. 生成全局唯一task_id
	taskID := h.taskCenterService.GenerateTaskID(models.TaskTypeCopywriting)
	
	// 2. 创建任务中心底表记录
	if err := h.taskCenterService.CreateBaseTask(taskID, models.TaskTypeCopywriting, operator); err != nil {
		log.Printf("Failed to create base task: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	
	// 3. 创建文案任务详细表记录
	urlsJSON, _ := json.Marshal(req.URLs)
	if err := h.copywritingTaskService.CreateTask(taskID, req.TaskName, string(urlsJSON), req.Model); err != nil {
		log.Printf("Failed to create copywriting task: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	
	log.Printf("Created copywriting task: task_id=%s, operator=%s", taskID, operator)

	// 4. 更新状态为进行中
	h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusOngoing)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 5. 执行分析
	analysis, err := h.copywritingService.AnalyzeCompetitors(ctx, req.URLs, req.Model)
	if err != nil {
		h.copywritingTaskService.SaveError(taskID, err.Error())
		h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 6. 保存分析结果
	analysisJSON, _ := json.Marshal(analysis)
	if err := h.copywritingTaskService.SaveAnalysisResult(taskID, string(analysisJSON)); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 7. 保持状态为ongoing，等待用户选择数据并生成文案
	// 状态会在GenerateCopy时更新为completed

	utils.RespondJSON(w, map[string]interface{}{
		"data":    analysis,
		"task_id": taskID,
	})
}

func (h *CopywritingHandler) GenerateCopy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TaskID                    string                     `json:"task_id"` // 改为string类型
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

	if req.TaskID == "" {
		utils.RespondError(w, fmt.Errorf("task_id is required"), http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. 保存用户选择的数据
	userSelectedJSON, _ := json.Marshal(map[string]interface{}{
		"selectedKeywords":       req.SelectedKeywords,
		"selectedSellingPoints":  req.SelectedSellingPoints,
		"selectedReviewInsights": req.SelectedReviewInsights,
		"selectedImageInsights":  req.SelectedImageInsights,
	})
	productDetailsJSON, _ := json.Marshal(req.ProductDetails)
	
	if err := h.copywritingTaskService.SaveUserSelectedData(req.TaskID, string(userSelectedJSON), string(productDetailsJSON)); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 2. 生成文案
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
		h.copywritingTaskService.SaveError(req.TaskID, err.Error())
		h.taskCenterService.UpdateTaskStatus(req.TaskID, models.TaskStatusFailed)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 3. 保存生成的文案
	copyJSON, _ := json.Marshal(copy)
	if err := h.copywritingTaskService.SaveGeneratedCopy(req.TaskID, string(copyJSON), req.Model); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 4. 更新任务状态为已完成
	h.taskCenterService.UpdateTaskStatus(req.TaskID, models.TaskStatusCompleted)

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
