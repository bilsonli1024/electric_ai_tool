-- 迁移: 重新设计任务中心架构
-- 原因: 采用统一底表+详细表的设计，职责更清晰
-- 日期: 2026-03-21 11:28:06

USE electric_ai_tool;

-- ============================================
-- 1. 创建任务中心统一底表
-- ============================================
CREATE TABLE IF NOT EXISTS task_center_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增主键',
    task_id VARCHAR(64) UNIQUE NOT NULL COMMENT '全局唯一任务ID，格式：taskType_timestamp_randomString',
    task_type VARCHAR(32) NOT NULL COMMENT '任务类型: copywriting(文案生成), image(图片生成)',
    task_status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '任务状态: pending(待处理), ongoing(进行中), completed(已完成), failed(失败)',
    operator VARCHAR(256) NOT NULL COMMENT '操作者邮箱',
    ctime BIGINT NOT NULL COMMENT '创建时间（秒级时间戳）',
    mtime BIGINT NOT NULL COMMENT '更新时间（秒级时间戳）',
    
    INDEX idx_task_id (task_id),
    INDEX idx_task_type (task_type),
    INDEX idx_task_status (task_status),
    INDEX idx_operator (operator),
    INDEX idx_ctime (ctime),
    INDEX idx_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务中心统一底表';

-- ============================================
-- 2. 修改文案生成任务表结构
-- ============================================
-- 删除旧表
DROP TABLE IF EXISTS copywriting_tasks_tab_old;
-- 重命名当前表为旧表
RENAME TABLE copywriting_tasks_tab TO copywriting_tasks_tab_old;

-- 创建新的文案生成任务表
CREATE TABLE IF NOT EXISTS copywriting_tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(64) NOT NULL COMMENT '关联task_center_tab的task_id',
    
    -- 分析阶段数据
    competitor_urls TEXT COMMENT '竞品链接JSON数组',
    analysis_result TEXT COMMENT '竞品分析结果JSON（AI初始生成）',
    analyze_model VARCHAR(32) DEFAULT 'gemini' COMMENT '分析使用的模型',
    
    -- 生成阶段数据
    user_selected_data TEXT COMMENT '用户选择后的数据JSON（用于生成）',
    product_details TEXT COMMENT '产品详情JSON',
    generated_copy TEXT COMMENT '生成的文案JSON',
    generate_model VARCHAR(32) DEFAULT 'gemini' COMMENT '生成使用的模型',
    
    -- 错误信息
    error_message TEXT,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_task_id (task_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文案生成任务详细表';

-- ============================================
-- 3. 修改图片生成任务表结构
-- ============================================
-- 删除旧表
DROP TABLE IF EXISTS tasks_tab_old;
-- 重命名当前表为旧表
RENAME TABLE tasks_tab TO tasks_tab_old;

-- 创建新的图片生成任务表
CREATE TABLE IF NOT EXISTS tasks_tab (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(64) NOT NULL COMMENT '关联task_center_tab的task_id',
    
    -- 任务参数
    sku VARCHAR(128),
    keywords VARCHAR(512),
    selling_points TEXT,
    competitor_link TEXT COMMENT '竞品链接',
    copywriting_task_id VARCHAR(64) COMMENT '关联的文案生成task_id（如果有）',
    
    -- 生成配置
    generate_model VARCHAR(32) DEFAULT 'gemini' COMMENT '生成使用的模型',
    aspect_ratio VARCHAR(16) DEFAULT '1:1' COMMENT '图片宽高比',
    
    -- 结果数据
    result_data TEXT COMMENT '生成结果JSON',
    generated_image_urls TEXT COMMENT '生成的图片URL列表JSON',
    
    -- 错误信息
    error_message TEXT,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_task_id (task_id),
    INDEX idx_created_at (created_at),
    INDEX idx_copywriting_task_id (copywriting_task_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='图片生成任务详细表';

-- ============================================
-- 4. 数据迁移（如果需要）
-- ============================================
-- 注意：由于项目未上线，如果没有重要数据可以跳过迁移
-- 如果需要迁移旧数据，请手动执行以下步骤：
-- 1. 从 copywriting_tasks_tab_old 迁移数据到新表
-- 2. 从 tasks_tab_old 迁移数据到新表
-- 3. 为迁移的数据创建 task_center_tab 记录
