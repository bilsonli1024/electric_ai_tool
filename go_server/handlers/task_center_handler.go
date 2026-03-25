package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type TaskCenterHandler struct {
	taskCenterService      *services.TaskCenterService
	authService            *services.AuthService
}

func NewTaskCenterHandler(
	taskCenterService *services.TaskCenterService,
	authService *services.AuthService,
) *TaskCenterHandler {
	return &TaskCenterHandler{
		taskCenterService: taskCenterService,
		authService:       authService,
	}
}

// GetTasks 获取任务列表
func (h *TaskCenterHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户信息
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

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to get user info"), http.StatusInternalServerError)
		return
	}

	// 解析查询参数
	query := r.URL.Query()
	filter := models.TaskCenterFilter{
		Limit:  20,
		Offset: 0,
	}

	// 分页参数 - 支持page_size/page_no和limit/offset两种方式
	pageSize := 20
	pageNo := 1
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	} else if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			pageSize = limit
		}
	}
	
	if pageNoStr := query.Get("page_no"); pageNoStr != "" {
		if pn, err := strconv.Atoi(pageNoStr); err == nil && pn > 0 {
			pageNo = pn
		}
	} else if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			pageNo = (offset / pageSize) + 1
		}
	}
	
	filter.Limit = pageSize
	filter.Offset = (pageNo - 1) * pageSize

	// 筛选条件
	if taskType := query.Get("task_type"); taskType != "" {
		if tt, err := strconv.Atoi(taskType); err == nil {
			filter.TaskType = tt
		}
	}

	if taskStatus := query.Get("task_status"); taskStatus != "" {
		if ts, err := strconv.Atoi(taskStatus); err == nil {
			filter.TaskStatus = ts
		}
	}
	
	if operator := query.Get("operator"); operator != "" {
		filter.Operator = operator
	}

	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil {
			filter.StartTime = startTime
		}
	}

	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil {
			filter.EndTime = endTime
		}
	}

	// 是否查看所有任务（管理员）
	viewAll := query.Get("view_all") == "true"
	
	// 如果不是查看所有任务，且用户没有指定operator筛选
	if !viewAll && filter.Operator == "" {
		// 默认只查看自己的任务（使用用户邮箱作为operator）
		filter.Operator = user.Email
		log.Printf("GetTasks: filtering by operator=%s (user email) for user_id=%d", filter.Operator, userID)
	} else if viewAll {
		log.Printf("GetTasks: view_all=true, showing all tasks")
	} else {
		log.Printf("GetTasks: using specified operator filter=%s", filter.Operator)
	}

	log.Printf("GetTasks: final filter=%+v, user.Email=%s", filter, user.Email)

	// 查询任务
	tasks, total, err := h.taskCenterService.ListTasks(filter)
	if err != nil {
		log.Printf("ListTasks error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	log.Printf("GetTasks: found %d tasks, total=%d", len(tasks), total)

	utils.RespondJSON(w, map[string]interface{}{
		"data":  tasks,
		"total": total,
	})
}

// GetTaskDetail 获取任务详情
func (h *TaskCenterHandler) GetTaskDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		utils.RespondError(w, fmt.Errorf("task_id is required"), http.StatusBadRequest)
		return
	}

	detail, err := h.taskCenterService.GetTaskDetail(taskID)
	if err != nil {
		log.Printf("GetTaskDetail error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": detail,
	})
}

// GetTaskStatistics 获取任务统计
func (h *TaskCenterHandler) GetTaskStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户信息
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}

	userID, ok := userIDValue.(int64)
	if !ok {
		utils.RespondError(w, fmt.Errorf("invalid user_id"), http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to get user info"), http.StatusInternalServerError)
		return
	}

	// 统计不同状态的任务数量
	stats := make(map[string]int)
	statuses := []int{
		models.TaskStatusPending,
		models.TaskStatusOngoing,
		models.TaskStatusCompleted,
		models.TaskStatusFailed,
	}

	for _, status := range statuses {
		filter := models.TaskCenterFilter{
			Operator:   user.Email,
			TaskStatus: status,
			Limit:      1,
		}
		_, total, err := h.taskCenterService.ListTasks(filter)
		if err != nil {
			log.Printf("GetTaskStatistics error for status %d: %v", status, err)
			continue
		}
		stats[models.TaskStatusToString(status)] = total
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": stats,
	})
}

// CopyTask 复制任务
func (h *TaskCenterHandler) CopyTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		utils.RespondError(w, fmt.Errorf("task_id is required"), http.StatusBadRequest)
		return
	}

	// 获取用户信息
	sessionID := r.Header.Get("Authorization")
	if sessionID != "" && len(sessionID) > 7 {
		sessionID = sessionID[7:]
	}

	userID, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to get user info"), http.StatusUnauthorized)
		return
	}

	// 获取原任务详情
	detail, err := h.taskCenterService.GetTaskDetail(taskID)
	if err != nil {
		log.Printf("CopyTask GetTaskDetail error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 生成新任务ID
	newTaskID := h.taskCenterService.GenerateTaskID(detail.TaskType)

	// 创建新的任务中心记录
	if err := h.taskCenterService.CreateBaseTask(newTaskID, detail.TaskType, user.Email); err != nil {
		log.Printf("CopyTask CreateBaseTask error: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 根据任务类型复制详细数据
	if detail.TaskType == models.TaskTypeCopywriting {
		copyDetail := detail.DetailData.(*models.CopywritingTaskDetail)
		if err := h.taskCenterService.CopyCopywritingTask(newTaskID, copyDetail); err != nil {
			log.Printf("CopyTask CopyCopywritingTask error: %v", err)
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
	} else if detail.TaskType == models.TaskTypeImage {
		imageDetail := detail.DetailData.(*models.ImageTaskDetail)
		if err := h.taskCenterService.CopyImageTask(newTaskID, imageDetail); err != nil {
			log.Printf("CopyTask CopyImageTask error: %v", err)
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Task copied successfully: %s -> %s", taskID, newTaskID)
	utils.RespondJSON(w, map[string]interface{}{
		"task_id": newTaskID,
		"message": "任务复制成功",
	})
}

