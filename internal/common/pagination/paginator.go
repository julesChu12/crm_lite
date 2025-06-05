package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
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
	pageStr := c.DefaultQuery("page", strconv.Itoa(DefaultPage))
	pageSizeStr := c.DefaultQuery("pageSize", strconv.Itoa(DefaultPageSize))

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = DefaultPage
	}

	pageSize, err = strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
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
