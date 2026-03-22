-- 添加本地文件路径字段
-- 用于保存图片文件的本地存储路径

USE electric_ai_tool;

-- 为tasks_tab添加本地文件路径字段
ALTER TABLE tasks_tab 
ADD COLUMN local_image_paths TEXT COMMENT '本地图片文件路径(JSON数组)' AFTER generated_image_urls;

-- 为copywriting_tasks_tab添加上传图片的本地路径字段
ALTER TABLE copywriting_tasks_tab 
ADD COLUMN uploaded_image_paths TEXT COMMENT '上传的产品图片本地路径(JSON数组)' AFTER task_name;

-- 说明：
-- 1. local_image_paths: 存储AI生成图片的本地文件路径
-- 2. uploaded_image_paths: 存储用户上传的产品图片本地路径  
-- 3. 两个字段都使用JSON数组格式存储多个文件路径
-- 4. generated_image_urls可以同时存储访问URL或CDN URL（兼容）
