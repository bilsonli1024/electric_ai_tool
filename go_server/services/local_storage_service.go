package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalStorageService struct {
	baseDir string
}

func NewLocalStorageService() *LocalStorageService {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
		workDir = "."
	}
	
	// 创建与logs同级的uploads目录
	baseDir := filepath.Join(workDir, "uploads")
	
	// 确保目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Printf("Failed to create uploads directory: %v", err)
	}
	
	log.Printf("Local storage initialized at: %s", baseDir)
	
	return &LocalStorageService{
		baseDir: baseDir,
	}
}

// SaveUploadedImage 保存上传的图片
func (s *LocalStorageService) SaveUploadedImage(data []byte, originalFilename string) (string, error) {
	// 提取文件扩展名
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = ".jpg" // 默认扩展名
	}
	
	// 生成唯一文件名：upload_年月日时分秒_随机字符串.扩展名
	timestamp := time.Now().Format("20060102_150405")
	randomStr := generateRandomString(8)
	filename := fmt.Sprintf("upload_%s_%s%s", timestamp, randomStr, ext)
	
	// 创建uploads子目录
	uploadDir := filepath.Join(s.baseDir, "images")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}
	
	// 完整文件路径
	fullPath := filepath.Join(uploadDir, filename)
	
	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	
	// 返回相对路径（用于数据库存储和URL访问）
	relativePath := filepath.Join("images", filename)
	log.Printf("Uploaded image saved: %s", relativePath)
	
	return relativePath, nil
}

// SaveGeneratedImage 保存AI生成的图片（从base64 data URL）
func (s *LocalStorageService) SaveGeneratedImage(dataURL string) (string, error) {
	// 解析data URL: data:image/png;base64,iVBORw0KGgo...
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid data URL format")
	}
	
	// 提取MIME类型和扩展名
	mimeType := strings.TrimPrefix(parts[0], "data:")
	mimeType = strings.TrimSuffix(mimeType, ";base64")
	
	ext := ".png" // 默认
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		ext = ".jpg"
	} else if strings.Contains(mimeType, "webp") {
		ext = ".webp"
	}
	
	// 解码base64
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	
	// 生成唯一文件名：generated_年月日时分秒_随机字符串.扩展名
	timestamp := time.Now().Format("20060102_150405")
	randomStr := generateRandomString(8)
	filename := fmt.Sprintf("generated_%s_%s%s", timestamp, randomStr, ext)
	
	// 创建generated子目录
	generatedDir := filepath.Join(s.baseDir, "generated")
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create generated directory: %w", err)
	}
	
	// 完整文件路径
	fullPath := filepath.Join(generatedDir, filename)
	
	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	
	// 返回相对路径
	relativePath := filepath.Join("generated", filename)
	log.Printf("Generated image saved: %s (size: %d bytes)", relativePath, len(data))
	
	return relativePath, nil
}

// GetFilePath 获取文件的完整路径
func (s *LocalStorageService) GetFilePath(relativePath string) string {
	return filepath.Join(s.baseDir, relativePath)
}

// GetFileURL 获取文件的访问URL
func (s *LocalStorageService) GetFileURL(relativePath string) string {
	// 返回相对于服务器根目录的URL路径
	return "/uploads/" + strings.ReplaceAll(relativePath, "\\", "/")
}

// DeleteFile 删除文件
func (s *LocalStorageService) DeleteFile(relativePath string) error {
	fullPath := s.GetFilePath(relativePath)
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	log.Printf("File deleted: %s", relativePath)
	return nil
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 降级方案：使用时间戳
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)[:length]
}

// SaveFileFromReader 从Reader保存文件
func (s *LocalStorageService) SaveFileFromReader(reader io.Reader, originalFilename string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}
	return s.SaveUploadedImage(data, originalFilename)
}
