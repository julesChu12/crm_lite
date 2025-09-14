package impl

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"crm_lite/internal/core/config"
	"crm_lite/internal/core/logger"
	"crm_lite/internal/core/resource"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/spf13/viper"
)

var (
	testRM  *resource.Manager
	cleanup func()
)

func TestMain(m *testing.M) {
	if os.Getenv("RUN_DB_TESTS") != "1" {
		log.Println("RUN_DB_TESTS is not set; skipping integration setup for billing truth tests")
		os.Exit(m.Run())
	}
	if err := setup(); err != nil {
		log.Fatalf("setup failed: %v", err)
	}
	code := m.Run()
	if err := teardown(); err != nil {
		log.Printf("teardown failed: %v", err)
	}
	os.Exit(code)
}

func setup() error {
	log.Println("[billing truth] setup starting...")
	cmd := exec.Command("docker-compose", "-f", "../../../../docker-compose.test.yaml", "up", "-d")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker-compose up: %w", err)
	}
	time.Sleep(10 * time.Second)

	vp := viper.New()
	vp.Set("database.driver", "mysql")
	vp.Set("database.host", "127.0.0.1")
	vp.Set("database.port", 3307)
	vp.Set("database.user", "testuser")
	vp.Set("database.password", "testpassword")
	vp.Set("database.dbname", "crm_test")
	vp.Set("database.debug", true)

	var opts config.Options
	if err := vp.Unmarshal(&opts); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	logger.InitGlobalLogger(&opts.Logger)

	dbRes := resource.NewDBResource(opts.Database)
	if err := dbRes.Initialize(context.Background()); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	sqlDB, err := dbRes.DB.DB()
	if err != nil {
		return fmt.Errorf("unwrap sql.DB: %w", err)
	}
	migrations := &migrate.FileMigrationSource{Dir: "../../../../db/migrations"}
	if _, err := migrate.Exec(sqlDB, "mysql", migrations, migrate.Up); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	testRM = resource.NewManager()
	_ = testRM.Register(resource.DBServiceKey, dbRes)
	cleanup = func() { _ = dbRes.Close(context.Background()) }
	log.Println("[billing truth] setup done.")
	return nil
}

func teardown() error {
	if cleanup != nil {
		cleanup()
	}
	cmd := exec.Command("docker-compose", "-f", "../../../../docker-compose.test.yaml", "down")
	return cmd.Run()
}

func TestTruthService_Credit_Debit_Refund_Idem(t *testing.T) {
	if os.Getenv("RUN_DB_TESTS") != "1" {
		t.Skip("integration only")
	}
	svc, err := NewTruthService(testRM)
	if err != nil {
		t.Fatalf("new truth service: %v", err)
	}

	ctx := context.Background()
	// 准备数据：创建 bil_wallets 记录（手工插入）
	dbRes, _ := resource.Get[*resource.DBResource](testRM, resource.DBServiceKey)
	w := &BilWallet{CustomerID: 777, Balance: 0, Status: 1, UpdatedAt: time.Now().Unix()}
	if err := dbRes.DB.WithContext(ctx).Create(w).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}

	// 1) 充值 1000 分
	if err := svc.Credit(ctx, 777, 1000, "recharge", "idem-1"); err != nil {
		t.Fatalf("credit: %v", err)
	}
	// 2) 幂等重放
	if err := svc.Credit(ctx, 777, 1000, "recharge", "idem-1"); err == nil {
		t.Fatalf("expected idem conflict")
	}
	// 3) 扣减 600 分
	if err := svc.DebitForOrder(ctx, 777, 9001, 600, "idem-2"); err != nil {
		t.Fatalf("debit: %v", err)
	}
	// 4) 退款 200 分
	if err := svc.CreditForRefund(ctx, 777, 9001, 200, "idem-3"); err != nil {
		t.Fatalf("refund: %v", err)
	}

	// 校验余额= 1000 - 600 + 200 = 600
	var bal int64
	if err := dbRes.DB.WithContext(ctx).Model(&BilWallet{}).Where("id = ?", w.ID).Select("balance").Scan(&bal).Error; err != nil {
		t.Fatalf("query bal: %v", err)
	}
	if bal != 600 {
		t.Fatalf("unexpected balance: %d", bal)
	}
}
