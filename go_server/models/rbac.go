package models

// ============================================================================
// RBAC相关数据结构
// ============================================================================

// Role 角色
type Role struct {
	ID         int64  `json:"id"`
	RoleName   string `json:"role_name"`
	RoleDesc   string `json:"role_desc"`
	RoleStatus int    `json:"role_status"`
	Ctime      int64  `json:"ctime"`
	Mtime      int64  `json:"mtime"`
}

// Permission 权限
type Permission struct {
	ID               int64  `json:"id"`
	PermissionCode   string `json:"permission_code"`
	PermissionName   string `json:"permission_name"`
	PermissionDesc   string `json:"permission_desc"`
	PermissionType   int    `json:"permission_type"`
	ParentID         int64  `json:"parent_id"`
	PermissionStatus int    `json:"permission_status"`
	Ctime            int64  `json:"ctime"`
	Mtime            int64  `json:"mtime"`
}

// UserRole 用户角色关系
type UserRole struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
	RoleID int64 `json:"role_id"`
	Ctime  int64 `json:"ctime"`
}

// RolePermission 角色权限关系
type RolePermission struct {
	ID           int64 `json:"id"`
	RoleID       int64 `json:"role_id"`
	PermissionID int64 `json:"permission_id"`
	Ctime        int64 `json:"ctime"`
}

// PermissionTreeNode 权限树节点（用于前端展示）
type PermissionTreeNode struct {
	ID               int64                  `json:"id"`
	PermissionCode   string                 `json:"permission_code"`
	PermissionName   string                 `json:"permission_name"`
	PermissionDesc   string                 `json:"permission_desc"`
	PermissionType   int                    `json:"permission_type"`
	PermissionStatus int                    `json:"permission_status"`
	Children         []*PermissionTreeNode  `json:"children,omitempty"`
}

// RoleWithPermissions 角色及其权限
type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}

// UserWithRoles 用户及其角色
type UserWithRoles struct {
	User
	Roles []Role `json:"roles"`
}
