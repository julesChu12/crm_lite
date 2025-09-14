package impl

import (
	"crm_lite/internal/dao/query"
	"crm_lite/internal/domains/catalog"
)

// NewCatalogService 创建 catalog 服务实例
func NewCatalogService(q *query.Query) catalog.Service {
	return New(q)
}
