package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

// TestChat 简单的聊天测试接口，用于模型连通性测试
func (s *MultiModelService) TestChat(ctx context.Context, model string, prompt string) (string, error) {
	// 默认使用Gemini
	return s.testChatWithGemini(ctx, prompt)
}

func (s *MultiModelService) AnalyzeSellingPoints(ctx context.Context, req models.AnalyzeRequest) ([]models.SellingPoint, error) {
	// 默认使用Gemini
	return s.analyzeWithGemini(ctx, req)
}

func (s *MultiModelService) GenerateImage(ctx context.Context, req models.GenerateImageRequest) (string, error) {
	model := req.Model
	if model == 0 {
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

	log.Printf("🤖 Gemini API Request - Model: gemini-3.1-flash-lite-preview, Prompt length: %d chars", len(prompt))
	
	resp, err := s.geminiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite-preview", contents, config)
	if err != nil {
		log.Printf("❌ Gemini API Error: %v", err)
		return nil, err
	}

	log.Printf("✅ Gemini API Response received - Candidates: %d", len(resp.Candidates))
	
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

	result, err := callOpenAICompatibleChat(apiBase, apiKey, prompt, "deepseek-chat")
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
	// 检查是否有产品图片
	if len(req.ProductImages) == 0 {
		return "", fmt.Errorf("Gemini图片生成需要至少一张产品图片作为参考。请上传产品图片后重试，或选择其他模型（如GPT或DeepSeek）")
	}
	
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
	for i, imageURL := range req.ProductImages {
		// 转换URL为data URL
		dataURL, err := utils.ConvertURLToDataURL(imageURL)
		if err != nil {
			return "", fmt.Errorf("failed to convert product image %d: %w", i+1, err)
		}
		
		part, err := utils.MakeImagePart(dataURL)
		if err != nil {
			return "", fmt.Errorf("failed to create image part %d: %w", i+1, err)
		}
		parts = append(parts, part)
	}

	if req.StyleRefImage != "" {
		// 转换风格参考图
		dataURL, err := utils.ConvertURLToDataURL(req.StyleRefImage)
		if err != nil {
			return "", fmt.Errorf("failed to convert style reference image: %w", err)
		}
		
		part, err := utils.MakeImagePart(dataURL)
		if err != nil {
			return "", fmt.Errorf("failed to create style reference part: %w", err)
		}
		parts = append(parts, part)
	}

	parts = append(parts, &genai.Part{Text: enhancedPrompt})

	contents := []*genai.Content{{Parts: parts}}
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	log.Printf("🖼️  Gemini Image API Request - Model: gemini-3.1-flash-lite-preview, Prompt length: %d chars, Images: %d", 
		len(enhancedPrompt), len(req.ProductImages))
	
	resp, err := s.geminiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite-preview", contents, config)
	if err != nil {
		log.Printf("❌ Gemini Image API Error: %v", err)
		return "", err
	}

	imageURL := utils.ExtractImageFromResponse(resp)
	if imageURL != "" {
		log.Printf("✅ Gemini Image API Response - Image generated successfully, data URL length: %d", len(imageURL))
	} else {
		log.Printf("⚠️  Gemini Image API Response - No image in response")
		return "", fmt.Errorf("Gemini未返回图片。可能需要提供产品图片，或尝试使用其他模型")
	}
	
	return imageURL, nil
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

	return callOpenAICompatibleChat(apiBase, apiKey, prompt, "gpt-4")
}

func callOpenAICompatibleChat(apiBase, apiKey, prompt string, modelName string) (string, error) {
	url := strings.TrimSuffix(apiBase, "/") + "/chat/completions"

	if modelName == "" {
		modelName = "gpt-4"
	}

	payload := map[string]interface{}{
		"model": modelName,
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

// 测试聊天方法

func (s *MultiModelService) testChatWithGemini(ctx context.Context, prompt string) (string, error) {
	parts := []*genai.Part{{Text: prompt}}
	contents := []*genai.Content{{Parts: parts}}

	log.Printf("💬 Gemini Chat Test - Prompt: %s", prompt)
	
	resp, err := s.geminiClient.Models.GenerateContent(ctx, "gemini-3.1-flash-lite-preview", contents, nil)
	if err != nil {
		log.Printf("❌ Gemini Chat Test Error: %v", err)
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if textPart := resp.Candidates[0].Content.Parts[0].Text; textPart != "" {
			log.Printf("✅ Gemini Chat Test Response: %s", textPart)
			return textPart, nil
		}
	}

	log.Printf("⚠️  Gemini Chat Test - No response text")
	return "", fmt.Errorf("empty response from Gemini")
}

func (s *MultiModelService) testChatWithGPT(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not configured")
	}

	return callOpenAIChat(apiKey, prompt)
}

func (s *MultiModelService) testChatWithDeepSeek(ctx context.Context, prompt string) (string, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("DEEPSEEK_API_KEY not configured")
	}

	apiBase := os.Getenv("DEEPSEEK_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.deepseek.com/v1"
	}

	return callOpenAICompatibleChat(apiBase, apiKey, prompt, "deepseek-chat")
}

