package handlers

import (
	"net/http"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/utils"
)

type EnumHandler struct{}

func NewEnumHandler() *EnumHandler {
	return &EnumHandler{}
}

// EnumItem 枚举项
type EnumItem struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

// EnumsResponse 枚举响应
type EnumsResponse struct {
	UserTypes              []EnumItem `json:"user_types"`
	UserStatuses           []EnumItem `json:"user_statuses"`
	TaskTypes              []EnumItem `json:"task_types"`
	TaskStatuses           []EnumItem `json:"task_statuses"`
	CopywritingStatuses    []EnumItem `json:"copywriting_statuses"`
	ImageStatuses          []EnumItem `json:"image_statuses"`
	Models                 []EnumItem `json:"models"`
	PermissionTypes        []EnumItem `json:"permission_types"`
	RoleStatuses           []EnumItem `json:"role_statuses"`
	PermissionStatuses     []EnumItem `json:"permission_statuses"`
}

// GetAllEnums 获取所有枚举值
func (h *EnumHandler) GetAllEnums(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := EnumsResponse{
		UserTypes: []EnumItem{
			{Value: models.UserTypeNormal, Label: models.UserTypeToString(models.UserTypeNormal)},
			{Value: models.UserTypeAdmin, Label: models.UserTypeToString(models.UserTypeAdmin)},
		},
		UserStatuses: []EnumItem{
			{Value: models.UserStatusPendingApproval, Label: models.UserStatusToString(models.UserStatusPendingApproval)},
			{Value: models.UserStatusNormal, Label: models.UserStatusToString(models.UserStatusNormal)},
			{Value: models.UserStatusDeleted, Label: models.UserStatusToString(models.UserStatusDeleted)},
		},
		TaskTypes: []EnumItem{
			{Value: models.TaskTypeCopywriting, Label: models.TaskTypeToString(models.TaskTypeCopywriting)},
			{Value: models.TaskTypeImage, Label: models.TaskTypeToString(models.TaskTypeImage)},
		},
		TaskStatuses: []EnumItem{
			{Value: models.TaskStatusPending, Label: models.TaskStatusToString(models.TaskStatusPending)},
			{Value: models.TaskStatusOngoing, Label: models.TaskStatusToString(models.TaskStatusOngoing)},
			{Value: models.TaskStatusCompleted, Label: models.TaskStatusToString(models.TaskStatusCompleted)},
			{Value: models.TaskStatusFailed, Label: models.TaskStatusToString(models.TaskStatusFailed)},
		},
		CopywritingStatuses: []EnumItem{
			{Value: models.CopywritingStatusPending, Label: models.CopywritingStatusToString(models.CopywritingStatusPending)},
			{Value: models.CopywritingStatusAnalyzing, Label: models.CopywritingStatusToString(models.CopywritingStatusAnalyzing)},
			{Value: models.CopywritingStatusAnalyzed, Label: models.CopywritingStatusToString(models.CopywritingStatusAnalyzed)},
			{Value: models.CopywritingStatusGenerating, Label: models.CopywritingStatusToString(models.CopywritingStatusGenerating)},
			{Value: models.CopywritingStatusCompleted, Label: models.CopywritingStatusToString(models.CopywritingStatusCompleted)},
			{Value: models.CopywritingStatusFailed, Label: models.CopywritingStatusToString(models.CopywritingStatusFailed)},
		},
		ImageStatuses: []EnumItem{
			{Value: models.ImageStatusPending, Label: models.ImageStatusToString(models.ImageStatusPending)},
			{Value: models.ImageStatusGenerating, Label: models.ImageStatusToString(models.ImageStatusGenerating)},
			{Value: models.ImageStatusCompleted, Label: models.ImageStatusToString(models.ImageStatusCompleted)},
			{Value: models.ImageStatusFailed, Label: models.ImageStatusToString(models.ImageStatusFailed)},
		},
		Models: []EnumItem{
			{Value: models.ModelGemini, Label: models.ModelToString(models.ModelGemini)},
			{Value: models.ModelGPT, Label: models.ModelToString(models.ModelGPT)},
			{Value: models.ModelDeepSeek, Label: models.ModelToString(models.ModelDeepSeek)},
		},
		PermissionTypes: []EnumItem{
			{Value: models.PermissionTypeMenu, Label: models.PermissionTypeToString(models.PermissionTypeMenu)},
			{Value: models.PermissionTypeButton, Label: models.PermissionTypeToString(models.PermissionTypeButton)},
			{Value: models.PermissionTypeAPI, Label: models.PermissionTypeToString(models.PermissionTypeAPI)},
		},
		RoleStatuses: []EnumItem{
			{Value: models.RoleStatusDisabled, Label: models.RoleStatusToString(models.RoleStatusDisabled)},
			{Value: models.RoleStatusEnabled, Label: models.RoleStatusToString(models.RoleStatusEnabled)},
		},
		PermissionStatuses: []EnumItem{
			{Value: models.PermissionStatusDisabled, Label: models.PermissionStatusToString(models.PermissionStatusDisabled)},
			{Value: models.PermissionStatusEnabled, Label: models.PermissionStatusToString(models.PermissionStatusEnabled)},
		},
	}

	utils.RespondJSON(w, response)
}
