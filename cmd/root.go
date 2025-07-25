package cmd

import (
	"crm_lite/internal/core/config"
	"fmt"
	"os"
	"path/filepath"
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
	// 优先级：--config flag > ./.env file (for APP_ENV) > ./config/app.{env}.yaml default path
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path (e.g. --config=config/app.prod.yaml)")
	cobra.OnInitialize(initConfig)
}

// 初始化系统配置
func initConfig() {
	// 优先从 .env 文件加载环境变量 (例如 APP_ENV)
	envFile, err := findConfigInProject("app.env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot locate app.env: %v\n", err)
	} else if err := godotenv.Load(envFile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load app.env: %v\n", err)
	}

	// 只有当 --config flag 未被设置时，我们才进行自动查找
	if configFile == "" {
		env := "dev"
		if loadedEnv := strings.ToLower(os.Getenv("ENV")); loadedEnv != "" {
			env = loadedEnv
		}
		configName := fmt.Sprintf("app.%s.yaml", env)
		// 优先级 1: 从 PROJECT_ROOT 环境变量指定的路径
		if projectRoot := os.Getenv("PROJECT_ROOT"); projectRoot != "" {
			path := filepath.Join(projectRoot, "config", configName)
			if _, err := os.Stat(path); err == nil {
				configFile = path
			}
		}

		// 优先级 2: 从容器/生产环境的约定路径
		if configFile == "" {
			containerConfigPath := fmt.Sprintf("/app/config/%s", configName)
			if _, err := os.Stat(containerConfigPath); err == nil {
				configFile = containerConfigPath
			}
		}

		// 优先级 3: 从项目根目录 (通过 go.mod 动态查找, 用于本地开发)
		if configFile == "" {
			projectConfigPath, err := findConfigInProject(configName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "config file not found in standard paths: %v\n", err)
				os.Exit(1)
			}
			configFile = projectConfigPath
		}
	}

	if err := config.InitOptions(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "init config failed: %v\n", err)
		os.Exit(1)
	}
}

// findConfigInProject 动态查找并返回项目内的配置文件绝对路径 (主要用于本地开发)
func findConfigInProject(configName string) (string, error) {
	// 尝试从当前工作目录向上查找 go.mod
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var projectRoot string
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir { // 到达文件系统根部
			return "", fmt.Errorf("cannot find go.mod in current directory or any parent")
		}
		dir = parent
	}

	// 构建配置文件的绝对路径
	configFilePath := filepath.Join(projectRoot, "config", configName)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("config file not found at %s", configFilePath)
	}

	return configFilePath, nil
}
