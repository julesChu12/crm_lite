package cmd

import (
	"crm_lite/pkg/config"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "crm_lite",
	Short: "CRM Lite is a lightweight CRM application",
	Long:  `A Fast and Flexible CRM application built with Go, Gin, and PostgreSQL.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	//优先级：--config 参数 > CRM_CONFIG 环境变量 > ./config/app.{env}.yaml
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file")
}

// 初始化系统配置
func initConfig() {
	if strings.TrimSpace(configFile) == "" {
		configFile = os.Getenv("CRM_CONFIG")
	}
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
