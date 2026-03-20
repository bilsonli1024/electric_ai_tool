package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"

	"google.golang.org/genai"
)

type MultiModelService struct {
	geminiClient *genai.Client
}

func NewMultiModelService(geminiClient *genai.Client) *MultiModelService {
	return &MultiModelService{
		geminiClient: geminiClient,
	}
}

func (s *MultiModelService) AnalyzeSellingPoints(ctx context.Context, req models.AnalyzeRequest) ([]models.SellingPoint, error) {
	model := req.Model
	if model == "" {
		model = models.ModelGemini
	}

	switch model {
	case models.ModelGemini:
		return s.analyzeWithGemini(ctx, req)
	case models.ModelGPT:
		return s.analyzeWithGPT(ctx, req)
	case models.ModelDeepSeek:
		return s.analyzeWithDeepSeek(ctx, req)
	default:
		return s.analyzeWithGemini(ctx, req)
	}
}

func (s *MultiModelService) GenerateImage(ctx context.Context, req models.GenerateImageRequest) (string, error) {
	model := req.Model
	if model == "" {
		model = models.ModelGemini
	}

	switch model {
	case models.ModelGemini:
		return s.generateWithGemini(ctx, req)
	case models.ModelGPT:
		return s.generateWithGPT(ctx, req)
	case models.ModelDeepSeek:
		return s.generateWithDeepSeek(ctx, req)
	default:
		return s.generateWithGemini(ctx, req)
	}
}

func (s *MultiModelService) analyzeWithGemini(ctx context.Context, req models.AnalyzeRequest) ([]models.SellingPoint, error) {
	competitorInfo := ""
	if req.CompetitorLink != "" {
		competitorInfo = "竞品参考: " + req.CompetitorLink
	}

	prompt := fmt.Sprintf(`
你是一个资深的亚马逊运营专家。请根据以下信息，提炼出9个最具吸引力的产品卖点（Selling Points）。
SKU: %s
核心关键词: %s
用户提供的卖点: %s
%s

请为每个卖点提供：
1. 英文标题 (title) 和 英文描述 (description) - 用于生成图片。描述中必须包含指令，要求生成模型"严格保留原产品的纹理、材质和细节特征"，确保产品看起来真实且与原图一致。
2. 中文标题 (title_cn) 和 中文描述 (description_cn) - 用于用户在网页上快速浏览。

请以JSON格式返回，包含一个数组，每个元素包含上述四个字段。
`, req.SKU, req.Keywords, req.SellingPoints, competitorInfo)

	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"sellingPoints": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"title":          {Type: genai.TypeString},
						"description":    {Type: genai.TypeString},
						"title_cn":       {Type: genai.TypeString},
						"description_cn": {Type: genai.TypeString},
					},
					Required: []string{"title", "description", "title_cn", "description_cn"},
				},
			},
		},
	}

	parts := []*genai.Part{{Text: prompt}}
	contents := []*genai.Content{{Parts: parts}}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	resp, err := s.geminiClient.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", contents, config)
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

	var result struct {
		SellingPoints []models.SellingPoint `json:"sellingPoints"`
	}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, err
	}

	return result.SellingPoints, nil
}

func (s *MultiModelService) analyzeWithGPT(ctx context.Context, req models.AnalyzeRequest) ([]models.SellingPoint, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not configured")
	}

	competitorInfo := ""
	if req.CompetitorLink != "" {
		competitorInfo = "竞品参考: " + req.CompetitorLink
	}

	prompt := fmt.Sprintf(`
你是一个资深的亚马逊运营专家。请根据以下信息，提炼出9个最具吸引力的产品卖点。
SKU: %s
核心关键词: %s
用户提供的卖点: %s
%s

请为每个卖点提供：英文标题(title)、英文描述(description)、中文标题(title_cn)、中文描述(description_cn)
返回JSON格式：{"sellingPoints": [{"title":"...","description":"...","title_cn":"...","description_cn":"..."}]}
`, req.SKU, req.Keywords, req.SellingPoints, competitorInfo)

	result, err := callOpenAIChat(apiKey, prompt)
	if err != nil {
		return nil, err
	}

	var response struct {
		SellingPoints []models.SellingPoint `json:"sellingPoints"`
	}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, err
	}

	return response.SellingPoints, nil
}

