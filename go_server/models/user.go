package models

import "time"

type User struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
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

type Task struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	TaskType       string     `json:"task_type"`
	SKU            string     `json:"sku,omitempty"`
	Keywords       string     `json:"keywords,omitempty"`
	SellingPoints  string     `json:"selling_points,omitempty"`
	CompetitorLink string     `json:"competitor_link,omitempty"`
	Status         string     `json:"status"`
	ResultData     string     `json:"result_data,omitempty"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Username       string     `json:"username,omitempty"`
}

type TaskHistory struct {
	ID                  int64     `json:"id"`
	TaskID              int64     `json:"task_id"`
	UserID              int64     `json:"user_id"`
	Version             int       `json:"version"`
	Prompt              string    `json:"prompt,omitempty"`
	AspectRatio         string    `json:"aspect_ratio,omitempty"`
	ProductImagesURLs   string    `json:"product_images_urls,omitempty"`
	StyleRefImageURL    string    `json:"style_ref_image_url,omitempty"`
	GeneratedImageURL   string    `json:"generated_image_url,omitempty"`
	EditInstruction     string    `json:"edit_instruction,omitempty"`
	Status              string    `json:"status"`
	ErrorMessage        string    `json:"error_message,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

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
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
