package services

import (
	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type ImageTaskService struct{}

func NewImageTaskService() *ImageTaskService {
	return &ImageTaskService{}
}

// CreateTask 创建图片生成任务详细记录
func (s *ImageTaskService) CreateTask(taskID, sku, keywords, sellingPoints, competitorLink, copywritingTaskID, generateModel, aspectRatio string) error {
	query := `
		INSERT INTO tasks_tab (
			task_id, sku, keywords, selling_points, competitor_link,
			copywriting_task_id, generate_model, aspect_ratio
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, taskID, sku, keywords, sellingPoints,
		competitorLink, copywritingTaskID, generateModel, aspectRatio)
	return err
}

// SaveResultData 保存生成结果
func (s *ImageTaskService) SaveResultData(taskID, resultData, imageURLs string) error {
	query := `
		UPDATE tasks_tab 
		SET result_data = ?, generated_image_urls = ?
		WHERE task_id = ?
	`
	_, err := config.DB.Exec(query, resultData, imageURLs, taskID)
	return err
}

// SaveError 保存错误信息
func (s *ImageTaskService) SaveError(taskID, errorMessage string) error {
	query := `UPDATE tasks_tab SET error_message = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, errorMessage, taskID)
	return err
}

// GetTaskByID 获取任务详情
func (s *ImageTaskService) GetTaskByID(taskID string) (*models.ImageTaskDetail, error) {
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
	
	var task models.ImageTaskDetail
	err := config.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.TaskID, &task.SKU, &task.Keywords,
		&task.SellingPoints, &task.CompetitorLink, &task.CopywritingTaskID,
		&task.GenerateModel, &task.AspectRatio, &task.ResultData,
		&task.GeneratedImageURLs, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &task, nil
}
