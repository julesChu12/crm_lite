// gorm_gen.go for the standalone gorm-gen tool.
package main

import (
	"crm_lite/internal/core/config"
	"flag"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"gorm.io/gen"
)

func main() {
	var env string
	flag.StringVar(&env, "env", "dev", "环境名 (dev/test/prod)")
	flag.Parse()

	fmt.Println("Starting GORM model generation...")

	// 1. 加载配置 - 根据环境选择配置文件
	configFile := fmt.Sprintf("config/app.%s.yaml", env)
	if err := config.InitOptions(configFile); err != nil {
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

	fmt.Printf("Found %d tables in database\n", len(tables))

	// 过滤掉不需要的表
	var filteredTables []string
	for _, table := range tables {
		// 跳过错误的 casbin_rule 单数表，只保留正确的 casbin_rules 复数表
		if table == "casbin_rule" {
			fmt.Printf("Skipping table: %s (using casbin_rules instead)\n", table)
			continue
		}
		filteredTables = append(filteredTables, table)
		fmt.Printf("Including table: %s\n", table)
	}

	fmt.Printf("Generating models and queries for %d filtered tables...\n", len(filteredTables))

	// 为过滤后的表生成模型
	var allModels []interface{}
	for _, tableName := range filteredTables {
		model := g.GenerateModel(tableName)
		allModels = append(allModels, model)
	}

	// 应用基础查询方法
	g.ApplyBasic(allModels...)

	// 5. 执行生成
	g.Execute()

	fmt.Println("GORM model and query generation finished successfully.")
}

// getDSN 根据数据库配置构建DSN字符串
func getDSN(dbOpts config.DBOptions) string {
	switch dbOpts.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbOpts.User, dbOpts.Password, dbOpts.Host, dbOpts.Port, dbOpts.DBName)
	default:
		return ""
	}
}

// ... existing code ...
