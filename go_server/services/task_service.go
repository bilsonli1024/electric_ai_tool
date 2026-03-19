package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

func (s *TaskService) CreateTask(userID int64, taskType string, sku string, keywords string, sellingPoints string, competitorLink string) (*models.Task, error) {
	query := `INSERT INTO tasks (user_id, task_type, sku, keywords, selling_points, competitor_link, status) 
              VALUES (?, ?, ?, ?, ?, ?, 'pending')`

	result, err := config.DB.Exec(query, userID, taskType, sku, keywords, sellingPoints, competitorLink)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	taskID, _ := result.LastInsertId()
	task := &models.Task{
		ID:             taskID,
		UserID:         userID,
		TaskType:       taskType,
		SKU:            sku,
		Keywords:       keywords,
		SellingPoints:  sellingPoints,
		CompetitorLink: competitorLink,
		Status:         "pending",
	}

	return task, nil
}

func (s *TaskService) UpdateTaskStatus(taskID int64, status string, resultData interface{}, errorMessage string) error {
	var resultJSON string
	if resultData != nil {
		data, err := json.Marshal(resultData)
		if err != nil {
			return fmt.Errorf("failed to marshal result data: %w", err)
		}
		resultJSON = string(data)
	}

	query := `UPDATE tasks SET status = ?, result_data = ?, error_message = ? WHERE id = ?`
	_, err := config.DB.Exec(query, status, resultJSON, errorMessage, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

func (s *TaskService) GetTaskByID(taskID int64) (*models.Task, error) {
	query := `SELECT t.id, t.user_id, t.task_type, t.sku, t.keywords, t.selling_points, t.competitor_link, 
              t.status, t.result_data, t.error_message, t.created_at, t.updated_at, u.username
              FROM tasks t
              LEFT JOIN users u ON t.user_id = u.id
              WHERE t.id = ?`

	task := &models.Task{}
	err := config.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.UserID, &task.TaskType, &task.SKU, &task.Keywords,
		&task.SellingPoints, &task.CompetitorLink, &task.Status, &task.ResultData,
		&task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt, &task.Username,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (s *TaskService) GetUserTasks(userID int64, taskType string, limit int, offset int) ([]models.Task, int, error) {
	var tasks []models.Task
	var total int

	countQuery := `SELECT COUNT(*) FROM tasks WHERE user_id = ?`
	args := []interface{}{userID}

	if taskType != "" {
		countQuery += ` AND task_type = ?`
		args = append(args, taskType)
	}

	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	query := `SELECT t.id, t.user_id, t.task_type, t.sku, t.keywords, t.selling_points, t.competitor_link, 
              t.status, t.result_data, t.error_message, t.created_at, t.updated_at, u.username
              FROM tasks t
              LEFT JOIN users u ON t.user_id = u.id
              WHERE t.user_id = ?`

	if taskType != "" {
		query += ` AND t.task_type = ?`
	}

	query += ` ORDER BY t.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.UserID, &task.TaskType, &task.SKU, &task.Keywords,
			&task.SellingPoints, &task.CompetitorLink, &task.Status, &task.ResultData,
			&task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt, &task.Username,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (s *TaskService) GetAllTasks(limit int, offset int) ([]models.Task, int, error) {
	var tasks []models.Task
	var total int

	countQuery := `SELECT COUNT(*) FROM tasks`
	err := config.DB.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	query := `SELECT t.id, t.user_id, t.task_type, t.sku, t.keywords, t.selling_points, t.competitor_link, 
              t.status, t.result_data, t.error_message, t.created_at, t.updated_at, u.username
              FROM tasks t
              LEFT JOIN users u ON t.user_id = u.id
              ORDER BY t.created_at DESC LIMIT ? OFFSET ?`

	rows, err := config.DB.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.UserID, &task.TaskType, &task.SKU, &task.Keywords,
			&task.SellingPoints, &task.CompetitorLink, &task.Status, &task.ResultData,
			&task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt, &task.Username,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}
