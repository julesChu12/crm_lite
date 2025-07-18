package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 定义统一的响应结构
// Code 为业务码，0 表示成功，其它非 0 表示错误
// Message 为人类可读信息
// Data 为实际返回数据，可选
// 我们始终返回 HTTP 200，由业务码区分是否成功，便于前端统一处理
// 如需严格遵守 HTTP 语义，可在需要时调整 statusCode 参数
// 但建议业务错误使用 200，系统错误使用 500 等。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 业务码常量
const (
	CodeSuccess       = 0
	CodeCreated       = 201
	CodeNoContent     = 204
	CodeInvalidParam  = 4001
	CodeUnauthorized  = 4010
	CodeForbidden     = 4030
	CodeNotFound      = 4040
	CodeConflict      = 4090 // 资源冲突，例如用户已存在
	CodeInternalError = 5000
)

// Success 统一成功返回
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithCode 成功返回，但使用指定的HTTP状态码
func SuccessWithCode(c *gin.Context, httpCode int, data interface{}) {
	// 对于 204 No Content，不应有响应体
	if httpCode == http.StatusNoContent {
		c.Status(httpCode)
		return
	}
	c.JSON(httpCode, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 统一错误返回，允许自定义业务码和信息
func Error(c *gin.Context, code int, msg string) {
	httpCode := http.StatusOK
	// 根据业务码决定HTTP状态码
	switch code {
	case CodeConflict:
		httpCode = http.StatusConflict
	case CodeNotFound:
		httpCode = http.StatusNotFound
	case CodeUnauthorized:
		httpCode = http.StatusUnauthorized
	case CodeForbidden:
		httpCode = http.StatusForbidden
	case CodeInvalidParam:
		httpCode = http.StatusBadRequest
	}

	c.JSON(httpCode, Response{
		Code:    code,
		Message: msg,
	})
}

// SystemError 系统级错误（HTTP 500）
func SystemError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeInternalError,
		Message: err.Error(),
	})
}
