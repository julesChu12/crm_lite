package migrations

import (
	"crm_lite/internal/migration"

	"gorm.io/gorm"
)

// Migration20250916001_AddCustomerIndexes 添加客户表索引
type Migration20250916001_AddCustomerIndexes struct {
	*migration.BaseMigration
}

// NewMigration20250916001_AddCustomerIndexes 创建迁移
func NewMigration20250916001_AddCustomerIndexes() migration.Migration {
	return &Migration20250916001_AddCustomerIndexes{
		BaseMigration: migration.NewBaseMigration("20250916001", "Add performance indexes for customers table"),
	}
}

// Up 执行迁移
func (m *Migration20250916001_AddCustomerIndexes) Up(db *gorm.DB) error {
	sqls := []string{
		"CREATE INDEX idx_customers_name ON customers(name)",
		"CREATE INDEX idx_customers_source_created ON customers(source, created_at)",
		"CREATE INDEX idx_customers_level_active ON customers(level, deleted_at)",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

// Down 回滚迁移
func (m *Migration20250916001_AddCustomerIndexes) Down(db *gorm.DB) error {
	sqls := []string{
		"DROP INDEX idx_customers_name ON customers",
		"DROP INDEX idx_customers_source_created ON customers",
		"DROP INDEX idx_customers_level_active ON customers",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

// Migration20250916002_AddOrderIndexes 添加订单表索引
type Migration20250916002_AddOrderIndexes struct {
	*migration.BaseMigration
}

// NewMigration20250916002_AddOrderIndexes 创建迁移
func NewMigration20250916002_AddOrderIndexes() migration.Migration {
	return &Migration20250916002_AddOrderIndexes{
		BaseMigration: migration.NewBaseMigration("20250916002", "Add performance indexes for orders table"),
	}
}

// Up 执行迁移
func (m *Migration20250916002_AddOrderIndexes) Up(db *gorm.DB) error {
	sqls := []string{
		"CREATE INDEX idx_orders_status_created ON orders(status, created_at)",
		"CREATE INDEX idx_orders_payment_status ON orders(payment_status, created_at)",
		"CREATE INDEX idx_orders_amount_range ON orders(final_amount, created_at)",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

// Down 回滚迁移
func (m *Migration20250916002_AddOrderIndexes) Down(db *gorm.DB) error {
	sqls := []string{
		"DROP INDEX idx_orders_status_created ON orders",
		"DROP INDEX idx_orders_payment_status ON orders",
		"DROP INDEX idx_orders_amount_range ON orders",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

// Migration20250916003_AddProductIndexes 添加产品表索引
type Migration20250916003_AddProductIndexes struct {
	*migration.BaseMigration
}

// NewMigration20250916003_AddProductIndexes 创建迁移
func NewMigration20250916003_AddProductIndexes() migration.Migration {
	return &Migration20250916003_AddProductIndexes{
		BaseMigration: migration.NewBaseMigration("20250916003", "Add performance indexes for products table"),
	}
}

// Up 执行迁移
func (m *Migration20250916003_AddProductIndexes) Up(db *gorm.DB) error {
	sqls := []string{
		"CREATE INDEX idx_products_category_active ON products(category, is_active)",
		"CREATE INDEX idx_products_price_range ON products(price, is_active)",
		"CREATE INDEX idx_products_stock_level ON products(stock_quantity, min_stock_level)",
		"CREATE INDEX idx_products_type_active ON products(type, is_active)",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

// Down 回滚迁移
func (m *Migration20250916003_AddProductIndexes) Down(db *gorm.DB) error {
	sqls := []string{
		"DROP INDEX idx_products_category_active ON products",
		"DROP INDEX idx_products_price_range ON products",
		"DROP INDEX idx_products_stock_level ON products",
		"DROP INDEX idx_products_type_active ON products",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}