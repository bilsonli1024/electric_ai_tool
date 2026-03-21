-- 迁移: 修改 tasks_tab 表的 competitor_link 字段类型
-- 原因: VARCHAR(512) 不足以存储长URL，改为 TEXT 类型
-- 日期: 2026-03-21 10:37:51

USE electric_ai_tool;

-- 修改 tasks_tab 表的 competitor_link 字段
ALTER TABLE tasks_tab 
MODIFY COLUMN competitor_link TEXT COMMENT '竞品链接，支持长URL';
