package handlers

import (
	"database/sql"
	"electric_ai_tool/go_server/utils"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PermissionHandler struct {
	db *sql.DB
}

func NewPermissionHandler(db *sql.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

// GetUserPermissions 获取用户的权限列表
func (h *PermissionHandler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
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
	err := h.db.QueryRow(`
		SELECT u.id, u.user_type
		FROM sessions_tab s
		INNER JOIN users_tab u ON s.user_id = u.id
		WHERE s.id = ? AND s.expires_at > ?
	`, sessionID, time.Now().Unix()).Scan(&userID, &userType)

	if err != nil {
		utils.RespondError(w, fmt.Errorf("认证失败"), http.StatusUnauthorized)
		return
	}

	// 如果是管理员，返回所有权限
	if userType == 99 {
		permissions, err := h.getAllPermissions()
		if err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		utils.RespondJSON(w, map[string]interface{}{
			"permissions": permissions,
			"is_admin":    true,
		})
		return
	}

	// 普通用户，查询其角色关联的权限
	permissions, err := h.getUserPermissionsByUserID(userID)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"permissions": permissions,
		"is_admin":    false,
	})
}

// getAllPermissions 获取所有权限（管理员）
func (h *PermissionHandler) getAllPermissions() ([]string, error) {
	rows, err := h.db.Query(`
		SELECT permission_code 
		FROM permissions_tab 
		WHERE permission_status = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}

	return permissions, nil
}

// getUserPermissionsByUserID 获取用户的权限（所有角色的并集）
func (h *PermissionHandler) getUserPermissionsByUserID(userID int64) ([]string, error) {
	rows, err := h.db.Query(`
		SELECT DISTINCT p.permission_code
		FROM user_roles_tab ur
		INNER JOIN role_permissions_tab rp ON ur.role_id = rp.role_id
		INNER JOIN permissions_tab p ON rp.permission_id = p.id
		INNER JOIN roles_tab r ON ur.role_id = r.id
		WHERE ur.user_id = ? 
		  AND r.role_status = 1 
		  AND p.permission_status = 1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		permissions = append(permissions, code)
	}

	return permissions, nil
}

// GetMenuTree 获取权限菜单树
func (h *PermissionHandler) GetMenuTree(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, permission_code, permission_name, permission_desc, 
		       permission_type, parent_id, permission_status
		FROM permissions_tab
		WHERE permission_status = 1
		ORDER BY parent_id, id
	`)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MenuItem struct {
		ID          int64       `json:"id"`
		Code        string      `json:"code"`
		Name        string      `json:"name"`
		Desc        string      `json:"desc"`
		Type        int         `json:"type"`
		ParentID    int64       `json:"parent_id"`
		Status      int         `json:"status"`
		Children    []MenuItem  `json:"children,omitempty"`
	}

	allItems := []MenuItem{}
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Desc, 
			&item.Type, &item.ParentID, &item.Status); err != nil {
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		allItems = append(allItems, item)
	}

	// 构建树形结构
	itemMap := make(map[int64]*MenuItem)
	for i := range allItems {
		itemMap[allItems[i].ID] = &allItems[i]
	}

	tree := []MenuItem{}
	for i := range allItems {
		if allItems[i].ParentID == 0 {
			tree = append(tree, allItems[i])
		} else {
			if parent, ok := itemMap[allItems[i].ParentID]; ok {
				if parent.Children == nil {
					parent.Children = []MenuItem{}
				}
				parent.Children = append(parent.Children, allItems[i])
			}
		}
	}

	utils.RespondJSON(w, map[string]interface{}{
		"menu_tree": tree,
	})
}
