package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateSalt 生成16位随机盐值（数字和字母）
func GenerateSalt() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	salt := make([]byte, 16)
	
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// 降级方案：使用时间戳
		return fmt.Sprintf("%016d", GetCurrentTimestamp())
	}
	
	for i := 0; i < 16; i++ {
		salt[i] = charset[randomBytes[i]%byte(len(charset))]
	}
	
	return string(salt)
}

// HashPasswordWithSalt 使用MD5哈希值和salt计算最终密码
// passwordMD5: 前端传来的MD5哈希值
// salt: 用户的盐值
// 返回: SHA256(passwordMD5 + salt)
func HashPasswordWithSalt(passwordMD5, salt string) string {
	combined := passwordMD5 + salt
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// VerifyPassword 验证密码
// passwordMD5: 前端传来的MD5哈希值
// salt: 数据库中存储的盐值
// hashedPassword: 数据库中存储的哈希密码
func VerifyPassword(passwordMD5, salt, hashedPassword string) bool {
	calculatedHash := HashPasswordWithSalt(passwordMD5, salt)
	return calculatedHash == hashedPassword
}
