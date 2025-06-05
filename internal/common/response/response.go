package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 定义了标准的API响应结构
type Response struct {
	Code    int         `json:"code"`           // 业务状态码，0表示成功，其他表示特定错误
	Message string      `json:"message"`        // 提示信息
	Data    interface{} `json:"data,omitempty"` // 响应数据，成功时可能包含，错误时可省略
}

// JSON 是一个通用的响应发送函数
func JSON(c *gin.Context, httpCode int, bizCode int, msg string, data interface{}) {
	c.JSON(httpCode, Response{
		Code:    bizCode,
		Message: msg,
		Data:    data,
	})
}

// Success 发送成功的响应
func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, 0, "成功", data)
}

// Error 发送业务错误的响应 (HTTP状态码通常是200 OK 或 400 BadRequest)
func Error(c *gin.Context, bizCode int, msg string) {
	// 根据实际情况决定httpCode，例如，参数错误用 http.StatusBadRequest
	// 此处示例为所有业务错误统一使用 http.StatusOK，由前端根据 bizCode 处理
	JSON(c, http.StatusOK, bizCode, msg, nil)
}

// Fail 发送系统级故障的响应 (HTTP状态码应为5xx)
func Fail(c *gin.Context, msg string) {
	JSON(c, http.StatusInternalServerError, 500, msg, nil) // 示例用500作为业务码，可自定义
}
