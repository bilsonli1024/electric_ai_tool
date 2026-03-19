package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("username, email and password are required")
	}

	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", req.Username, req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("username or email already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	result, err := config.DB.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		req.Username, req.Email, string(passwordHash),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	userID, _ := result.LastInsertId()
	user := &models.User{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
		Status:   1,
	}

	return user, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.User, string, error) {
	if req.Username == "" || req.Password == "" {
		return nil, "", fmt.Errorf("username and password are required")
	}

	var user models.User
	err := config.DB.QueryRow(
		"SELECT id, username, email, password_hash, created_at, updated_at, last_login_at, status FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.Status)

	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("invalid username or password")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != 1 {
		return nil, "", fmt.Errorf("user is inactive")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, "", fmt.Errorf("invalid username or password")
	}

	sessionID, err := s.CreateSession(user.ID)
	if err != nil {
		return nil, "", err
	}

	config.DB.Exec("UPDATE users SET last_login_at = ? WHERE id = ?", time.Now(), user.ID)

	return &user, sessionID, nil
}

func (s *AuthService) CreateSession(userID int64) (string, error) {
	sessionID := s.generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := config.DB.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

func (s *AuthService) ValidateSession(sessionID string) (*models.User, error) {
	var session models.Session
	err := config.DB.QueryRow(
		"SELECT id, user_id, expires_at FROM sessions WHERE id = ?",
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid session")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	if session.ExpiresAt.Before(time.Now()) {
		config.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
		return nil, fmt.Errorf("session expired")
	}

	var user models.User
	err = config.DB.QueryRow(
		"SELECT id, username, email, created_at, updated_at, last_login_at, status FROM users WHERE id = ?",
		session.UserID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt, &user.Status)

	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != 1 {
		return nil, fmt.Errorf("user is inactive")
	}

	return &user, nil
}

func (s *AuthService) Logout(sessionID string) error {
	_, err := config.DB.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (s *AuthService) generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
