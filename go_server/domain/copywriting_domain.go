package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type CopywritingDomain struct {
	copywritingService     *services.CopywritingService
	authService            *services.AuthService
	unifiedTaskService     *services.UnifiedTaskService
	taskCenterService      *services.TaskCenterService
	copywritingTaskService *services.CopywritingTaskService
}

func NewCopywritingDomain(
	copywritingService *services.CopywritingService,
	authService *services.AuthService,
	unifiedTaskService *services.UnifiedTaskService,
	taskCenterService *services.TaskCenterService,
	copywritingTaskService *services.CopywritingTaskService,
) *CopywritingDomain {
	return &CopywritingDomain{
		copywritingService:     copywritingService,
		authService:            authService,
		unifiedTaskService:     unifiedTaskService,
		taskCenterService:      taskCenterService,
		copywritingTaskService: copywritingTaskService,
	}
}

func (d *CopywritingDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	copywritingHandler := handlers.NewCopywritingHandler(
		d.copywritingService,
		d.authService,
		d.unifiedTaskService,
		d.taskCenterService,
		d.copywritingTaskService,
	)

	http.HandleFunc("/api/copywriting/analyze", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(copywritingHandler.AnalyzeCompetitors))))
	http.HandleFunc("/api/copywriting/generate", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(copywritingHandler.GenerateCopy))))
	http.HandleFunc("/api/copywriting/tasks", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(copywritingHandler.GetTasks))))
	http.HandleFunc("/api/copywriting/task", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(copywritingHandler.GetTask))))
	http.HandleFunc("/api/copywriting/search", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(copywritingHandler.SearchTasks))))
}
