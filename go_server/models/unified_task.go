package models

import "time"

// UnifiedTask 统一任务模型
type UnifiedTask struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	TaskName  string    `json:"task_name"`
	TaskType  string    `json:"task_type"` // copywriting, image
	Status    int       `json:"status"`    // 0:分析中, 1:分析完成/待生成, 2:生成中, 3:已完成, 10:分析失败, 11:生成失败
	
	TaskConfig       string `json:"task_config,omitempty"`        // JSON格式的任务配置
	AnalysisResult   string `json:"analysis_result,omitempty"`    // JSON格式的分析结果
	GenerationResult string `json:"generation_result,omitempty"`  // JSON格式的生成结果
	
	AnalyzeModel  string `json:"analyze_model"`
	GenerateModel string `json:"generate_model"`
	
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// TaskFilter 任务筛选条件
type TaskFilter struct {
	UserID    *int64  // 按创建者筛选
	TaskType  *string // 按任务类型筛选
	Status    *int    // 按状态筛选
	StartTime *time.Time // 开始时间
	EndTime   *time.Time // 结束时间
	Limit     int
	Offset    int
}

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TotalTasks      int `json:"total_tasks"`
	CompletedTasks  int `json:"completed_tasks"`
	ProcessingTasks int `json:"processing_tasks"`
	FailedTasks     int `json:"failed_tasks"`
	CopywritingTasks int `json:"copywriting_tasks"`
	ImageTasks      int `json:"image_tasks"`
}
