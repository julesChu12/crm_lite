package main

import (
	"crm_lite/internal/bootstrap"
	"crm_lite/internal/core/config"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/policy"
	"crm_lite/internal/routes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // Blank import for database driver
	"github.com/joho/godotenv"
)

// virtualRoleForAllApis 是一个特殊的、不存在的角色，用作占位符
// 用于在 Casbin 中对所有可发现的 API 端点进行分组。
// 管理员界面可以获取这个"角色"的所有策略，以获得所有
// 可分配给真实角色的可用权限列表。
const virtualRoleForAllApis = "_all_apis_"

// main 是一个工具，用于发现所有注册的 API 端点并将它们添加为 casbin 策略。
// 这允许管理员管理角色的权限，而无需手动输入 API 路径。
func main() {
	// 1. 初始化配置
	if err := initConfigForTool(); err != nil {
		log.Fatalf("failed to initialize configuration: %v", err)
	}
	fmt.Println("Configuration loaded successfully.")

	// 2. 初始化应用程序引导程序
	resManager, logCleaner, cleanup, err := bootstrap.Bootstrap()
	if err != nil {
		log.Fatalf("failed to bootstrap application: %v", err)
	}
	defer cleanup() // 确保资源被清理
	fmt.Println("Application bootstrapped successfully.")

	// 3. 获取 Casbin Enforcer
	casbinRes, err := resource.Get[*resource.CasbinResource](resManager, resource.CasbinServiceKey)
	if err != nil {
		log.Fatalf("failed to get casbin resource: %v", err)
	}
	enforcer := casbinRes.GetEnforcer()
	fmt.Println("Casbin enforcer obtained.")

	// 4. 初始化一个临时的 Gin Router 来获取路由列表
	// 注意：这里我们不需要启动 HTTP 服务器，只需要路由定义
	gin.SetMode(gin.ReleaseMode)
	router := routes.NewRouter(resManager, logCleaner)
	fmt.Println("Router initialized for route discovery.")

	// 5. 执行发现和写入操作
	if err := discoverAndSeed(router, enforcer); err != nil {
		log.Fatalf("failed to discover and seed APIs: %v", err)
	}

	fmt.Println("✅ API discovery and seeding completed successfully!")
}

// initConfigForTool 参照 cmd/root.go 的逻辑来初始化配置
func initConfigForTool() error {
	// 优先从 .env 文件加载环境变量 (例如 ENV)
	_ = godotenv.Load()

	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		env = "dev" // 默认环境
	}
	fmt.Printf("Running tool in [%s] environment.\n", env)

	configName := fmt.Sprintf("app.%s.yaml", env)
	configFilePath, err := findConfigInProject(configName)
	if err != nil {
		return fmt.Errorf("config file not found in project: %w", err)
	}

	fmt.Printf("Using config file: %s\n", configFilePath)

	if err := config.InitOptions(configFilePath); err != nil {
		return fmt.Errorf("init config failed: %w", err)
	}
	return nil
}

// findConfigInProject 动态查找并返回项目内的配置文件绝对路径 (主要用于本地开发)
// 从 cmd/root.go 复制而来，以保证行为一致。
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

// discoverAndSeed 发现所有在 /api/v1/ 前缀下的路由，并将它们作为资源策略写入 Casbin。
// 这个工具的目的是确保所有潜在的受保护端点都在 Casbin 中有记录，以便管理员可以为它们分配权限。
// 运行时的实际访问控制由中间件（JWT+Casbin）处理，包括白名单的放行。
func discoverAndSeed(router *gin.Engine, enforcer *casbin.Enforcer) error {
	allRoutes := router.Routes()
	var policiesAdded int

	// 1. 加载白名单路由并转换为方便查找的 map
	publicRoutes := policy.GetPublicRoutes()
	whitelist := make(map[string]map[string]struct{})
	for _, route := range publicRoutes {
		if _, ok := whitelist[route.Path]; !ok {
			whitelist[route.Path] = make(map[string]struct{})
		}
		whitelist[route.Path][route.Method] = struct{}{}
	}
	fmt.Printf("Loaded %d public routes into the whitelist.\n", len(publicRoutes))

	fmt.Printf("Found %d total routes. Filtering for API endpoints under /api/v1/...\n", len(allRoutes))

	for _, routeInfo := range allRoutes {
		// 我们只关心 /api/v1/ 下的可管理API路由
		if !strings.HasPrefix(routeInfo.Path, "/api/v1/") {
			continue
		}

		// 2. 检查当前路由是否在白名单中
		if methods, ok := whitelist[routeInfo.Path]; ok {
			if _, ok := methods[routeInfo.Method]; ok {
				fmt.Printf("  -> Skipping whitelisted route: { %s, %s }\n", routeInfo.Method, routeInfo.Path)
				continue
			}
		}

		// 所有 /api/v1/ 路由都应被视为资源
		path := routeInfo.Path
		method := routeInfo.Method

		// 检查策略是否已存在
		has, err := enforcer.HasPolicy(virtualRoleForAllApis, path, method)
		if err != nil {
			log.Printf("Error checking policy existence for %s %s: %v", method, path, err)
			continue
		}

		if !has {
			// 添加策略 (p, _all_apis_, /api/v1/customers, GET)
			if _, err := enforcer.AddPolicy(virtualRoleForAllApis, path, method); err != nil {
				log.Printf("Error adding policy for %s %s: %v", method, path, err)
				continue
			}
			policiesAdded++
			fmt.Printf("  -> Added policy for resource: { %s, %s, %s }\n", virtualRoleForAllApis, path, method)
		}
	}

	if policiesAdded > 0 {
		fmt.Printf("Added %d new API policies. Saving to database...\n", policiesAdded)
		// 保存所有新添加的策略到数据库
		if err := enforcer.SavePolicy(); err != nil {
			return fmt.Errorf("failed to save policies to database: %w", err)
		}
		fmt.Println("Policies saved successfully.")
	} else {
		fmt.Println("No new API policies to add. Everything is up to date.")
	}

	return nil
}
