package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

// ConvertURLToDataURL 将HTTP URL转换为data URL
func ConvertURLToDataURL(url string) (string, error) {
	// 如果已经是data URL，直接返回
	if strings.HasPrefix(url, "data:") {
		return url, nil
	}
	
	LogInfo("Converting URL to data URL: %s", url)
	
	// 从HTTP URL下载图片
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}
	
	// 读取图片数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}
	
	// 获取content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // 默认
	}
	
	// 编码为base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// 构造data URL
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
	
	LogInfo("Successfully converted URL to data URL (size: %d bytes)", len(dataURL))
	
	return dataURL, nil
}

func MakeImagePart(dataURL string) (*genai.Part, error) {
	mimeType := GetMimeType(dataURL)
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format")
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
