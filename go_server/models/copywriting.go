package models

type Keyword struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
	Category    string `json:"category"` // core, attribute, extension
}

type BilingualText struct {
	Original    string `json:"original"`
	Translation string `json:"translation"`
}

type CompetitorAnalysis struct {
	Keywords       []Keyword       `json:"keywords"`
	SellingPoints  []BilingualText `json:"sellingPoints"`
	ReviewInsights []BilingualText `json:"reviewInsights"`
	ImageInsights  []BilingualText `json:"imageInsights"`
}

type ProductDetails struct {
	Size          string `json:"size"`
	Color         string `json:"color"`
	Quantity      string `json:"quantity"`
	Function      string `json:"function"`
	Scenario      string `json:"scenario"`
	Audience      string `json:"audience"`
	Material      string `json:"material"`
	SellingPoints string `json:"sellingPoints"`
	Keywords      string `json:"keywords"`
}

type GeneratedCopy struct {
	Title        string   `json:"title"`
	BulletPoints []string `json:"bulletPoints"`
	Description  string   `json:"description"`
	SearchTerms  string   `json:"searchTerms"`
}

type CopywritingTask struct {
	ID             int64  `json:"id"`
	UserID         int64  `json:"user_id"`
	TaskName       string `json:"task_name"`
	CompetitorURLs string `json:"competitor_urls"`
	AnalysisResult string `json:"analysis_result"`
	ProductDetails string `json:"product_details"`
	GeneratedCopy  string `json:"generated_copy"`
	Status         int    `json:"status"`
	AnalyzeModel   string `json:"analyze_model"`
	GenerateModel  string `json:"generate_model"`
	ErrorMessage   string `json:"error_message"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type AnalyzeCompetitorsRequest struct {
	URLs     []string `json:"urls"`
	Model    string   `json:"model"`
	TaskName string   `json:"task_name"`
}

type GenerateCopyRequest struct {
	SelectedKeywords       []string       `json:"selectedKeywords"`
	SelectedSellingPoints  []string       `json:"selectedSellingPoints"`
	SelectedReviewInsights []string       `json:"selectedReviewInsights"`
	SelectedImageInsights  []string       `json:"selectedImageInsights"`
	ProductDetails         ProductDetails `json:"productDetails"`
	Model                  string         `json:"model"`
}

const (
	CopyStatusAnalyzing       = 0
	CopyStatusAnalyzed        = 1
	CopyStatusGenerating      = 2
	CopyStatusCompleted       = 3
	CopyStatusAnalyzeFailed   = 10
	CopyStatusGenerateFailed  = 11
)
