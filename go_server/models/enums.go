package models

// ============================================================================
// 用户相关枚举
// ============================================================================

// 用户类型枚举
const (
	UserTypeNormal = 0  // 普通用户
	UserTypeAdmin  = 99 // 管理员
)

// 用户状态枚举
const (
	UserStatusPendingApproval = 0 // 待审批
	UserStatusNormal          = 1 // 正常
	UserStatusDeleted         = 2 // 已删除
)

// ============================================================================
// 任务相关枚举
// ============================================================================

// 任务类型枚举
const (
	TaskTypeCopywriting = 1 // 文案生成
	TaskTypeImage       = 2 // 图片生成
)

// 任务中心状态枚举
const (
	TaskStatusPending   = 0 // 待处理
	TaskStatusOngoing   = 1 // 进行中
	TaskStatusCompleted = 2 // 已完成
	TaskStatusFailed    = 3 // 失败
)

// 文案生成详细状态枚举
const (
	CopywritingStatusPending    = 0 // 待处理
	CopywritingStatusAnalyzing  = 1 // 分析中
	CopywritingStatusAnalyzed   = 2 // 分析完成
	CopywritingStatusGenerating = 3 // 生成中
	CopywritingStatusCompleted  = 4 // 已完成
	CopywritingStatusFailed     = 5 // 失败
)

// 图片生成详细状态枚举
const (
	ImageStatusPending    = 0 // 待处理
	ImageStatusGenerating = 1 // 生成中
	ImageStatusCompleted  = 2 // 已完成
	ImageStatusFailed     = 3 // 失败
)

// AI模型枚举
const (
	ModelGemini   = 1 // Gemini
	ModelGPT      = 2 // GPT
	ModelDeepSeek = 3 // DeepSeek
)

// ============================================================================
// 权限相关枚举
// ============================================================================

// 权限类型枚举
const (
	PermissionTypeMenu   = 1 // 菜单
	PermissionTypeButton = 2 // 按钮
	PermissionTypeAPI    = 3 // API
)

// 角色状态枚举
const (
	RoleStatusDisabled = 0 // 禁用
	RoleStatusEnabled  = 1 // 启用
)

// 权限状态枚举
const (
	PermissionStatusDisabled = 0 // 禁用
	PermissionStatusEnabled  = 1 // 启用
)

// ============================================================================
// 枚举映射函数
// ============================================================================

// MapDetailStatusToTaskStatus 将详细状态映射到任务中心状态
func MapDetailStatusToTaskStatus(taskType, detailStatus int) int {
	if taskType == TaskTypeCopywriting {
		// 文案生成任务
		switch detailStatus {
		case CopywritingStatusPending:
			return TaskStatusPending
		case CopywritingStatusAnalyzing, CopywritingStatusAnalyzed, CopywritingStatusGenerating:
			return TaskStatusOngoing
		case CopywritingStatusCompleted:
			return TaskStatusCompleted
		case CopywritingStatusFailed:
			return TaskStatusFailed
		default:
			return TaskStatusPending
		}
	} else if taskType == TaskTypeImage {
		// 图片生成任务
		switch detailStatus {
		case ImageStatusPending:
			return TaskStatusPending
		case ImageStatusGenerating:
			return TaskStatusOngoing
		case ImageStatusCompleted:
			return TaskStatusCompleted
		case ImageStatusFailed:
			return TaskStatusFailed
		default:
			return TaskStatusPending
		}
	}
	return TaskStatusPending
}

// ============================================================================
// 枚举转字符串函数（用于前端展示）
// ============================================================================

// 用户类型转字符串
func UserTypeToString(userType int) string {
	switch userType {
	case UserTypeNormal:
		return "普通用户"
	case UserTypeAdmin:
		return "管理员"
	default:
		return "未知"
	}
}

// 用户状态转字符串
func UserStatusToString(status int) string {
	switch status {
	case UserStatusPendingApproval:
		return "待审批"
	case UserStatusNormal:
		return "正常"
	case UserStatusDeleted:
		return "已删除"
	default:
		return "未知"
	}
}

// 任务类型转字符串
func TaskTypeToString(taskType int) string {
	switch taskType {
	case TaskTypeCopywriting:
		return "文案生成"
	case TaskTypeImage:
		return "图片生成"
	default:
		return "未知"
	}
}

// 任务状态转字符串
func TaskStatusToString(status int) string {
	switch status {
	case TaskStatusPending:
		return "待处理"
	case TaskStatusOngoing:
		return "进行中"
	case TaskStatusCompleted:
		return "已完成"
	case TaskStatusFailed:
		return "失败"
	default:
		return "未知"
	}
}

// 文案详细状态转字符串
func CopywritingStatusToString(status int) string {
	switch status {
	case CopywritingStatusPending:
		return "待处理"
	case CopywritingStatusAnalyzing:
		return "分析中"
	case CopywritingStatusAnalyzed:
		return "分析完成"
	case CopywritingStatusGenerating:
		return "生成中"
	case CopywritingStatusCompleted:
		return "已完成"
	case CopywritingStatusFailed:
		return "失败"
	default:
		return "未知"
	}
}

// 图片详细状态转字符串
func ImageStatusToString(status int) string {
	switch status {
	case ImageStatusPending:
		return "待处理"
	case ImageStatusGenerating:
		return "生成中"
	case ImageStatusCompleted:
		return "已完成"
	case ImageStatusFailed:
		return "失败"
	default:
		return "未知"
	}
}

// AI模型转字符串
func ModelToString(model int) string {
	switch model {
	case ModelGemini:
		return "Gemini"
	case ModelGPT:
		return "GPT"
	case ModelDeepSeek:
		return "DeepSeek"
	default:
		return "未知"
	}
}

// 权限类型转字符串
func PermissionTypeToString(permType int) string {
	switch permType {
	case PermissionTypeMenu:
		return "菜单"
	case PermissionTypeButton:
		return "按钮"
	case PermissionTypeAPI:
		return "API"
	default:
		return "未知"
	}
}

// 角色状态转字符串
func RoleStatusToString(status int) string {
	switch status {
	case RoleStatusDisabled:
		return "禁用"
	case RoleStatusEnabled:
		return "启用"
	default:
		return "未知"
	}
}

// 权限状态转字符串
func PermissionStatusToString(status int) string {
	switch status {
	case PermissionStatusDisabled:
		return "禁用"
	case PermissionStatusEnabled:
		return "启用"
	default:
		return "未知"
	}
}
