package errcode

// Error 定义了自定义错误类型，包含业务码和错误信息
type Error struct {
	Code    int
	Message string
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.Message
}

// New 创建一个新的 Error 实例
func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// 预定义的常用错误码
var (
	Success          = New(0, "成功")
	ServerError      = New(1001, "服务器内部错误")     // 通常对应HTTP 500
	InvalidParams    = New(1002, "参数无效")        // 通常对应HTTP 400
	NotFound         = New(1003, "资源不存在")       // 通常对应HTTP 404
	Unauthorized     = New(1004, "未授权或Token无效") // 通常对应HTTP 401
	PermissionDenied = New(1005, "无权限访问")       // 通常对应HTTP 403
	TooManyRequests  = New(1006, "请求过于频繁")      // 通常对应HTTP 429
)

// IsCode 判断一个error是否是指定的业务错误码
func IsCode(err error, code int) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}
