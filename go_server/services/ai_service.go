package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"

	"google.golang.org/genai"
)

type AIService struct {
	client *genai.Client
}

func NewAIService(client *genai.Client) *AIService {
	return &AIService{client: client}
}

func (s *AIService) AnalyzeSellingPoints(ctx context.Context, req models.AnalyzeRequest) ([]models.SellingPoint, error) {
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

	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
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

func (s *AIService) GenerateImage(ctx context.Context, req models.GenerateImageRequest) (string, error) {
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

	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
	if err != nil {
		return "", err
	}

	return utils.ExtractImageFromResponse(resp), nil
}

func (s *AIService) EditImage(ctx context.Context, req models.EditImageRequest) (string, error) {
	aspectRatio := req.AspectRatio
	if aspectRatio == "" {
		aspectRatio = "1:1"
	}

	aspectHint := "square format (1:1)"
	if aspectRatio == "4:5" {
		aspectHint = "portrait format (4:5)"
	}

	fullInstruction := fmt.Sprintf("%s Output in %s. Maintain photorealistic quality and preserve all product details.",
		req.Instruction, aspectHint)

	basePart, err := utils.MakeImagePart(req.BaseImage)
	if err != nil {
		return "", err
	}

	parts := []*genai.Part{basePart, {Text: fullInstruction}}
	contents := []*genai.Content{{Parts: parts}}
	config := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
	if err != nil {
		return "", err
	}

	return utils.ExtractImageFromResponse(resp), nil
}

func (s *AIService) GenerateAPlusContent(ctx context.Context, req models.APlusContentRequest) ([]models.APlusModule, error) {
	template := req.Template
	if template == "" {
		template = "standard"
	}

	templateDescriptions := map[string]string{
		"standard":   "标准5模块：包含品牌故事、核心功能、场景化展示、细节展示、对比表。",
		"visual":     "视觉导向4模块：包含超大首图、三列功能展示、大图场景、细节放大。",
		"technical":  "技术详尽6模块：包含顶部横幅、爆炸图展示、材质细节、使用指南、安全说明、对比表。",
		"minimalist": "极简3模块：包含干净的顶部图、场景网格、核心参数。",
	}

	parts := []*genai.Part{}
	for _, dataURL := range req.RefImages {
		part, err := utils.MakeImagePart(dataURL)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	refImageHint := ""
	if len(req.RefImages) > 0 {
		refImageHint = "参考图片已提供，请在策划模块和生成图片提示词时，参考这些竞品或参考图的风格、排版和视觉逻辑。"
	}

	prompt := fmt.Sprintf(`
你是一个亚马逊高级A+页面设计专家。请根据以下产品卖点和选定的模板，策划一套符合亚马逊"高级A+ (Premium A+)"要求的页面方案。
SKU: %s
关键词: %s
核心卖点: %s
选定模板: %s (%s)

%s

要求：
1. 展示逻辑清晰：根据模板要求，从品牌心智到核心功能，再到场景体验和细节参数。
2. 文字简洁有力：包含必要的关键词、卖点和属性词，符合亚马逊合规要求。
3. 视觉引导：为每个模块提供高质量的图片生成指令。指令中必须包含"严格保留原产品的纹理、材质和细节特征"。

请以JSON格式返回，包含：
- modules: 数组，每个模块包含 type, title, description, imagePrompt (用于生成该模块图片的提示词)。
`, req.SKU, req.Keywords, strings.Join(req.SellingPoints, ", "), template,
		templateDescriptions[template], refImageHint)

	parts = append(parts, &genai.Part{Text: prompt})

	contents := []*genai.Content{{Parts: parts}}
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"modules": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"type":        {Type: genai.TypeString},
						"title":       {Type: genai.TypeString},
						"description": {Type: genai.TypeString},
						"imagePrompt": {Type: genai.TypeString},
					},
					Required: []string{"type", "title", "description", "imagePrompt"},
				},
			},
		},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
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
		Modules []models.APlusModule `json:"modules"`
	}
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, err
	}

	return result.Modules, nil
}
