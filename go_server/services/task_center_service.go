package services

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"
)

type TaskCenterService struct{}

func NewTaskCenterService() *TaskCenterService {
	return &TaskCenterService{}
}

// GenerateTaskID 生成任务ID
// 格式: 任务类型缩写 + YYYYMMDDHHMMSS + 5位随机数
func (s *TaskCenterService) GenerateTaskID(taskType int) string {
	var prefix string
	switch taskType {
	case models.TaskTypeCopywriting:
		prefix = "CP"
	case models.TaskTypeImage:
		prefix = "IG"
	default:
		prefix = "TK"
	}

	now := time.Now()
	timestamp := now.Format("20060102150405")
	randomNum := rand.Intn(100000)

	return fmt.Sprintf("%s%s%05d", prefix, timestamp, randomNum)
}

// CreateBaseTask 创建任务中心底表记录
func (s *TaskCenterService) CreateBaseTask(taskID string, taskType int, operator string) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `
		INSERT INTO task_center_tab (task_id, task_type, task_status, operator, ctime, mtime)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, taskID, taskType, models.TaskStatusPending, operator, currentTime, currentTime)
	return err
}

// UpdateTaskStatus 更新任务状态
func (s *TaskCenterService) UpdateTaskStatus(taskID string, taskStatus int) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `UPDATE task_center_tab SET task_status = ?, mtime = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, taskStatus, currentTime, taskID)
	return err
}

