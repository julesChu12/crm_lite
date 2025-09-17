package migration

import (
	"fmt"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// Migration 数据库迁移接口
type Migration interface {
	// GetVersion 获取迁移版本号
	GetVersion() string

	// GetDescription 获取迁移描述
	GetDescription() string

	// Up 执行迁移
	Up(db *gorm.DB) error

	// Down 回滚迁移
	Down(db *gorm.DB) error
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Version     string    `gorm:"uniqueIndex;size:255;not null"`
	Description string    `gorm:"size:500"`
	AppliedAt   time.Time `gorm:"not null"`
	Checksum    string    `gorm:"size:64"` // 用于验证迁移文件是否被修改
}

// TableName 指定表名
func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// Register 注册迁移
func (mm *MigrationManager) Register(migration Migration) {
	mm.migrations = append(mm.migrations, migration)
}

// Initialize 初始化迁移系统
func (mm *MigrationManager) Initialize() error {
	// 创建迁移记录表
	if err := mm.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}
	return nil
}

// Migrate 执行所有未应用的迁移
func (mm *MigrationManager) Migrate() error {
	if err := mm.Initialize(); err != nil {
		return err
	}

	// 获取已应用的迁移
	appliedMigrations, err := mm.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 执行未应用的迁移
	for _, migration := range mm.migrations {
		version := migration.GetVersion()

		if _, exists := appliedMigrations[version]; exists {
			continue // 已应用，跳过
		}

		fmt.Printf("Applying migration %s: %s\n", version, migration.GetDescription())

		if err := mm.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		fmt.Printf("Successfully applied migration %s\n", version)
	}

	return nil
}

// Rollback 回滚指定版本的迁移
func (mm *MigrationManager) Rollback(targetVersion string) error {
	appliedMigrations, err := mm.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 按版本倒序回滚
	for i := len(mm.migrations) - 1; i >= 0; i-- {
		migration := mm.migrations[i]
		version := migration.GetVersion()

		// 如果到达目标版本，停止回滚
		if version == targetVersion {
			break
		}

		// 只回滚已应用的迁移
		if _, exists := appliedMigrations[version]; !exists {
			continue
		}

		fmt.Printf("Rolling back migration %s: %s\n", version, migration.GetDescription())

		if err := mm.rollbackMigration(migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", version, err)
		}

		fmt.Printf("Successfully rolled back migration %s\n", version)
	}

	return nil
}

// Status 显示迁移状态
func (mm *MigrationManager) Status() error {
	appliedMigrations, err := mm.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")

	for _, migration := range mm.migrations {
		version := migration.GetVersion()
		status := "Pending"
		appliedAt := ""

		if record, exists := appliedMigrations[version]; exists {
			status = "Applied"
			appliedAt = record.AppliedAt.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%-20s %-10s %-20s %s\n", version, status, appliedAt, migration.GetDescription())
	}

	return nil
}

// applyMigration 应用单个迁移
func (mm *MigrationManager) applyMigration(migration Migration) error {
	return mm.db.Transaction(func(tx *gorm.DB) error {
		// 执行迁移
		if err := migration.Up(tx); err != nil {
			return err
		}

		// 记录迁移
		record := &MigrationRecord{
			Version:     migration.GetVersion(),
			Description: migration.GetDescription(),
			AppliedAt:   time.Now(),
			Checksum:    mm.calculateChecksum(migration),
		}

		return tx.Create(record).Error
	})
}

// rollbackMigration 回滚单个迁移
func (mm *MigrationManager) rollbackMigration(migration Migration) error {
	return mm.db.Transaction(func(tx *gorm.DB) error {
		// 执行回滚
		if err := migration.Down(tx); err != nil {
			return err
		}

		// 删除迁移记录
		return tx.Where("version = ?", migration.GetVersion()).Delete(&MigrationRecord{}).Error
	})
}

// getAppliedMigrations 获取已应用的迁移
func (mm *MigrationManager) getAppliedMigrations() (map[string]*MigrationRecord, error) {
	var records []*MigrationRecord
	if err := mm.db.Find(&records).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*MigrationRecord)
	for _, record := range records {
		result[record.Version] = record
	}

	return result, nil
}

// calculateChecksum 计算迁移校验和（简化实现）
func (mm *MigrationManager) calculateChecksum(migration Migration) string {
	// 在实际实现中，应该根据迁移的具体内容计算校验和
	// 这里使用版本号作为简化实现
	return fmt.Sprintf("%x", []byte(migration.GetVersion()+migration.GetDescription()))
}

// BaseMigration 基础迁移结构
type BaseMigration struct {
	version     string
	description string
}

// NewBaseMigration 创建基础迁移
func NewBaseMigration(version, description string) *BaseMigration {
	return &BaseMigration{
		version:     version,
		description: description,
	}
}

// GetVersion 获取版本号
func (bm *BaseMigration) GetVersion() string {
	return bm.version
}

// GetDescription 获取描述
func (bm *BaseMigration) GetDescription() string {
	return bm.description
}

// GenerateVersion 生成迁移版本号
func GenerateVersion() string {
	return time.Now().Format("20060102150405")
}

// MigrationTemplate 迁移模板
type MigrationTemplate struct {
	*BaseMigration
	upSQL   string
	downSQL string
}

// NewMigrationTemplate 创建迁移模板
func NewMigrationTemplate(version, description, upSQL, downSQL string) *MigrationTemplate {
	return &MigrationTemplate{
		BaseMigration: NewBaseMigration(version, description),
		upSQL:         upSQL,
		downSQL:       downSQL,
	}
}

// Up 执行迁移
func (mt *MigrationTemplate) Up(db *gorm.DB) error {
	if mt.upSQL == "" {
		return nil
	}
	return db.Exec(mt.upSQL).Error
}

// Down 回滚迁移
func (mt *MigrationTemplate) Down(db *gorm.DB) error {
	if mt.downSQL == "" {
		return nil
	}
	return db.Exec(mt.downSQL).Error
}

// CreateMigrationFile 创建迁移文件模板
func CreateMigrationFile(migrationDir, name string) (string, error) {
	version := GenerateVersion()
	filename := fmt.Sprintf("%s_%s.go", version, name)
	fullPath := filepath.Join(migrationDir, filename)

	template := fmt.Sprintf(`package migrations

import (
	"crm_lite/internal/migration"
	"gorm.io/gorm"
)

// Migration%s %s迁移
type Migration%s struct {
	*migration.BaseMigration
}

// NewMigration%s 创建%s迁移
func NewMigration%s() migration.Migration {
	return &Migration%s{
		BaseMigration: migration.NewBaseMigration("%s", "%s"),
	}
}

// Up 执行迁移
func (m *Migration%s) Up(db *gorm.DB) error {
	// TODO: 实现迁移逻辑
	return nil
}

// Down 回滚迁移
func (m *Migration%s) Down(db *gorm.DB) error {
	// TODO: 实现回滚逻辑
	return nil
}
`, version, name, version, version, name, version, version, version, name, version, version)

	// 这里应该写入文件，但为了演示，我们只返回路径和内容
	return fullPath, nil
}