func (s *MultiModelService) analyzeWithDeepSeek(ctx context.Context, req models.AnalyzeRequest) ([]models.SellingPoint, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY not configured")
	}

	apiBase := os.Getenv("DEEPSEEK_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.deepseek.com/v1"
	}

	competitorInfo := ""
	if req.CompetitorLink != "" {
		competitorInfo = "竞品参考: " + req.CompetitorLink
	}

	prompt := fmt.Sprintf(`
你是一个资深的亚马逊运营专家。请根据以下信息，提炼出9个最具吸引力的产品卖点。
SKU: %s
核心关键词: %s
用户提供的卖点: %s
%s

请为每个卖点提供：英文标题(title)、英文描述(description)、中文标题(title_cn)、中文描述(description_cn)
返回JSON格式：{"sellingPoints": [{"title":"...","description":"...","title_cn":"...","description_cn":"..."}]}
`, req.SKU, req.Keywords, req.SellingPoints, competitorInfo)

	result, err := callOpenAICompatibleChat(apiBase, apiKey, prompt)
	if err != nil {
		return nil, err
	}

	var response struct {
		SellingPoints []models.SellingPoint `json:"sellingPoints"`
	}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, err
	}

	return response.SellingPoints, nil
}

func (s *MultiModelService) generateWithGemini(ctx context.Context, req models.GenerateImageRequest) (string, error) {
	aspectHint := "square format (1:1)"
	if req.AspectRatio == "4:5" {
		aspectHint = "portrait format (4:5)"
	}

	stylePrompt := ""
	if req.StyleRefImage != "" {
		stylePrompt = " Follow the style, composition, and lighting of the provided style reference image."
	}

	enhancedPrompt := fmt.Sprintf("%s.%s Generate in %s. Photorealistic, high quality, maintain the original product's texture, material, and fine details exactly as in the product images. Ensure the product appears consistent with all provided reference images. No distortion of product features.",
		req.Prompt, stylePrompt, aspectHint)

	parts := []*genai.Part{}
	for _, dataURL := range req.ProductImages {
		part, err := utils.MakeImagePart(dataURL)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}

	if req.StyleRefImage != "" {
		part, err := utils.MakeImagePart(req.StyleRefImage)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}

	parts = append(parts, &genai.Part{Text: enhancedPrompt})

	contents := []*genai.Content{{Parts: parts}}
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	resp, err := s.geminiClient.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", contents, config)
	if err != nil {
		return "", err
	}

	return utils.ExtractImageFromResponse(resp), nil
}

func (s *MultiModelService) generateWithGPT(ctx context.Context, req models.GenerateImageRequest) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not configured")
	}

	return "", fmt.Errorf("GPT image generation not yet implemented - use DALL-E API")
}

func (s *MultiModelService) generateWithDeepSeek(ctx context.Context, req models.GenerateImageRequest) (string, error) {
	return "", fmt.Errorf("DeepSeek does not support image generation")
}

func callOpenAIChat(apiKey, prompt string) (string, error) {
	apiBase := os.Getenv("OPENAI_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}

	return callOpenAICompatibleChat(apiBase, apiKey, prompt)
}

func callOpenAICompatibleChat(apiBase, apiKey, prompt string) (string, error) {
	url := strings.TrimSuffix(apiBase, "/") + "/chat/completions"

	payload := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}

	payloadBytes, _ := json.Marshal(payload)

	result, err := utils.HTTPPost(url, apiKey, payloadBytes)
	if err != nil {
		return "", err
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(result, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Choices[0].Message.Content, nil
}
