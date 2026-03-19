package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

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
			utils.RespondError(w, errors.New("unauthorized"), http.StatusUnauthorized)
			return
		}

		user, err := m.authService.ValidateSession(sessionID)
		if err != nil {
			utils.RespondError(w, err, http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", fmt.Sprintf("%d", user.ID))
		r.Header.Set("X-Username", user.Username)

		ctx := context.WithValue(r.Context(), UserIDKey, user.ID)
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
