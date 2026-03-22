package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"electric_ai_tool/go_server/services"
	"electric_ai_tool/go_server/utils"
)

type UploadHandler struct {
	localStorageService *services.LocalStorageService
	authService         *services.AuthService
}

func NewUploadHandler(localStorageService *services.LocalStorageService, authService *services.AuthService) *UploadHandler {
	return &UploadHandler{
		localStorageService: localStorageService,
		authService:         authService,
	}
}

// UploadImage 上传单张图片
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析multipart form (最大32MB)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to parse form: %w", err), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to get file: %w", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("📤 Uploading image: %s (size: %d bytes)", header.Filename, header.Size)

	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		utils.RespondError(w, fmt.Errorf("failed to read file: %w", err), http.StatusInternalServerError)
		return
	}

	// 保存到本地存储
	relativePath, err := h.localStorageService.SaveUploadedImage(data, header.Filename)
	if err != nil {
		log.Printf("Failed to save uploaded image: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 生成访问URL
	imageURL := h.localStorageService.GetFileURL(relativePath)

	log.Printf("✅ Image uploaded successfully: %s", imageURL)

	utils.RespondJSON(w, map[string]interface{}{
		"url":       imageURL,
		"path":      relativePath,
		"filename":  header.Filename,
		"size":      header.Size,
		"message":   "Image uploaded successfully",
	})
}

// UploadImageBase64 上传base64编码的图片
func (h *UploadHandler) UploadImageBase64(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Image    string `json:"image"`    // base64 data URL
		Filename string `json:"filename"` // 可选的原始文件名
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, err, http.StatusBadRequest)
		return
	}

	if req.Image == "" {
		utils.RespondError(w, fmt.Errorf("image data is required"), http.StatusBadRequest)
		return
	}

	// 解析data URL
	parts := strings.SplitN(req.Image, ",", 2)
	if len(parts) != 2 {
		utils.RespondError(w, fmt.Errorf("invalid image data format"), http.StatusBadRequest)
		return
	}

	log.Printf("📤 Uploading base64 image (size: ~%d bytes)", len(req.Image))

	// 使用SaveGeneratedImage方法（它会处理base64解码）
	relativePath, err := h.localStorageService.SaveGeneratedImage(req.Image)
	if err != nil {
		log.Printf("Failed to save base64 image: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 修改路径：从generated移到images目录
	// 因为这是用户上传的，不是AI生成的
	relativePath = strings.Replace(relativePath, "generated/", "images/", 1)

	// 生成访问URL
	imageURL := h.localStorageService.GetFileURL(relativePath)

	log.Printf("✅ Base64 image uploaded successfully: %s", imageURL)

	utils.RespondJSON(w, map[string]interface{}{
		"url":      imageURL,
		"path":     relativePath,
		"message":  "Image uploaded successfully",
	})
}
