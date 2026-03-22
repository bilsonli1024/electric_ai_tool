package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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
	multiModelService   *services.MultiModelService
	taskService         *services.TaskService
	taskHistoryService  *services.TaskHistoryService
	cdnService          *services.CDNService
	authService         *services.AuthService
	unifiedTaskService  *services.UnifiedTaskService
	taskCenterService   *services.TaskCenterService
	imageTaskService    *services.ImageTaskService
	localStorageService *services.LocalStorageService
}

func NewTaskHandler(multiModelService *services.MultiModelService, taskService *services.TaskService,
	taskHistoryService *services.TaskHistoryService, cdnService *services.CDNService,
	authService *services.AuthService, unifiedTaskService *services.UnifiedTaskService,
	taskCenterService *services.TaskCenterService, imageTaskService *services.ImageTaskService,
	localStorageService *services.LocalStorageService) *TaskHandler {
	return &TaskHandler{
		multiModelService:   multiModelService,
		taskService:         taskService,
		taskHistoryService:  taskHistoryService,
		cdnService:          cdnService,
		authService:         authService,
		unifiedTaskService:  unifiedTaskService,
		taskCenterService:   taskCenterService,
		imageTaskService:    imageTaskService,
		localStorageService: localStorageService,
	}
}

func (h *TaskHandler) getUserIDAndUsername(r *http.Request) (int64, string, error) {
	sessionID := r.Header.Get("Authorization")
	if sessionID != "" && len(sessionID) > 7 {
		sessionID = sessionID[7:]
	}

	user, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		return 0, "", err
	}

	return user.ID, user.Email, nil
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

	h.taskService.UpdateTaskStatus(task.ID, models.LegacyTaskStatusAnalyzing, nil, "")

	ctx := context.Background()
	sellingPoints, err := h.multiModelService.AnalyzeSellingPoints(ctx, req)
	if err != nil {
		h.taskService.UpdateTaskStatus(task.ID, models.LegacyTaskStatusAnalyzeFailed, nil, err.Error())
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	h.taskService.UpdateTaskStatus(task.ID, models.LegacyTaskStatusAnalyzed, sellingPoints, "")

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

	_, operator, err := h.getUserIDAndUsername(r)
	if err != nil {
		utils.RespondError(w, err, http.StatusUnauthorized)
		return
	}

	var req struct {
		SKU                  string `json:"sku"`
		Keywords             string `json:"keywords"`
		SellingPoints        string `json:"sellingPoints"`
		CompetitorLink       string `json:"competitorLink"`
		Model                string `json:"model"`
		TaskName             string `json:"taskName"`
		CopywritingTaskID    string `json:"copywritingTaskId"` // 改为string类型的task_id
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		req.Model = models.ModelGemini
	}

	if req.TaskName == "" {
		req.TaskName = fmt.Sprintf("图片生成_%d", time.Now().Unix())
	}

	// 1. 生成任务ID
	taskID := h.taskCenterService.GenerateTaskID(models.TaskTypeImage)
	
	// 2. 创建任务中心底表记录
	if err := h.taskCenterService.CreateBaseTask(taskID, models.TaskTypeImage, operator); err != nil {
		log.Printf("Failed to create task center base: %v", err)
		utils.RespondError(w, fmt.Errorf("failed to create task: %w", err), http.StatusInternalServerError)
		return
	}
	
	// 3. 创建图片生成任务详细记录
	if err := h.imageTaskService.CreateTask(taskID, req.SKU, req.Keywords, req.SellingPoints, 
		req.CompetitorLink, req.CopywritingTaskID, req.Model, "1:1"); err != nil {
		log.Printf("Failed to create image task: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}
	
	log.Printf("Created image generation task: task_id=%s, sku=%s, operator=%s", taskID, req.SKU, operator)

	// 4. 响应客户端
	utils.RespondJSON(w, map[string]interface{}{
		"task_id": taskID,
		"message": "图片生成任务已创建，正在处理中",
	})

	// 5. 异步处理图片生成
	go func() {
		// 添加panic恢复
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in image generation goroutine for task %s: %v", taskID, r)
				h.imageTaskService.SaveError(taskID, fmt.Sprintf("Internal error: %v", r))
				h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			}
		}()
		
		ctx := context.Background()
		
		log.Printf("Starting image generation for task %s", taskID)
		
		// 更新状态为进行中
		if err := h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusOngoing); err != nil {
			log.Printf("Failed to update status to ongoing for task %s: %v", taskID, err)
		}

		// 构建AI生成提示词
		prompt := h.buildImageGenerationPrompt(req.SKU, req.Keywords, req.SellingPoints)
		log.Printf("Generated prompt for task %s: %s", taskID, prompt)
		
		// 调用AI模型生成图片
		imageReq := models.GenerateImageRequest{
			Prompt:      prompt,
			AspectRatio: "1:1",
			Model:       req.Model,
		}
		
		log.Printf("Calling AI service to generate image for task %s with model %s", taskID, req.Model)
		generatedDataURL, err := h.multiModelService.GenerateImage(ctx, imageReq)
		if err != nil {
			log.Printf("Image generation failed for task %s: %v", taskID, err)
			h.imageTaskService.SaveError(taskID, err.Error())
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}

		log.Printf("Image generated successfully for task %s, data URL length: %d", taskID, len(generatedDataURL))

		// 保存图片到本地文件
		localPath, err := h.localStorageService.SaveGeneratedImage(generatedDataURL)
		if err != nil {
			log.Printf("Failed to save image to local storage for task %s: %v", taskID, err)
			h.imageTaskService.SaveError(taskID, fmt.Sprintf("Failed to save image: %v", err))
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}
		
		// 生成访问URL
		imageURL := h.localStorageService.GetFileURL(localPath)
		log.Printf("Image saved to local storage: %s, access URL: %s", localPath, imageURL)

		// 保存结果（同时保存本地路径和访问URL）
		resultData := map[string]interface{}{
			"image_url":   imageURL,
			"local_path":  localPath,
			"prompt":      prompt,
			"data_url":    generatedDataURL[:100] + "...", // 只保存前100个字符作为记录
		}
		resultJSON, _ := json.Marshal(resultData)
		
		log.Printf("Saving result data for task %s", taskID)
		if err := h.imageTaskService.SaveResultData(taskID, string(resultJSON), imageURL); err != nil {
			log.Printf("Failed to save result for task %s: %v", taskID, err)
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}
		
		// 更新状态为已完成
		if err := h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusCompleted); err != nil {
			log.Printf("Failed to update status to completed for task %s: %v", taskID, err)
		}
		log.Printf("Task %s completed successfully", taskID)
	}()
}

func (h *TaskHandler) buildImageGenerationPrompt(sku, keywords, sellingPoints string) string {
	prompt := "Create a professional product image for Amazon listing.\n"
	
	if sku != "" {
		prompt += fmt.Sprintf("Product SKU: %s\n", sku)
	}
	
	if keywords != "" {
		prompt += fmt.Sprintf("Keywords: %s\n", keywords)
	}
	
	if sellingPoints != "" {
		prompt += fmt.Sprintf("Product Features: %s\n", sellingPoints)
	}
	
	prompt += "\nRequirements:\n"
	prompt += "- High quality, professional photography style\n"
	prompt += "- Clean white background\n"
	prompt += "- Product should be centered and well-lit\n"
	prompt += "- Show product details clearly\n"
	prompt += "- Suitable for e-commerce listing\n"
	
	return prompt
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
