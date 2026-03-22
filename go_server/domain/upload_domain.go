package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type UploadDomain struct {
	localStorageService *services.LocalStorageService
	authService         *services.AuthService
}

func NewUploadDomain(
	localStorageService *services.LocalStorageService,
	authService *services.AuthService,
) *UploadDomain {
	return &UploadDomain{
		localStorageService: localStorageService,
		authService:         authService,
	}
}

func (d *UploadDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	handler := handlers.NewUploadHandler(d.localStorageService, d.authService)

	// 上传图片接口
	http.HandleFunc("/api/upload/image", 
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.UploadImage))))
	
	// 上传base64图片接口
	http.HandleFunc("/api/upload/image-base64", 
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.UploadImageBase64))))
}
