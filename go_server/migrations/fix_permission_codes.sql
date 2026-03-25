-- 更新权限码格式（从下划线改为冒号）
-- 执行此脚本前请备份数据库

-- 更新菜单权限
UPDATE permissions_tab SET permission_code = 'menu:copywriting' WHERE permission_code = 'menu_copywriting';
UPDATE permissions_tab SET permission_code = 'menu:image' WHERE permission_code = 'menu_image';
UPDATE permissions_tab SET permission_code = 'menu:task-center' WHERE permission_code = 'menu_task_center';
UPDATE permissions_tab SET permission_code = 'menu:user-management' WHERE permission_code = 'menu_user_management';
UPDATE permissions_tab SET permission_code = 'menu:admin' WHERE permission_code = 'menu_admin';
UPDATE permissions_tab SET permission_code = 'menu:admin:users' WHERE permission_code = 'menu_admin_users';
UPDATE permissions_tab SET permission_code = 'menu:admin:roles' WHERE permission_code = 'menu_admin_roles';
UPDATE permissions_tab SET permission_code = 'menu:admin:permissions' WHERE permission_code = 'menu_admin_permissions';
UPDATE permissions_tab SET permission_code = 'menu:admin:role-permissions' WHERE permission_code = 'menu_admin_role_permissions';
UPDATE permissions_tab SET permission_code = 'menu:admin:user-roles' WHERE permission_code = 'menu_admin_user_roles';

-- 更新按钮权限
UPDATE permissions_tab SET permission_code = 'btn:task:view-all' WHERE permission_code = 'task_view_all';
UPDATE permissions_tab SET permission_code = 'btn:task:copy' WHERE permission_code = 'task_copy';
UPDATE permissions_tab SET permission_code = 'btn:user:approve' WHERE permission_code = 'user_approve';

-- 更新功能权限（如果存在）
UPDATE permissions_tab SET permission_code = 'copywriting:analyze' WHERE permission_code = 'copywriting_analyze';
UPDATE permissions_tab SET permission_code = 'copywriting:generate' WHERE permission_code = 'copywriting_generate';
UPDATE permissions_tab SET permission_code = 'copywriting:view' WHERE permission_code = 'copywriting_view';
UPDATE permissions_tab SET permission_code = 'image:generate' WHERE permission_code = 'image_generate';
UPDATE permissions_tab SET permission_code = 'image:view' WHERE permission_code = 'image_view';
UPDATE permissions_tab SET permission_code = 'task:list' WHERE permission_code = 'task_list';
UPDATE permissions_tab SET permission_code = 'task:detail' WHERE permission_code = 'task_detail';
UPDATE permissions_tab SET permission_code = 'task:copy' WHERE permission_code = 'task_copy';

-- 显示更新结果
SELECT permission_code, permission_name FROM permissions_tab ORDER BY id;
