package cmd

import (
	"crm_lite/internal/core/config"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "crm_lite",
	Short: "CRM Lite is a lightweight CRM application",
	Long:  `A Fast and Flexible CRM application built with Go, Gin, and MariaDB.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	// 优先级：--config flag > ./.env file (for APP_ENV) > ./config/app.{env}.yaml default path
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path (e.g. --config=config/app.prod.yaml)")
}

// 初始化系统配置
func initConfig() {
	_ = godotenv.Load()

	if configFile == "" {
		env := strings.ToLower(os.Getenv("APP_ENV"))
		if env == "" {
			env = "dev"
		}
		configFile = fmt.Sprintf("./config/app.%s.yaml", env)
	}

	if err := config.InitOptions(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "init config failed: %v\n", err)
		os.Exit(1)
	}
}
