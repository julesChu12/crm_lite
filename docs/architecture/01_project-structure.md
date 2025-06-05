# 🧱 项目结构说明文档 - CRM Lite

本项目是一个为理发店、洗车店等小微企业打造的轻量级客户关系管理系统，后端采用 Go 实现，前后端分离，支持模块化、容器化部署，具有良好的拓展性。

---

## 📌 项目概览

- 类型：SaaS 型 CRM 系统（支持单体部署 + 插件扩展）
- 后端：Go + Gin + PostgreSQL + Redis + Cobra + Zap
- 安全：JWT + Casbin
- 前端：Vue3 + Pinia + WebSocket
- 部署：Docker Compose + GitHub Actions（CI）

---

## 🧱 顶层目录结构说明

.
├── bootstrap/                 # 应用初始化（配置加载、数据库连接等）
├── cmd/                       # CLI 命令入口（如 cobra 命令）
│   └── tools/
│       └── db/                # 数据库相关命令（迁移、种子数据等）
|   router.go
|   root.go
|   start.go
|   stop.go
├── config/                    # 配置文件（如 casbin 权限配置）
├── deploy/                    # 部署相关文件
│   ├── docker/
│   │   ├── base/              # 基础镜像（如数据库、缓存）
│   │   ├── compose/           # Docker Compose 文件
│   │   └── overlays/          # 环境覆盖配置（dev、prod）
│   └── scripts/               # 部署脚本
├── docs/                       # 项目文档
│   ├── architecture/          # 架构设计文档
│   └── database/              # 数据库设计文档（ER 图、迁移说明等）
├── internal/                  # 内部应用逻辑
│   ├── controller/            # 控制器（处理 HTTP 请求）
│   ├── model/                 # 数据模型（gorm/gen 生成）
│   ├── query/                 # 查询接口（gorm/gen 生成）
│   ├── service/               # 业务逻辑处理
│   └── middleware/            # 中间件（如认证、日志）
├── logs/                      # 日志文件
├── pkg/                       # 公共库
│   ├── auth/                  # 认证相关（JWT、RBAC）
│   ├── cache/                 # 缓存处理（Redis）
│   ├── config/                # 配置管理
│   ├── db/                    # 数据库连接与操作
│   ├── errors/                # 错误处理
│   ├── logger/                # 日志工具（如 zap）
│   ├── middlewares/           # 公共中间件
│   ├── mq/                    # 消息队列处理
│   ├── process/               # 后台任务处理
│   ├── resource/              # 静态资源管理
│   ├── response/              # 响应处理
│   ├── scheduler/             # 定时任务调度
│   └── utils/                 # 工具函数
├── tmp/                       # 临时文件
└── web/                       # 前端代码（Vue3 + Pinia）

## 🗂 internal/ 模块结构（基于 MVC）

| 模块 | 说明 |
|------|------|
| `controller/`     | HTTP 层，处理请求路由与响应返回（即 Controller） |
| `service/`        | 业务逻辑处理层，协调 model、缓存、事务等操作 |
| `model/`          | GORM 实体定义，包含表结构及方法（即 Model） |
| `middleware/`     | Gin 中间件（JWT验证、日志、恢复、限流等） |
| `router/`         | Gin 路由初始化与模块注册 |
| `validator/`      | 参数校验器（可选，集成 binding & validation） |
| `websocket/`      | WebSocket 消息推送服务（事件发布/订阅） |
| `job/`            | 异步任务处理模块，如定时营销回访任务 |
| `startup/`        | 应用启动器（组合初始化逻辑） |

### MVC 结构说明

- `controller` 负责接收请求并调用 `service`
- `service` 执行具体业务逻辑并调用 `model`
- `model` 封装数据库结构及相关方法

---

## ⚙️ 使用技术栈与依赖说明

| 技术组件 | 用途 | 包名 |
|----------|------|------|
| **Gin** | Web框架，处理 REST 接口与中间件 | `github.com/gin-gonic/gin` |
| **GORM** | ORM 框架，连接 PostgreSQL | `gorm.io/gorm` / `gorm.io/driver/postgres` |
| **JWT** | 鉴权令牌生成与校验 | `github.com/golang-jwt/jwt/v5` |
| **Redis** | 缓存、验证码、临时令牌存储 | `github.com/redis/go-redis/v9` |
| **Zap** | 日志框架，结构化日志输出 | `go.uber.org/zap` |
| **Viper** | 配置加载器 | `github.com/spf13/viper` |
| **Cobra** | 命令行启动与子命令注册 | `github.com/spf13/cobra` |
| **Swaggo** | Swagger API 自动文档生成 | `github.com/swaggo/gin-swagger` |
| **Casbin** | 权限模型（RBAC）管理 | `github.com/casbin/casbin/v2` |
| **WebSocket** | 实时事件推送 | `github.com/gorilla/websocket` |
| **uuid** | 主键生成工具 | `github.com/google/uuid` |

---

## 🚀 Cobra 命令结构

```bash
go run main.go serve      # 启动 Web API 服务
go run main.go migrate    # 数据库迁移
go run main.go seed       # 写入测试数据
go run main.go version    # 输出构建信息
