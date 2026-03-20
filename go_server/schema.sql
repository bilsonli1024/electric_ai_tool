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
    INDEX idx_session_id (session_id),
    FOREIGN KEY (user_id) REFERENCES users_tab(id) ON DELETE CASCADE
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
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users_tab(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='密码重置令牌表';

-- 会话表
CREATE TABLE IF NOT EXISTS sessions_tab (
    id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_active_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users_tab(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- 任务表
CREATE TABLE IF NOT EXISTS tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    sku VARCHAR(128),
    keywords VARCHAR(512),
    selling_points TEXT,
    competitor_link VARCHAR(512),
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
    FOREIGN KEY (user_id) REFERENCES users_tab(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

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
    INDEX idx_created_at (created_at),
    FOREIGN KEY (task_id) REFERENCES tasks_tab(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users_tab(id) ON DELETE CASCADE
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
    INDEX idx_image_type (image_type),
    FOREIGN KEY (user_id) REFERENCES users_tab(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='CDN图片记录表';
