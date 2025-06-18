package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	migrate "github.com/rubenv/sql-migrate"
	"gopkg.in/yaml.v2"
)

type DBConfig struct {
	Dialect    string `yaml:"dialect"`
	Datasource string `yaml:"datasource"`
	Dir        string `yaml:"dir"`
	Table      string `yaml:"table"`
}

func loadConfig(env string) (*DBConfig, error) {
	data, err := os.ReadFile("db/dbconfig.yml")
	if err != nil {
		return nil, err
	}
	var raw map[string]DBConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	cfg, ok := raw[env]
	if !ok {
		return nil, fmt.Errorf("env %s not found in dbconfig.yml", env)
	}

	// 展开 datasource 中的环境变量
	cfg.Datasource = os.ExpandEnv(cfg.Datasource)

	return &cfg, nil
}

func main() {
	var env string
	var steps int
	var direction string

	flag.StringVar(&env, "env", "development", "环境名，对应 dbconfig.yml")
	flag.StringVar(&direction, "direction", "up", "迁移方向 up/down")
	flag.IntVar(&steps, "steps", 0, "执行多少步（0=全部）")
	flag.Parse()

	cfg, err := loadConfig(env)
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: cfg.Dir,
	}

	db, err := sql.Open(cfg.Dialect, cfg.Datasource)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 设置迁移表名
	migrate.SetTable(cfg.Table)

	var n int
	if direction == "up" {
		n, err = migrate.ExecMax(db, cfg.Dialect, migrations, migrate.Up, steps)
	} else {
		n, err = migrate.ExecMax(db, cfg.Dialect, migrations, migrate.Down, steps)
	}
	if err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	fmt.Printf("迁移完成，执行了 %d 步\n", n)
}
