package services

import (
	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"
)

type CopywritingTaskService struct{}

func NewCopywritingTaskService() *CopywritingTaskService {
	return &CopywritingTaskService{}
}

// CreateTask 创建文案生成任务详细记录
func (s *CopywritingTaskService) CreateTask(taskID, taskName, competitorURLs string, analyzeModel int) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `
		INSERT INTO copywriting_tasks_tab (
			task_id, task_name, competitor_urls, analyze_model, 
			detail_status, ctime, mtime
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := config.DB.Exec(query, taskID, taskName, competitorURLs, analyzeModel, 
		models.CopywritingStatusPending, currentTime, currentTime)
	return err
}

// SaveAnalysisResult 保存分析结果
func (s *CopywritingTaskService) SaveAnalysisResult(taskID, analysisResult string) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `UPDATE copywriting_tasks_tab SET analysis_result = ?, mtime = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, analysisResult, currentTime, taskID)
	return err
}

// SaveUserSelectedData 保存用户选择的数据
func (s *CopywritingTaskService) SaveUserSelectedData(taskID, userSelectedData, productDetails string) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `
		UPDATE copywriting_tasks_tab 
		SET user_selected_data = ?, product_details = ?, mtime = ?
		WHERE task_id = ?
	`
	_, err := config.DB.Exec(query, userSelectedData, productDetails, currentTime, taskID)
	return err
}

// SaveGeneratedCopy 保存生成的文案
func (s *CopywritingTaskService) SaveGeneratedCopy(taskID, generatedCopy string, generateModel int) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `
		UPDATE copywriting_tasks_tab 
		SET generated_copy = ?, generate_model = ?, mtime = ?
		WHERE task_id = ?
	`
	_, err := config.DB.Exec(query, generatedCopy, generateModel, currentTime, taskID)
	return err
}

// SaveError 保存错误信息
func (s *CopywritingTaskService) SaveError(taskID, errorMessage string) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `UPDATE copywriting_tasks_tab SET error_message = ?, fail_msg = ?, mtime = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, errorMessage, errorMessage, currentTime, taskID)
	return err
}

// UpdateDetailStatus 更新详细状态
func (s *CopywritingTaskService) UpdateDetailStatus(taskID string, detailStatus int) error {
	currentTime := utils.GetCurrentTimestamp()
	query := `UPDATE copywriting_tasks_tab SET detail_status = ?, mtime = ? WHERE task_id = ?`
	_, err := config.DB.Exec(query, detailStatus, currentTime, taskID)
	return err
}

// GetTaskByID 获取任务详情
func (s *CopywritingTaskService) GetTaskByID(taskID string) (*models.CopywritingTaskDetail, error) {
	query := `
		SELECT id, task_id, COALESCE(task_name, ''), COALESCE(detail_status, 0),
		       competitor_urls, 
		       COALESCE(analysis_result, ''), COALESCE(analyze_model, 0),
		       COALESCE(user_selected_data, ''), COALESCE(product_details, ''), 
		       COALESCE(generated_copy, ''), COALESCE(generate_model, 0),
		       COALESCE(error_message, ''), COALESCE(fail_msg, ''), 
		       ctime, mtime
		FROM copywriting_tasks_tab
		WHERE task_id = ?
	`
	
	var task models.CopywritingTaskDetail
	err := config.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.TaskID, &task.TaskName, &task.DetailStatus,
		&task.CompetitorURLs, &task.AnalysisResult,
		&task.AnalyzeModel, &task.UserSelectedData, &task.ProductDetails,
		&task.GeneratedCopy, &task.GenerateModel, &task.ErrorMessage, &task.FailMsg,
		&task.Ctime, &task.Mtime,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &task, nil
}
