// Package analytics 数据分析域服务接口
// 职责：业务数据分析、仪表盘报告、层级关系查询、统计报表
// 核心原则：数据驱动决策、实时分析、跨域聚合
package analytics

import "context"

// Overview 业务总览
type Overview struct {
	TotalCustomers   int64   `json:"total_customers"`   // 总客户数
	TotalOrders      int64   `json:"total_orders"`      // 总订单数
	TotalProducts    int64   `json:"total_products"`    // 总产品数
	TotalRevenue     float64 `json:"total_revenue"`     // 总收入
	MonthlyCustomers int64   `json:"monthly_customers"` // 当月新增客户
	MonthlyOrders    int64   `json:"monthly_orders"`    // 当月订单数
	MonthlyRevenue   float64 `json:"monthly_revenue"`   // 当月收入
	GrowthRate       float64 `json:"growth_rate"`       // 增长率
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Time  string  `json:"time"`  // 时间标签
	Value float64 `json:"value"` // 数值
}

// SalesAnalysis 销售分析
type SalesAnalysis struct {
	TotalRevenue   float64          `json:"total_revenue"`   // 总收入
	OrderCount     int64            `json:"order_count"`     // 订单数量
	AverageValue   float64          `json:"average_value"`   // 平均订单价值
	TopProducts    []ProductSales   `json:"top_products"`    // 热销产品
	SalesTrend     []TimeSeriesData `json:"sales_trend"`     // 销售趋势
	ConversionRate float64          `json:"conversion_rate"` // 转化率
}

// ProductSales 产品销售统计
type ProductSales struct {
	ProductID   int64   `json:"product_id"`   // 产品ID
	ProductName string  `json:"product_name"` // 产品名称
	SalesCount  int64   `json:"sales_count"`  // 销售数量
	Revenue     float64 `json:"revenue"`      // 销售收入
	Percentage  float64 `json:"percentage"`   // 占比
}

// CustomerAnalysis 客户分析
type CustomerAnalysis struct {
	TotalCustomers   int64             `json:"total_customers"`   // 总客户数
	ActiveCustomers  int64             `json:"active_customers"`  // 活跃客户数
	NewCustomers     int64             `json:"new_customers"`     // 新客户数
	CustomerSegments []CustomerSegment `json:"customer_segments"` // 客户分层
	GrowthTrend      []TimeSeriesData  `json:"growth_trend"`      // 增长趋势
	RetentionRate    float64           `json:"retention_rate"`    // 客户留存率
}

// CustomerSegment 客户分层
type CustomerSegment struct {
	Level      string  `json:"level"`      // 等级：金牌/银牌/普通
	Count      int64   `json:"count"`      // 数量
	Revenue    float64 `json:"revenue"`    // 贡献收入
	Percentage float64 `json:"percentage"` // 占比
}

// HierarchyNode 层级节点
type HierarchyNode struct {
	ID       int64  `json:"id"`        // 节点ID
	Name     string `json:"name"`      // 节点名称
	Level    int    `json:"level"`     // 层级深度
	ParentID int64  `json:"parent_id"` // 父节点ID
}

// HierarchyTree 层级树
type HierarchyTree struct {
	Root     *HierarchyNode   `json:"root"`     // 根节点
	Children []*HierarchyTree `json:"children"` // 子节点列表
}

// DashboardService 仪表盘服务接口
// 提供业务总览和核心指标分析
type DashboardService interface {
	// GetOverview 获取业务总览
	GetOverview(ctx context.Context) (*Overview, error)

	// GetSalesAnalysis 获取销售分析
	GetSalesAnalysis(ctx context.Context, days int) (*SalesAnalysis, error)

	// GetCustomerAnalysis 获取客户分析
	GetCustomerAnalysis(ctx context.Context, days int) (*CustomerAnalysis, error)

	// GetRevenueTree 获取收入趋势
	GetRevenueTrend(ctx context.Context, days int) ([]TimeSeriesData, error)

	// GetTopPerformers 获取业绩排行
	GetTopPerformers(ctx context.Context, metric string, limit int) ([]interface{}, error)
}

// HierarchyService 层级关系服务接口
// 提供组织架构和上下级关系查询
type HierarchyService interface {
	// GetSubordinates 获取下属列表
	// 返回指定管理者的所有下属ID，支持多级查询
	GetSubordinates(ctx context.Context, managerID int64) ([]int64, error)

	// GetSuperiors 获取上级列表
	// 返回指定员工的所有上级ID，支持多级查询
	GetSuperiors(ctx context.Context, employeeID int64) ([]int64, error)

	// GetHierarchyTree 获取层级树
	// 返回以指定节点为根的完整层级结构
	GetHierarchyTree(ctx context.Context, rootID int64) (*HierarchyTree, error)

	// GetDirectReports 获取直接下属
	// 只返回一级直接下属，不递归查询
	GetDirectReports(ctx context.Context, managerID int64) ([]int64, error)

	// IsSubordinate 检查下属关系
	// 检查 employeeID 是否是 managerID 的下属（支持多级）
	IsSubordinate(ctx context.Context, managerID, employeeID int64) (bool, error)

	// RefreshCache 刷新层级缓存
	// 当组织架构发生变更时调用，重新计算和缓存层级关系
	RefreshCache(ctx context.Context) error
}

// ReportService 报表服务接口
// 提供各类业务报表和数据导出
type ReportService interface {
	// GenerateCustomerReport 生成客户报表
	GenerateCustomerReport(ctx context.Context, startTime, endTime string) ([]byte, error)

	// GenerateSalesReport 生成销售报表
	GenerateSalesReport(ctx context.Context, startTime, endTime string) ([]byte, error)

	// GenerateMarketingReport 生成营销报表
	GenerateMarketingReport(ctx context.Context, startTime, endTime string) ([]byte, error)

	// ExportData 导出业务数据
	ExportData(ctx context.Context, dataType string, format string) ([]byte, error)

	// ScheduleReport 定时报表
	ScheduleReport(ctx context.Context, reportType string, schedule string, recipients []string) error
}

// Service 数据分析域统一服务接口
// 整合仪表盘、层级关系、报表生成的完整功能
type Service interface {
	DashboardService
	HierarchyService
	ReportService
}
