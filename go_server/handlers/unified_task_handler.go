package handlers

import (
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

type UnifiedTaskHandler struct {
	taskService *services.UnifiedTaskService
	authService *services.AuthService
}

func NewUnifiedTaskHandler(taskService *services.UnifiedTaskService, authService *services.AuthService) *UnifiedTaskHandler {
	return &UnifiedTaskHandler{
		taskService: taskService,
		authService: authService,
	}
}

// getUserInfo 从context中获取用户信息
func (h *UnifiedTaskHandler) getUserInfo(r *http.Request) (int64, string, error) {
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		return 0, "", fmt.Errorf("user_id not found in context")
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		return 0, "", fmt.Errorf("invalid user_id type")
	}

	// 获取用户名（从header或者查询数据库）
	username := r.Header.Get("X-Username")
	if username == "" {
		// 如果header中没有，从数据库查询
		user, err := h.authService.GetUserByID(userID)
		if err == nil && user != nil {
			username = user.Username
		}
	}

	return userID, username, nil
}

// GetTasks 获取任务列表（支持多条件筛选）
func (h *UnifiedTaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _, err := h.getUserInfo(r)
	if err != nil {
		log.Printf("GetTasks error: %v", err)
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	// 解析查询参数
	query := r.URL.Query()
	
	filter := models.TaskFilter{
		Limit:  20,
		Offset: 0,
	}

	// limit和offset
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// 任务类型筛选
	if taskType := query.Get("task_type"); taskType != "" {
		filter.TaskType = &taskType
	}

	// 状态筛选
	if statusStr := query.Get("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			filter.Status = &status
		}
	}

	// 时间范围筛选
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			filter.StartTime = &startTime
		}
	}
	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			filter.EndTime = &endTime
		}
	}

	// 是否查看所有用户的任务
	viewAll := query.Get("view_all") == "true"
	if !viewAll {
		filter.UserID = &userID
	}

	tasks, total, err := h.taskService.GetTasks(filter)
	if err != nil {
		log.Printf("GetTasks error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  tasks,
		"total": total,
	})
}

// GetTaskByID 根据ID获取任务详情
func (h *UnifiedTaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, _, err := h.getUserInfo(r)
	if err != nil {
		log.Printf("GetTaskByID error: %v", err)
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	taskIDStr := r.URL.Query().Get("id")
	if taskIDStr == "" {
		utils.RespondError(w, fmt.Errorf("task id is required"), http.StatusBadRequest)
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("invalid task id"), http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		log.Printf("GetTaskByID error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": task,
	})
}

// GetTaskStatistics 获取任务统计信息
func (h *UnifiedTaskHandler) GetTaskStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _, err := h.getUserInfo(r)
	if err != nil {
		log.Printf("GetTaskStatistics error: %v", err)
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	// 是否查看所有用户的统计
	var userIDPtr *int64
	if r.URL.Query().Get("view_all") != "true" {
		userIDPtr = &userID
	}

	stats, err := h.taskService.GetTaskStatistics(userIDPtr)
	if err != nil {
		log.Printf("GetTaskStatistics error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": stats,
	})
}

// CreateTask 创建任务（通用接口）
func (h *UnifiedTaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, username, err := h.getUserInfo(r)
	if err != nil {
		log.Printf("CreateTask error: %v", err)
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	var req struct {
		TaskName      string                 `json:"task_name"`
		TaskType      string                 `json:"task_type"` // copywriting, image
		TaskConfig    map[string]interface{} `json:"task_config"`
		AnalyzeModel  string                 `json:"analyze_model"`
		GenerateModel string                 `json:"generate_model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	// 验证任务类型
	if req.TaskType != "copywriting" && req.TaskType != "image" {
		utils.RespondError(w, fmt.Errorf("invalid task_type"), http.StatusBadRequest)
		return
	}

	// 默认模型
	if req.AnalyzeModel == "" {
		req.AnalyzeModel = "gemini"
	}
	if req.GenerateModel == "" {
		req.GenerateModel = "gemini"
	}

	// 将配置转为JSON
	configJSON, err := json.Marshal(req.TaskConfig)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	task := &models.UnifiedTask{
		UserID:        userID,
		Username:      username,
		TaskName:      req.TaskName,
		TaskType:      req.TaskType,
		Status:        0, // 初始状态：分析中
		TaskConfig:    string(configJSON),
		AnalyzeModel:  req.AnalyzeModel,
		GenerateModel: req.GenerateModel,
	}

	taskID, err := h.taskService.CreateTask(task)
	if err != nil {
		log.Printf("CreateTask error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"task_id": taskID,
		"message": "任务创建成功",
	})
}
