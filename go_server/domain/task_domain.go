package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type TaskDomain struct {
	multiModelService  *services.MultiModelService
	taskService        *services.TaskService
	taskHistoryService *services.TaskHistoryService
	cdnService         *services.CDNService
	authService        *services.AuthService
	unifiedTaskService *services.UnifiedTaskService
	taskCenterService  *services.TaskCenterService
	imageTaskService   *services.ImageTaskService
}

func NewTaskDomain(
	multiModelService *services.MultiModelService,
	taskService *services.TaskService,
	taskHistoryService *services.TaskHistoryService,
	cdnService *services.CDNService,
	authService *services.AuthService,
	unifiedTaskService *services.UnifiedTaskService,
	taskCenterService *services.TaskCenterService,
	imageTaskService *services.ImageTaskService,
) *TaskDomain {
	return &TaskDomain{
		multiModelService:  multiModelService,
		taskService:        taskService,
		taskHistoryService: taskHistoryService,
		cdnService:         cdnService,
		authService:        authService,
		unifiedTaskService: unifiedTaskService,
		taskCenterService:  taskCenterService,
		imageTaskService:   imageTaskService,
	}
}

func (d *TaskDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	taskHandler := handlers.NewTaskHandler(
		d.multiModelService,
		d.taskService,
		d.taskHistoryService,
		d.cdnService,
		d.authService,
		d.unifiedTaskService,
		d.taskCenterService,
		d.imageTaskService,
	)

	http.HandleFunc("/api/tasks/analyze", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.AnalyzeWithTask))))
	http.HandleFunc("/api/tasks/generate-image", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GenerateImageWithTask))))
	http.HandleFunc("/api/tasks", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetTasks))))
	http.HandleFunc("/api/tasks/all", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetAllTasks))))
	http.HandleFunc("/api/tasks/history", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetTaskHistory))))
}
