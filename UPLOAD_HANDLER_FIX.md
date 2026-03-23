# 🔧 Nil Pointer Panic 修复指南

## 问题

```
runtime error: invalid memory address or nil pointer dereference
```

发生在 `/api/upload/image-base64` 接口

## 根本原因

`UploadImageBase64` 方法中，在第120行对相对路径进行了字符串替换（从`generated/`改为`images/`），但实际文件仍保存在`generated/`目录。当调用`GetFileURL()`时使用了不存在的路径，可能导致nil pointer错误。

## 修复内容

已修改 `handlers/upload_handler.go` 的 `UploadImageBase64` 方法：

### 修改前（错误的逻辑）
```go
// 使用SaveGeneratedImage方法（它会处理base64解码）
relativePath, err := h.localStorageService.SaveGeneratedImage(req.Image)
if err != nil {
    log.Printf("Failed to save base64 image: %v", err)
    utils.RespondError(w, err, http.StatusInternalServerError)
    return
}

// 修改路径：从generated移到images目录
// ⚠️ 问题：只修改了字符串，没有移动文件！
relativePath = strings.Replace(relativePath, "generated/", "images/", 1)

// 生成访问URL
imageURL := h.localStorageService.GetFileURL(relativePath)
```

### 修改后（正确的逻辑）
```go
// 解码base64数据
imageData, err := base64.StdEncoding.DecodeString(parts[1])
if err != nil {
    log.Printf("Failed to decode base64: %v", err)
    utils.RespondError(w, fmt.Errorf("invalid base64 data: %w", err), http.StatusBadRequest)
    return
}

// 确定文件名
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

// ✅ 直接使用SaveUploadedImage保存到images目录
relativePath, err := h.localStorageService.SaveUploadedImage(imageData, filename)
if err != nil {
    log.Printf("Failed to save base64 image: %v", err)
    utils.RespondError(w, err, http.StatusInternalServerError)
    return
}

// 生成访问URL
imageURL := h.localStorageService.GetFileURL(relativePath)
```

## 部署步骤（在服务器上执行）

### 1. 备份当前运行的文件

```bash
cd /home/lc/electric_ai_tool/go_server
cp handlers/upload_handler.go handlers/upload_handler.go.backup
```

### 2. 应用修复

将修复后的 `handlers/upload_handler.go` 文件上传到服务器，或者手动编辑：

```bash
vi handlers/upload_handler.go
```

找到 `UploadImageBase64` 函数（约第80行），将第108-123行替换为上面"修改后"的代码。

同时确保文件顶部的import包含：
```go
import (
    "encoding/base64"  // ← 确保有这一行
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"

    "electric_ai_tool/go_server/services"
    "electric_ai_tool/go_server/utils"
)
```

### 3. 重新编译

```bash
cd /home/lc/electric_ai_tool/go_server
go build -o electric_ai_tool .
```

如果遇到编译错误，可能需要：
```bash
go clean -cache
go mod tidy
go build -o electric_ai_tool .
```

### 4. 停止旧服务

```bash
# 查找进程
ps aux | grep electric_ai_tool

# 停止进程（替换PID）
kill <PID>

# 或者使用pkill
pkill -f electric_ai_tool
```

### 5. 启动新服务

```bash
cd /home/lc/electric_ai_tool/go_server
nohup ./electric_ai_tool > logs/server.log 2>&1 &

# 检查是否运行
ps aux | grep electric_ai_tool
```

### 6. 验证修复

```bash
# 检查日志
tail -f logs/server.log

# 测试接口（需要有效的session_id）
curl -X POST http://localhost:4002/api/upload/image-base64 \
  -H "Content-Type: application/json" \
  -H "Session-ID: YOUR_SESSION_ID" \
  -d '{
    "image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "filename": "test.png"
  }'
```

期望响应：
```json
{
  "url": "/uploads/images/upload_20260323_223000_abc123.png",
  "path": "images/upload_20260323_223000_abc123.png",
  "message": "Image uploaded successfully"
}
```

## 回滚（如果需要）

```bash
cd /home/lc/electric_ai_tool/go_server
cp handlers/upload_handler.go.backup handlers/upload_handler.go
go build -o electric_ai_tool .
pkill -f electric_ai_tool
./electric_ai_tool &
```

## 相关问题

这个修复同时解决了：
1. ❌ Nil pointer dereference panic
2. ❌ 文件保存路径与访问路径不匹配
3. ✅ 用户上传的图片现在正确保存到 `uploads/images/` 目录
4. ✅ AI生成的图片保存到 `uploads/generated/` 目录（通过 `SaveGeneratedImage`）

## 其他待修复问题

根据日志，还有一个数据库表结构问题需要修复：

```
Auth failed: session validation error: sql: Scan error on column index 1, 
name "expires_at": converting driver.Value type time.Time to a int64
```

参考 `SESSION_TYPE_FIX.md` 文档解决。

---

**修复日期**: 2026-03-23  
**影响范围**: `/api/upload/image-base64` 接口  
**需要重启**: 是
