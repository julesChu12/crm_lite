# CRM Lite - 轻量级客户关系管理系统

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![Gin](https://img.shields.io/badge/Gin-v1.9-0089D6?style=for-the-badge)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-v1.25-9B4F96?style=for-the-badge)](https://gorm.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)

**CRM Lite** 是一个使用 Go 语言构建的、为小微企业（如理发店、洗车店）设计的轻量级客户关系管理 (CRM) 后端服务。它提供了一套完整的 API，用于管理客户、产品、订单等核心业务数据。

## ✨ 核心功能

- **身份认证与授权**: 基于 JWT 的用户认证和 Casbin 的 RBAC 权限控制。
- **用户与角色管理**: 支持多用户、多角色的管理体系。
- **客户管理**: 完整的客户信息 CRUD 和批量查询功能。
- **产品管理**: 管理可销售的产品或服务，包括库存。
- **订单管理**: 支持事务性的订单创建和丰富的查询功能。
- **API 文档**: 通过 Swagger (OpenAPI) 自动生成并提供交互式 API 文档。
- **其他模块**: 包含钱包、营销等模块的基础结构，可按需扩展。

## 🛠️ 技术栈

- **后端**: Go, Gin
- **数据库**: GORM, MariaDB
- **缓存**: Redis
- **命令行**: Cobra
- **安全**: JWT, Casbin
- **日志**: Zap
- **容器化**: Docker, Docker Compose

## 🚀 快速开始

### 1. 环境准备

- [Go](https://go.dev/doc/install) 1.24+
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- [swag](https://github.com/swaggo/swag) CLI (用于生成 API 文档)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 环境配置

复制配置文件模板，并根据需要进行修改。

```bash
cp .env.example .env
cp config/app.test.yaml config/app.prod.yaml
```

> **注意**: `.env` 文件包含了数据库密码等敏感信息，已在 `.gitignore` 中忽略。`config/app.prod.yaml` 是生产环境的最终配置文件。

### 3. 启动依赖服务

使用 Docker Compose 一键启动数据库和缓存服务。

```bash
docker-compose up -d
```

这将启动 MariaDB, Redis, 和一个可选的 phpMyAdmin。

### 4. 数据库迁移

运行 `db:migrate` 命令来初始化数据库表结构。

```bash
go run main.go tools db:migrate
```

### 5. 启动应用

现在，可以启动 CRM Lite 的 API 服务了。

```bash
go run main.go start
```

服务启动后，你可以在 `http://localhost:8080` 访问 API。

### 6. API 文档

我们使用 `swag` 根据代码注释自动生成 Swagger 文档。要查看或更新文档：

```bash
# 生成/更新 docs 目录下的 swagger.json, swagger.yaml, docs.go
swag init

# 启动服务后，访问以下 URL 查看交互式 API 文档
# http://localhost:8080/swagger/index.html
```

## 🏗️ 项目结构

CRM Lite 遵循了清晰的分层架构，主要的应用逻辑位于 `internal/` 目录下：

```text
internal/
├── bootstrap/     # 应用启动和初始化逻辑
├── controller/    # HTTP 控制器，处理请求和响应
├── core/          # 核心组件，如配置、日志、资源管理器
├── dao/           # 数据访问对象 (DAO)，由 GORM Gen 生成
├── dto/           # 数据传输对象 (DTO)，用于 API 的输入和输出
├── middleware/    # Gin 中间件，如认证、日志记录
├── policy/        # 权限策略相关，如 Casbin 白名单
├── routes/        # API 路由注册
└── service/       # 业务逻辑层，处理核心业务流程
```

## 📚 架构文档

想深入了解每个模块的设计细节吗？请查阅我们的**架构文档**：

- [**`docs/architecture/README.md`**](./docs/architecture/README.md)

## 🤝 贡献

我们欢迎任何形式的贡献！无论是提交 Issue、发起 Pull Request，还是改进文档。

## 📄 许可证

本项目基于 MIT License 开源。

## Database Migrations

Use the built-in migration tool to apply migrations.

Dev environment:

```
RUN_DB_TESTS=0 go run cmd/tools/db/migrate.go -env=dev -direction=up -steps=0
```

Test environment:

```
RUN_DB_TESTS=0 go run cmd/tools/db/migrate.go -env=test -direction=up -steps=0
```

Notes:
- New domain-related migrations added:
  - `db/migrations/20250913_01_billing_tables.sql`
  - `db/migrations/20250913_02_sales_item_snapshots.sql`
  - `db/migrations/20250913_03_sys_outbox.sql`
- These are additive and safe to apply. Rollback with `-direction=down` if needed.
