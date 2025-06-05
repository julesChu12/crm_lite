# 👥 客户模块设计文档（Customer Module） - CRM Lite

## 📌 模块概述

客户模块是系统的核心基础模块之一，主要负责管理客户的基本信息、联系资料、标签、等级等数据内容。所有订单、回访、营销等业务均以客户为基础构建。

---

## 🧱 数据模型设计

### 客户数据表（customer）

| 字段名         | 类型       | 说明                         |
|----------------|------------|------------------------------|
| `id`           | UUID       | 客户唯一标识                 |
| `name`         | string     | 客户姓名                     |
| `phone`        | string     | 手机号（唯一）              |
| `email`        | string     | 邮箱（可选）                |
| `gender`       | string     | 性别（male/female/unknown） |
| `birthday`     | date       | 出生日期（可选）            |
| `level`        | string     | 客户等级（如 VIP、普通）     |
| `tags`         | []string   | 标签字段（如 JSONB 存储）    |
| `note`         | text       | 客户备注                     |
| `created_at`   | datetime   | 创建时间                     |
| `updated_at`   | datetime   | 更新时间                     |
| `deleted_at`   | datetime   | 软删除字段（gorm 支持）     |

---

## 📂 模块目录结构

internal/customer/
├── controller/
│   └── customer_controller.go    # HTTP 路由处理
├── service/
│   └── customer_service.go       # 业务逻辑处理
├── model/
│   └── customer.go               # GORM 模型定义
├── query/
│   └── customer_query.go         # 查询封装（可选）
├── router/
│   └── customer_router.go        # 路由注册
├── validator/
│   └── customer_validator.go     # 请求参数校验器（可选）

---

## 🔌 接口设计（RESTful）

### 📄 客户信息接口定义

#### GET `/api/customers`

- 描述：获取客户分页列表
- 请求参数：
  - `name`（可选）：模糊搜索
  - `level`（可选）：等级筛选
  - `page`, `page_size`
- 返回：

```json
{
  "list": [
    {
      "id": "uuid",
      "name": "张三",
      "phone": "138****8888",
      "level": "VIP"
    }
  ],
  "pagination": {
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

#### GET `/api/customers/:id`

- 描述：根据 ID 获取客户详细信息
- 请求参数：无
- 返回：

```json
{
  "id": "uuid",
  "name": "张三",
  "phone": "138****8888",
  "email": "zhangsan@example.com",
  "gender": "male",
  "birthday": "1990-01-01",
  "level": "VIP",
  "tags": ["老客户", "高价值"],
  "note": "这是一个重要的客户。",
  "created_at": "2023-01-15T10:00:00Z",
  "updated_at": "2023-01-16T14:30:00Z"
}
```

#### POST `/api/customers`

- 描述：新增客户
- 请求体 (JSON)：

```json
{
  "name": "李四",
  "phone": "13912345678",
  "email": "lisi@example.com",
  "gender": "male",
  "birthday": "1992-05-20",
  "level": "普通",
  "tags": ["回头客", "会员"],
  "note": "新注册会员"
}
```

- 返回：成功时返回新创建的客户信息，类似 `GET /api/customers/:id` 的结构。

#### PUT `/api/customers/:id`

- 描述：更新客户信息
- 请求体 (JSON)：（包含需要更新的字段）

```json
{
  "name": "李四新",
  "email": "lisi.new@example.com",
  "level": "VIP"
}
```

- 返回：成功时返回更新后的客户信息。

#### DELETE `/api/customers/:id`

- 描述：软删除客户记录
- 请求参数：无
- 返回：成功时返回 204 No Content 或操作成功信息。

---

### 🔐 权限设计建议

- 普通员工可管理自己的客户
- 管理员可查看与编辑所有客户
- 删除接口需要 `customer:delete` 权限控制

---

### 🧪 单元测试建议

- 创建客户成功与失败用例
- 更新无效字段验证
- 模糊查询与分页准确性验证

---

### 📎 依赖说明

- GORM（自动迁移、CRUD）
- Casbin（RBAC 权限控制）
- Zap 日志记录
- Redis（可用于手机号去重缓存）

---

### 💡 扩展建议

如需扩展客户画像、积分、会员卡等字段，建议通过“附加属性表”或“JSONB字段”方式以保证表结构整洁灵活。
