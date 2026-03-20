-- Migration: Add email verification code table
-- Created: 2026-03-20

-- 邮箱验证码表
CREATE TABLE IF NOT EXISTS email_verification_codes_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(256) NOT NULL COMMENT '邮箱地址',
    code VARCHAR(6) NOT NULL COMMENT '6位验证码',
    purpose VARCHAR(32) NOT NULL COMMENT 'register:注册, reset:重置密码',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    used TINYINT DEFAULT 0 COMMENT '0:未使用, 1:已使用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_code (code),
    INDEX idx_expires_at (expires_at),
    INDEX idx_purpose (purpose)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='邮箱验证码表';
