package services

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type TaskCenterService struct{}

func init() {
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
}

func NewTaskCenterService() *TaskCenterService {
	return &TaskCenterService{}
}

// GenerateTaskID 生成全局唯一的任务ID
// 格式：缩写 + yyyyMMddHHmmss + 5位随机数字
// 例如：CP20260321135530 12345
func (s *TaskCenterService) GenerateTaskID(taskType string) string {
	// 任务类型缩写
	var prefix string
	switch taskType {
	case models.TaskTypeCopywriting:
		prefix = "CP"
	case models.TaskTypeImage:
		prefix = "IG"
	default:
		prefix = "TK"
	}
	
	// 时间格式：yyyyMMddHHmmss
	timestamp := time.Now().Format("20060102150405")
	
	// 生成5位随机数字 (10000-99999)
	randomNum := rand.Intn(90000) + 10000
	
	return fmt.Sprintf("%s%s%05d", prefix, timestamp, randomNum)
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

// GetTasks 获取任务列表（带task_name和sku字段）
func (s *TaskCenterService) GetTasks(filter models.TaskCenterFilter) ([]*models.TaskCenterListItem, int, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	
	if filter.Operator != "" {
		whereClause += " AND t.operator = ?"
		args = append(args, filter.Operator)
	}
	
	if filter.TaskType != "" {
		whereClause += " AND t.task_type = ?"
		args = append(args, filter.TaskType)
	}
	
	if filter.TaskStatus != "" {
		whereClause += " AND t.task_status = ?"
		args = append(args, filter.TaskStatus)
	}
	
	if filter.StartTime > 0 {
		whereClause += " AND t.ctime >= ?"
		args = append(args, filter.StartTime)
	}
	
	if filter.EndTime > 0 {
		whereClause += " AND t.ctime <= ?"
		args = append(args, filter.EndTime)
	}
	
	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_center_tab t %s", whereClause)
	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// 查询数据（LEFT JOIN详细表获取task_name和sku）
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	
	query := fmt.Sprintf(`
		SELECT t.id, t.task_id, t.task_type, t.task_status, t.operator, t.ctime, t.mtime,
		       COALESCE(c.task_name, '') as task_name,
		       COALESCE(i.sku, '') as sku
		FROM task_center_tab t
		LEFT JOIN copywriting_tasks_tab c ON t.task_id = c.task_id AND t.task_type = 'copywriting'
		LEFT JOIN tasks_tab i ON t.task_id = i.task_id AND t.task_type = 'image'
		%s
		ORDER BY t.ctime DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, filter.Limit, filter.Offset)
	
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	tasks := []*models.TaskCenterListItem{}
	for rows.Next() {
		var task models.TaskCenterListItem
		err := rows.Scan(
			&task.ID, &task.TaskID, &task.TaskType, &task.TaskStatus,
			&task.Operator, &task.Ctime, &task.Mtime,
			&task.TaskName, &task.SKU,
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
		SELECT id, task_id, 
		       COALESCE(task_name, '') as task_name,
		       competitor_urls, 
		       COALESCE(analysis_result, ''), COALESCE(analyze_model, ''),
		       COALESCE(user_selected_data, ''), COALESCE(product_details, ''), 
		       COALESCE(generated_copy, ''), COALESCE(generate_model, ''),
		       COALESCE(error_message, ''), created_at, updated_at
		FROM copywriting_tasks_tab
		WHERE task_id = ?
	`
	
	var detail models.CopywritingTaskDetail

	err := config.DB.QueryRow(query, taskID).Scan(
		&detail.ID, &detail.TaskID, &detail.TaskName,
		&detail.CompetitorURLs, &detail.AnalysisResult,
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
		SELECT id, task_id, 
		       COALESCE(sku, ''), COALESCE(keywords, ''), 
		       COALESCE(selling_points, ''), COALESCE(competitor_link, ''),
		       COALESCE(copywriting_task_id, ''), COALESCE(generate_model, ''), 
		       COALESCE(aspect_ratio, ''),
		       COALESCE(result_data, ''), COALESCE(generated_image_urls, ''), 
		       COALESCE(error_message, ''),
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

// CopyCopywritingTask 复制文案生成任务（只复制输入参数，清空结果）
func (s *TaskCenterService) CopyCopywritingTask(newTaskID string, originalTask *models.CopywritingTaskDetail) error {
	query := `
		INSERT INTO copywriting_tasks_tab (
			task_id, task_name, competitor_urls
		) VALUES (?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, newTaskID, originalTask.TaskName, originalTask.CompetitorURLs)
	return err
}

// CopyImageTask 复制图片生成任务（只复制输入参数，清空结果）
func (s *TaskCenterService) CopyImageTask(newTaskID string, originalTask *models.ImageTaskDetail) error {
	query := `
		INSERT INTO tasks_tab (
			task_id, sku, keywords, selling_points, competitor_link,
			copywriting_task_id, generate_model, aspect_ratio
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, newTaskID, originalTask.SKU, originalTask.Keywords,
		originalTask.SellingPoints, originalTask.CompetitorLink, originalTask.CopywritingTaskID,
		originalTask.GenerateModel, originalTask.AspectRatio)
	return err
}

