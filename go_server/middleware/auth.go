package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := m.getSessionID(r)
		if sessionID == "" {
			log.Printf("Auth failed: no session ID found in request from %s", r.RemoteAddr)
			utils.RespondError(w, errors.New("unauthorized"), http.StatusUnauthorized)
			return
		}

	log.Printf("Auth: validating session %s... (length: %d)", sessionID[:min(8, len(sessionID))], len(sessionID))
	userID, err := m.authService.ValidateSession(sessionID)
	if err != nil {
		log.Printf("Auth failed: session validation error: %v", err)
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	log.Printf("Auth success: user_id=%d", userID)
	r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	ctx := context.WithValue(r.Context(), "user_id", userID)
	next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) getSessionID(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	cookie, err := r.Cookie("session_id")
	if err == nil {
		return cookie.Value
	}

	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
