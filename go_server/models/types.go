package models

type AnalyzeRequest struct {
	Keywords       string `json:"keywords"`
	SellingPoints  string `json:"sellingPoints"`
	CompetitorLink string `json:"competitorLink"`
	SKU            string `json:"sku"`
	Model          string `json:"model,omitempty"`
}

type SellingPoint struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	TitleCN       string `json:"title_cn"`
	DescriptionCN string `json:"description_cn"`
}

type AnalyzeResponse struct {
	Data []SellingPoint `json:"data"`
}

type GenerateImageRequest struct {
	Prompt        string   `json:"prompt"`
	AspectRatio   string   `json:"aspectRatio"`
	ProductImages []string `json:"productImages"`
	StyleRefImage string   `json:"styleRefImage,omitempty"`
	Model         string   `json:"model,omitempty"`
}

type EditImageRequest struct {
	BaseImage   string `json:"baseImage"`
	Instruction string `json:"instruction"`
	AspectRatio string `json:"aspectRatio"`
}

type APlusContentRequest struct {
	Keywords      string   `json:"keywords"`
	SellingPoints []string `json:"sellingPoints"`
	SKU           string   `json:"sku,omitempty"`
	Template      string   `json:"template,omitempty"`
	RefImages     []string `json:"refImages,omitempty"`
}

type APlusModule struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImagePrompt string `json:"imagePrompt"`
}

type APlusContentResponse struct {
	Data []APlusModule `json:"data"`
}

type ImageResponse struct {
	Data string `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Port   string `json:"port"`
}
