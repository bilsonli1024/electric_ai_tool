package handlers

import (
	"encoding/base64"
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

	// 解码base64数据
	imageData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		log.Printf("Failed to decode base64: %v", err)
		utils.RespondError(w, fmt.Errorf("invalid base64 data: %w", err), http.StatusBadRequest)
		return
	}

	// 确定文件名（如果提供了原始文件名）
	filename := req.Filename
	if filename == "" {
		// 从data URL中提取扩展名
		mimeType := strings.TrimPrefix(parts[0], "data:")
		mimeType = strings.TrimSuffix(mimeType, ";base64")
		if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
			filename = "image.jpg"
		} else if strings.Contains(mimeType, "png") {
			filename = "image.png"
		} else if strings.Contains(mimeType, "webp") {
			filename = "image.webp"
		} else {
			filename = "image.jpg" // 默认
		}
	}

	// 使用SaveUploadedImage保存（直接保存到images目录）
	relativePath, err := h.localStorageService.SaveUploadedImage(imageData, filename)
	if err != nil {
		log.Printf("Failed to save base64 image: %v", err)
		utils.RespondError(w, err, http.StatusInternalServerError)
		return
	}

	// 生成访问URL
	imageURL := h.localStorageService.GetFileURL(relativePath)

	log.Printf("✅ Base64 image uploaded successfully: %s", imageURL)

	utils.RespondJSON(w, map[string]interface{}{
		"url":      imageURL,
		"path":     relativePath,
		"message":  "Image uploaded successfully",
	})
}
