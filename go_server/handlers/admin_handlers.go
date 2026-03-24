package handlers

import (
	"database/sql"
	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// GetUsers 获取所有用户列表
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, email, username, user_type, user_status, ctime, mtime
		FROM users_tab
		ORDER BY ctime DESC
	`)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []models.User{} // 初始化为空数组而不是nil
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Username, &user.UserType, &user.UserStatus, &user.Ctime, &user.Mtime); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  users,
		"total": len(users),
	})
}

// ApproveUser 审批通过用户
func (h *AdminHandler) ApproveUser(w http.ResponseWriter, r *http.Request) {
	// 从请求体解析用户ID
	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	_, err := h.db.Exec(`
		UPDATE users_tab 
		SET user_status = 1, mtime = ? 
		WHERE id = ? AND user_status = 0
	`, now, req.UserID)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]string{"message": "用户已批准"})
}

// RejectUser 拒绝用户
func (h *AdminHandler) RejectUser(w http.ResponseWriter, r *http.Request) {
	// 从请求体解析用户ID
	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	_, err := h.db.Exec(`
		UPDATE users_tab 
		SET user_status = 2, mtime = ? 
		WHERE id = ? AND user_status = 0
	`, now, req.UserID)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]string{"message": "用户已拒绝"})
}

// GetRoles 获取所有角色列表
func (h *AdminHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, role_name, role_desc, role_status, ctime, mtime
		FROM roles_tab
		ORDER BY ctime DESC
	`)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Role struct {
		ID         int64  `json:"id"`
		RoleName   string `json:"role_name"`
		RoleDesc   string `json:"role_desc"`
		RoleStatus int    `json:"role_status"`
		Ctime      int64  `json:"ctime"`
		Mtime      int64  `json:"mtime"`
	}

	roles := []Role{} // 初始化为空数组
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.RoleName, &role.RoleDesc, &role.RoleStatus, &role.Ctime, &role.Mtime); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		roles = append(roles, role)
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  roles,
		"total": len(roles),
	})
}

// GetPermissions 获取所有权限列表
func (h *AdminHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT p.id, p.permission_code, p.permission_name, p.permission_desc, 
		       p.permission_type, p.parent_id, p.permission_status, p.ctime, p.mtime,
		       COALESCE(p2.permission_name, '') as parent_name
		FROM permissions_tab p
		LEFT JOIN permissions_tab p2 ON p.parent_id = p2.id
		ORDER BY p.parent_id, p.ctime DESC
	`)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Permission struct {
		ID               int64  `json:"id"`
		PermissionCode   string `json:"permission_code"`
		PermissionName   string `json:"permission_name"`
		PermissionDesc   string `json:"permission_desc"`
		PermissionType   int    `json:"permission_type"`
		ParentID         int64  `json:"parent_id"`
		ParentName       string `json:"parent_name"`
		PermissionStatus int    `json:"permission_status"`
		Ctime            int64  `json:"ctime"`
		Mtime            int64  `json:"mtime"`
	}

	permissions := []Permission{} // 初始化为空数组
	for rows.Next() {
		var perm Permission
		if err := rows.Scan(&perm.ID, &perm.PermissionCode, &perm.PermissionName, &perm.PermissionDesc,
			&perm.PermissionType, &perm.ParentID, &perm.PermissionStatus, &perm.Ctime, &perm.Mtime, &perm.ParentName); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		permissions = append(permissions, perm)
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  permissions,
		"total": len(permissions),
	})
}

// GetRolePermissions 获取角色权限列表
func (h *AdminHandler) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT rp.id, rp.role_id, r.role_name, rp.permission_id, p.permission_name, rp.ctime
		FROM role_permissions_tab rp
		INNER JOIN roles_tab r ON rp.role_id = r.id
		INNER JOIN permissions_tab p ON rp.permission_id = p.id
		ORDER BY rp.role_id, rp.ctime DESC
	`)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type RolePermission struct {
		ID             int64  `json:"id"`
		RoleID         int64  `json:"role_id"`
		RoleName       string `json:"role_name"`
		PermissionID   int64  `json:"permission_id"`
		PermissionName string `json:"permission_name"`
		Ctime          int64  `json:"ctime"`
	}

	rolePermissions := []RolePermission{} // 初始化为空数组
	for rows.Next() {
		var rp RolePermission
		if err := rows.Scan(&rp.ID, &rp.RoleID, &rp.RoleName, &rp.PermissionID, &rp.PermissionName, &rp.Ctime); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		rolePermissions = append(rolePermissions, rp)
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  rolePermissions,
		"total": len(rolePermissions),
	})
}

// AdminMiddleware 管理员权限验证中间件
func AdminMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			utils.RespondError(w, fmt.Errorf("database connection is nil"), http.StatusInternalServerError)
			return
		}
		
		sessionID := r.Header.Get("Authorization")
		if sessionID == "" {
			utils.RespondError(w, fmt.Errorf("未授权访问"), http.StatusUnauthorized)
			return
		}

		// 移除 "Bearer " 前缀
		if strings.HasPrefix(sessionID, "Bearer ") {
			sessionID = sessionID[7:]
		}

		// 验证session并获取用户信息
		var userID int64
		var userType int
		err := db.QueryRow(`
			SELECT u.id, u.user_type
			FROM sessions_tab s
			INNER JOIN users_tab u ON s.user_id = u.id
			WHERE s.id = ? AND s.expires_at > ?
		`, sessionID, time.Now().Unix()).Scan(&userID, &userType)

		if err != nil {
			utils.RespondError(w, fmt.Errorf("认证失败"), http.StatusUnauthorized)
			return
		}

		// 检查是否是管理员
		if userType != 99 {
			utils.RespondError(w, fmt.Errorf("需要管理员权限"), http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
