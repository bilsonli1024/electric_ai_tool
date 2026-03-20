package services

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"electric_ai_tool/go_server/config"
	"electric_ai_tool/go_server/models"
)

type CDNService struct {
	cdnEndpoint string
	cdnBucket   string
	cdnAccessKey string
	cdnSecretKey string
}

func NewCDNService() *CDNService {
	return &CDNService{
		cdnEndpoint:  os.Getenv("CDN_ENDPOINT"),
		cdnBucket:    os.Getenv("CDN_BUCKET"),
		cdnAccessKey: os.Getenv("CDN_ACCESS_KEY"),
		cdnSecretKey: os.Getenv("CDN_SECRET_KEY"),
	}
}

func (s *CDNService) UploadImage(userID int64, dataURL string, imageType string) (*models.CDNImage, error) {
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format")
	}

	mimeType := s.extractMimeType(dataURL)
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	hash := md5.Sum(data)
	cdnKey := fmt.Sprintf("images/%d/%s/%s.%s", userID, time.Now().Format("2006/01/02"), hex.EncodeToString(hash[:]), s.getExtension(mimeType))

	cdnURL, err := s.uploadToCDN(cdnKey, data, mimeType)
	if err != nil {
		return nil, err
	}

	cdnImage := &models.CDNImage{
		UserID:    userID,
		CDNURL:    cdnURL,
		CDNKey:    cdnKey,
		FileSize:  int64(len(data)),
		MimeType:  mimeType,
		ImageType: imageType,
		CreatedAt: time.Now(),
	}

	query := `INSERT INTO cdn_images_tab (user_id, cdn_url, cdn_key, file_size, mime_type, image_type) 
              VALUES (?, ?, ?, ?, ?, ?)`
	result, err := config.DB.Exec(query, cdnImage.UserID, cdnImage.CDNURL, cdnImage.CDNKey, cdnImage.FileSize, cdnImage.MimeType, cdnImage.ImageType)
	if err != nil {
		return nil, fmt.Errorf("failed to save CDN image record: %w", err)
	}

	id, _ := result.LastInsertId()
	cdnImage.ID = id

	return cdnImage, nil
}

func (s *CDNService) UploadFromMultipart(userID int64, file multipart.File, header *multipart.FileHeader, imageType string) (*models.CDNImage, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	hash := md5.Sum(data)
	cdnKey := fmt.Sprintf("images/%d/%s/%s.%s", userID, time.Now().Format("2006/01/02"), hex.EncodeToString(hash[:]), s.getExtension(mimeType))

	cdnURL, err := s.uploadToCDN(cdnKey, data, mimeType)
	if err != nil {
		return nil, err
	}

	cdnImage := &models.CDNImage{
		UserID:           userID,
		OriginalFilename: header.Filename,
		CDNURL:           cdnURL,
		CDNKey:           cdnKey,
		FileSize:         int64(len(data)),
		MimeType:         mimeType,
		ImageType:        imageType,
		CreatedAt:        time.Now(),
	}

	query := `INSERT INTO cdn_images_tab (user_id, original_filename, cdn_url, cdn_key, file_size, mime_type, image_type) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := config.DB.Exec(query, cdnImage.UserID, cdnImage.OriginalFilename, cdnImage.CDNURL, cdnImage.CDNKey, cdnImage.FileSize, cdnImage.MimeType, cdnImage.ImageType)
	if err != nil {
		return nil, fmt.Errorf("failed to save CDN image record: %w", err)
	}

	id, _ := result.LastInsertId()
	cdnImage.ID = id

	return cdnImage, nil
}

func (s *CDNService) uploadToCDN(key string, data []byte, mimeType string) (string, error) {
	if s.cdnEndpoint == "" {
		localPath := fmt.Sprintf("/tmp/cdn_uploads/%s", key)
		os.MkdirAll(strings.TrimSuffix(localPath, "/"+strings.Split(key, "/")[len(strings.Split(key, "/"))-1]), 0755)
		if err := os.WriteFile(localPath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to save local file: %w", err)
		}
		return fmt.Sprintf("file://%s", localPath), nil
	}

	url := fmt.Sprintf("%s/%s/%s", s.cdnEndpoint, s.cdnBucket, key)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", mimeType)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to CDN: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("CDN upload failed with status: %d", resp.StatusCode)
	}

	return url, nil
}

func (s *CDNService) extractMimeType(dataURL string) string {
	if strings.HasPrefix(dataURL, "data:") {
		end := strings.Index(dataURL, ";")
		if end > 5 {
			return dataURL[5:end]
		}
	}
	return "image/png"
}

func (s *CDNService) getExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	default:
		return "png"
	}
}
