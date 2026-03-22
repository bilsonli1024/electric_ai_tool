package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type TaskCenterDomain struct {
	taskCenterService *services.TaskCenterService
	authService       *services.AuthService
}

func NewTaskCenterDomain(
	taskCenterService *services.TaskCenterService,
	authService *services.AuthService,
) *TaskCenterDomain {
	return &TaskCenterDomain{
		taskCenterService: taskCenterService,
		authService:       authService,
	}
}

func (d *TaskCenterDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	handler := handlers.NewTaskCenterHandler(d.taskCenterService, d.authService)

	// 任务列表
	http.HandleFunc("/api/task-center/list", 
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetTasks))))
	
	// 任务详情
	http.HandleFunc("/api/task-center/detail", 
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetTaskDetail))))
	
	// 任务统计
	http.HandleFunc("/api/task-center/statistics", 
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetTaskStatistics))))
	
	// 复制任务
	http.HandleFunc("/api/task-center/copy", 
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.CopyTask))))
}
