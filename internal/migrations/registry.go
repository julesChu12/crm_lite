package migrations

import (
	"crm_lite/internal/migration"
)

// RegisterAllMigrations 注册所有迁移
func RegisterAllMigrations(manager *migration.MigrationManager) {
	// 性能优化相关迁移
	manager.Register(NewMigration20250916001_AddCustomerIndexes())
	manager.Register(NewMigration20250916002_AddOrderIndexes())
	manager.Register(NewMigration20250916003_AddProductIndexes())

	// 可以在这里添加更多迁移...
}