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
	// Gemini 2.0 Flash支持原生的多模态图片生成
	// 可以接收文字+图片输入，直接输出图片
	
	aspectHint := "square format (1:1 ratio)"
	if req.AspectRatio == "4:5" {
		aspectHint = "portrait format (4:5 ratio)"
	}

	stylePrompt := ""
	if req.StyleRefImage != "" {
		stylePrompt = " Follow the style, composition, and lighting similar to the reference image provided."
	}

	// 构建prompt
	enhancedPrompt := fmt.Sprintf(`Create a professional e-commerce product image based on the provided product image(s).

Requirements: %s

Style: Professional e-commerce photography with clean background, studio lighting, high quality, photorealistic.
Format: %s
%s

Maintain the product's original features, texture, and details. Ensure the product is clearly visible and appealing for online shopping.`, 
		req.Prompt, aspectHint, stylePrompt)

	// 准备multimodal parts
	parts := []*genai.Part{}
	
	// 添加产品图片
	for i, imageURL := range req.ProductImages {
		dataURL, err := utils.ConvertURLToDataURL(imageURL)
		if err != nil {
			return "", fmt.Errorf("转换产品图片 %d 失败: %w", i+1, err)
		}
		
		part, err := utils.MakeImagePart(dataURL)
		if err != nil {
			return "", fmt.Errorf("创建图片部分 %d 失败: %w", i+1, err)
		}
		parts = append(parts, part)
		
		// 打印图片信息
		urlPreview := imageURL
		if len(urlPreview) > 80 {
			urlPreview = urlPreview[:80] + "..."
		}
		log.Printf("🖼️  Product image[%d]: %s", i, urlPreview)
	}
	
	// 添加风格参考图
	if req.StyleRefImage != "" {
		dataURL, err := utils.ConvertURLToDataURL(req.StyleRefImage)
		if err != nil {
			return "", fmt.Errorf("转换风格参考图失败: %w", err)
		}
		
		part, err := utils.MakeImagePart(dataURL)
		if err != nil {
			return "", fmt.Errorf("创建风格参考部分失败: %w", err)
		}
		parts = append(parts, part)
		log.Printf("🎨 Style reference image added")
	}
	
	// 添加文字prompt
	parts = append(parts, &genai.Part{Text: enhancedPrompt})

	contents := []*genai.Content{{Parts: parts}}
	
	// 使用Gemini 3.1 Flash Image模型（优先）或3 Pro Image Preview作为备选
	modelName := "gemini-3.1-flash-image"
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE"}, // 只要求图片输出
	}

	log.Printf("🖼️  Gemini Native Image Generation Request")
	log.Printf("📝 Model: %s, Product images: %d, Prompt length: %d", 
		modelName, len(req.ProductImages), len(enhancedPrompt))
	
	promptPreview := enhancedPrompt
	if len(promptPreview) > 300 {
		promptPreview = promptPreview[:300] + "..."
	}
	log.Printf("📝 Prompt: %s", promptPreview)

	resp, err := s.geminiClient.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		log.Printf("⚠️  %s failed: %v, trying fallback model", modelName, err)
		
		// 尝试备选模型
		modelName = "gemini-3-pro-image-preview"
		log.Printf("🔄 Retrying with fallback model: %s", modelName)
		
		resp, err = s.geminiClient.Models.GenerateContent(ctx, modelName, contents, config)
		if err != nil {
			log.Printf("❌ Gemini Image Generation Error (both models failed): %v", err)
			return "", fmt.Errorf("Gemini图片生成失败: %w", err)
		}
	}

	// 打印响应信息
	candidatesCount := 0
	if resp.Candidates != nil {
		candidatesCount = len(resp.Candidates)
	}
	log.Printf("🖼️  Gemini Response - Model: %s, Candidates: %d", modelName, candidatesCount)

	// 提取图片
	imageDataURL := utils.ExtractImageFromResponse(resp)
	if imageDataURL == "" {
		log.Printf("❌ No image found in Gemini response")
		return "", fmt.Errorf("Gemini未返回图片。这可能是因为内容不符合安全政策，或模型暂时不可用")
	}

	log.Printf("✅ Gemini native image generation successful - Model: %s, Data URL length: %d", modelName, len(imageDataURL))
	return imageDataURL, nil
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
