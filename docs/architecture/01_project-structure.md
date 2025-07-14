# 🏗️ 项目结构与模块划分

本文档详细描述了 CRM Lite 项目的目录结构和代码组织方式，旨在帮助开发人员快速理解项目全貌。

## 🧱 顶层目录结构说明
```
.
├── cmd/             # CLI 命令入口 (Cobra)
├── config/          # 配置文件 (YAML)
├── db/              # 数据库迁移脚本 (.sql)
├── docs/            # 项目文档 (Markdown)
├── internal/        # 核心应用代码 (私有)
├── logs/            # 运行时日志
├── pkg/             # 可供外部应用引用的公共库
├── tmp/             # 临时文件
├── .env.example     # 环境变量示例
├── docker-compose.yaml # Docker Compose 配置
├── Dockerfile       # 生产环境 Dockerfile
├── go.mod           # Go 模块依赖
└── main.go          # 项目主入口
```

## 📂 `internal` 目录详解

`internal` 目录是项目的心脏，包含了所有不对外暴露的业务逻辑代码，并遵循清晰的分层架构。

```
internal/
├── bootstrap/       # 应用初始化，在 main.go 调用，负责加载配置、连接资源
├── controller/      # HTTP 控制器 (Controller)，负责解析请求、调用 Service、返回响应
├── core/            # 核心引擎与资源管理
│   ├── config/      # 配置加载与解析 (Viper)
│   ├── logger/      # 日志组件 (Zap)
│   └── resource/    # 资源管理器，统一管理 DB, Cache 等共享资源
├── dao/             # 数据访问对象 (DAO)，由 GORM Gen 自动生成，禁止手动修改
│   ├── model/       # 数据库表模型
│   └── query/       # 类型安全的查询代码
├── dto/             # 数据传输对象 (DTO)，定义 API 的请求和响应结构体
├── middleware/      # Gin 中间件，如 JWT 认证、Casbin 鉴权、日志记录
├── policy/          # 权限相关策略，如 Casbin 的 API 白名单
├── routes/          # API 路由注册，将 URL 路径与 Controller 方法绑定
├── service/         # 业务逻辑层 (Service)，处理核心业务流程，是功能的主要实现者
└── startup/         # 封装应用启动和关闭的逻辑
```

## 📦 `pkg` 目录详解

`pkg` 目录存放可以被外部应用安全引用的公共代码库。

```
pkg/
├── process/         # 进程管理工具，如 PID 文件处理
├── resp/            # 统一的 API 响应格式封装
└── utils/           # 通用工具函数，如 JWT、上下文处理、随机数等
```

---

## 🚀 推荐开发流程

1.  **定义模型**: 在 `db/migrations` 中创建或修改表的 SQL 文件。
2.  **生成代码**: 运行 `go run main.go tools gen`，GORM Gen 会自动更新 `internal/dao` 下的代码。
3.  **定义 DTO**: 在 `internal/dto` 中为新功能定义请求和响应的数据结构。
4.  **编写服务**: 在 `internal/service` 中创建或修改服务，实现核心业务逻辑。
5.  **创建控制器**: 在 `internal/controller` 中添加方法，调用服务并处理 HTTP IO。
6.  **注册路由**: 在 `internal/routes` 中将新的 URL 路径指向控制器方法。
7.  **更新文档**: 如果有必要，更新 `docs/architecture` 中的相关模块文档和 `README.md`。
