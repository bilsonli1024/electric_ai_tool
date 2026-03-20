package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"

	"google.golang.org/genai"
)

type CopywritingService struct {
	multiModelService *MultiModelService
}

func NewCopywritingService(multiModelService *MultiModelService) *CopywritingService {
	return &CopywritingService{
		multiModelService: multiModelService,
	}
}

func (s *CopywritingService) AnalyzeCompetitors(ctx context.Context, urls []string, model string) (*models.CompetitorAnalysis, error) {
	if model == "" {
		model = models.ModelGemini
	}

	urlsStr := strings.Join(urls, ", ")
	
	prompt := fmt.Sprintf(`
Analyze the following Amazon competitor product pages. 
Tasks:
1. Extract at least 50 high-traffic keywords based on Amazon search logic. Categorize them into:
   - "core": Main product name and high-volume search terms.
   - "attribute": Words describing features, materials, sizes, colors.
   - "extension": Long-tail keywords, usage scenarios, and related search terms.
2. Summarize the main product selling points (benefits and features) from listing copy.
3. Analyze customer reviews to identify what consumers care about most (pain points, desired features, common praise).
4. Analyze visual elements (implied from listing context) to identify key scenes and selling points shown in images.
5. For each item, provide a precise Chinese translation.

URLs: %s

Return JSON with this structure:
{
  "keywords": [{"original": "keyword", "translation": "关键词", "category": "core|attribute|extension"}, ...],
  "sellingPoints": [{"original": "selling point", "translation": "卖点翻译"}, ...],
  "reviewInsights": [{"original": "insight", "translation": "洞察翻译"}, ...],
  "imageInsights": [{"original": "insight", "translation": "图片卖点翻译"}, ...]
}
`, urlsStr)

	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"keywords": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"original":    {Type: genai.TypeString},
						"translation": {Type: genai.TypeString},
						"category":    {Type: genai.TypeString},
					},
					Required: []string{"original", "translation", "category"},
				},
			},
			"sellingPoints": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"original":    {Type: genai.TypeString},
						"translation": {Type: genai.TypeString},
					},
					Required: []string{"original", "translation"},
				},
			},
			"reviewInsights": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"original":    {Type: genai.TypeString},
						"translation": {Type: genai.TypeString},
					},
					Required: []string{"original", "translation"},
				},
			},
			"imageInsights": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"original":    {Type: genai.TypeString},
						"translation": {Type: genai.TypeString},
					},
					Required: []string{"original", "translation"},
				},
			},
		},
		Required: []string{"keywords", "sellingPoints", "reviewInsights", "imageInsights"},
	}

	parts := []*genai.Part{{Text: prompt}}
	contents := []*genai.Content{{Parts: parts}}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	resp, err := s.multiModelService.geminiClient.Models.GenerateContent(ctx, "gemini-1.5-flash", contents, config)
	if err != nil {
		return nil, err
	}

	var responseText string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if textPart := resp.Candidates[0].Content.Parts[0].Text; textPart != "" {
			responseText = textPart
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("empty response from AI")
	}

	var result models.CompetitorAnalysis
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *CopywritingService) GenerateCopy(ctx context.Context, req models.GenerateCopyRequest) (*models.GeneratedCopy, error) {
	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	allKeywords := append(req.SelectedKeywords, req.SelectedReviewInsights...)
	allKeywords = append(allKeywords, req.SelectedImageInsights...)

	prompt := fmt.Sprintf(`
You are an expert Amazon copywriter. Based on the following information, generate a high-converting Amazon product listing.

Selected Competitor Keywords: %s
Selected Competitor Selling Points: %s

My Product Details:
- Size: %s
- Color: %s
- Quantity: %s
- Function: %s
- Usage Scenario: %s
- Target Audience: %s
- Material: %s
- Main Selling Points: %s
- Target Keywords: %s

Requirements:
1. Title: Catchy, keyword-rich (200 characters max), and follows Amazon best practices. Prioritize the most important keywords at the beginning.
2. 5 Bullet Points: Each point should start with a bolded summary. Highlight benefits, use selected keywords naturally, and address customer pain points.
3. Description: Detailed, persuasive, and formatted for readability using HTML-like tags (e.g., <p>, <b>, <br>) or Markdown. Focus on storytelling and emotional connection.
4. Search Terms (ST): A list of relevant keywords for backend search, optimized for maximum traffic. Do not repeat keywords from the title or bullets. Limit to 249 bytes.

Return JSON:
{
  "title": "Product Title",
  "bulletPoints": ["Point 1", "Point 2", "Point 3", "Point 4", "Point 5"],
  "description": "Product Description",
  "searchTerms": "keyword1 keyword2 keyword3 ..."
}
`, strings.Join(allKeywords, ", "), strings.Join(req.SelectedSellingPoints, ", "),
		req.ProductDetails.Size, req.ProductDetails.Color, req.ProductDetails.Quantity,
		req.ProductDetails.Function, req.ProductDetails.Scenario, req.ProductDetails.Audience,
		req.ProductDetails.Material, req.ProductDetails.SellingPoints, req.ProductDetails.Keywords)

	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title": {Type: genai.TypeString},
			"bulletPoints": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
			"description": {Type: genai.TypeString},
			"searchTerms": {Type: genai.TypeString},
		},
		Required: []string{"title", "bulletPoints", "description", "searchTerms"},
	}

	parts := []*genai.Part{{Text: prompt}}
	contents := []*genai.Content{{Parts: parts}}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	resp, err := s.multiModelService.geminiClient.Models.GenerateContent(ctx, "gemini-1.5-flash", contents, config)
	if err != nil {
		return nil, err
	}

	var responseText string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if textPart := resp.Candidates[0].Content.Parts[0].Text; textPart != "" {
			responseText = textPart
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("empty response from AI")
	}

	var result models.GeneratedCopy
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *CopywritingService) CreateTask(userID int64, competitorURLs []string, model string, taskName string) (int64, error) {
	db := config.GetDB()
	if db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	urlsJSON, _ := json.Marshal(competitorURLs)

	if taskName == "" {
		taskName = fmt.Sprintf("文案任务_%d", time.Now().Unix())
	}

	result, err := db.Exec(
		`INSERT INTO copywriting_tasks_tab (user_id, task_name, competitor_urls, status, analyze_model, generate_model) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		userID, taskName, string(urlsJSON), models.CopyStatusAnalyzing, model, model,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *CopywritingService) UpdateTaskStatus(taskID int64, status int, errorMsg string) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(
		`UPDATE copywriting_tasks_tab SET status = ?, error_message = ? WHERE id = ?`,
		status, errorMsg, taskID,
	)
	return err
}

func (s *CopywritingService) SaveAnalysisResult(taskID int64, analysis *models.CompetitorAnalysis) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	analysisJSON, _ := json.Marshal(analysis)

	_, err := db.Exec(
		`UPDATE copywriting_tasks_tab SET analysis_result = ?, status = ? WHERE id = ?`,
		string(analysisJSON), models.CopyStatusAnalyzed, taskID,
	)
	return err
}

func (s *CopywritingService) SaveGeneratedCopy(taskID int64, copy *models.GeneratedCopy, productDetails *models.ProductDetails) error {
	db := config.GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	copyJSON, _ := json.Marshal(copy)
	detailsJSON, _ := json.Marshal(productDetails)

	_, err := db.Exec(
		`UPDATE copywriting_tasks_tab SET generated_copy = ?, product_details = ?, status = ? WHERE id = ?`,
		string(copyJSON), string(detailsJSON), models.CopyStatusCompleted, taskID,
	)
	return err
}

func (s *CopywritingService) GetTaskByID(taskID int64) (*models.CopywritingTask, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	task := &models.CopywritingTask{}
	err := db.QueryRow(
		`SELECT id, user_id, task_name, competitor_urls, analysis_result, product_details, generated_copy, 
		 status, analyze_model, generate_model, error_message, created_at, updated_at 
		 FROM copywriting_tasks_tab WHERE id = ?`,
		taskID,
	).Scan(
		&task.ID, &task.UserID, &task.TaskName, &task.CompetitorURLs, &task.AnalysisResult,
		&task.ProductDetails, &task.GeneratedCopy, &task.Status,
		&task.AnalyzeModel, &task.GenerateModel, &task.ErrorMessage,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *CopywritingService) GetUserTasks(userID int64, limit, offset int) ([]*models.CopywritingTask, int, error) {
	db := config.GetDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database not initialized")
	}

	var total int
	err := db.QueryRow("SELECT COUNT(*) FROM copywriting_tasks_tab WHERE user_id = ?", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := db.Query(
		`SELECT id, user_id, task_name, competitor_urls, analysis_result, product_details, generated_copy, 
		 status, analyze_model, generate_model, error_message, created_at, updated_at 
		 FROM copywriting_tasks_tab WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []*models.CopywritingTask{}
	for rows.Next() {
		task := &models.CopywritingTask{}
		err := rows.Scan(
			&task.ID, &task.UserID, &task.TaskName, &task.CompetitorURLs, &task.AnalysisResult,
			&task.ProductDetails, &task.GeneratedCopy, &task.Status,
			&task.AnalyzeModel, &task.GenerateModel, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (s *CopywritingService) SearchCompletedTasks(userID int64, keyword string, limit int) ([]*models.CopywritingTask, error) {
	db := config.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `SELECT id, user_id, task_name, competitor_urls, analysis_result, product_details, generated_copy, 
		 status, analyze_model, generate_model, error_message, created_at, updated_at 
		 FROM copywriting_tasks_tab 
		 WHERE user_id = ? AND status = ? AND (task_name LIKE ? OR generated_copy LIKE ?) 
		 ORDER BY created_at DESC LIMIT ?`

	searchPattern := "%" + keyword + "%"
	rows, err := db.Query(query, userID, models.CopyStatusCompleted, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*models.CopywritingTask{}
	for rows.Next() {
		task := &models.CopywritingTask{}
		err := rows.Scan(
			&task.ID, &task.UserID, &task.TaskName, &task.CompetitorURLs, &task.AnalysisResult,
			&task.ProductDetails, &task.GeneratedCopy, &task.Status,
			&task.AnalyzeModel, &task.GenerateModel, &task.ErrorMessage,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