// ListTasks 查询任务列表（带筛选和分页）
func (s *TaskCenterService) ListTasks(filter models.TaskCenterFilter) ([]models.TaskCenterListItem, int, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filter.Operator != "" {
		whereClause += " AND tc.operator = ?"
		args = append(args, filter.Operator)
	}

	if filter.TaskType > 0 {
		whereClause += " AND tc.task_type = ?"
		args = append(args, filter.TaskType)
	}

	if filter.TaskStatus > 0 {
		whereClause += " AND tc.task_status = ?"
		args = append(args, filter.TaskStatus)
	}

	if filter.StartTime > 0 {
		whereClause += " AND tc.ctime >= ?"
		args = append(args, filter.StartTime)
	}

	if filter.EndTime > 0 {
		whereClause += " AND tc.ctime <= ?"
		args = append(args, filter.EndTime)
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_center_tab tc %s", whereClause)
	var total int
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表数据
	query := fmt.Sprintf(`
		SELECT 
			tc.id, tc.task_id, tc.task_type, tc.task_status, 
			tc.operator, tc.ctime, tc.mtime,
			CASE 
				WHEN tc.task_type = %d THEN COALESCE(ct.task_name, '')
				ELSE ''
			END as task_name,
			CASE 
				WHEN tc.task_type = %d THEN COALESCE(it.sku, '')
				ELSE ''
			END as sku
		FROM task_center_tab tc
		LEFT JOIN copywriting_tasks_tab ct ON tc.task_id = ct.task_id AND tc.task_type = %d
		LEFT JOIN tasks_tab it ON tc.task_id = it.task_id AND tc.task_type = %d
		%s
		ORDER BY tc.ctime DESC
		LIMIT ? OFFSET ?
	`, models.TaskTypeCopywriting, models.TaskTypeImage, 
	   models.TaskTypeCopywriting, models.TaskTypeImage, whereClause)

	args = append(args, filter.Limit, filter.Offset)
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []models.TaskCenterListItem{} // 初始化为空数组
	for rows.Next() {
		var item models.TaskCenterListItem
		err := rows.Scan(
			&item.ID, &item.TaskID, &item.TaskType, &item.TaskStatus,
			&item.Operator, &item.Ctime, &item.Mtime,
			&item.TaskName, &item.SKU,
		)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, nil
}

// GetTaskDetail 获取任务详情（包含详细数据）
func (s *TaskCenterService) GetTaskDetail(taskID string) (*models.TaskCenterDetail, error) {
	// 1. 查询底表
	baseQuery := `
		SELECT id, task_id, task_type, task_status, operator, ctime, mtime
		FROM task_center_tab
		WHERE task_id = ?
	`
	
	var detail models.TaskCenterDetail
	err := config.DB.QueryRow(baseQuery, taskID).Scan(
		&detail.ID, &detail.TaskID, &detail.TaskType, &detail.TaskStatus,
		&detail.Operator, &detail.Ctime, &detail.Mtime,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, err
	}

	// 2. 根据任务类型查询详细数据
	switch detail.TaskType {
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
		
	default:
		return nil, fmt.Errorf("unknown task type: %d", detail.TaskType)
	}

	return &detail, nil
}

// getCopywritingDetail 获取文案生成详细数据
func (s *TaskCenterService) getCopywritingDetail(taskID string) (*models.CopywritingTaskDetail, error) {
	query := `
		SELECT id, task_id, 
		       COALESCE(task_name, '') as task_name,
		       COALESCE(detail_status, 0) as detail_status,
		       competitor_urls, 
		       COALESCE(analysis_result, ''), COALESCE(analyze_model, 0),
		       COALESCE(user_selected_data, ''), COALESCE(product_details, ''), 
		       COALESCE(generated_copy, ''), COALESCE(generate_model, 0),
		       COALESCE(error_message, ''), COALESCE(fail_msg, ''), 
		       ctime, mtime
		FROM copywriting_tasks_tab
		WHERE task_id = ?
	`
	
	var detail models.CopywritingTaskDetail

	err := config.DB.QueryRow(query, taskID).Scan(
		&detail.ID, &detail.TaskID, &detail.TaskName, &detail.DetailStatus,
		&detail.CompetitorURLs, &detail.AnalysisResult,
		&detail.AnalyzeModel, &detail.UserSelectedData, &detail.ProductDetails,
		&detail.GeneratedCopy, &detail.GenerateModel, &detail.ErrorMessage, &detail.FailMsg,
		&detail.Ctime, &detail.Mtime,
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
		SELECT id, task_id, COALESCE(detail_status, 0) as detail_status,
		       COALESCE(sku, ''), COALESCE(keywords, ''), 
		       COALESCE(selling_points, ''), COALESCE(competitor_link, ''),
		       COALESCE(copywriting_task_id, ''), COALESCE(generate_model, 0), 
		       COALESCE(aspect_ratio, ''),
		       COALESCE(result_data, ''), COALESCE(generated_image_urls, ''), 
		       COALESCE(error_message, ''), COALESCE(fail_msg, ''),
		       ctime, mtime
		FROM tasks_tab
		WHERE task_id = ?
	`
	
	var detail models.ImageTaskDetail
	err := config.DB.QueryRow(query, taskID).Scan(
		&detail.ID, &detail.TaskID, &detail.DetailStatus, &detail.SKU, &detail.Keywords,
		&detail.SellingPoints, &detail.CompetitorLink, &detail.CopywritingTaskID,
		&detail.GenerateModel, &detail.AspectRatio, &detail.ResultData,
		&detail.GeneratedImageURLs, &detail.ErrorMessage, &detail.FailMsg,
		&detail.Ctime, &detail.Mtime,
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
	currentTime := utils.GetCurrentTimestamp()
	query := `
		INSERT INTO copywriting_tasks_tab (
			task_id, task_name, competitor_urls, analyze_model, detail_status, ctime, mtime
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, newTaskID, originalTask.TaskName, originalTask.CompetitorURLs, 
		originalTask.AnalyzeModel, models.CopywritingStatusPending, currentTime, currentTime)
	return err
}

// CopyImageTask 复制图片生成任务（只复制输入参数，清空结果）
func (s *TaskCenterService) CopyImageTask(newTaskID string, originalTask *models.ImageTaskDetail) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `
		INSERT INTO tasks_tab (
			task_id, sku, keywords, selling_points, competitor_link,
			copywriting_task_id, generate_model, aspect_ratio, detail_status, ctime, mtime
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, newTaskID, originalTask.SKU, originalTask.Keywords,
		originalTask.SellingPoints, originalTask.CompetitorLink, originalTask.CopywritingTaskID,
		originalTask.GenerateModel, originalTask.AspectRatio, models.ImageStatusPending, 
		currentTime, currentTime)
	return err
}
