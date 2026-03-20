package services

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.User, error) {
	if req.Email == "" || req.PasswordHash == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	if !s.isValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM users_tab WHERE email = ?", req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists")
	}

	salt := s.generateSalt()
	finalHash := s.hashPasswordWithSalt(req.PasswordHash, salt)

	username := s.generateUsernameFromEmail(req.Email)
	username = s.ensureUniqueUsername(username)

	result, err := config.DB.Exec(
		"INSERT INTO users_tab (username, email, password_hash, salt) VALUES (?, ?, ?, ?)",
		username, req.Email, finalHash, salt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	userID, _ := result.LastInsertId()
	user := &models.User{
		ID:       userID,
		Username: username,
		Email:    req.Email,
		Status:   1,
	}

	// Assign default user role
	rbacService := NewRBACService()
	if userRole, err := rbacService.GetRoleByCode(models.RoleUser); err == nil {
		rbacService.AssignRoleToUser(userID, userRole.ID)
	}

	return user, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.User, string, error) {
	if req.LoginID == "" || req.PasswordHash == "" {
		return nil, "", fmt.Errorf("login ID and password are required")
	}

	var user models.User
	var query string
	
	if s.isValidEmail(req.LoginID) {
		query = "SELECT id, username, email, password_hash, salt, created_at, updated_at, status FROM users_tab WHERE email = ?"
	} else {
		query = "SELECT id, username, email, password_hash, salt, created_at, updated_at, status FROM users_tab WHERE username = ?"
	}

	var lastLoginAt sql.NullTime
	err := config.DB.QueryRow(query, req.LoginID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Salt,
		&user.CreatedAt, &user.UpdatedAt, &user.Status,
	)

	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != 1 {
		return nil, "", fmt.Errorf("user is inactive")
	}

	expectedHash := s.hashPasswordWithSalt(req.PasswordHash, user.Salt)
	if expectedHash != user.PasswordHash {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	sessionID, err := s.CreateSession(user.ID)
	if err != nil {
		return nil, "", err
	}

	s.logUserLogin(user.ID, sessionID, models.LoginTypeLogin, req.LoginIP, req.UserAgent)

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, sessionID, nil
}

func (s *AuthService) ForgotPassword(req models.ForgotPasswordRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if !s.isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	var userID int64
	err := config.DB.QueryRow("SELECT id FROM users_tab WHERE email = ?", req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to query user: %w", err)
	}

	token := s.generateResetToken()
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = config.DB.Exec(
		"INSERT INTO password_reset_tokens_tab (user_id, token, expires_at) VALUES (?, ?, ?)",
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// TODO: 在生产环境中，这里应该发送真实的邮件
	// 目前仅记录到日志，开发时可以从日志中复制token
	log.Printf("Password reset token for user %d: %s", userID, token)
	log.Printf("Reset link: http://localhost:5173/?reset_token=%s", token)

	return nil
}

func (s *AuthService) ResetPassword(req models.ResetPasswordRequest) error {
	if req.Token == "" || req.NewPasswordHash == "" {
		return fmt.Errorf("token and new password are required")
	}

	var tokenRecord models.PasswordResetToken
	err := config.DB.QueryRow(
		"SELECT id, user_id, token, expires_at, used FROM password_reset_tokens_tab WHERE token = ?",
		req.Token,
	).Scan(&tokenRecord.ID, &tokenRecord.UserID, &tokenRecord.Token, &tokenRecord.ExpiresAt, &tokenRecord.Used)

	if err == sql.ErrNoRows {
		return fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		return fmt.Errorf("failed to query token: %w", err)
	}

	if tokenRecord.Used {
		return fmt.Errorf("token has already been used")
	}

	if tokenRecord.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	salt := s.generateSalt()
	finalHash := s.hashPasswordWithSalt(req.NewPasswordHash, salt)

	tx, err := config.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE users_tab SET password_hash = ?, salt = ? WHERE id = ?",
		finalHash, salt, tokenRecord.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	_, err = tx.Exec(
		"UPDATE password_reset_tokens_tab SET used = 1 WHERE id = ?",
		tokenRecord.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	_, err = tx.Exec("DELETE FROM sessions_tab WHERE user_id = ?", tokenRecord.UserID)
	if err != nil {
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AuthService) CreateSession(userID int64) (string, error) {
	sessionID := s.generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := config.DB.Exec(
		"INSERT INTO sessions_tab (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return sessionID, nil
}

func (s *AuthService) ValidateSession(sessionID string) (*models.User, error) {
	log.Printf("ValidateSession: checking session %s...", sessionID[:min(8, len(sessionID))])
	
	var session models.Session
	err := config.DB.QueryRow(
		"SELECT id, user_id, expires_at FROM sessions_tab WHERE id = ?",
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt)

	if err == sql.ErrNoRows {
		log.Printf("ValidateSession: session not found in database")
		return nil, fmt.Errorf("invalid session")
	}
	if err != nil {
		log.Printf("ValidateSession: database error: %v", err)
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	log.Printf("ValidateSession: found session for user_id=%d, expires_at=%v", session.UserID, session.ExpiresAt)

	if session.ExpiresAt.Before(time.Now()) {
		log.Printf("ValidateSession: session expired at %v", session.ExpiresAt)
		config.DB.Exec("DELETE FROM sessions_tab WHERE id = ?", sessionID)
		return nil, fmt.Errorf("session expired")
	}

	var user models.User
	var lastLoginAt sql.NullTime
	err = config.DB.QueryRow(
		"SELECT id, username, email, created_at, updated_at, status FROM users_tab WHERE id = ?",
		session.UserID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.Status)

	if err != nil {
		log.Printf("ValidateSession: failed to query user: %v", err)
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status != 1 {
		log.Printf("ValidateSession: user status is %d (inactive)", user.Status)
		return nil, fmt.Errorf("user is inactive")
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	log.Printf("ValidateSession: validation successful for user %s (id=%d)", user.Username, user.ID)
	return &user, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *AuthService) Logout(sessionID string) error {
	var userID int64
	config.DB.QueryRow("SELECT user_id FROM sessions_tab WHERE id = ?", sessionID).Scan(&userID)
	
	_, err := config.DB.Exec("DELETE FROM sessions_tab WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	if userID > 0 {
		s.logUserLogin(userID, sessionID, models.LoginTypeLogout, "", "")
	}
	
	return nil
}

func (s *AuthService) SwitchUser(oldSessionID string, newUserID int64) error {
	var oldUserID int64
	config.DB.QueryRow("SELECT user_id FROM sessions_tab WHERE id = ?", oldSessionID).Scan(&oldUserID)
	
	if oldUserID > 0 {
		s.logUserLogin(oldUserID, oldSessionID, models.LoginTypeSwitch, "", "")
	}
	
	return nil
}

func (s *AuthService) logUserLogin(userID int64, sessionID string, loginType int, loginIP string, userAgent string) {
	config.DB.Exec(
		"INSERT INTO user_login_log_tab (user_id, login_type, login_ip, user_agent, session_id) VALUES (?, ?, ?, ?, ?)",
		userID, loginType, loginIP, userAgent, sessionID,
	)
}

func (s *AuthService) hashPasswordWithSalt(passwordHash string, salt string) string {
	combined := passwordHash + salt
	hash := md5.Sum([]byte(combined))
	return hex.EncodeToString(hash[:])
}

func (s *AuthService) generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) generateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *AuthService) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (s *AuthService) generateUsernameFromEmail(email string) string {
	parts := regexp.MustCompile(`[@.]`).Split(email, -1)
	if len(parts) > 0 {
		return parts[0]
	}
	return "user"
}

func (s *AuthService) ensureUniqueUsername(username string) string {
	var count int
	config.DB.QueryRow("SELECT COUNT(*) FROM users_tab WHERE username = ?", username).Scan(&count)
	
	if count == 0 {
		return username
	}

	for i := 1; i < 10000; i++ {
		testUsername := fmt.Sprintf("%s%d", username, i)
		config.DB.QueryRow("SELECT COUNT(*) FROM users_tab WHERE username = ?", testUsername).Scan(&count)
		if count == 0 {
			return testUsername
		}
	}

	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%s_%s", username, hex.EncodeToString(b))
}
