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
		filter.TaskType = taskType
	}

	if taskStatus := query.Get("task_status"); taskStatus != "" {
		filter.TaskStatus = taskStatus
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
	if !viewAll && filter.Operator == "" {
		// 默认只查看自己的任务
		filter.Operator = user.Email
	}

	// 查询任务
	tasks, total, err := h.taskCenterService.GetTasks(filter)
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
	statuses := []string{
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
		_, total, err := h.taskCenterService.GetTasks(filter)
		if err != nil {
			log.Printf("GetTaskStatistics error for status %s: %v", status, err)
			continue
		}
		stats[status] = total
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": stats,
	})
}
