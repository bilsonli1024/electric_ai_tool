-- 清理并重新初始化权限系统
-- 执行此脚本前请备份数据库

-- 选择数据库
USE electric_ai_tool;

-- 1. 清理现有数据（按照外键依赖顺序删除）
DELETE FROM role_permissions_tab;
DELETE FROM user_roles_tab;
DELETE FROM permissions_tab;
DELETE FROM roles_tab;

-- 2. 重置自增ID
ALTER TABLE permissions_tab AUTO_INCREMENT = 1;
ALTER TABLE roles_tab AUTO_INCREMENT = 1;
ALTER TABLE role_permissions_tab AUTO_INCREMENT = 1;
ALTER TABLE user_roles_tab AUTO_INCREMENT = 1;

-- 3. 插入基础菜单权限
INSERT INTO permissions_tab (permission_code, permission_name, permission_desc, permission_type, parent_id, permission_status) VALUES
('menu:copywriting', '文案生成', '文案生成菜单', 1, 0, 1),
('menu:image', '图片生成', '图片生成菜单', 1, 0, 1),
('menu:task-center', '任务中心', '任务中心菜单', 1, 0, 1),
('menu:user-management', '用户管理', '用户管理菜单', 1, 0, 1),
('menu:admin', '管理员功能', '管理员功能菜单', 1, 0, 1);

-- 4. 插入管理员子菜单权限（parent_id=5，即menu:admin的ID）
INSERT INTO permissions_tab (permission_code, permission_name, permission_desc, permission_type, parent_id, permission_status) VALUES
('menu:admin:users', '用户列表', '用户列表菜单', 1, 5, 1),
('menu:admin:roles', '角色列表', '角色列表菜单', 1, 5, 1),
('menu:admin:permissions', '权限列表', '权限列表菜单', 1, 5, 1),
('menu:admin:role-permissions', '角色权限列表', '角色权限列表菜单', 1, 5, 1),
('menu:admin:user-roles', '用户角色管理', '用户角色管理菜单', 1, 5, 1);

-- 5. 插入按钮权限
INSERT INTO permissions_tab (permission_code, permission_name, permission_desc, permission_type, parent_id, permission_status) VALUES
('btn:task:view-all', '查看所有任务', '查看所有任务按钮', 2, 0, 1),
('btn:task:copy', '复制任务', '复制任务按钮', 2, 0, 1),
('btn:user:approve', '审批用户', '审批用户按钮', 2, 0, 1);

-- 6. 插入普通用户角色
INSERT INTO roles_tab (role_name, role_desc, role_status) VALUES
('普通用户', '普通用户角色', 1);

-- 7. 为普通用户角色分配权限（ID：1,2,3,4,12）
INSERT INTO role_permissions_tab (role_id, permission_id) VALUES
(1, 1),  -- menu:copywriting
(1, 2),  -- menu:image
(1, 3),  -- menu:task-center
(1, 4),  -- menu:user-management
(1, 12); -- btn:task:copy

-- 8. 为现有普通用户分配角色（排除管理员用户）
INSERT INTO user_roles_tab (user_id, role_id, ctime)
SELECT id, 1, UNIX_TIMESTAMP()
FROM users_tab
WHERE user_type != 99 AND user_status = 1
AND id NOT IN (SELECT user_id FROM user_roles_tab WHERE role_id = 1);

-- 显示结果
SELECT '=== 权限列表 ===' as Info;
SELECT id, permission_code, permission_name, parent_id FROM permissions_tab ORDER BY id;

SELECT '=== 角色列表 ===' as Info;
SELECT id, role_name, role_desc FROM roles_tab;

SELECT '=== 角色权限关系 ===' as Info;
SELECT rp.id, r.role_name, p.permission_code 
FROM role_permissions_tab rp
INNER JOIN roles_tab r ON rp.role_id = r.id
INNER JOIN permissions_tab p ON rp.permission_id = p.id
ORDER BY rp.id;

SELECT '=== 用户角色关系 ===' as Info;
SELECT ur.id, u.email, r.role_name
FROM user_roles_tab ur
INNER JOIN users_tab u ON ur.user_id = u.id
INNER JOIN roles_tab r ON ur.role_id = r.id
ORDER BY ur.id;
