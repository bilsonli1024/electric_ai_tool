package models

// User 用户
type User struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"-"`              // 密码（加密后，不返回给前端）
	Username   string `json:"username"`
	UserType   int    `json:"user_type"`      // 用户类型：0=普通用户, 99=管理员
	UserStatus int    `json:"user_status"`    // 用户状态：0=待审批, 1=正常, 2=已删除
	Ctime      int64  `json:"ctime"`          // 创建时间(UNIX时间戳)
	Mtime      int64  `json:"mtime"`          // 更新时间(UNIX时间戳)
}

// Session 会话
type Session struct {
	ID        string `json:"id"`
	UserID    int64  `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"` // 过期时间(UNIX时间戳)
	Ctime     int64  `json:"ctime"`      // 创建时间(UNIX时间戳)
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username"`
	IsAdmin     bool   `json:"is_admin"`      // 是否申请管理员权限
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	LoginIP   string `json:"login_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User      User   `json:"user"`
	SessionID string `json:"session_id"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Data  []User `json:"data"`
	Total int    `json:"total"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

