# 🧰 通用模块设计文档（Common Module） - CRM Lite

## 📌 模块目标

通用模块汇集系统中多个业务模块共用的功能组件，包括日志管理、统一响应格式、错误处理、分页工具、配置加载等内容，目的是提高代码复用性、统一处理标准与系统健壮性。

---

## 📦 模块组成说明

### 响应封装 (response)

- 作用：标准化所有 API 接口输出格式。
- 功能：提供快捷响应函数（如：Success、Error、Fail）。

示例结构：

```go
type Response struct {
  Code    int         `json:"code"`
  Message string      `json:"message"`
  Data    interface{} `json:"data,omitempty"`
}
```

### 错误处理 (errcode)

- 作用：自定义错误码与错误对象。
- 功能：统一错误码定义，方便维护和问题定位。

示例：

```go
// 错误码定义示例 (errcode.go)
var (
  Success        = New(0, "成功")
  ServerError    = New(1001, "内部服务器错误")
  InvalidParams  = New(1002, "参数错误")
  NotFound       = New(1003, "资源不存在")
)
```

### 日志 (logger)

- 作用：使用 `zap` 实现统一的、高性能的日志记录。
- 功能：支持按模块、级别记录，建议配置日志切割、轮转等。

### 分页工具 (pagination)

- 作用：封装常用的分页结构体与计算逻辑。
- 功能：支持 Page、PageSize 输入，Total 输出，方便列表接口使用。

### 时间工具 (timeutil)

- 作用：封装时间格式化、解析、UTC转换等常用操作。
- 功能：提供统一的时间处理函数，避免重复编码和格式不一致问题。

---

## 📂 建议的目录结构

`internal/common/`
├── `response/`
│   └── `response.go`
├── `errcode/`
│   └── `errcode.go`
├── `logger/`
│   └── `logger.go`
├── `pagination/`
│   └── `paginator.go`
├── `timeutil/`
│   └── `datetime.go`

---

## 💡 通用组件设计建议

- **无状态性**：所有通用工具函数和方法应尽量设计为无状态的（pure function），不依赖外部可变状态，易于测试和复用。
- **命名空间**：公共错误码应避免与具体业务模块的错误码冲突，建议定义独立的命名空间或使用统一前缀。
- **可配置性**：对于如日志、时间格式等，应考虑通过配置进行管理，而不是硬编码。
- **多语言支持**：对于API响应消息，可以考虑支持多语言（i18n），例如通过加载不同语言的资源文件实现。

---

## ⚙️ 代码实现参考

以下是 CRM Lite 项目通用模块部分组件的骨架代码参考，可放置于 `internal/common` 目录下对应的文件中。

### `internal/common/response/response.go`

```go
package response

import (
 "github.com/gin-gonic/gin"
 "net/http"
)

// Response 定义了标准的API响应结构
type Response struct {
 Code    int         `json:"code"`
 Message string      `json:"message"`
 Data    interface{} `json:"data,omitempty"`
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
```

---

### `internal/common/errcode/errcode.go`

```go
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
 Success         = New(0, "成功")
 ServerError     = New(1001, "服务器内部错误") // 通常对应HTTP 500
 InvalidParams   = New(1002, "参数无效")       // 通常对应HTTP 400
 NotFound        = New(1003, "资源不存在")     // 通常对应HTTP 404
 Unauthorized    = New(1004, "未授权或Token无效") // 通常对应HTTP 401
 PermissionDenied = New(1005, "无权限访问")     // 通常对应HTTP 403
 TooManyRequests = New(1006, "请求过于频繁")   // 通常对应HTTP 429
)

// IsCode 判断一个error是否是指定的业务错误码
func IsCode(err error, code int) bool {
 if e, ok := err.(*Error); ok {
  return e.Code == code
 }
 return false
}
```

---

### `internal/common/logger/logger.go`

