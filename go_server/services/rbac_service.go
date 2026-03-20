package services

import (
	"fmt"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type RBACService struct{}

func NewRBACService() *RBACService {
	return &RBACService{}
}

func (s *RBACService) CreateRole(role *models.Role) (int64, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	result, err := db.Exec(
		`INSERT INTO roles_tab (role_name, role_code, description, status) VALUES (?, ?, ?, ?)`,
		role.RoleName, role.RoleCode, role.Description, role.Status,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *RBACService) GetRoleByCode(roleCode string) (*models.Role, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	role := &models.Role{}
	err := db.QueryRow(
		`SELECT id, role_name, role_code, description, status, created_at, updated_at 
		 FROM roles_tab WHERE role_code = ?`,
		roleCode,
	).Scan(&role.ID, &role.RoleName, &role.RoleCode, &role.Description, &role.Status, &role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return role, nil
}

func (s *RBACService) GetAllRoles() ([]*models.Role, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		`SELECT id, role_name, role_code, description, status, created_at, updated_at 
		 FROM roles_tab ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []*models.Role{}
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(&role.ID, &role.RoleName, &role.RoleCode, &role.Description, &role.Status, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (s *RBACService) CreatePermission(perm *models.Permission) (int64, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	result, err := db.Exec(
		`INSERT INTO permissions_tab (permission_name, permission_code, resource_type, action, description) 
		 VALUES (?, ?, ?, ?, ?)`,
		perm.PermissionName, perm.PermissionCode, perm.ResourceType, perm.Action, perm.Description,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *RBACService) GetAllPermissions() ([]*models.Permission, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		`SELECT id, permission_name, permission_code, resource_type, action, description, created_at 
		 FROM permissions_tab ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*models.Permission{}
	for rows.Next() {
		perm := &models.Permission{}
		err := rows.Scan(&perm.ID, &perm.PermissionName, &perm.PermissionCode, &perm.ResourceType, 
			&perm.Action, &perm.Description, &perm.CreatedAt)
		if err != nil {
			continue
		}
		perms = append(perms, perm)
	}

	return perms, nil
}

func (s *RBACService) AssignRoleToUser(userID, roleID int64) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(
		`INSERT IGNORE INTO user_roles_tab (user_id, role_id) VALUES (?, ?)`,
		userID, roleID,
	)
	return err
}

func (s *RBACService) RemoveRoleFromUser(userID, roleID int64) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(
		`DELETE FROM user_roles_tab WHERE user_id = ? AND role_id = ?`,
		userID, roleID,
	)
	return err
}

func (s *RBACService) GetUserRoles(userID int64) ([]*models.Role, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		`SELECT r.id, r.role_name, r.role_code, r.description, r.status, r.created_at, r.updated_at 
		 FROM roles_tab r
		 INNER JOIN user_roles_tab ur ON r.id = ur.role_id
		 WHERE ur.user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []*models.Role{}
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(&role.ID, &role.RoleName, &role.RoleCode, &role.Description, &role.Status, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			continue
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (s *RBACService) AssignPermissionToRole(roleID, permissionID int64) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(
		`INSERT IGNORE INTO role_permissions_tab (role_id, permission_id) VALUES (?, ?)`,
		roleID, permissionID,
	)
	return err
}

func (s *RBACService) GetRolePermissions(roleID int64) ([]*models.Permission, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		`SELECT p.id, p.permission_name, p.permission_code, p.resource_type, p.action, p.description, p.created_at 
		 FROM permissions_tab p
		 INNER JOIN role_permissions_tab rp ON p.id = rp.permission_id
		 WHERE rp.role_id = ?`,
		roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*models.Permission{}
	for rows.Next() {
		perm := &models.Permission{}
		err := rows.Scan(&perm.ID, &perm.PermissionName, &perm.PermissionCode, &perm.ResourceType,
			&perm.Action, &perm.Description, &perm.CreatedAt)
		if err != nil {
			continue
		}
		perms = append(perms, perm)
	}

	return perms, nil
}

func (s *RBACService) GetUserPermissions(userID int64) ([]*models.Permission, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		`SELECT DISTINCT p.id, p.permission_name, p.permission_code, p.resource_type, p.action, p.description, p.created_at 
		 FROM permissions_tab p
		 INNER JOIN role_permissions_tab rp ON p.id = rp.permission_id
		 INNER JOIN user_roles_tab ur ON rp.role_id = ur.role_id
		 WHERE ur.user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms := []*models.Permission{}
	for rows.Next() {
		perm := &models.Permission{}
		err := rows.Scan(&perm.ID, &perm.PermissionName, &perm.PermissionCode, &perm.ResourceType,
			&perm.Action, &perm.Description, &perm.CreatedAt)
		if err != nil {
			continue
		}
		perms = append(perms, perm)
	}

	return perms, nil
}

func (s *RBACService) HasPermission(userID int64, resourceType, action string) (bool, error) {
	perms, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range perms {
		if perm.ResourceType == resourceType && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

func (s *RBACService) InitializeDefaultRolesAndPermissions() error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Create default roles
	roles := []models.Role{
		{RoleName: "管理员", RoleCode: models.RoleAdmin, Description: "系统管理员，拥有所有权限", Status: models.RoleStatusActive},
		{RoleName: "普通用户", RoleCode: models.RoleUser, Description: "普通用户，基本权限", Status: models.RoleStatusActive},
		{RoleName: "版主", RoleCode: models.RoleModerator, Description: "版主，部分管理权限", Status: models.RoleStatusActive},
	}

	for _, role := range roles {
		_, err := db.Exec(
			`INSERT IGNORE INTO roles_tab (role_name, role_code, description, status) VALUES (?, ?, ?, ?)`,
			role.RoleName, role.RoleCode, role.Description, role.Status,
		)
		if err != nil {
			return err
		}
	}

	// Create default permissions
	permissions := []models.Permission{
		{PermissionName: "创建文案", PermissionCode: "copywriting:create", ResourceType: models.ResourceCopywriting, Action: models.ActionCreate},
		{PermissionName: "查看文案", PermissionCode: "copywriting:read", ResourceType: models.ResourceCopywriting, Action: models.ActionRead},
		{PermissionName: "更新文案", PermissionCode: "copywriting:update", ResourceType: models.ResourceCopywriting, Action: models.ActionUpdate},
		{PermissionName: "删除文案", PermissionCode: "copywriting:delete", ResourceType: models.ResourceCopywriting, Action: models.ActionDelete},
		
		{PermissionName: "创建图片生成", PermissionCode: "image_generation:create", ResourceType: models.ResourceImageGeneration, Action: models.ActionCreate},
		{PermissionName: "查看图片生成", PermissionCode: "image_generation:read", ResourceType: models.ResourceImageGeneration, Action: models.ActionRead},
		{PermissionName: "更新图片生成", PermissionCode: "image_generation:update", ResourceType: models.ResourceImageGeneration, Action: models.ActionUpdate},
		{PermissionName: "删除图片生成", PermissionCode: "image_generation:delete", ResourceType: models.ResourceImageGeneration, Action: models.ActionDelete},
		
		{PermissionName: "执行模型测试", PermissionCode: "model_test:execute", ResourceType: models.ResourceModelTest, Action: models.ActionExecute},
		
		{PermissionName: "管理用户", PermissionCode: "user:update", ResourceType: models.ResourceUser, Action: models.ActionUpdate},
		{PermissionName: "管理角色", PermissionCode: "role:update", ResourceType: models.ResourceRole, Action: models.ActionUpdate},
	}

	for _, perm := range permissions {
		_, err := db.Exec(
			`INSERT IGNORE INTO permissions_tab (permission_name, permission_code, resource_type, action, description) 
			 VALUES (?, ?, ?, ?, ?)`,
			perm.PermissionName, perm.PermissionCode, perm.ResourceType, perm.Action, perm.Description,
		)
		if err != nil {
			return err
		}
	}

	// Assign all permissions to admin role
	adminRole, err := s.GetRoleByCode(models.RoleAdmin)
	if err == nil {
		allPerms, _ := s.GetAllPermissions()
		for _, perm := range allPerms {
			s.AssignPermissionToRole(adminRole.ID, perm.ID)
		}
	}

	// Assign basic permissions to user role
	userRole, err := s.GetRoleByCode(models.RoleUser)
	if err == nil {
		basicPermCodes := []string{
			"copywriting:create", "copywriting:read", "copywriting:update",
			"image_generation:create", "image_generation:read", "image_generation:update",
			"model_test:execute",
		}
		allPerms, _ := s.GetAllPermissions()
		for _, perm := range allPerms {
			for _, code := range basicPermCodes {
				if perm.PermissionCode == code {
					s.AssignPermissionToRole(userRole.ID, perm.ID)
					break
				}
			}
		}
	}

	return nil
}
