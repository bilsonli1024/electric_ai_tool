package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type UnifiedTaskService struct{}

func NewUnifiedTaskService() *UnifiedTaskService {
	return &UnifiedTaskService{}
}

// CreateTask 创建任务
func (s *UnifiedTaskService) CreateTask(task *models.UnifiedTask) (int64, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	result, err := db.Exec(`
		INSERT INTO unified_tasks_tab (
			user_id, username, task_name, task_type, status,
			task_config, analyze_model, generate_model
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		task.UserID, task.Username, task.TaskName, task.TaskType, task.Status,
		task.TaskConfig, task.AnalyzeModel, task.GenerateModel,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get task ID: %w", err)
	}

	log.Printf("Created task: id=%d, type=%s, name=%s, user=%s", taskID, task.TaskType, task.TaskName, task.Username)
	return taskID, nil
}

// UpdateTaskStatus 更新任务状态
func (s *UnifiedTaskService) UpdateTaskStatus(taskID int64, status int, errorMessage string) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(`
		UPDATE unified_tasks_tab 
		SET status = ?, error_message = ?, updated_at = NOW()
		WHERE id = ?`,
		status, errorMessage, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	log.Printf("Updated task %d status to %d", taskID, status)
	return nil
}

// UpdateTaskResult 更新任务结果
func (s *UnifiedTaskService) UpdateTaskResult(taskID int64, analysisResult, generationResult string, status int) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(`
		UPDATE unified_tasks_tab 
		SET analysis_result = ?, generation_result = ?, status = ?, updated_at = NOW()
		WHERE id = ?`,
		analysisResult, generationResult, status, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}

	log.Printf("Updated task %d results, status=%d", taskID, status)
	return nil
}

// GetTaskByID 根据ID获取任务
func (s *UnifiedTaskService) GetTaskByID(taskID int64) (*models.UnifiedTask, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	task := &models.UnifiedTask{}
	err := db.QueryRow(`
		SELECT id, user_id, username, task_name, task_type, status,
			task_config, analysis_result, generation_result,
			analyze_model, generate_model, error_message,
			created_at, updated_at
		FROM unified_tasks_tab
		WHERE id = ?`,
		taskID,
	).Scan(
		&task.ID, &task.UserID, &task.Username, &task.TaskName, &task.TaskType, &task.Status,
		&task.TaskConfig, &task.AnalysisResult, &task.GenerationResult,
		&task.AnalyzeModel, &task.GenerateModel, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}

	return task, nil
}

// GetTasks 获取任务列表（支持多条件筛选）
func (s *UnifiedTaskService) GetTasks(filter models.TaskFilter) ([]*models.UnifiedTask, int, error) {
	db := config.GetDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	// 构建查询条件
	whereConditions := []string{}
	args := []interface{}{}

	if filter.UserID != nil {
		whereConditions = append(whereConditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.TaskType != nil {
		whereConditions = append(whereConditions, "task_type = ?")
		args = append(args, *filter.TaskType)
	}

	if filter.Status != nil {
		whereConditions = append(whereConditions, "status = ?")
		args = append(args, *filter.Status)
	}

	if filter.StartTime != nil {
		whereConditions = append(whereConditions, "created_at >= ?")
		args = append(args, *filter.StartTime)
	}

	if filter.EndTime != nil {
		whereConditions = append(whereConditions, "created_at <= ?")
		args = append(args, *filter.EndTime)
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM unified_tasks_tab %s", whereClause)
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// 查询任务列表
	query := fmt.Sprintf(`
		SELECT id, user_id, username, task_name, task_type, status,
			task_config, analysis_result, generation_result,
			analyze_model, generate_model, error_message,
			created_at, updated_at
		FROM unified_tasks_tab
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		whereClause,
	)

	args = append(args, filter.Limit, filter.Offset)
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	tasks := []*models.UnifiedTask{}
	for rows.Next() {
		task := &models.UnifiedTask{}
		err := rows.Scan(
			&task.ID, &task.UserID, &task.Username, &task.TaskName, &task.TaskType, &task.Status,
			&task.TaskConfig, &task.AnalysisResult, &task.GenerationResult,
			&task.AnalyzeModel, &task.GenerateModel, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning task row: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

// GetTaskStatistics 获取任务统计信息
func (s *UnifiedTaskService) GetTaskStatistics(userID *int64) (*models.TaskStatistics, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	whereClause := ""
	args := []interface{}{}
	if userID != nil {
		whereClause = "WHERE user_id = ?"
		args = append(args, *userID)
	}

	stats := &models.TaskStatistics{}
	
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status IN (0, 1, 2) THEN 1 ELSE 0 END) as processing,
			SUM(CASE WHEN status IN (10, 11) THEN 1 ELSE 0 END) as failed,
			SUM(CASE WHEN task_type = 'copywriting' THEN 1 ELSE 0 END) as copywriting,
			SUM(CASE WHEN task_type = 'image' THEN 1 ELSE 0 END) as image
		FROM unified_tasks_tab
		%s`,
		whereClause,
	)

	err := db.QueryRow(query, args...).Scan(
		&stats.TotalTasks,
		&stats.CompletedTasks,
		&stats.ProcessingTasks,
		&stats.FailedTasks,
		&stats.CopywritingTasks,
		&stats.ImageTasks,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query statistics: %w", err)
	}

	return stats, nil
}
