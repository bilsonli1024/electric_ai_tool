package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type TaskCenterService struct{}

func NewTaskCenterService() *TaskCenterService {
	return &TaskCenterService{}
}

// GenerateTaskID 生成全局唯一的任务ID
// 格式：taskType_timestamp_randomString
func (s *TaskCenterService) GenerateTaskID(taskType string) string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomString := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s_%d_%s", taskType, timestamp, randomString)
}

// CreateBaseTask 创建任务中心底表记录
func (s *TaskCenterService) CreateBaseTask(taskID, taskType, operator string) error {
	now := time.Now().Unix()
	
	query := `
		INSERT INTO task_center_tab (task_id, task_type, task_status, operator, ctime, mtime)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, taskID, taskType, models.TaskStatusPending, operator, now, now)
	return err
}

// UpdateTaskStatus 更新任务状态
func (s *TaskCenterService) UpdateTaskStatus(taskID, status string) error {
	now := time.Now().Unix()
	
	query := `UPDATE task_center_tab SET task_status = ?, mtime = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, status, now, taskID)
	return err
}

// GetTaskByID 获取任务详情
func (s *TaskCenterService) GetTaskByID(taskID string) (*models.TaskCenterBase, error) {
	query := `SELECT id, task_id, task_type, task_status, operator, ctime, mtime FROM task_center_tab WHERE task_id = ?`
	
	var task models.TaskCenterBase
	err := config.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.TaskID, &task.TaskType, &task.TaskStatus,
		&task.Operator, &task.Ctime, &task.Mtime,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, err
	}
	
	return &task, nil
}

// GetTasks 获取任务列表
func (s *TaskCenterService) GetTasks(filter models.TaskCenterFilter) ([]*models.TaskCenterBase, int, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	
	if filter.Operator != "" {
		whereClause += " AND operator = ?"
		args = append(args, filter.Operator)
	}
	
	if filter.TaskType != "" {
		whereClause += " AND task_type = ?"
		args = append(args, filter.TaskType)
	}
	
	if filter.TaskStatus != "" {
		whereClause += " AND task_status = ?"
		args = append(args, filter.TaskStatus)
	}
	
	if filter.StartTime > 0 {
		whereClause += " AND ctime >= ?"
		args = append(args, filter.StartTime)
	}
	
	if filter.EndTime > 0 {
		whereClause += " AND ctime <= ?"
		args = append(args, filter.EndTime)
	}
	
	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_center_tab %s", whereClause)
	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	
	query := fmt.Sprintf(`
		SELECT id, task_id, task_type, task_status, operator, ctime, mtime
		FROM task_center_tab
		%s
		ORDER BY ctime DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, filter.Limit, filter.Offset)
	
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	tasks := []*models.TaskCenterBase{}
	for rows.Next() {
		var task models.TaskCenterBase
		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskType, &task.TaskStatus,
			&task.Operator, &task.Ctime, &task.Mtime,
		)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, &task)
	}
	
	return tasks, total, nil
}

// GetTaskDetail 获取任务完整详情（底表+详细表）
func (s *TaskCenterService) GetTaskDetail(taskID string) (*models.TaskCenterDetail, error) {
	// 获取底表数据
	baseTask, err := s.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	
	detail := &models.TaskCenterDetail{
		TaskCenterBase: *baseTask,
	}
	
	// 根据任务类型获取详细数据
	switch baseTask.TaskType {
	case models.TaskTypeCopywriting:
		copyDetail, err := s.getCopywritingDetail(taskID)
		if err != nil {
			return nil, err
		}
		detail.DetailData = copyDetail
		
	case models.TaskTypeImage:
		imageDetail, err := s.getImageDetail(taskID)
		if err != nil {
			return nil, err
		}
		detail.DetailData = imageDetail
	}
	
	return detail, nil
}

// getCopywritingDetail 获取文案生成详细数据
func (s *TaskCenterService) getCopywritingDetail(taskID string) (*models.CopywritingTaskDetail, error) {
	query := `
		SELECT id, task_id, competitor_urls, analysis_result, analyze_model,
		       user_selected_data, product_details, generated_copy, generate_model,
		       error_message, created_at, updated_at
		FROM copywriting_tasks_tab
		WHERE task_id = ?
	`
	
	var detail models.CopywritingTaskDetail
	err := config.DB.QueryRow(query, taskID).Scan(
		&detail.ID, &detail.TaskID, &detail.CompetitorURLs, &detail.AnalysisResult,
		&detail.AnalyzeModel, &detail.UserSelectedData, &detail.ProductDetails,
		&detail.GeneratedCopy, &detail.GenerateModel, &detail.ErrorMessage,
		&detail.CreatedAt, &detail.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("copywriting task detail not found")
	}
	if err != nil {
		return nil, err
	}
	
	return &detail, nil
}

// getImageDetail 获取图片生成详细数据
func (s *TaskCenterService) getImageDetail(taskID string) (*models.ImageTaskDetail, error) {
	query := `
		SELECT id, task_id, sku, keywords, selling_points, competitor_link,
		       copywriting_task_id, generate_model, aspect_ratio,
		       result_data, generated_image_urls, error_message,
		       created_at, updated_at
		FROM tasks_tab
		WHERE task_id = ?
	`
	
	var detail models.ImageTaskDetail
	err := config.DB.QueryRow(query, taskID).Scan(
		&detail.ID, &detail.TaskID, &detail.SKU, &detail.Keywords,
		&detail.SellingPoints, &detail.CompetitorLink, &detail.CopywritingTaskID,
		&detail.GenerateModel, &detail.AspectRatio, &detail.ResultData,
		&detail.GeneratedImageURLs, &detail.ErrorMessage,
		&detail.CreatedAt, &detail.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("image task detail not found")
	}
	if err != nil {
		return nil, err
	}
	
	return &detail, nil
}
