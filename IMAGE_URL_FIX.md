# 图片生成数据URL格式错误修复

## 问题描述
```
auth_task_handlers.go:430: Image generation failed for task IG2026032322034874015: 
invalid data URL format
```

## 根本原因
`MakeImagePart` 函数在解析 data URL 时缺少详细的错误信息，无法准确定位问题所在。可能的原因包括：
1. 前端传递的图片URL不是有效的 data URL 格式
2. 图片URL可能是本地文件路径而不是HTTP URL
3. HTTP URL下载失败或返回空数据

## 修复内容

### 1. 增强 `MakeImagePart` 函数 (`utils/image.go`)

**改进**:
- 添加data URL格式验证（必须以 `data:` 开头）
- 改进错误消息，显示实际接收到的URL前缀
- 添加 `min()` 辅助函数避免日志中打印过长字符串

**修复后的代码**:
```go
func MakeImagePart(dataURL string) (*genai.Part, error) {
	// 验证dataURL格式
	if !strings.HasPrefix(dataURL, "data:") {
		return nil, fmt.Errorf("invalid data URL format: must start with 'data:', got: %s", 
			dataURL[:min(50, len(dataURL))])
	}
	
	mimeType := GetMimeType(dataURL)
	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data URL format: missing comma separator, dataURL prefix: %s", 
			dataURL[:min(100, len(dataURL))])
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
```

### 2. 增强 `ConvertURLToDataURL` 函数 (`utils/image.go`)

**新增功能**:
- 验证输入URL不为空
- 检测并拒绝本地文件路径（以 `/uploads/` 开头）
- 验证HTTP/HTTPS协议
- 验证下载的数据不为空
- 添加详细的日志信息
- 改进错误消息，包含具体的URL信息

**修复后的代码**:
```go
func ConvertURLToDataURL(url string) (string, error) {
	// 如果已经是data URL，直接返回
	if strings.HasPrefix(url, "data:") {
		LogInfo("URL is already a data URL")
		return url, nil
	}
	
	// 验证非空
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}
	
	LogInfo("Converting URL to data URL: %s", url)
	
	// 检查是否是本地文件路径
	if strings.HasPrefix(url, "/uploads/") || strings.HasPrefix(url, "./uploads/") {
		return "", fmt.Errorf("local file path detected (%s), please provide HTTP URL or data URL", url)
	}
	
	// 验证HTTP/HTTPS协议
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("invalid URL format: must be HTTP/HTTPS URL or data URL, got: %s", 
			url[:min(100, len(url))])
	}
	
	// 下载图片
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image from %s: %w", url, err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image from %s: HTTP status %d", url, resp.StatusCode)
	}
	
	// 读取数据
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}
	
	// 验证数据大小
	if len(data) == 0 {
		return "", fmt.Errorf("downloaded image is empty")
	}
	
	LogInfo("Downloaded image: %d bytes", len(data))
	
	// 构造data URL
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	
	encoded := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", contentType, encoded)
	
	LogInfo("Successfully converted URL to data URL (size: %d bytes, data URL length: %d)", 
		len(data), len(dataURL))
	
	return dataURL, nil
}
```

## 错误诊断

现在当图片生成失败时，日志会提供更详细的信息：

### 情况1: 本地文件路径
```
Error: local file path detected (/uploads/image.jpg), please provide HTTP URL or data URL
```

### 情况2: 无效URL格式
```
Error: invalid URL format: must be HTTP/HTTPS URL or data URL, got: file:///path/to/image.jpg
```

### 情况3: 下载失败
```
Error: failed to download image from http://example.com/image.jpg: HTTP status 404
```

### 情况4: Data URL格式错误
```
Error: invalid data URL format: must start with 'data:', got: http://example.com/image.jpg
```

### 情况5: Data URL缺少分隔符
```
Error: invalid data URL format: missing comma separator, dataURL prefix: data:image/jpeg;base64
```

## 使用建议

### 前端需要确保：

1. **如果上传文件**：需要将文件转换为data URL
```javascript
const file = event.target.files[0];
const reader = new FileReader();
reader.onload = (e) => {
  const dataURL = e.target.result; // 格式: data:image/jpeg;base64,/9j/4AAQ...
  // 发送dataURL到后端
};
reader.readAsDataURL(file);
```

2. **如果提供外部URL**：必须是完整的HTTP/HTTPS URL
```javascript
const imageUrl = "https://example.com/product-image.jpg"; // ✅ 正确
// 不要使用: "/uploads/image.jpg" ❌ 错误
// 不要使用: "./uploads/image.jpg" ❌ 错误
```

3. **如果使用已上传的图片**：需要构造完整URL
```javascript
const serverUrl = process.env.VITE_SERVER_GO_LINK_IP;
const fullUrl = `http://${serverUrl}:4002/uploads/${filename}`; // ✅ 正确
```

## 测试验证

1. **编译状态**: ✅ 成功
2. **文件大小**: 21MB
3. **Go版本**: 1.24.6

## 下一步排查

如果问题仍然存在，请：

1. 检查前端发送的请求体，确认 `product_images` 数组的格式
2. 查看完整的错误日志，现在会包含URL的前50-100个字符
3. 确认图片URL是否可以在浏览器中直接访问
4. 如果使用本地上传的图片，确保构造完整的HTTP URL

## 文件修改

- `go_server/utils/image.go`: 增强错误处理和验证逻辑
- 编译产物: `go_server/electric_ai_tool` (21MB)

---

**修复时间**: 2026-03-23 22:18  
**编译状态**: ✅ 成功
