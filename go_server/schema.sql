-- 创建数据库
CREATE DATABASE IF NOT EXISTS electric_ai_tool DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE electric_ai_tool;

-- 用户表
CREATE TABLE IF NOT EXISTS users_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(128) UNIQUE NOT NULL,
    email VARCHAR(256) UNIQUE NOT NULL,
    password_hash VARCHAR(64) NOT NULL COMMENT 'MD5加盐后的密码',
    salt VARCHAR(32) NOT NULL COMMENT '密码盐值',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    status TINYINT DEFAULT 1 COMMENT '1:激活, 0:禁用',
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户账号表';

-- 角色表
CREATE TABLE IF NOT EXISTS roles_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    role_name VARCHAR(64) UNIQUE NOT NULL,
    role_code VARCHAR(64) UNIQUE NOT NULL COMMENT '角色代码',
    description VARCHAR(256),
    status TINYINT DEFAULT 1 COMMENT '1:启用, 0:禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_role_code (role_code),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 权限表
CREATE TABLE IF NOT EXISTS permissions_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    permission_name VARCHAR(128) NOT NULL,
    permission_code VARCHAR(128) UNIQUE NOT NULL COMMENT '权限代码',
    resource_type VARCHAR(64) COMMENT '资源类型:copywriting,image_generation,model_test',
    action VARCHAR(32) COMMENT '操作:create,read,update,delete',
    description VARCHAR(256),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_permission_code (permission_code),
    INDEX idx_resource_type (resource_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    INDEX idx_permission_id (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user_id (user_id),
    INDEX idx_role_id (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 用户登录日志表
CREATE TABLE IF NOT EXISTS user_login_log_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    login_type TINYINT NOT NULL COMMENT '1:登录, 2:登出, 3:切换用户',
    login_ip VARCHAR(64),
    user_agent VARCHAR(512),
    session_id VARCHAR(64),
    login_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_login_time (login_time),
    INDEX idx_session_id (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户登录日志表';

-- 密码重置令牌表
CREATE TABLE IF NOT EXISTS password_reset_tokens_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used TINYINT DEFAULT 0 COMMENT '0:未使用, 1:已使用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码重置令牌表';

-- 会话表
CREATE TABLE IF NOT EXISTS sessions_tab (
    id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- 任务表（图片生成任务）
CREATE TABLE IF NOT EXISTS tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    sku VARCHAR(128),
    keywords VARCHAR(512),
    selling_points TEXT,
    competitor_link VARCHAR(512),
    copywriting_task_id BIGINT COMMENT '关联的文案生成任务ID',
    analyze_model VARCHAR(32) DEFAULT 'gemini' COMMENT '分析使用的模型',
    generate_model VARCHAR(32) DEFAULT 'gemini' COMMENT '生成使用的模型',
    status TINYINT DEFAULT 0 COMMENT '0:分析中, 1:分析完成, 2:生成图片中, 3:已完成, 10:分析失败, 11:生成失败',
    result_data TEXT COMMENT '分析结果JSON',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_copywriting_task_id (copywriting_task_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片生成任务表';

-- 任务历史表
-- 字段说明：
-- task_id: 关联的任务ID
-- user_id: 创建该历史版本的用户ID
-- version: 版本号，同一任务的第N次生成
-- prompt: 生成图片使用的提示词（AI生成指令）
-- aspect_ratio: 图片宽高比，如"1:1"或"4:5"
-- product_images_urls: 用户上传的产品白底图CDN链接，JSON数组格式存储多张图片
-- style_ref_image_url: 风格参考图的CDN链接（可选）
-- generated_image_url: AI生成的最终图片CDN链接
-- edit_instruction: 图片编辑指令（如果是编辑操作）
-- status: 该版本的生成状态（成功/失败）
-- error_message: 失败时的错误信息
CREATE TABLE IF NOT EXISTS task_history_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    version INT DEFAULT 1 COMMENT '版本号',
    model VARCHAR(32) DEFAULT 'gemini' COMMENT '使用的AI模型',
    prompt TEXT COMMENT 'AI生成提示词',
    aspect_ratio VARCHAR(16) COMMENT '图片宽高比',
    product_images_urls TEXT COMMENT '产品图CDN链接JSON数组',
    style_ref_image_url VARCHAR(512) COMMENT '风格参考图CDN链接',
    generated_image_url VARCHAR(512) COMMENT '生成图片CDN链接',
    edit_instruction TEXT COMMENT '编辑指令',
    status TINYINT DEFAULT 1 COMMENT '0:失败, 1:成功',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_id (task_id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务历史版本表';

-- CDN图片记录表
CREATE TABLE IF NOT EXISTS cdn_images_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    original_filename VARCHAR(256),
    cdn_url VARCHAR(512) NOT NULL,
    cdn_key VARCHAR(512) NOT NULL COMMENT 'CDN存储key',
    file_size BIGINT,
    mime_type VARCHAR(64),
    image_type VARCHAR(32) COMMENT 'product:产品图, style_ref:风格参考图, generated:生成图',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_cdn_key (cdn_key),
    INDEX idx_image_type (image_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CDN图片记录表';

-- 文案生成任务表
CREATE TABLE IF NOT EXISTS copywriting_tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    task_name VARCHAR(256) COMMENT '任务名称',
    competitor_urls TEXT COMMENT '竞品链接JSON数组',
    analysis_result TEXT COMMENT '竞品分析结果JSON',
    product_details TEXT COMMENT '产品详情JSON',
    generated_copy TEXT COMMENT '生成的文案JSON',
    status TINYINT DEFAULT 0 COMMENT '0:分析中, 1:分析完成, 2:生成中, 3:已完成, 10:分析失败, 11:生成失败',
    analyze_model VARCHAR(32) DEFAULT 'gemini' COMMENT '分析使用的模型',
    generate_model VARCHAR(32) DEFAULT 'gemini' COMMENT '生成使用的模型',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_task_name (task_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文案生成任务表';
