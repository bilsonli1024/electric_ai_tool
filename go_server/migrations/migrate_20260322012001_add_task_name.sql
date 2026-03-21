-- 添加task_name字段到copywriting_tasks_tab
-- 用于任务中心列表显示

ALTER TABLE copywriting_tasks_tab 
ADD COLUMN task_name VARCHAR(255) DEFAULT '' COMMENT '任务名称' AFTER task_id;