```go
package logger

import (
 "go.uber.org/zap"
 "go.uber.org/zap/zapcore"
 "os"
)

var Log *zap.SugaredLogger

// Init 初始化 zap 日志记录器
// level: 日志级别 (debug, info, warn, error, panic, fatal)
// filePath: 日志文件路径，如果为空则输出到控制台
func Init(level string, filePath string) {
 var zapLevel zapcore.Level
 err := zapLevel.UnmarshalText([]byte(level))
 if err != nil {
  zapLevel = zapcore.InfoLevel // 默认级别
 }

 encoderConfig := zap.NewProductionEncoderConfig()
 encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
 encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

 var core zapcore.Core
 if filePath != "" {
  fileWriter := zapcore.AddSync(&zapcore.BufferedWriteSyncer{
   WS: zapcore.AddSync(mustOpenFile(filePath)),
   //BufferSize: 256 * 1024, // 256KB
   //FlushInterval: 30 * time.Second,
  })
  core = zapcore.NewCore(
   zapcore.NewJSONEncoder(encoderConfig),
   fileWriter,
   zapLevel,
  )
 } else {
  core = zapcore.NewCore(
   zapcore.NewConsoleEncoder(encoderConfig),
   zapcore.Lock(os.Stdout),
   zapLevel,
  )
 }

 rawLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)) // AddCallerSkip(1) 以便封装后调用栈正确
 Log = rawLogger.Sugar()
 Log.Info("Logger initialized", "level", zapLevel.String(), "filePath", filePath)
}

func mustOpenFile(filePath string) *os.File {
 file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
 if err != nil {
  panic("Failed to open log file: " + err.Error())
 }
 return file
}

// Sync flushes any buffered log entries.
func Sync() {
 if Log != nil {
  _ = Log.Sync()
 }
}
```

---

### `internal/common/pagination/paginator.go`

```go
package pagination

import (
 "github.com/gin-gonic/gin"
 "strconv"
)

const (
 DefaultPage     = 1
 DefaultPageSize = 10
 MaxPageSize     = 100
)

// PagerInfo 包含分页结果的详细信息
type PagerInfo struct {
 Page     int   `json:"page"`      // 当前页码
 PageSize int   `json:"page_size"` // 每页数量
 Total    int64 `json:"total"`     // 总记录数
 // TotalPages int `json:"total_pages"` // 总页数 (可选)
}

// Request 分页请求参数
type Request struct {
 Page     int `form:"page"`
 PageSize int `form:"pageSize"`
}

// GetPage 从 Gin Context 获取分页参数
func GetPage(c *gin.Context) (page int, pageSize int) {
 page, _ = strconv.Atoi(c.Query("page"))
 pageSize, _ = strconv.Atoi(c.Query("pageSize"))

 if page <= 0 {
  page = DefaultPage
 }
 if pageSize <= 0 {
  pageSize = DefaultPageSize
 }
 if pageSize > MaxPageSize {
  pageSize = MaxPageSize
 }
 return page, pageSize
}

// New 创建分页结果信息
func New(page, pageSize int, total int64) *PagerInfo {
 return &PagerInfo{
  Page:     page,
  PageSize: pageSize,
  Total:    total,
  // TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
 }
}

// GetOffset 计算数据库查询的偏移量
func (r *Request) GetOffset() int {
 p := r.Page
 ps := r.PageSize
 if p <= 0 {
  p = DefaultPage
 }
 if ps <= 0 {
  ps = DefaultPageSize
 }
 return (p - 1) * ps
}
```

---

### `internal/common/timeutil/datetime.go`

```go
package timeutil

import (
 "time"
)

const (
 DefaultFormat = "2006-01-02 15:04:05"
 DateFormat    = "2006-01-02"
)

// FormatTime 将 time.Time 格式化为标准字符串 (YYYY-MM-DD HH:MM:SS)
func FormatTime(t time.Time) string {
 return t.Format(DefaultFormat)
}

// FormatDate 将 time.Time 格式化为日期字符串 (YYYY-MM-DD)
func FormatDate(t time.Time) string {
 return t.Format(DateFormat)
}

// ParseTime 将标准时间字符串解析为 time.Time
func ParseTime(s string) (time.Time, error) {
 return time.ParseInLocation(DefaultFormat, s, time.Local) // 使用本地时区
}

// ParseDate 将日期字符串解析为 time.Time
func ParseDate(s string) (time.Time, error) {
 return time.ParseInLocation(DateFormat, s, time.Local)
}

// GetCurrentTime 获取当前时间字符串
func GetCurrentTime() string {
 return FormatTime(time.Now())
}

// GetCurrentDate 获取当前日期字符串
func GetCurrentDate() string {
 return FormatDate(time.Now())
}
```

---
