package utils

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

const (
	// 日志级别
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// LogInfo 记录INFO级别日志
func LogInfo(format string, args ...interface{}) {
	logWithLevel(LogLevelInfo, format, args...)
}

// LogError 记录ERROR级别日志
func LogError(format string, args ...interface{}) {
	logWithLevel(LogLevelError, format, args...)
}

// LogWarn 记录WARN级别日志
func LogWarn(format string, args ...interface{}) {
	logWithLevel(LogLevelWarn, format, args...)
}

// LogDebug 记录DEBUG级别日志
func LogDebug(format string, args ...interface{}) {
	logWithLevel(LogLevelDebug, format, args...)
}

func logWithLevel(level string, format string, args ...interface{}) {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2)
	caller := ""
	if ok {
		// 只保留文件名，不要完整路径
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}
	
	// 根据级别选择前缀
	prefix := ""
	switch level {
	case LogLevelError:
		prefix = "❌ [ERROR]"
	case LogLevelWarn:
		prefix = "⚠️  [WARN]"
	case LogLevelInfo:
		prefix = "ℹ️  [INFO]"
	case LogLevelDebug:
		prefix = "🔍 [DEBUG]"
	}
	
	// 格式化消息
	message := fmt.Sprintf(format, args...)
	
	// 输出日志
	if caller != "" {
		log.Printf("%s [%s] %s", prefix, caller, message)
	} else {
		log.Printf("%s %s", prefix, message)
	}
}
