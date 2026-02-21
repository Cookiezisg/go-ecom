package utils

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, &Response{
		Code:    0,
		Message: "成功",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(w http.ResponseWriter, message string, data interface{}) {
	httpx.OkJson(w, &Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func Error(w http.ResponseWriter, code int, message string) {
	httpx.WriteJson(w, http.StatusOK, &Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithHTTPCode 错误响应（自定义HTTP状态码）
func ErrorWithHTTPCode(w http.ResponseWriter, httpCode, code int, message string) {
	httpx.WriteJson(w, httpCode, &Response{
		Code:    code,
		Message: message,
	})
}
