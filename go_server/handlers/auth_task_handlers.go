package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"electric_ai_tool/go_server/models"
	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type AuthHandler struct {
	authService  *services.AuthService
	emailService *services.EmailService
}

func NewAuthHandler(authService *services.AuthService, emailService *services.EmailService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.VerificationCode == "" {
		utils.RespondError(w, fmt.Errorf("verification code is required"), http.StatusBadRequest)
		return
	}

	valid, err := h.emailService.VerifyCode(req.Email, req.VerificationCode, "register")
	if err != nil || !valid {
		utils.RespondError(w, fmt.Errorf("invalid or expired verification code"), http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	sessionID, err := h.authService.CreateSession(user.ID)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.AuthResponse{User: *user, SessionID: sessionID})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	user, sessionID, err := h.authService.Login(req)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	utils.RespondJSON(w, models.AuthResponse{User: *user, SessionID: sessionID})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	err := h.authService.ForgotPassword(req)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]string{
		"message": "如果该邮箱存在，重置链接已发送",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	err := h.authService.ResetPassword(req)
	if err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, map[string]string{"message": "密码重置成功"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.Header.Get("Authorization")
	if sessionID != "" && len(sessionID) > 7 {
		sessionID = sessionID[7:]
	}

	if sessionID != "" {
		h.authService.Logout(sessionID)
	}

	utils.RespondJSON(w, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) SendVerificationCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email   string `json:"email"`
		Purpose string `json:"purpose"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Purpose != "register" && req.Purpose != "reset" {
		utils.RespondError(w, fmt.Errorf("invalid purpose"), http.StatusBadRequest)
		return
	}

	code := h.emailService.GenerateCode()
	
	if err := h.emailService.SaveVerificationCode(req.Email, code, req.Purpose); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	if err := h.emailService.SendVerificationCode(req.Email, code, req.Purpose); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]string{"message": "验证码已发送"})
}

func (h *AuthHandler) TestSendVerificationCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	code := h.emailService.GenerateCode()
	
	if err := h.emailService.SendTestEmail(req.Email, code); err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, map[string]interface{}{
		"message": "测试验证码已生成（请查看后端日志）",
		"code":    code,
		"email":   req.Email,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Authorization")
	if sessionID != "" && len(sessionID) > 7 {
		sessionID = sessionID[7:]
	}

	user, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	utils.RespondJSON(w, user)
}

type TaskHandler struct {
	multiModelService  *services.MultiModelService
	taskService        *services.TaskService
	taskHistoryService *services.TaskHistoryService
	cdnService         *services.CDNService
	authService        *services.AuthService
}

func NewTaskHandler(multiModelService *services.MultiModelService, taskService *services.TaskService,
	taskHistoryService *services.TaskHistoryService, cdnService *services.CDNService,
	authService *services.AuthService) *TaskHandler {
	return &TaskHandler{
		multiModelService:  multiModelService,
		taskService:        taskService,
		taskHistoryService: taskHistoryService,
		cdnService:         cdnService,
		authService:        authService,
	}
}

func (h *TaskHandler) getUserID(r *http.Request) (int64, error) {
	sessionID := r.Header.Get("Authorization")
	if sessionID != "" && len(sessionID) > 7 {
		sessionID = sessionID[7:]
	}

	user, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

func (h *TaskHandler) AnalyzeWithTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserID(r)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	var req models.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	task, err := h.taskService.CreateTask(userID, req.SKU, req.Keywords, req.SellingPoints, req.CompetitorLink, req.Model, models.ModelGemini)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusAnalyzing, nil, "")

	ctx := context.Background()
	sellingPoints, err := h.multiModelService.AnalyzeSellingPoints(ctx, req)
	if err != nil {
		h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusAnalyzeFailed, nil, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusAnalyzed, sellingPoints, "")

	utils.RespondJSON(w, map[string]interface{}{
		"data":    sellingPoints,
		"task_id": task.ID,
	})
}

func (h *TaskHandler) GenerateImageWithTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserID(r)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	var req models.GenerateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	task, err := h.taskService.CreateTask(userID, "", "", req.Prompt, "", models.ModelGemini, req.Model)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusGenerating, nil, "")

	productImageURLs, err := h.taskHistoryService.SaveProductImagesToCDN(userID, req.ProductImages, h.cdnService)
	if err != nil {
		h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusGenerateFailed, nil, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	var styleRefURL string
	if req.StyleRefImage != "" {
		cdnImage, err := h.cdnService.UploadImage(userID, req.StyleRefImage, "style_ref")
		if err != nil {
			h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusGenerateFailed, nil, err.Error())
			utils.RespondError(w, err, http.StatusInternalServerError)
			return
		}
		styleRefURL = cdnImage.CDNURL
	}

	ctx := context.Background()
	generatedDataURL, err := h.multiModelService.GenerateImage(ctx, req)
	if err != nil {
		h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusGenerateFailed, nil, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	generatedCDNURL, err := h.taskHistoryService.SaveGeneratedImageToCDN(userID, generatedDataURL, h.cdnService)
	if err != nil {
		h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusGenerateFailed, nil, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	history := &models.TaskHistory{
		TaskID:            task.ID,
		UserID:            userID,
		Model:             req.Model,
		Prompt:            req.Prompt,
		AspectRatio:       req.AspectRatio,
		ProductImagesURLs: h.taskHistoryService.ConvertURLsToJSON(productImageURLs),
		StyleRefImageURL:  styleRefURL,
		GeneratedImageURL: generatedCDNURL,
		Status:            models.TaskHistoryStatusSuccess,
	}

	if err := h.taskHistoryService.CreateHistory(history); err != nil {
		h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusGenerateFailed, nil, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	h.taskService.UpdateTaskStatus(task.ID, models.TaskStatusCompleted, map[string]interface{}{
		"generated_image_url": generatedCDNURL,
	}, "")

	utils.RespondJSON(w, map[string]interface{}{
		"data":    generatedCDNURL,
		"task_id": task.ID,
	})
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserID(r)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	statusStr := r.URL.Query().Get("status")

	limit := 20
	offset := 0
	statusFilter := -1

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			statusFilter = s
		}
	}

	tasks, total, err := h.taskService.GetUserTasks(userID, statusFilter, limit, offset)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.TaskListResponse{Data: tasks, Total: total})
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	tasks, total, err := h.taskService.GetAllTasks(limit, offset)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.TaskListResponse{Data: tasks, Total: total})
}

func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		utils.RespondError(w, fmt.Errorf("task_id is required"), http.StatusBadRequest)
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("invalid task_id"), http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	histories, total, err := h.taskHistoryService.GetTaskHistory(taskID, limit, offset)
	if err != nil {
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	utils.RespondJSON(w, models.TaskHistoryListResponse{Data: histories, Total: total})
}
