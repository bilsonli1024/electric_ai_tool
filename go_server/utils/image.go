package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

// ConvertURLToDataURL 将HTTP URL或本地文件路径转换为data URL
func ConvertURLToDataURL(url string) (string, error) {
	// 如果已经是data URL，直接返回
	if strings.HasPrefix(url, "data:") {
		LogInfo("URL is already a data URL")
		return url, nil
	}
	
	// 如果是空字符串
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}
	
	LogInfo("Converting URL to data URL: %s", url)
	
	// 检查是否是本地文件路径
	if strings.HasPrefix(url, "/uploads/") || strings.HasPrefix(url, "./uploads/") || strings.HasPrefix(url, "uploads/") {
		// 读取本地文件
		return convertLocalFileToDataURL(url)
	}
	
	// 验证是否是有效的HTTP URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("invalid URL format: must be HTTP/HTTPS URL or data URL, got: %s", url[:min(100, len(url))])
	}
	
	// 从HTTP URL下载图片
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image from %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image from %s: HTTP status %d", url, resp.StatusCode)
	}
	
	// 读取图片数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}
	
	// 验证数据大小
	if len(data) == 0 {
		return "", fmt.Errorf("downloaded image is empty")
	}
	
	LogInfo("Downloaded image: %d bytes", len(data))
	
	// 获取content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // 默认
	}
	
	// 编码为base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// 构造data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
	
	LogInfo("Successfully converted URL to data URL (size: %d bytes, data URL length: %d)", len(data), len(dataURL))
	
	return dataURL, nil
}

// convertLocalFileToDataURL 将本地文件转换为data URL
func convertLocalFileToDataURL(localPath string) (string, error) {
	// 移除开头的 ./ 或 /
	localPath = strings.TrimPrefix(localPath, "./")
	localPath = strings.TrimPrefix(localPath, "/")
	
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	
	// 构造完整路径
	fullPath := filepath.Join(workDir, localPath)
	LogInfo("Reading local file: %s", fullPath)
	
	// 读取文件
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read local file %s: %w", fullPath, err)
	}
	
	// 验证数据大小
	if len(data) == 0 {
		return "", fmt.Errorf("local file is empty: %s", fullPath)
	}
	
	LogInfo("Read local file: %d bytes", len(data))
	
	// 根据文件扩展名判断content type
	ext := strings.ToLower(filepath.Ext(fullPath))
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}
	
	// 编码为base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// 构造data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
	
	LogInfo("Successfully converted local file to data URL (size: %d bytes, data URL length: %d)", len(data), len(dataURL))
	
	return dataURL, nil
}

func MakeImagePart(dataURL string) (*genai.Part, error) {
	// 验证dataURL格式
	if !strings.HasPrefix(dataURL, "data:") {
		return nil, fmt.Errorf("invalid data URL format: must start with 'data:', got: %s", dataURL[:min(50, len(dataURL))])
	}
	
	mimeType := GetMimeType(dataURL)
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format: missing comma separator, dataURL prefix: %s", dataURL[:min(100, len(dataURL))])
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return &genai.Part{
		InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     data,
		},
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetMimeType(dataURL string) string {
	re := regexp.MustCompile(`^data:([^;]+);`)
	matches := re.FindStringSubmatch(dataURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return "image/png"
}

func ExtractImageFromResponse(resp *genai.GenerateContentResponse) string {
	if resp.Candidates == nil || len(resp.Candidates) == 0 {
		return ""
	}

	for _, part := range resp.Candidates[0].Content.Parts {
		if part.InlineData != nil {
			encoded := base64.StdEncoding.EncodeToString(part.InlineData.Data)
			return fmt.Sprintf("data:%s;base64,%s", part.InlineData.MIMEType, encoded)
		}
	}
	return ""
}
