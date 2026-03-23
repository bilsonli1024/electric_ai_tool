package domain

import (
	"net/http"

	"electric_ai_tool/go_server/handlers"
	"electric_ai_tool/go_server/middleware"
	"electric_ai_tool/go_server/services"
)

type AuthDomain struct {
	authService  *services.AuthService
	emailService *services.EmailService
}

func NewAuthDomain(authService *services.AuthService, emailService *services.EmailService) *AuthDomain {
	return &AuthDomain{
		authService:  authService,
		emailService: emailService,
	}
}

func (d *AuthDomain) RegisterRoutes(authMiddleware *middleware.AuthMiddleware) {
	authHandler := handlers.NewAuthHandler(d.authService, d.emailService)

	http.HandleFunc("/api/auth/register", middleware.LoggingMiddleware(middleware.CORS(authHandler.Register)))
	http.HandleFunc("/api/auth/login", middleware.LoggingMiddleware(middleware.CORS(authHandler.Login)))
	http.HandleFunc("/api/auth/logout", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(authHandler.Logout))))
	http.HandleFunc("/api/auth/me", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(authHandler.Me))))
	http.HandleFunc("/api/auth/forgot-password", middleware.LoggingMiddleware(middleware.CORS(authHandler.ForgotPassword)))
	http.HandleFunc("/api/auth/reset-password", middleware.LoggingMiddleware(middleware.CORS(authHandler.ResetPassword)))
	http.HandleFunc("/api/auth/change-password", middleware.LoggingMiddleware(middleware.CORS(authMiddleware.RequireAuth(authHandler.ChangePassword))))
	http.HandleFunc("/api/auth/send-verification-code", middleware.LoggingMiddleware(middleware.CORS(authHandler.SendVerificationCode)))
	http.HandleFunc("/api/auth/test-send-verification-code", middleware.LoggingMiddleware(middleware.CORS(authHandler.TestSendVerificationCode)))
}
