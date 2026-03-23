package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// StandardResponse 统一响应格式
type StandardResponse struct {
	Code    int         `json:"code"`    // 0=成功，其他=失败
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 响应数据
}

// RespondSuccess 成功响应
func RespondSuccess(w http.ResponseWriter, data interface{}, message string) {
	if message == "" {
		message = "操作成功"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StandardResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// RespondSuccessWithMsg 成功响应（只有消息，无数据）
func RespondSuccessWithMsg(w http.ResponseWriter, message string) {
	if message == "" {
		message = "操作成功"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StandardResponse{
		Code:    0,
		Message: message,
		Data:    nil,
	})
}

// RespondFail 失败响应
func RespondFail(w http.ResponseWriter, code int, message string) {
	if code == 0 {
		code = 1 // 默认错误码
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StandardResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// RespondErrorWithCode 错误响应（带自定义错误码）
func RespondErrorWithCode(w http.ResponseWriter, err error, code int) {
	log.Printf("Error: %v", err)
	if code == 0 {
		code = 1
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StandardResponse{
		Code:    code,
		Message: err.Error(),
		Data:    nil,
	})
}

// ============= 兼容旧代码的函数 =============

// RespondJSON 返回JSON数据（兼容旧代码，现在使用标准格式）
func RespondJSON(w http.ResponseWriter, data interface{}) {
	RespondSuccess(w, data, "")
}

// RespondError 返回错误（兼容旧代码，现在使用标准格式）
func RespondError(w http.ResponseWriter, err error, status int) {
	log.Printf("Error: %v", err)
	// 根据HTTP状态码映射到自定义错误码
	code := httpStatusToCode(status)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StandardResponse{
		Code:    code,
		Message: err.Error(),
		Data:    nil,
	})
}

// httpStatusToCode 将HTTP状态码映射到自定义错误码
func httpStatusToCode(status int) int {
	switch status {
	case http.StatusBadRequest:
		return 400
	case http.StatusUnauthorized:
		return 401
	case http.StatusForbidden:
		return 403
	case http.StatusNotFound:
		return 404
	case http.StatusMethodNotAllowed:
		return 405
	case http.StatusInternalServerError:
		return 500
	default:
		return 1 // 默认错误码
	}
}

