package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type TaskHistoryService struct{}

func NewTaskHistoryService() *TaskHistoryService {
	return &TaskHistoryService{}
}

func (s *TaskHistoryService) CreateHistory(history *models.TaskHistory) error {
	query := `SELECT COALESCE(MAX(version), 0) + 1 FROM task_history_tab WHERE task_id = ?`
	err := config.DB.QueryRow(query, history.TaskID).Scan(&history.Version)
	if err != nil {
		history.Version = 1
	}

	insertQuery := `INSERT INTO task_history_tab (task_id, user_id, version, prompt, aspect_ratio, 
                    product_images_urls, style_ref_image_url, generated_image_url, edit_instruction, status, error_message) 
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := config.DB.Exec(insertQuery,
		history.TaskID, history.UserID, history.Version, history.Prompt, history.AspectRatio,
		history.ProductImagesURLs, history.StyleRefImageURL, history.GeneratedImageURL,
		history.EditInstruction, history.Status, history.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("failed to create task history: %w", err)
	}

	id, _ := result.LastInsertId()
	history.ID = id

	return nil
}

func (s *TaskHistoryService) GetTaskHistory(taskID int64, limit int, offset int) ([]models.TaskHistory, int, error) {
	var histories []models.TaskHistory
	var total int

	countQuery := `SELECT COUNT(*) FROM task_history_tab WHERE task_id = ?`
	err := config.DB.QueryRow(countQuery, taskID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count history: %w", err)
	}

	query := `SELECT id, task_id, user_id, version, prompt, aspect_ratio, 
              product_images_urls, style_ref_image_url, generated_image_url, 
              edit_instruction, status, error_message, created_at 
              FROM task_history_tab WHERE task_id = ? 
              ORDER BY version DESC LIMIT ? OFFSET ?`

	rows, err := config.DB.Query(query, taskID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var history models.TaskHistory
		var productImagesURLs, styleRefImageURL, generatedImageURL, editInstruction, errorMessage sql.NullString

		err := rows.Scan(
			&history.ID, &history.TaskID, &history.UserID, &history.Version,
			&history.Prompt, &history.AspectRatio, &productImagesURLs,
			&styleRefImageURL, &generatedImageURL, &editInstruction,
			&history.Status, &errorMessage, &history.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan history: %w", err)
		}

		if productImagesURLs.Valid {
			history.ProductImagesURLs = productImagesURLs.String
		}
		if styleRefImageURL.Valid {
			history.StyleRefImageURL = styleRefImageURL.String
		}
		if generatedImageURL.Valid {
			history.GeneratedImageURL = generatedImageURL.String
		}
		if editInstruction.Valid {
			history.EditInstruction = editInstruction.String
		}
		if errorMessage.Valid {
			history.ErrorMessage = errorMessage.String
		}

		histories = append(histories, history)
	}

	return histories, total, nil
}

func (s *TaskHistoryService) GetLatestHistory(taskID int64) (*models.TaskHistory, error) {
	query := `SELECT id, task_id, user_id, version, prompt, aspect_ratio, 
              product_images_urls, style_ref_image_url, generated_image_url, 
              edit_instruction, status, error_message, created_at 
              FROM task_history_tab WHERE task_id = ? 
              ORDER BY version DESC LIMIT 1`

	history := &models.TaskHistory{}
	var productImagesURLs, styleRefImageURL, generatedImageURL, editInstruction, errorMessage sql.NullString

	err := config.DB.QueryRow(query, taskID).Scan(
		&history.ID, &history.TaskID, &history.UserID, &history.Version,
		&history.Prompt, &history.AspectRatio, &productImagesURLs,
		&styleRefImageURL, &generatedImageURL, &editInstruction,
		&history.Status, &errorMessage, &history.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no history found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest history: %w", err)
	}

	if productImagesURLs.Valid {
		history.ProductImagesURLs = productImagesURLs.String
	}
	if styleRefImageURL.Valid {
		history.StyleRefImageURL = styleRefImageURL.String
	}
	if generatedImageURL.Valid {
		history.GeneratedImageURL = generatedImageURL.String
	}
	if editInstruction.Valid {
		history.EditInstruction = editInstruction.String
	}
	if errorMessage.Valid {
		history.ErrorMessage = errorMessage.String
	}

	return history, nil
}

func (s *TaskHistoryService) SaveProductImagesToCDN(userID int64, productImages []string, cdnService *CDNService) ([]string, error) {
	var cdnURLs []string

	for _, dataURL := range productImages {
		cdnImage, err := cdnService.UploadImage(userID, dataURL, "product")
		if err != nil {
			return nil, fmt.Errorf("failed to upload product image: %w", err)
		}
		cdnURLs = append(cdnURLs, cdnImage.CDNURL)
	}

	return cdnURLs, nil
}

func (s *TaskHistoryService) SaveGeneratedImageToCDN(userID int64, imageDataURL string, cdnService *CDNService) (string, error) {
	cdnImage, err := cdnService.UploadImage(userID, imageDataURL, "generated")
	if err != nil {
		return "", fmt.Errorf("failed to upload generated image: %w", err)
	}

	return cdnImage.CDNURL, nil
}

func (s *TaskHistoryService) ConvertURLsToJSON(urls []string) string {
	data, _ := json.Marshal(urls)
	return string(data)
}
