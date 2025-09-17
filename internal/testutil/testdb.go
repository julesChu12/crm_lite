package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	testContainer  *mysql.MySQLContainer
	testDB         *gorm.DB
	containerOnce  sync.Once
	containerError error
)

// SetupTestDatabase 设置测试数据库容器
func SetupTestDatabase() (*gorm.DB, func(), error) {
	containerOnce.Do(func() {
		ctx := context.Background()

		// 启动MySQL容器
		container, err := mysql.RunContainer(ctx,
			testcontainers.WithImage("mysql:8.0"),
			mysql.WithDatabase("crm_test"),
			mysql.WithUsername("root"),
			mysql.WithPassword("testpass"),
		)
		if err != nil {
			containerError = fmt.Errorf("failed to start MySQL container: %w", err)
			return
		}

		testContainer = container

		// 等待容器启动
		time.Sleep(5 * time.Second)

		// 获取连接信息
		host, err := container.Host(ctx)
		if err != nil {
			containerError = fmt.Errorf("failed to get container host: %w", err)
			return
		}

		port, err := container.MappedPort(ctx, "3306")
		if err != nil {
			containerError = fmt.Errorf("failed to get container port: %w", err)
			return
		}

		// 构建DSN
		dsn := fmt.Sprintf("root:testpass@tcp(%s:%s)/crm_test?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
			host, port.Port())

		// 等待数据库就绪
		var db *sql.DB
		for i := 0; i < 30; i++ {
			db, err = sql.Open("mysql", dsn)
			if err == nil {
				if err = db.Ping(); err == nil {
					break
				}
			}
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			containerError = fmt.Errorf("failed to connect to test database: %w", err)
			return
		}
		db.Close()

		// 创建GORM连接
		testDB, err = gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			containerError = fmt.Errorf("failed to create GORM connection: %w", err)
			return
		}

		log.Println("Test database container is ready")
	})

	if containerError != nil {
		return nil, nil, containerError
	}

	// 返回清理函数
	cleanup := func() {
		if testContainer != nil {
			ctx := context.Background()
			if err := testContainer.Terminate(ctx); err != nil {
				log.Printf("Failed to terminate test container: %v", err)
			}
		}
	}

	return testDB, cleanup, nil
}

// GetTestDB 获取测试数据库连接（复用容器）
func GetTestDB() (*gorm.DB, error) {
	if testDB == nil {
		return nil, fmt.Errorf("test database not initialized, call SetupTestDatabase first")
	}
	return testDB, nil
}

// CleanupTestData 清理测试数据
func CleanupTestData(db *gorm.DB, tables ...string) error {
	// 禁用外键约束检查
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// 清理指定表
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// 重新启用外键约束检查
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
		return fmt.Errorf("failed to enable foreign key checks: %w", err)
	}

	return nil
}