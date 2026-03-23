package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// Register 用户注册（不自动登录）
func (s *AuthService) Register(req models.RegisterRequest) error {
	if req.Email == "" || req.Password == "" {
		return fmt.Errorf("邮箱和密码不能为空")
	}

	if !s.isValidEmail(req.Email) {
		return fmt.Errorf("邮箱格式不正确")
	}

	// 检查邮箱是否已存在
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM users_tab WHERE email = ?", req.Email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("该邮箱已被注册")
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 确定用户类型
	userType := models.UserTypeNormal
	if req.IsAdmin {
		userType = models.UserTypeAdmin
	}

	// 生成用户名
	username := req.Username
	if username == "" {
		username = s.generateUsernameFromEmail(req.Email)
	}
	username = s.ensureUniqueUsername(username)

	// 创建用户（状态为待审批）
	currentTime := utils.GetCurrentTimestamp()
	_, err = config.DB.Exec(
		`INSERT INTO users_tab (email, password, username, user_type, user_status, ctime, mtime) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.Email, string(hashedPassword), username, userType, models.UserStatusPendingApproval, currentTime, currentTime,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	utils.LogInfo("User registered: %s (type: %d, status: pending approval)", req.Email, userType)
	return nil
}

// Login 用户登录
func (s *AuthService) Login(req models.LoginRequest) (*models.User, string, error) {
	if req.Email == "" || req.Password == "" {
		return nil, "", fmt.Errorf("邮箱和密码不能为空")
	}

	// 查询用户
	var user models.User
	var password string
	err := config.DB.QueryRow(
		`SELECT id, email, password, username, user_type, user_status, ctime, mtime 
		 FROM users_tab WHERE email = ?`,
		req.Email,
	).Scan(
		&user.ID, &user.Email, &password, &user.Username, &user.UserType, &user.UserStatus,
		&user.Ctime, &user.Mtime,
	)

	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("邮箱或密码错误")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to query user: %w", err)
	}

	// 检查用户状态
	if user.UserStatus == models.UserStatusPendingApproval {
		return nil, "", fmt.Errorf("您的账号正在审核中，请等待管理员审批")
	}
	if user.UserStatus == models.UserStatusDeleted {
		return nil, "", fmt.Errorf("您的账号已被删除")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password))
	if err != nil {
		return nil, "", fmt.Errorf("邮箱或密码错误")
	}

	// 创建会话
	sessionID, err := s.CreateSession(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	utils.LogInfo("User logged in: %s (id: %d)", user.Email, user.ID)
	return &user, sessionID, nil
}

// CreateSession 创建会话
func (s *AuthService) CreateSession(userID int64) (string, error) {
	sessionID := s.generateSessionID()
	expiresAt := utils.GetCurrentTimestamp() + 7*24*60*60 // 7天后过期
	ctime := utils.GetCurrentTimestamp()

	_, err := config.DB.Exec(
		`INSERT INTO sessions_tab (id, user_id, expires_at, ctime) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE expires_at = ?, ctime = ?`,
		sessionID, userID, expiresAt, ctime, expiresAt, ctime,
	)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// ValidateSession 验证会话
func (s *AuthService) ValidateSession(sessionID string) (int64, error) {
	var userID int64
	var expiresAt int64

	err := config.DB.QueryRow(
		"SELECT user_id, expires_at FROM sessions_tab WHERE id = ?",
		sessionID,
	).Scan(&userID, &expiresAt)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid session")
	}
	if err != nil {
		return 0, err
	}

	// 检查是否过期
	if expiresAt < utils.GetCurrentTimestamp() {
		return 0, fmt.Errorf("session expired")
	}

	return userID, nil
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	err := config.DB.QueryRow(
		`SELECT id, email, username, user_type, user_status, ctime, mtime 
		 FROM users_tab WHERE id = ?`,
		userID,
	).Scan(
		&user.ID, &user.Email, &user.Username, &user.UserType, &user.UserStatus,
		&user.Ctime, &user.Mtime,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := config.DB.QueryRow(
		`SELECT id, email, username, user_type, user_status, ctime, mtime 
		 FROM users_tab WHERE email = ?`,
		email,
	).Scan(
		&user.ID, &user.Email, &user.Username, &user.UserType, &user.UserStatus,
		&user.Ctime, &user.Mtime,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Logout 用户登出
func (s *AuthService) Logout(sessionID string) error {
	_, err := config.DB.Exec("DELETE FROM sessions_tab WHERE id = ?", sessionID)
	return err
}

// ForgotPassword 忘记密码 - 生成重置token并发送邮件
func (s *AuthService) ForgotPassword(email string) error {
	// 1. 检查用户是否存在
	var userID int64
	var userStatus int
	err := config.DB.QueryRow(
		"SELECT id, user_status FROM users_tab WHERE email = ?",
		email,
	).Scan(&userID, &userStatus)
	
	if err == sql.ErrNoRows {
		// 为了安全，即使用户不存在也返回成功
		return nil
	}
	if err != nil {
		return err
	}

	// 检查用户状态
	if userStatus != models.UserStatusNormal {
		return fmt.Errorf("用户状态异常")
	}

	// 2. 生成重置token
	token := s.generateResetToken()
	expiresAt := utils.GetCurrentTimestamp() + 3600 // 1小时后过期
	ctime := utils.GetCurrentTimestamp()

	// 3. 保存token到数据库
	_, err = config.DB.Exec(
		`INSERT INTO password_reset_tokens_tab (user_id, token, expires_at, used, ctime)
		 VALUES (?, ?, ?, 0, ?)`,
		userID, token, expiresAt, ctime,
	)
	if err != nil {
		return err
	}

	// 4. TODO: 发送重置邮件（这里暂时只记录日志）
	log.Printf("Password reset token for %s: %s (expires at %d)", email, token, expiresAt)
	
	return nil
}

// ResetPassword 重置密码
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// 1. 验证token
	var userID int64
	var expiresAt int64
	var used int
	err := config.DB.QueryRow(
		"SELECT user_id, expires_at, used FROM password_reset_tokens_tab WHERE token = ?",
		token,
	).Scan(&userID, &expiresAt, &used)

	if err == sql.ErrNoRows {
		return fmt.Errorf("无效的重置token")
	}
	if err != nil {
		return err
	}

	// 2. 检查token是否已使用
	if used == 1 {
		return fmt.Errorf("该重置链接已使用")
	}

	// 3. 检查token是否过期
	if utils.GetCurrentTimestamp() > expiresAt {
		return fmt.Errorf("重置链接已过期")
	}

	// 4. 验证新密码
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少为6位")
	}

	// 5. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 6. 更新用户密码
	mtime := utils.GetCurrentTimestamp()
	_, err = config.DB.Exec(
		"UPDATE users_tab SET password = ?, mtime = ? WHERE id = ?",
		string(hashedPassword), mtime, userID,
	)
	if err != nil {
		return err
	}

	// 7. 标记token为已使用
	_, err = config.DB.Exec(
		"UPDATE password_reset_tokens_tab SET used = 1 WHERE token = ?",
		token,
	)
	if err != nil {
		return err
	}

	// 8. 删除该用户的所有会话（强制重新登录）
	_, err = config.DB.Exec("DELETE FROM sessions_tab WHERE user_id = ?", userID)
	if err != nil {
		log.Printf("Failed to delete sessions after password reset: %v", err)
	}

	return nil
}

// ChangePassword 修改密码（需要验证旧密码）
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// 1. 获取用户当前密码
	var currentPassword string
	err := config.DB.QueryRow(
		"SELECT password FROM users_tab WHERE id = ?",
		userID,
	).Scan(&currentPassword)

	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 2. 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(oldPassword)); err != nil {
		return fmt.Errorf("原密码错误")
	}

	// 3. 验证新密码
	if len(newPassword) < 6 {
		return fmt.Errorf("新密码长度至少为6位")
	}

	// 4. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 5. 更新密码
	mtime := utils.GetCurrentTimestamp()
	_, err = config.DB.Exec(
		"UPDATE users_tab SET password = ?, mtime = ? WHERE id = ?",
		string(hashedPassword), mtime, userID,
	)
	if err != nil {
		return err
	}

	// 6. 删除该用户的其他会话（保留当前会话需要前端传sessionID，这里简单处理）
	// TODO: 可以改进为保留当前会话
	
	return nil
}

// ============================================================================
// 辅助函数
// ============================================================================

func (s *AuthService) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (s *AuthService) generateUsernameFromEmail(email string) string {
	re := regexp.MustCompile(`^([^@]+)@`)
	matches := re.FindStringSubmatch(email)
	if len(matches) > 1 {
		return matches[1]
	}
	return "user"
}

func (s *AuthService) ensureUniqueUsername(username string) string {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM users_tab WHERE username = ?", username).Scan(&count)
	if err != nil || count == 0 {
		return username
	}

	// 添加数字后缀
	for i := 1; i < 1000; i++ {
		newUsername := fmt.Sprintf("%s%d", username, i)
		err := config.DB.QueryRow("SELECT COUNT(*) FROM users_tab WHERE username = ?", newUsername).Scan(&count)
		if err == nil && count == 0 {
			return newUsername
		}
	}

	return username
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

