package utils

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/genai"
)

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
