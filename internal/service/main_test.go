package service

import (
	"context"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
)

// 全局变量，用于在测试套件之间共享资源
var (
	testResManager *resource.Manager
	testCleanup    func()
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	code := m.Run()

	if err := teardown(); err != nil {
		log.Printf("Failed to teardown test environment: %v", err)
	}

	os.Exit(code)
}

func setup() error {
	log.Println("Setting up test environment...")
	cmd := exec.Command("docker-compose", "-f", "../../docker-compose.test.yaml", "up", "-d")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not start docker-compose containers: %w", err)
	}
	time.Sleep(10 * time.Second) // 等待数据库就绪

	// ------------------ 手动初始化配置和资源 ------------------
	// 1. 创建一个新的 Viper 实例用于测试
	vp := viper.New()
	overrideConfigForTest(vp)

	// 2. 将配置绑定到结构体
	var opts config.Options
	if err := vp.Unmarshal(&opts); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 3. 初始化全局 Logger (必须在初始化其他资源之前)
	logger.InitGlobalLogger(&opts.Logger)

	// 4. 手动初始化资源
	dbResource := resource.NewDBResource(opts.Database)
	if err := dbResource.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize db resource: %w", err)
	}

	// 5. 运行数据库迁移
	sqlDB, err := dbResource.DB.DB()
	if err != nil {
		return fmt.Errorf("could not get underlying sql.DB: %w", err)
	}
	migrations := &migrate.FileMigrationSource{Dir: "../../db/migrations"}
	if _, err := migrate.Exec(sqlDB, "mysql", migrations, migrate.Up); err != nil {
		return fmt.Errorf("migrations failed: %w", err)
	}

	// 6. 将初始化好的资源保存到全局变量
	testResManager = resource.NewManager()
	testResManager.Register(resource.DBServiceKey, dbResource)
	// 如果需要，也可以在这里注册其他资源，如 Cache
	// cacheResource := resource.NewCacheResource(opts.Cache)
	// cacheResource.Initialize(context.Background())
	// testResManager.Register(resource.CacheServiceKey, cacheResource)

	testCleanup = func() {
		dbResource.Close(context.Background())
	}

	log.Println("Test environment setup complete.")
	return nil
}

func teardown() error {
	log.Println("Tearing down test environment...")
	if testCleanup != nil {
		testCleanup()
	}
	cmd := exec.Command("docker-compose", "-f", "../../docker-compose.test.yaml", "down")
	return cmd.Run()
}

func overrideConfigForTest(vp *viper.Viper) {
	vp.Set("database.driver", "mysql")
	vp.Set("database.host", "127.0.0.1")
	vp.Set("database.port", 3307)
	vp.Set("database.user", "testuser")
	vp.Set("database.password", "testpassword")
	vp.Set("database.dbname", "crm_test")
	vp.Set("database.debug", true)
	vp.Set("cache.redis.host", "127.0.0.1")
	vp.Set("cache.redis.port", 6380)
}
