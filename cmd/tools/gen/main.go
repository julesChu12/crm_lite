// main.go for the standalone generator tool.
package main

import (
	"crm_lite/internal/core/config"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gorm.io/gen"
)

func main() {
	fmt.Println("Starting GORM model generation...")

	// 1. 加载配置 - 假设从项目根目录运行
	if err := config.InitOptions("config/app.dev.yaml"); err != nil {
		fmt.Printf("Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}
	opts := config.GetInstance()
	dbOpts := opts.Database

	// 2. 连接数据库
	var db *gorm.DB
	var err error
	dsn := getDSN(dbOpts)

	switch dbOpts.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(dsn))
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn))
	default:
		fmt.Printf("Unsupported database driver: %s\n", dbOpts.Driver)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully connected to the database.")

	// 初始化生成器 - 专门用于生成查询文件
	g := gen.NewGenerator(gen.Config{
		OutPath:           "./internal/dao/query",
		ModelPkgPath:      "./internal/dao/model",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		FieldNullable:     false,
		FieldCoverable:    false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})
	g.UseDB(db)

	// 获取所有表并生成模型和查询
	tables, err := db.Migrator().GetTables()
	if err != nil {
		fmt.Printf("Failed to get tables: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generating models and queries for %d tables...\n", len(tables))

	// 使用 GenerateAllTable 生成所有表的模型和查询
	allModels := g.GenerateAllTable()

	// 应用基础查询方法
	g.ApplyBasic(allModels...)

	// 5. 执行生成
	g.Execute()

	fmt.Println("GORM model and query generation finished successfully.")
}

// getDSN 根据数据库配置构建DSN字符串
func getDSN(dbOpts config.DBOptions) string {
	switch dbOpts.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			dbOpts.Host, dbOpts.User, dbOpts.Password, dbOpts.DBName, dbOpts.Port, dbOpts.SSLMode, dbOpts.TimeZone)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbOpts.User, dbOpts.Password, dbOpts.Host, dbOpts.Port, dbOpts.DBName)
	default:
		return ""
	}
}
