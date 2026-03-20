package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type ModelDomain struct {
	multiModelService *services.MultiModelService
}

func NewModelDomain(multiModelService *services.MultiModelService) *ModelDomain {
	return &ModelDomain{
		multiModelService: multiModelService,
	}
}

func (d *ModelDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	modelTestHandler := handlers.NewModelTestHandler(d.multiModelService)

	http.HandleFunc("/api/models/test", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(modelTestHandler.TestModel))))
	http.HandleFunc("/api/models/test-all", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(modelTestHandler.TestAllModels))))
}
