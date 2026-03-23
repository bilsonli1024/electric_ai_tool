package utils

import (
	"time"
)

// 东八区时区
var ChinaTimezone = time.FixedZone("CST", 8*3600)

// GetCurrentTimestamp 获取当前时间的UNIX时间戳（秒）
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentTime 获取当前东八区时间
func GetCurrentTime() time.Time {
	return time.Now().In(ChinaTimezone)
}

// TimestampToTime 将UNIX时间戳转换为东八区Time对象
func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0).In(ChinaTimezone)
}

// TimestampToString 将UNIX时间戳转换为东八区格式化字符串
// 格式: 2026-03-23 12:00:00
func TimestampToString(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(timestamp, 0).In(ChinaTimezone).Format("2006-01-02 15:04:05")
}

// StringToTimestamp 将东八区格式化字符串转换为UNIX时间戳
// 格式: 2026-03-23 12:00:00
func StringToTimestamp(timeStr string) (int64, error) {
	if timeStr == "" {
		return 0, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, ChinaTimezone)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// TimeToTimestamp 将Time对象转换为UNIX时间戳
func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}

// FormatTimestamp 将时间戳格式化为指定格式的字符串（东八区）
func FormatTimestamp(timestamp int64, layout string) string {
	if timestamp == 0 {
		return ""
	}
	return time.Unix(timestamp, 0).In(ChinaTimezone).Format(layout)
}
