package main

import (
	"crm_lite/internal/bootstrap"
	"crm_lite/internal/core/resource"
	"crm_lite/internal/routes"
	"fmt"
	"log"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// virtualRoleForAllApis 是一个特殊的、不存在的角色，用作占位符
// 用于在 Casbin 中对所有可发现的 API 端点进行分组。
// 管理员界面可以获取这个"角色"的所有策略，以获得所有
// 可分配给真实角色的可用权限列表。
const virtualRoleForAllApis = "_all_apis_"

// main 是一个工具，用于发现所有注册的 API 端点并将它们添加为 casbin 策略。
// 这允许管理员管理角色的权限，而无需手动输入 API 路径。
func main() {
	// 1. 初始化应用程序引导程序
	resManager, cleanup, err := bootstrap.Bootstrap()
	if err != nil {
		log.Fatalf("failed to bootstrap application: %v", err)
	}
	defer cleanup() // 确保资源被清理
	fmt.Println("Application bootstrapped successfully.")

	// 2. 获取 Casbin Enforcer
	casbinRes, err := resource.Get[*resource.CasbinResource](resManager, resource.CasbinServiceKey)
	if err != nil {
		log.Fatalf("failed to get casbin resource: %v", err)
	}
	enforcer := casbinRes.GetEnforcer()
	fmt.Println("Casbin enforcer obtained.")

	// 3. 初始化一个临时的 Gin Router 来获取路由列表
	// 注意：这里我们不需要启动 HTTP 服务器，只需要路由定义
	gin.SetMode(gin.ReleaseMode)
	router := routes.NewRouter(resManager)
	fmt.Println("Router initialized for route discovery.")

	// 4. 执行发现和写入操作
	if err := discoverAndSeed(router, enforcer); err != nil {
		log.Fatalf("failed to discover and seed APIs: %v", err)
	}

	fmt.Println("✅ API discovery and seeding completed successfully!")
}

// discoverAndSeed 发现所有在 /api/v1/ 前缀下的路由，并将它们作为资源策略写入 Casbin。
// 这个工具的目的是确保所有潜在的受保护端点都在 Casbin 中有记录，以便管理员可以为它们分配权限。
// 运行时的实际访问控制由中间件（JWT+Casbin）处理，包括白名单的放行。
func discoverAndSeed(router *gin.Engine, enforcer *casbin.Enforcer) error {
	allRoutes := router.Routes()
	var policiesAdded int

	fmt.Printf("Found %d total routes. Filtering for API endpoints under /api/v1/...\n", len(allRoutes))

	for _, routeInfo := range allRoutes {
		// 我们只关心 /api/v1/ 下的可管理API路由
		if !strings.HasPrefix(routeInfo.Path, "/api/v1/") {
			continue
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
