package handlers

import (
	"database/sql"
	"electric_ai_tool/go_server/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type UserRoleHandler struct {
	db *sql.DB
}

func NewUserRoleHandler(db *sql.DB) *UserRoleHandler {
	return &UserRoleHandler{db: db}
}

// UserRoleItem 用户角色关系项
type UserRoleItem struct {
	ID       int64  `json:"id"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleID   int64  `json:"role_id"`
	RoleName string `json:"role_name"`
	Ctime    int64  `json:"ctime"`
}

// GetUserRoles 获取所有用户角色关系
func (h *UserRoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := `
		SELECT 
			ur.id, ur.user_id, u.username, u.email, 
			ur.role_id, r.role_name, ur.ctime
		FROM user_roles_tab ur
		INNER JOIN users_tab u ON ur.user_id = u.id
		INNER JOIN roles_tab r ON ur.role_id = r.id
		ORDER BY ur.ctime DESC
	`

	rows, err := h.db.Query(query)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []UserRoleItem{}
	for rows.Next() {
		var item UserRoleItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Username, &item.Email,
			&item.RoleID, &item.RoleName, &item.Ctime); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data":  items,
		"total": len(items),
	})
}

// GetUserRolesByUserID 获取指定用户的角色列表
func (h *UserRoleHandler) GetUserRolesByUserID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.RespondError(w, fmt.Errorf("缺少user_id参数"), http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			ur.id, ur.user_id, u.username, u.email, 
			ur.role_id, r.role_name, ur.ctime
		FROM user_roles_tab ur
		INNER JOIN users_tab u ON ur.user_id = u.id
		INNER JOIN roles_tab r ON ur.role_id = r.id
		WHERE ur.user_id = ?
		ORDER BY ur.ctime DESC
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []UserRoleItem{}
	for rows.Next() {
		var item UserRoleItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.Username, &item.Email,
			&item.RoleID, &item.RoleName, &item.Ctime); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	utils.RespondJSON(w, map[string]interface{}{
		"data": items,
	})
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	UserID int64 `json:"user_id"`
	RoleID int64 `json:"role_id"`
}

// AssignRole 为用户分配角色
func (h *UserRoleHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 || req.RoleID <= 0 {
		utils.RespondError(w, fmt.Errorf("无效的用户ID或角色ID"), http.StatusBadRequest)
		return
	}

	// 检查是否已存在
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM user_roles_tab WHERE user_id = ? AND role_id = ?",
		req.UserID, req.RoleID).Scan(&count)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	if count > 0 {
		utils.RespondError(w, fmt.Errorf("该用户已拥有此角色"), http.StatusBadRequest)
		return
	}

	// 插入用户角色关系
	currentTime := time.Now().Unix()
	_, err = h.db.Exec(
		"INSERT INTO user_roles_tab (user_id, role_id, ctime) VALUES (?, ?, ?)",
		req.UserID, req.RoleID, currentTime,
	)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"message": "角色分配成功",
	})
}

// RemoveRoleRequest 移除角色请求
type RemoveRoleRequest struct {
	UserID int64 `json:"user_id"`
	RoleID int64 `json:"role_id"`
}

// RemoveRole 移除用户的角色
func (h *UserRoleHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RemoveRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.UserID <= 0 || req.RoleID <= 0 {
		utils.RespondError(w, fmt.Errorf("无效的用户ID或角色ID"), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(
		"DELETE FROM user_roles_tab WHERE user_id = ? AND role_id = ?",
		req.UserID, req.RoleID,
	)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		utils.RespondError(w, fmt.Errorf("未找到该用户角色关系"), http.StatusNotFound)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"message": "角色移除成功",
	})
}
