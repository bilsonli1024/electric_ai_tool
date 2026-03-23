package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type TaskDomain struct {
	multiModelService   *services.MultiModelService
	authService         *services.AuthService
	taskCenterService   *services.TaskCenterService
	imageTaskService    *services.ImageTaskService
	localStorageService *services.LocalStorageService
}

func NewTaskDomain(
	multiModelService *services.MultiModelService,
	authService *services.AuthService,
	taskCenterService *services.TaskCenterService,
	imageTaskService *services.ImageTaskService,
	localStorageService *services.LocalStorageService,
) *TaskDomain {
	return &TaskDomain{
		multiModelService:   multiModelService,
		authService:         authService,
		taskCenterService:   taskCenterService,
		imageTaskService:    imageTaskService,
		localStorageService: localStorageService,
	}
}

func (d *TaskDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	taskHandler := handlers.NewTaskHandler(
		d.multiModelService,
		d.authService,
		d.taskCenterService,
		d.imageTaskService,
		d.localStorageService,
	)

	http.HandleFunc("/api/tasks/analyze", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.AnalyzeWithTask))))
	http.HandleFunc("/api/tasks/generate-image", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GenerateImageWithTask))))
	http.HandleFunc("/api/tasks", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetTasks))))
	http.HandleFunc("/api/tasks/all", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetAllTasks))))
	http.HandleFunc("/api/tasks/history", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(taskHandler.GetTaskHistory))))
}
