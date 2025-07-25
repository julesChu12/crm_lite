package cmd

import (
	"fmt"
	"os"

	"crm_lite/internal/bootstrap"
	"crm_lite/internal/routes"
	"crm_lite/internal/startup"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the crm_lite server",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 引导程序, 初始化资源
		resManager, logCleaner, cleanup, err := bootstrap.Bootstrap()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to bootstrap application: %v\n", err)
			os.Exit(1)
		}

		// 2. 初始化路由，传入资源管理器和日志清理器
		router := routes.NewRouter(resManager, logCleaner)

		// 3. 启动服务
		startup.Start(router, cleanup)
	},
}
