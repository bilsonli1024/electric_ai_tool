-- Electric AI Tool Database Schema
-- 完整的数据库结构定义
-- 所有状态字段使用INT枚举，所有时间字段使用INT(UNIX时间戳)

CREATE DATABASE IF NOT EXISTS electric_ai_tool;
USE electric_ai_tool;

-- ============================================================================
-- 用户管理相关表
-- ============================================================================

-- 用户表
DROP TABLE IF EXISTS users_tab;
CREATE TABLE users_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    email VARCHAR(255) NOT NULL UNIQUE COMMENT '邮箱',
    password VARCHAR(255) NOT NULL COMMENT '密码（加密后）',
    username VARCHAR(100) COMMENT '用户名',
    user_type TINYINT NOT NULL DEFAULT 0 COMMENT '用户类型: 0=普通用户, 99=管理员',
    user_status TINYINT NOT NULL DEFAULT 0 COMMENT '用户状态: 0=待审批, 1=正常, 2=已删除',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    mtime INT NOT NULL COMMENT '更新时间(UNIX时间戳)',
    INDEX idx_email (email),
    INDEX idx_user_type (user_type),
    INDEX idx_user_status (user_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 角色表
DROP TABLE IF EXISTS roles_tab;
CREATE TABLE roles_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '角色ID',
    role_name VARCHAR(100) NOT NULL UNIQUE COMMENT '角色名称',
    role_desc VARCHAR(500) COMMENT '角色描述',
    role_status TINYINT NOT NULL DEFAULT 1 COMMENT '角色状态: 0=禁用, 1=启用',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    mtime INT NOT NULL COMMENT '更新时间(UNIX时间戳)',
    INDEX idx_role_name (role_name),
    INDEX idx_role_status (role_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- 权限表
DROP TABLE IF EXISTS permissions_tab;
CREATE TABLE permissions_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '权限ID',
    permission_code VARCHAR(100) NOT NULL UNIQUE COMMENT '权限代码',
    permission_name VARCHAR(100) NOT NULL COMMENT '权限名称',
    permission_desc VARCHAR(500) COMMENT '权限描述',
    permission_type TINYINT NOT NULL DEFAULT 1 COMMENT '权限类型: 1=菜单, 2=按钮, 3=API',
    parent_id BIGINT DEFAULT 0 COMMENT '父权限ID, 0表示顶级',
    permission_status TINYINT NOT NULL DEFAULT 1 COMMENT '权限状态: 0=禁用, 1=启用',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    mtime INT NOT NULL COMMENT '更新时间(UNIX时间戳)',
    INDEX idx_permission_code (permission_code),
    INDEX idx_parent_id (parent_id),
    INDEX idx_permission_status (permission_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- 用户角色关系表
DROP TABLE IF EXISTS user_roles_tab;
CREATE TABLE user_roles_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '关系ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    role_id BIGINT NOT NULL COMMENT '角色ID',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关系表';

-- 角色权限关系表
DROP TABLE IF EXISTS role_permissions_tab;
CREATE TABLE role_permissions_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '关系ID',
    role_id BIGINT NOT NULL COMMENT '角色ID',
    permission_id BIGINT NOT NULL COMMENT '权限ID',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    UNIQUE KEY uk_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关系表';

-- 会话表
DROP TABLE IF EXISTS sessions_tab;
CREATE TABLE sessions_tab (
    id VARCHAR(64) PRIMARY KEY COMMENT '会话ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    expires_at INT NOT NULL COMMENT '过期时间(UNIX时间戳)',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';

-- 密码重置token表
DROP TABLE IF EXISTS password_reset_tokens_tab;
CREATE TABLE password_reset_tokens_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    token VARCHAR(64) NOT NULL UNIQUE COMMENT '重置token',
    expires_at INT NOT NULL COMMENT '过期时间(UNIX时间戳)',
    used TINYINT NOT NULL DEFAULT 0 COMMENT '是否已使用: 0=未使用, 1=已使用',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    INDEX idx_user_id (user_id),
    INDEX idx_token (token),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='密码重置token表';

-- 邮箱验证码表
DROP TABLE IF EXISTS email_verification_codes_tab;
CREATE TABLE email_verification_codes_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT 'ID',
    email VARCHAR(255) NOT NULL COMMENT '邮箱地址',
    code VARCHAR(10) NOT NULL COMMENT '验证码',
    purpose VARCHAR(50) NOT NULL COMMENT '用途: register=注册, reset=重置密码',
    expires_at DATETIME NOT NULL COMMENT '过期时间',
    used TINYINT NOT NULL DEFAULT 0 COMMENT '是否已使用: 0=未使用, 1=已使用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_email (email),
    INDEX idx_code (code),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱验证码表';

-- ============================================================================
-- 任务中心相关表
-- ============================================================================

-- 任务中心底表（统一任务管理）
DROP TABLE IF EXISTS task_center_tab;
CREATE TABLE task_center_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    task_id VARCHAR(50) NOT NULL UNIQUE COMMENT '任务ID',
    task_type TINYINT NOT NULL COMMENT '任务类型: 1=文案生成, 2=图片生成',
    task_status TINYINT NOT NULL DEFAULT 0 COMMENT '任务状态: 0=待处理, 1=进行中, 2=已完成, 3=失败',
    operator VARCHAR(255) NOT NULL COMMENT '操作人邮箱',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    mtime INT NOT NULL COMMENT '更新时间(UNIX时间戳)',
    INDEX idx_task_id (task_id),
    INDEX idx_task_type (task_type),
    INDEX idx_task_status (task_status),
    INDEX idx_operator (operator),
    INDEX idx_ctime (ctime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务中心底表';

-- 文案生成任务详细表
DROP TABLE IF EXISTS copywriting_tasks_tab;
CREATE TABLE copywriting_tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    task_id VARCHAR(50) NOT NULL UNIQUE COMMENT '任务ID（关联task_center_tab）',
    task_name VARCHAR(200) DEFAULT '' COMMENT '任务名称',
    detail_status TINYINT NOT NULL DEFAULT 0 COMMENT '详细状态: 0=待处理, 1=分析中, 2=分析完成, 3=生成中, 4=已完成, 5=失败',
    competitor_urls TEXT COMMENT '竞品链接(JSON数组)',
    analysis_result LONGTEXT COMMENT 'AI分析结果(JSON)',
    analyze_model TINYINT DEFAULT 1 COMMENT '分析模型: 1=Gemini, 2=GPT, 3=DeepSeek',
    user_selected_data TEXT COMMENT '用户选择的数据(JSON)',
    product_details TEXT COMMENT '产品详情(JSON)',
    uploaded_image_paths TEXT COMMENT '上传的产品图片本地路径(JSON数组)',
    generated_copy LONGTEXT COMMENT '生成的文案(JSON)',
    generate_model TINYINT DEFAULT 1 COMMENT '生成模型: 1=Gemini, 2=GPT, 3=DeepSeek',
    error_message TEXT COMMENT '错误信息',
    fail_msg TEXT COMMENT '失败原因（用户友好）',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    mtime INT NOT NULL COMMENT '更新时间(UNIX时间戳)',
    INDEX idx_task_id (task_id),
    INDEX idx_detail_status (detail_status),
    INDEX idx_ctime (ctime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文案生成任务表';

-- 图片生成任务详细表
DROP TABLE IF EXISTS tasks_tab;
CREATE TABLE tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
    task_id VARCHAR(50) NOT NULL UNIQUE COMMENT '任务ID（关联task_center_tab）',
    detail_status TINYINT NOT NULL DEFAULT 0 COMMENT '详细状态: 0=待处理, 1=生成中, 2=已完成, 3=失败',
    sku VARCHAR(100) COMMENT 'SKU',
    keywords TEXT COMMENT '关键词',
    selling_points TEXT COMMENT '卖点',
    competitor_link VARCHAR(1000) COMMENT '竞品链接',
    copywriting_task_id VARCHAR(50) COMMENT '关联的文案任务ID',
    generate_model TINYINT DEFAULT 1 COMMENT '生成模型: 1=Gemini, 2=GPT, 3=DeepSeek',
    aspect_ratio VARCHAR(20) DEFAULT '1:1' COMMENT '图片宽高比',
    result_data LONGTEXT COMMENT '生成结果数据(JSON)',
    generated_image_urls TEXT COMMENT '生成的图片URL(逗号分隔)',
    local_image_paths TEXT COMMENT '本地图片文件路径(JSON数组)',
    error_message TEXT COMMENT '错误信息',
    fail_msg TEXT COMMENT '失败原因（用户友好）',
    ctime INT NOT NULL COMMENT '创建时间(UNIX时间戳)',
    mtime INT NOT NULL COMMENT '更新时间(UNIX时间戳)',
    INDEX idx_task_id (task_id),
    INDEX idx_detail_status (detail_status),
    INDEX idx_sku (sku),
    INDEX idx_copywriting_task_id (copywriting_task_id),
    INDEX idx_ctime (ctime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图片生成任务表';

-- ============================================================================
-- 初始数据插入
-- ============================================================================

-- 插入管理员用户（密码: 123456，需要在应用层加密）
INSERT INTO users_tab (email, password, username, user_type, user_status, ctime, mtime) 
VALUES ('admin@gmail.com', '$2a$10$placeholder_will_be_replaced', 'Administrator', 99, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入默认角色
INSERT INTO roles_tab (role_name, role_desc, role_status, ctime, mtime) VALUES
('超级管理员', '拥有所有权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('普通用户', '基础使用权限', 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入默认权限
INSERT INTO permissions_tab (permission_code, permission_name, permission_desc, permission_type, parent_id, permission_status, ctime, mtime) VALUES
-- 一级菜单
('menu_copywriting', '文案生成', '文案生成功能模块', 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('menu_image', '图片生成', '图片生成功能模块', 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('menu_task_center', '任务中心', '任务中心功能模块', 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('menu_user_mgmt', '用户管理', '用户管理功能模块（仅管理员）', 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('menu_role_mgmt', '角色管理', '角色管理功能模块（仅管理员）', 1, 0, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

-- 文案生成子权限
('copywriting_analyze', '竞品分析', '竞品分析功能', 2, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('copywriting_generate', '文案生成', '文案生成功能', 2, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('copywriting_view', '文案查看', '查看生成的文案', 2, 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

-- 图片生成子权限
('image_generate', '图片生成', '图片生成功能', 2, 2, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('image_view', '图片查看', '查看生成的图片', 2, 2, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

-- 任务中心子权限
('task_list', '任务列表', '查看任务列表', 2, 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('task_detail', '任务详情', '查看任务详情', 2, 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('task_copy', '任务复制', '复制任务', 2, 3, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

-- 用户管理子权限
('user_list', '用户列表', '查看用户列表', 2, 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('user_approve', '用户审批', '审批用户注册', 2, 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('user_edit', '用户编辑', '编辑用户信息', 2, 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('user_delete', '用户删除', '删除用户', 2, 4, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),

-- 角色管理子权限
('role_list', '角色列表', '查看角色列表', 2, 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('role_create', '角色创建', '创建新角色', 2, 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('role_edit', '角色编辑', '编辑角色信息', 2, 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('role_delete', '角色删除', '删除角色', 2, 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('role_permission', '角色权限', '管理角色权限', 2, 5, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 分配管理员用户所有权限
-- 首先将管理员用户关联到超级管理员角色
INSERT INTO user_roles_tab (user_id, role_id, ctime)
SELECT u.id, r.id, UNIX_TIMESTAMP()
FROM users_tab u, roles_tab r
WHERE u.email = 'admin@gmail.com' AND r.role_name = '超级管理员';

-- 将所有权限分配给超级管理员角色
INSERT INTO role_permissions_tab (role_id, permission_id, ctime)
SELECT r.id, p.id, UNIX_TIMESTAMP()
FROM roles_tab r, permissions_tab p
WHERE r.role_name = '超级管理员';

-- 为普通用户角色分配基础权限
INSERT INTO role_permissions_tab (role_id, permission_id, ctime)
SELECT r.id, p.id, UNIX_TIMESTAMP()
FROM roles_tab r, permissions_tab p
WHERE r.role_name = '普通用户' 
AND p.permission_code IN (
    'menu_copywriting', 'copywriting_analyze', 'copywriting_generate', 'copywriting_view',
    'menu_image', 'image_generate', 'image_view',
    'menu_task_center', 'task_list', 'task_detail', 'task_copy'
);

-- ============================================================================
-- 枚举值说明（注释）
-- ============================================================================

/*
用户类型枚举 (user_type):
0  - 普通用户
99 - 管理员

用户状态枚举 (user_status):
0 - 待审批
1 - 正常
2 - 已删除

任务类型枚举 (task_type):
1 - 文案生成
2 - 图片生成

任务中心状态枚举 (task_status):
0 - 待处理
1 - 进行中
2 - 已完成
3 - 失败

文案生成详细状态枚举 (copywriting detail_status):
0 - 待处理
1 - 分析中
2 - 分析完成
3 - 生成中
4 - 已完成
5 - 失败

图片生成详细状态枚举 (image detail_status):
0 - 待处理
1 - 生成中
2 - 已完成
3 - 失败

AI模型枚举 (analyze_model, generate_model):
1 - Gemini
2 - GPT
3 - DeepSeek

权限类型枚举 (permission_type):
1 - 菜单
2 - 按钮
3 - API

角色状态枚举 (role_status):
0 - 禁用
1 - 启用

权限状态枚举 (permission_status):
0 - 禁用
1 - 启用
*/
