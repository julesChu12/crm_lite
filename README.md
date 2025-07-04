# CRM Lite

一个为理发店、洗车店等小微企业打造的轻量级客户关系管理系统。

## 技术栈

- **后端**: Go + Gin + MariaDB + Redis + Cobra + Zap
- **安全**: JWT + Casbin
- **前端**: Vue3 + Pinia + WebSocket
- **部署**: Docker Compose + GitHub Actions（CI）

## 项目结构

```
.
├── bootstrap/                 # 应用初始化
├── cmd/                       # CLI 命令入口
├── config/                    # 配置文件
├── db/                        # 数据库相关
│   └── migrations/            # 数据库迁移文件
├── deploy/                    # 部署相关文件
├── docs/                      # 项目文档
│   └── architecture/          # 架构设计文档
├── internal/                  # 内部应用逻辑
│   ├── common/                # 通用组件
│   ├── controller/            # 控制器
│   ├── model/                 # 数据模型
│   ├── service/               # 业务逻辑
│   └── ...
├── logs/                      # 日志文件
├── pkg/                       # 公共库
├── tmp/                       # 临时文件
└── web/                       # 前端代码
```

## 数据库设计

### 设计原则

- **逻辑外键设计**: 使用逻辑外键而非物理外键约束，提高性能和扩展性
- **软删除支持**: 所有核心表支持软删除，保证数据安全
- **UUID主键**: 使用UUID作为主键，支持分布式环境
- **索引优化**: 合理设计索引，优化查询性能

### 核心表结构

1. **权限管理**
   - `casbin_rules` - Casbin权限规则
   - `roles` - 角色表
   - `admin_users` - 管理员用户
   - `admin_user_roles` - 用户角色关联

2. **客户管理**
   - `customers` - 客户主档
   - `contacts` - 联系人（一个客户可有多联系人）

3. **产品订单**
   - `products` - 产品/服务
   - `orders` - 订单主表
   - `order_items` - 订单明细

4. **资金管理**
   - `wallets` - 钱包/储值
   - `wallet_transactions` - 钱包流水

5. **客户运营**
   - `activities` - 客户互动记录
   - `marketing_campaigns` - 营销活动
   - `marketing_records` - 营销记录

### 逻辑外键关系说明

由于采用逻辑外键设计，数据完整性需要在应用层保证：

```
customers.assigned_to -> admin_users.id
contacts.customer_id -> customers.id
orders.customer_id -> customers.id
orders.contact_id -> contacts.id
orders.assigned_to -> admin_users.id
orders.created_by -> admin_users.id
order_items.order_id -> orders.id
order_items.product_id -> products.id
wallets.customer_id -> customers.id
wallet_transactions.wallet_id -> wallets.id
wallet_transactions.operator_id -> admin_users.id
activities.customer_id -> customers.id
activities.contact_id -> contacts.id
activities.assigned_to -> admin_users.id
activities.created_by -> admin_users.id
marketing_campaigns.created_by -> admin_users.id
marketing_campaigns.updated_by -> admin_users.id
marketing_records.campaign_id -> marketing_campaigns.id
marketing_records.customer_id -> customers.id
marketing_records.contact_id -> contacts.id
admin_user_roles.admin_user_id -> admin_users.id
admin_user_roles.role_id -> roles.id
```

## 快速开始

### 1. 环境准备

确保已安装：

- Go 1.21+
- Docker & Docker Compose
- MariaDB 10.8+
- Redis 6+

### 2. 启动数据库

```bash
# 启动 MariaDB 和 Redis
docker-compose up -d mariadb redis

# 可选：启动 phpMyAdmin 数据库管理工具
docker-compose up -d phpmyadmin
```

### 3. 运行迁移

```bash
# 运行数据库迁移
go run main.go migrate
```

### 4. 启动服务

```bash
# 启动API服务器
go run main.go serve
```

### 5. 访问服务

- API服务器: <http://localhost:8080>
- phpMyAdmin数据库管理: <http://localhost:8081>
  - 服务器: `mariadb`
  - 用户名: `crm_user`
  - 密码: `crm_pass`
  - 数据库: `crm_db`

## 开发指南

### 可用命令

```bash
# 查看帮助
go run main.go --help

# 启动服务器
go run main.go serve

# 运行数据库迁移
go run main.go migrate
```

### 数据完整性保证

由于使用逻辑外键，需要在应用层保证数据完整性：

1. **删除操作**: 在删除父记录前，检查是否存在子记录
2. **插入操作**: 在插入子记录时，验证父记录是否存在
3. **更新操作**: 在更新外键字段时，验证目标记录是否存在

### 开发流程建议

1. **权限表优先**: 先完善 `roles`、`admin_users`、`casbin_rules` 权限核心
2. **最小闭环**: `customers` → `contacts` → `orders`/`products` 形成基础业务闭环
3. **按需扩展**: 钱包、活动、营销功能可按业务需求逐步添加

## 本地调试 GitHub Actions（act）

本项目所有 CI/CD workflow 支持使用 [act](https://github.com/nektos/act) 进行本地调试。

### 步骤

1. 安装 act：
   ```bash
   brew install act # macOS
   # 或参考官方文档安装
   ```
2. 在项目根目录新建 `.secrets` 文件，内容参考下方示例。
3. 运行本地 workflow，例如：
   ```bash
   act -j deploy-production -W .github/workflows/release.yml
   # 或
   act -j build-dev -W .github/workflows/dev.yml
   ```

### .secrets 文件示例

```
PROD_HOST=your-production-host
PROD_USER=your-ssh-user
PROD_SSH_KEY=your-ssh-private-key
# 其他 workflow 用到的 secrets 也可在此补充
```

> `.secrets` 文件内容需与 workflow 中用到的 `${{ secrets.XXX }}` 保持一致。
> act 会自动加载该文件并注入到 workflow 运行环境。

## 文档

详细的架构设计文档请查看 `docs/architecture/` 目录：

- [项目结构说明](docs/architecture/01_project-structure.md)
- [客户模块设计](docs/architecture/02_module_customer.md)
- [订单模块设计](docs/architecture/03_module_order.md)
- [钱包模块设计](docs/architecture/04_module_wallet.md)
- [营销模块设计](docs/architecture/05_module_marketing.md)
- [通用模块设计](docs/architecture/06_module_common.md)

## 项目进度

- **v0.1 (已完成)**
  - [x] 搭建项目基础框架 (Gin, Gorm, Casbin)
  - [x] 实现用户认证与RBAC权限管理
  - [x] 实现用户、角色、权限管理API
  - [x] 实现客户管理核心API

- **v0.2 (进行中)**
  - [ ] **产品与订单模块**: 打通 `客户 -> 产品 -> 订单` 的核心业务流
  - [ ] 完善产品、订单相关的 `Service` 和 `Controller`
  - [ ] 注册新模块的API路由

- **v0.3 (计划中)**
  - [ ] **资金钱包模块**: 实现客户储值、消费与流水功能

- **v0.4 (计划中)**
  - [ ] **客户运营模块**: 实现营销活动与客户互动功能

- **v0.5 (计划中)**
  - [ ] **测试与完善**: 补充单元测试，与前端完成对接

## 许可证

[MIT License](LICENSE)
