package services

import (
	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type CopywritingTaskService struct{}

func NewCopywritingTaskService() *CopywritingTaskService {
	return &CopywritingTaskService{}
}

// CreateTask 创建文案生成任务详细记录
func (s *CopywritingTaskService) CreateTask(taskID, taskName, competitorURLs, analyzeModel string) error {
	query := `
		INSERT INTO copywriting_tasks_tab (task_id, task_name, competitor_urls, analyze_model)
		VALUES (?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, taskID, taskName, competitorURLs, analyzeModel)
	return err
}

// SaveAnalysisResult 保存分析结果
func (s *CopywritingTaskService) SaveAnalysisResult(taskID, analysisResult string) error {
	query := `UPDATE copywriting_tasks_tab SET analysis_result = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, analysisResult, taskID)
	return err
}

// SaveUserSelectedData 保存用户选择的数据
func (s *CopywritingTaskService) SaveUserSelectedData(taskID, userSelectedData, productDetails string) error {
	query := `
		UPDATE copywriting_tasks_tab 
		SET user_selected_data = ?, product_details = ?
		WHERE task_id = ?
	`
	_, err := config.DB.Exec(query, userSelectedData, productDetails, taskID)
	return err
}

// SaveGeneratedCopy 保存生成的文案
func (s *CopywritingTaskService) SaveGeneratedCopy(taskID, generatedCopy, generateModel string) error {
	query := `
		UPDATE copywriting_tasks_tab 
		SET generated_copy = ?, generate_model = ?
		WHERE task_id = ?
	`
	_, err := config.DB.Exec(query, generatedCopy, generateModel, taskID)
	return err
}

// SaveError 保存错误信息
func (s *CopywritingTaskService) SaveError(taskID, errorMessage string) error {
	query := `UPDATE copywriting_tasks_tab SET error_message = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, errorMessage, taskID)
	return err
}

// GetTaskByID 获取任务详情
func (s *CopywritingTaskService) GetTaskByID(taskID string) (*models.CopywritingTaskDetail, error) {
	query := `
		SELECT id, task_id, competitor_urls, 
		       COALESCE(analysis_result, ''), COALESCE(analyze_model, ''),
		       COALESCE(user_selected_data, ''), COALESCE(product_details, ''), 
		       COALESCE(generated_copy, ''), COALESCE(generate_model, ''),
		       COALESCE(error_message, ''), created_at, updated_at
		FROM copywriting_tasks_tab
		WHERE task_id = ?
	`
	
	var task models.CopywritingTaskDetail
	err := config.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.TaskID, &task.CompetitorURLs, &task.AnalysisResult,
		&task.AnalyzeModel, &task.UserSelectedData, &task.ProductDetails,
		&task.GeneratedCopy, &task.GenerateModel, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &task, nil
}
