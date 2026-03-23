package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// 注册用户（不自动登录）
	err := h.authService.Register(req)
	if err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	// 返回成功消息，提示等待审批
	utils.RespondJSON(w, map[string]interface{}{
		"message": "注册成功，请等待管理员审批后登录",
		"success": true,
	})
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

	if req.Email == "" {
		utils.RespondError(w, fmt.Errorf("邮箱不能为空"), http.StatusBadRequest)
		return
	}

	// 发送重置邮件
	err := h.authService.ForgotPassword(req.Email)
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

	if req.Token == "" || req.NewPassword == "" {
		utils.RespondError(w, fmt.Errorf("token和新密码不能为空"), http.StatusBadRequest)
		return
	}

	// 重置密码
	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, map[string]string{
		"message": "密码重置成功，请重新登录",
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取用户ID
	userIDValue := r.Context().Value("user_id")
	if userIDValue == nil {
		utils.RespondError(w, fmt.Errorf("unauthorized"), http.StatusUnauthorized)
		return
	}
	userID := userIDValue.(int64)

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		utils.RespondError(w, fmt.Errorf("旧密码和新密码不能为空"), http.StatusBadRequest)
		return
	}

	// 修改密码
	err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	utils.RespondJSON(w, map[string]string{
		"message": "密码修改成功",
	})
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
	authService         *services.AuthService
	taskCenterService   *services.TaskCenterService
	imageTaskService    *services.ImageTaskService
	localStorageService *services.LocalStorageService
}

func NewTaskHandler(multiModelService *services.MultiModelService,
	authService *services.AuthService,
	taskCenterService *services.TaskCenterService, imageTaskService *services.ImageTaskService,
	localStorageService *services.LocalStorageService) *TaskHandler {
	return &TaskHandler{
		multiModelService:   multiModelService,
		authService:         authService,
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

	userID, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		return 0, "", err
	}

	// 获取用户详细信息
	user, err := h.authService.GetUserByID(userID)
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

	userID, err := h.authService.ValidateSession(sessionID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (h *TaskHandler) AnalyzeWithTask(w http.ResponseWriter, r *http.Request) {
	// 该功能已迁移到新的文案生成API
	// 请使用 /api/copywriting/analyze
	utils.RespondJSON(w, map[string]string{
		"message": "该API已废弃，请使用 /api/copywriting/analyze",
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
		SKU                  string   `json:"sku"`
		Keywords             string   `json:"keywords"`
		SellingPoints        string   `json:"sellingPoints"`
		CompetitorLink       string   `json:"competitorLink"`
		Model                int      `json:"model"`
		TaskName             string   `json:"taskName"`
		CopywritingTaskID    string   `json:"copywritingTaskId"`
		ProductImages        []string `json:"productImages"` // 产品图片URL数组
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Model == 0 {
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
		
		utils.LogInfo("Starting image generation for task %s", taskID)
		
		// 更新状态为generating（生成中）
		if err := h.imageTaskService.UpdateDetailStatus(taskID, models.ImageStatusGenerating); err != nil {
			utils.LogError("Failed to update detail status to generating for task %s: %v", taskID, err)
		}
		if err := h.taskCenterService.UpdateTaskStatus(taskID, models.MapDetailStatusToTaskStatus(models.TaskTypeImage, models.ImageStatusGenerating)); err != nil {
			utils.LogError("Failed to update task status to ongoing for task %s: %v", taskID, err)
		}

		// 构建AI生成提示词
		prompt := h.buildImageGenerationPrompt(req.SKU, req.Keywords, req.SellingPoints)
		utils.LogInfo("Generated prompt for task %s: %s", taskID, prompt)
		
		// 调用AI模型生成图片
		imageReq := models.GenerateImageRequest{
			Prompt:        prompt,
			AspectRatio:   "1:1",
			Model:         req.Model,
			ProductImages: req.ProductImages, // 传递产品图片
		}
		
		utils.LogInfo("Calling AI service to generate image for task %s with model %s (product images: %d)", 
			taskID, req.Model, len(req.ProductImages))
		generatedDataURL, err := h.multiModelService.GenerateImage(ctx, imageReq)
		if err != nil {
			utils.LogError("Image generation failed for task %s: %v", taskID, err)
			h.imageTaskService.SaveError(taskID, err.Error())
			h.imageTaskService.UpdateDetailStatus(taskID, models.ImageStatusFailed)
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}
		
		// 验证返回的数据
		if generatedDataURL == "" {
			errMsg := "AI返回了空的图片数据"
			utils.LogError("Image generation failed for task %s: %s", taskID, errMsg)
			h.imageTaskService.SaveError(taskID, errMsg)
			h.imageTaskService.UpdateDetailStatus(taskID, models.ImageStatusFailed)
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}

		utils.LogInfo("Image generated successfully for task %s, data URL length: %d", taskID, len(generatedDataURL))

		// 保存图片到本地文件
		localPath, err := h.localStorageService.SaveGeneratedImage(generatedDataURL)
		if err != nil {
			utils.LogError("Failed to save image to local storage for task %s: %v", taskID, err)
			h.imageTaskService.SaveError(taskID, fmt.Sprintf("Failed to save image: %v", err))
			h.imageTaskService.UpdateDetailStatus(taskID, models.ImageStatusFailed)
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}
		
		// 生成访问URL
		imageURL := h.localStorageService.GetFileURL(localPath)
		utils.LogInfo("Image saved to local storage: %s, access URL: %s", localPath, imageURL)

		// 保存结果（同时保存本地路径和访问URL）
		resultData := map[string]interface{}{
			"image_url":   imageURL,
			"local_path":  localPath,
			"prompt":      prompt,
			"data_url":    generatedDataURL[:100] + "...", // 只保存前100个字符作为记录
		}
		resultJSON, _ := json.Marshal(resultData)
		
		utils.LogInfo("Saving result data for task %s", taskID)
		if err := h.imageTaskService.SaveResultData(taskID, string(resultJSON), imageURL); err != nil {
			utils.LogError("Failed to save result for task %s: %v", taskID, err)
			h.imageTaskService.UpdateDetailStatus(taskID, models.ImageStatusFailed)
			h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusFailed)
			return
		}
		
		// 更新状态为已完成
		h.imageTaskService.UpdateDetailStatus(taskID, models.ImageStatusCompleted)
		if err := h.taskCenterService.UpdateTaskStatus(taskID, models.TaskStatusCompleted); err != nil {
			utils.LogError("Failed to update status to completed for task %s: %v", taskID, err)
		}
		utils.LogInfo("Task %s completed successfully", taskID)
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
	// 该功能已迁移到新的任务中心API
	// 请使用 /api/task-center/list
	utils.RespondJSON(w, map[string]string{
		"message": "该API已废弃，请使用 /api/task-center/list",
	})
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	// 该功能已迁移到新的任务中心API
	// 请使用 /api/task-center/list
	utils.RespondJSON(w, map[string]string{
		"message": "该API已废弃，请使用 /api/task-center/list",
	})
}

func (h *TaskHandler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	// 任务历史功能已废弃
	utils.RespondJSON(w, map[string]string{
		"message": "该功能已废弃",
	})
}
