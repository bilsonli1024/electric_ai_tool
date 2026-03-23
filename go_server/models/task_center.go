package models

// TaskCenterBase 任务中心统一底表
type TaskCenterBase struct {
	ID         int64  `json:"id"`
	TaskID     string `json:"task_id"`      // 全局唯一任务ID
	TaskType   int    `json:"task_type"`    // 1=文案生成, 2=图片生成
	TaskStatus int    `json:"task_status"`  // 0=待处理, 1=进行中, 2=已完成, 3=失败
	Operator   string `json:"operator"`     // 操作者邮箱
	Ctime      int64  `json:"ctime"`        // 创建时间(UNIX时间戳)
	Mtime      int64  `json:"mtime"`        // 更新时间(UNIX时间戳)
}

// TaskCenterListItem 任务中心列表项（包含详细字段）
type TaskCenterListItem struct {
	ID         int64  `json:"id"`
	TaskID     string `json:"task_id"`
	TaskType   int    `json:"task_type"`
	TaskStatus int    `json:"task_status"`
	Operator   string `json:"operator"`
	Ctime      int64  `json:"ctime"`
	Mtime      int64  `json:"mtime"`
	TaskName   string `json:"task_name,omitempty"` // 文案生成任务名称
	SKU        string `json:"sku,omitempty"`       // 图片生成SKU
}

// TaskCenterFilter 任务筛选条件
type TaskCenterFilter struct {
	Operator   string // 按操作者筛选
	TaskType   int    // 按任务类型筛选
	TaskStatus int    // 按任务状态筛选
	StartTime  int64  // 开始时间(UNIX时间戳)
	EndTime    int64  // 结束时间(UNIX时间戳)
	Limit      int    // 分页大小
	Offset     int    // 分页偏移
}

// TaskCenterDetail 任务详情（底表+详细表的联合数据）
type TaskCenterDetail struct {
	TaskCenterBase
	DetailData interface{} `json:"detail_data"` // 详细数据，根据task_type不同而不同
}

// CopywritingTaskDetail 文案生成任务详细数据
type CopywritingTaskDetail struct {
	ID               int64  `json:"id"`
	TaskID           string `json:"task_id"`
	TaskName         string `json:"task_name"`          // 任务名称
	DetailStatus     int    `json:"detail_status"`      // 详细状态：0=待处理, 1=分析中, 2=分析完成, 3=生成中, 4=已完成, 5=失败
	CompetitorURLs   string `json:"competitor_urls"`    // JSON数组
	AnalysisResult   string `json:"analysis_result"`    // AI初始分析结果
	AnalyzeModel     int    `json:"analyze_model"`      // 分析模型：1=Gemini, 2=GPT, 3=DeepSeek
	UserSelectedData string `json:"user_selected_data"` // 用户选择后的数据
	ProductDetails   string `json:"product_details"`
	GeneratedCopy    string `json:"generated_copy"`
	GenerateModel    int    `json:"generate_model"` // 生成模型：1=Gemini, 2=GPT, 3=DeepSeek
	ErrorMessage     string `json:"error_message"`
	FailMsg          string `json:"fail_msg"` // 失败原因（用户友好）
	Ctime            int64  `json:"ctime"`    // 创建时间(UNIX时间戳)
	Mtime            int64  `json:"mtime"`    // 更新时间(UNIX时间戳)
}

// ImageTaskDetail 图片生成任务详细数据
type ImageTaskDetail struct {
	ID                 int64  `json:"id"`
	TaskID             string `json:"task_id"`
	DetailStatus       int    `json:"detail_status"`       // 详细状态：0=待处理, 1=生成中, 2=已完成, 3=失败
	SKU                string `json:"sku"`
	Keywords           string `json:"keywords"`
	SellingPoints      string `json:"selling_points"`
	CompetitorLink     string `json:"competitor_link"`
	CopywritingTaskID  string `json:"copywriting_task_id"`
	GenerateModel      int    `json:"generate_model"` // 生成模型：1=Gemini, 2=GPT, 3=DeepSeek
	AspectRatio        string `json:"aspect_ratio"`
	ResultData         string `json:"result_data"`
	GeneratedImageURLs string `json:"generated_image_urls"`
	ErrorMessage       string `json:"error_message"`
	FailMsg            string `json:"fail_msg"` // 失败原因（用户友好）
	Ctime              int64  `json:"ctime"`    // 创建时间(UNIX时间戳)
	Mtime              int64  `json:"mtime"`    // 更新时间(UNIX时间戳)
}

