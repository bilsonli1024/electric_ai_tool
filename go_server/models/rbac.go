package models

type Role struct {
	ID          int64  `json:"id"`
	RoleName    string `json:"role_name"`
	RoleCode    string `json:"role_code"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Permission struct {
	ID             int64  `json:"id"`
	PermissionName string `json:"permission_name"`
	PermissionCode string `json:"permission_code"`
	ResourceType   string `json:"resource_type"`
	Action         string `json:"action"`
	Description    string `json:"description"`
	CreatedAt      string `json:"created_at"`
}

type UserRole struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	RoleID    int64  `json:"role_id"`
	CreatedAt string `json:"created_at"`
}

type RolePermission struct {
	ID           int64  `json:"id"`
	RoleID       int64  `json:"role_id"`
	PermissionID int64  `json:"permission_id"`
	CreatedAt    string `json:"created_at"`
}

type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}

type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}

const (
	RoleStatusActive   = 1
	RoleStatusInactive = 0
)

const (
	// Default roles
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleModerator = "moderator"
)

const (
	// Resource types
	ResourceCopywriting     = "copywriting"
	ResourceImageGeneration = "image_generation"
	ResourceModelTest       = "model_test"
	ResourceUser            = "user"
	ResourceRole            = "role"
)

const (
	// Actions
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionExecute = "execute"
)
