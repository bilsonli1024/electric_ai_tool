package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type UnifiedTaskDomain struct {
	taskService *services.UnifiedTaskService
	authService *services.AuthService
}

func NewUnifiedTaskDomain(taskService *services.UnifiedTaskService, authService *services.AuthService) *UnifiedTaskDomain {
	return &UnifiedTaskDomain{
		taskService: taskService,
		authService: authService,
	}
}

func (d *UnifiedTaskDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	handler := handlers.NewUnifiedTaskHandler(d.taskService, d.authService)

	// 任务列表（支持筛选）
	http.HandleFunc("/api/unified-tasks", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetTasks))))
	
	// 任务详情
	http.HandleFunc("/api/unified-tasks/detail", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetTaskByID))))
	
	// 任务统计
	http.HandleFunc("/api/unified-tasks/statistics", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetTaskStatistics))))
	
	// 创建任务
	http.HandleFunc("/api/unified-tasks/create", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.CreateTask))))
}
