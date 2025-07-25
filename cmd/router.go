package cmd

import (
	"crm_lite/internal/bootstrap"
	"crm_lite/internal/routes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// 列出项目中所有路由

func init() {
	rootCmd.AddCommand(routerListCmd)
}

var routerListCmd = &cobra.Command{
	Use:   "router:list",
	Short: "List all registered routes",
	Long:  `Display all registered routes in the application, similar to Laravel's route:list command`,
	Run:   listRoutes,
}

func listRoutes(cmd *cobra.Command, args []string) {
	// 初始化资源管理器（仅用于路由列表，不需要完整启动）
	resManager, logCleaner, cleanup, err := bootstrap.Bootstrap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bootstrap application: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// 初始化 Gin 引擎并注册路由
	r := routes.NewRouter(resManager, logCleaner)

	// 获取所有路由
	routesInfo := r.Routes()

	// 转换为自定义结构以便排序和格式化
	type routeInfo struct {
		Method  string
		Path    string
		Handler string
	}

	routeInfos := make([]routeInfo, 0, len(routesInfo))
	for _, route := range routesInfo {
		routeInfos = append(routeInfos, routeInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		})
	}

	// 按路径排序
	sort.Slice(routeInfos, func(i, j int) bool {
		return routeInfos[i].Path < routeInfos[j].Path
	})

	// 计算列宽
	methodWidth := 8 // 默认最小宽度
	pathWidth := 20  // 默认最小宽度
	for _, route := range routeInfos {
		if len(route.Method) > methodWidth {
			methodWidth = len(route.Method)
		}
		if len(route.Path) > pathWidth {
			pathWidth = len(route.Path)
		}
	}

	// 打印表头
	headerFormat := fmt.Sprintf("%%-%ds | %%-%ds | %%s\n", methodWidth, pathWidth)
	dividerLine := strings.Repeat("-", methodWidth+pathWidth+45)

	fmt.Println(dividerLine)
	fmt.Printf(headerFormat, "Method", "Path", "Handler")
	fmt.Println(dividerLine)

	// 打印路由信息
	format := fmt.Sprintf("%%-%ds | %%-%ds | %%s\n", methodWidth, pathWidth)
	for _, route := range routeInfos {
		handlerName := getHandlerFuncName(route.Handler)
		fmt.Printf(format, route.Method, route.Path, handlerName)
	}
	fmt.Println(dividerLine)
}

// getHandlerFuncName 尝试获取处理函数的名称
func getHandlerFuncName(handler string) string {
	// 处理函数名通常是包含包路径的完整名称
	parts := strings.Split(handler, ".")
	if len(parts) > 0 {
		// 返回最后一部分作为函数名
		// 加上 "()" 来表示它是一个函数
		lastPart := parts[len(parts)-1]
		if strings.HasPrefix(lastPart, "func1") {
			return "Closure"
		}
		return lastPart
	}
	return handler
}
