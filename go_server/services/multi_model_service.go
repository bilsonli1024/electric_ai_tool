package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	// Gemini图片生成使用Imagen模型，通过GenerateImages API
	// 注意：Imagen不支持使用产品图片作为输入，只能基于文本prompt生成
	// 如果需要基于产品图片生成，应该使用其他模型
	
	log.Printf("ℹ️  Gemini/Imagen: Note that Imagen generates images from text prompts only. Product images in request will be ignored.")
	
	aspectRatio := "1:1"
	if req.AspectRatio == "4:5" {
		aspectRatio = "3:4"  // Imagen支持的最接近比例
	}

	stylePrompt := ""
	if req.StyleRefImage != "" {
		stylePrompt = " Use a style similar to: professional e-commerce product photography with clean background."
	}

	// Imagen只支持文本prompt，不能输入图片
	enhancedPrompt := fmt.Sprintf("%s.%s Generate in photorealistic style, high quality, professional e-commerce product image. Clean background, studio lighting, show product details clearly.",
		req.Prompt, stylePrompt)

	// 使用Imagen 4 Fast模型（最快的Imagen版本）
	modelName := "imagen-4.0-fast-generate-001"
	config := &genai.GenerateImagesConfig{
		NumberOfImages: 1,  // 生成1张图片
		AspectRatio:    aspectRatio,
	}

	log.Printf("🖼️  Imagen API Request - Model: %s, Prompt length: %d chars, AspectRatio: %s", 
		modelName, len(enhancedPrompt), aspectRatio)
	
	// 打印prompt预览
	promptPreview := enhancedPrompt
	if len(promptPreview) > 200 {
		promptPreview = promptPreview[:200] + "..."
	}
	log.Printf("🖼️  Prompt preview: %s", promptPreview)
	
	if len(req.ProductImages) > 0 {
		log.Printf("⚠️  Warning: Imagen does not support product image inputs. %d product image(s) will be ignored. Consider using GPT or DeepSeek for image-to-image generation.", len(req.ProductImages))
	}
	
	// 使用GenerateImages API
	resp, err := s.geminiClient.Models.GenerateImages(ctx, modelName, enhancedPrompt, config)
	if err != nil {
		log.Printf("❌ Imagen API Error: %v", err)
		return "", fmt.Errorf("Imagen图片生成失败: %w。Imagen只支持文本生成图片，不支持基于产品图片的变换。建议使用GPT或DeepSeek模型", err)
	}
	
	// 打印响应基本信息
	imagesCount := 0
	if resp.GeneratedImages != nil {
		imagesCount = len(resp.GeneratedImages)
	}
	log.Printf("🖼️  Imagen API Response received - Generated images: %d", imagesCount)

	// 检查是否生成了图片
	if imagesCount == 0 || resp.GeneratedImages[0].Image == nil {
		log.Printf("❌ Imagen returned no images")
		return "", fmt.Errorf("Imagen未返回图片。这可能是因为prompt不符合内容政策，或API配置问题")
	}

	// 获取第一张图片
	imageBytes := resp.GeneratedImages[0].Image.ImageBytes
	if len(imageBytes) == 0 {
		log.Printf("❌ Imagen returned empty image data")
		return "", fmt.Errorf("Imagen返回的图片数据为空")
	}

	log.Printf("✅ Imagen generated image successfully - Size: %d bytes", len(imageBytes))

	// 确保目录存在（使用相对路径）
	uploadDir := "./uploads/images"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("❌ Failed to create upload directory: %v", err)
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// 保存图片到本地文件
	timestamp := time.Now().Format("20060102_150405")
	randomSuffix := make([]byte, 4)
	rand.Read(randomSuffix)
	filename := fmt.Sprintf("imagen_%s_%x.png", timestamp, randomSuffix)
	filepath := uploadDir + "/" + filename

	if err := os.WriteFile(filepath, imageBytes, 0644); err != nil {
		log.Printf("❌ Failed to save Imagen image: %v", err)
		return "", fmt.Errorf("failed to save generated image: %w", err)
	}

	log.Printf("✅ Imagen image saved to: %s", filepath)

	// 返回相对路径
	return "/uploads/images/" + filename, nil
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

