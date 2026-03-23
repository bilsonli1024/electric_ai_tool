package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
)

type EnumDomain struct{}

func NewEnumDomain() *EnumDomain {
	return &EnumDomain{}
}

func (d *EnumDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	handler := handlers.NewEnumHandler()

	// 获取所有枚举（需要登录）
	http.HandleFunc("/api/domain/enums",
		middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(handler.GetAllEnums))))
}
