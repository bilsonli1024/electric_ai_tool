package models

import "time"

type User struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Salt         string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	Status       int        `json:"status"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type PasswordResetToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	SKU            string     `json:"sku,omitempty"`
	Keywords       string     `json:"keywords,omitempty"`
	SellingPoints  string     `json:"selling_points,omitempty"`
	CompetitorLink string     `json:"competitor_link,omitempty"`
	AnalyzeModel   string     `json:"analyze_model"`
	GenerateModel  string     `json:"generate_model"`
	Status         int        `json:"status"`
	ResultData     string     `json:"result_data,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Username       string     `json:"username,omitempty"`
}

const (
	TaskStatusAnalyzing      = 0  // 分析中
	TaskStatusAnalyzed       = 1  // 分析完成
	TaskStatusGenerating     = 2  // 生成图片中
	TaskStatusCompleted      = 3  // 已完成
	TaskStatusAnalyzeFailed  = 10 // 分析失败
	TaskStatusGenerateFailed = 11 // 生成失败
)

const (
	ModelGemini   = "gemini"
	ModelGPT      = "gpt"
	ModelDeepSeek = "deepseek"
)

type TaskHistory struct {
	ID                  int64     `json:"id"`
	TaskID              int64     `json:"task_id"`
	UserID              int64     `json:"user_id"`
	Version             int       `json:"version"`
	Model               string    `json:"model"`
	Prompt              string    `json:"prompt,omitempty"`
	AspectRatio         string    `json:"aspect_ratio,omitempty"`
	ProductImagesURLs   string    `json:"product_images_urls,omitempty"`
	StyleRefImageURL    string    `json:"style_ref_image_url,omitempty"`
	GeneratedImageURL   string    `json:"generated_image_url,omitempty"`
	EditInstruction     string    `json:"edit_instruction,omitempty"`
	Status              int       `json:"status"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

const (
	TaskHistoryStatusFailed  = 0 // 失败
	TaskHistoryStatusSuccess = 1 // 成功
)

type UserLoginLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	LoginType  int       `json:"login_type"`
	LoginIP    string    `json:"login_ip,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	SessionID  string    `json:"session_id,omitempty"`
	LoginTime  time.Time `json:"login_time"`
}

const (
	LoginTypeLogin  = 1 // 登录
	LoginTypeLogout = 2 // 登出
	LoginTypeSwitch = 3 // 切换用户
)

type CDNImage struct {
	ID               int64     `json:"id"`
	UserID           int64     `json:"user_id"`
	OriginalFilename string    `json:"original_filename,omitempty"`
	CDNURL           string    `json:"cdn_url"`
	CDNKey           string    `json:"cdn_key"`
	FileSize         int64     `json:"file_size"`
	MimeType         string    `json:"mime_type"`
	ImageType        string    `json:"image_type"`
	CreatedAt        time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email            string `json:"email"`
	PasswordHash     string `json:"password_hash"`
	VerificationCode string `json:"verification_code"`
}

type LoginRequest struct {
	LoginID      string `json:"login_id"`
	PasswordHash string `json:"password_hash"`
	LoginIP      string `json:"login_ip,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token"`
	NewPasswordHash string `json:"new_password_hash"`
}

type AuthResponse struct {
	User      User   `json:"user"`
	SessionID string `json:"session_id"`
}

type TaskListResponse struct {
	Data  []Task `json:"data"`
	Total int    `json:"total"`
}

type TaskHistoryListResponse struct {
	Data  []TaskHistory `json:"data"`
	Total int           `json:"total"`
}
