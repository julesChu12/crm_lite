# 📦 订单模块设计文档（Order Module） - CRM Lite

## 📌 模块概述

订单模块负责记录客户的交易行为与服务记录，是客户价值评估、业绩统计与业务分析的重要数据来源。模块支持订单的创建、修改、查询与删除，关联客户信息及订单明细。

---

## 🧱 数据模型设计

### 订单主表（order）

| 字段名          | 类型       | 描述                   |
|------------------|------------|------------------------|
| `id`             | UUID       | 订单唯一标识           |
| `customer_id`    | UUID       | 所属客户ID             |
| `order_date`     | datetime   | 下单日期               |
| `status`         | string     | 订单状态（如：待支付, 已支付, 处理中, 已完成, 已取消, 退款中, 已退款）   |
| `total_amount`   | float      | 总金额（含折扣）(注意：使用 float 类型处理金额需注意精度问题，或考虑使用定点数/整数存储分)       |
| `remark`         | text       | 订单备注信息           |
| `created_at`     | datetime   | 创建时间               |
| `updated_at`     | datetime   | 更新时间               |
| `deleted_at`     | datetime   | 软删除字段             |

### 订单明细表（order_item）

| 字段名         | 类型     | 描述                     |
|----------------|----------|--------------------------|
| `id`           | UUID     | 明细唯一标识             |
| `order_id`     | UUID     | 所属订单ID               |
| `item_name`    | string   | 商品或服务名称           |
| `quantity`     | int      | 数量                     |
| `unit_price`   | float    | 单价 (注意：使用 float 类型处理金额需注意精度问题，或考虑使用定点数/整数存储分)                    |
| `total_price`  | float    | 小计（数量×单价）(注意：使用 float 类型处理金额需注意精度问题，或考虑使用定点数/整数存储分)        |

---

## 📂 模块目录结构

internal/order/
├── controller/
│   └── order_controller.go      # 订单路由处理
├── service/
│   └── order_service.go         # 核心业务逻辑
├── model/
│   ├── order.go                 # 订单主表模型
│   └── order_item.go            # 订单明细模型
├── query/
│   └── order_query.go           # 高级筛选逻辑封装
├── router/
│   └── order_router.go          # 接口路由注册
├── validator/
│   └── order_validator.go       # 请求验证结构体

---

## 🔌 接口设计（RESTful）

### 📄 订单接口定义

#### GET `/api/orders`

- 描述：获取订单分页列表
- 请求参数：
  - `customer_id` (string, 可选)：筛选指定客户订单
  - `status` (string, 可选)：订单状态筛选
  - `start_date` (date, 可选)：时间区间筛选（起始日期）
  - `end_date` (date, 可选)：时间区间筛选（结束日期）
  - `page` (int, 可选)：页码
  - `page_size` (int, 可选)：每页数量
- 返回：

```json
{
  "list": [
    {
      "id": "uuid-order-1",
      "customer_id": "uuid-customer-1",
      "order_date": "2024-05-20T10:00:00Z",
      "status": "已支付",
      "total_amount": 199.00,
      "remark": "客户首次消费"
    }
  ],
  "pagination": {
    "total": 50,
    "page": 1,
    "page_size": 10
  }
}
```

#### GET `/api/orders/:id`

- 描述：获取订单详情（含订单明细）
- 请求参数：无
- 返回：

```json
{
  "id": "uuid-order-1",
  "customer_id": "uuid-customer-1",
  "order_date": "2024-05-20T10:00:00Z",
  "status": "已支付",
  "total_amount": 199.00,
  "remark": "客户首次消费",
  "created_at": "2024-05-20T10:00:00Z",
  "updated_at": "2024-05-20T10:05:00Z",
  "items": [
    {
      "id": "uuid-item-1",
      "item_name": "精致洗车服务",
      "quantity": 1,
      "unit_price": 199.00,
      "total_price": 199.00
    }
  ]
}
```

#### POST `/api/orders`

- 描述：创建订单（含订单明细）。后端通常会基于 `items` 重新计算或校验 `total_amount`。
- 请求体 (JSON):

```json
{
  "customer_id": "uuid-customer-1",
  "order_date": "2024-06-01T14:30:00Z",
  "remark": "新订单",
  "items": [
    {
      "item_name": "VIP月卡",
      "quantity": 1,
      "unit_price": 299.00
    },
    {
      "item_name": "内饰清洁",
      "quantity": 1,
      "unit_price": 100.00
    }
  ]
  // total_amount 建议由后端计算，或前端传入后后端校验
}
```

- 返回：成功时返回新创建的订单详情，类似 `GET /api/orders/:id` 的结构。

#### PUT `/api/orders/:id`

- 描述：更新订单信息（主要用于更新状态、备注等非核心信息）。核心交易信息（如商品、金额）的修改通常有特定流程或限制。
- 请求体 (JSON)：(包含需要更新的字段)

```json
{
  "status": "已完成",
  "remark": "服务已完成，客户满意"
}
```

- 返回：成功时返回更新后的订单详情。

#### DELETE `/api/orders/:id`

- 描述：软删除订单。
- 请求参数：无
- 返回：成功时返回 204 No Content 或操作成功信息。

---

### 🔐 权限设计建议

- 订单的访问可能需要用户对关联客户拥有相应权限。同时，订单自身操作（如创建、修改、删除）需要如 `order:create`, `order:write`, `order:delete` 等特定权限。
- 金额字段建议对低权限用户进行脱敏显示或隐藏。

---

### 🧪 单元测试建议

- 订单金额（基于订单明细）计算的准确性。
- 各种订单状态流转的逻辑正确性。
- 订单创建、更新（特别是状态变更）成功与失败用例。
- 基于客户ID、日期范围、状态等条件的查询与分页准确性验证。

---

### 📎 依赖说明

- GORM（模型管理、事务处理）
- Validator（输入校验，如 `github.com/go-playground/validator`）
- Zap（日志记录）
- Casbin（权限控制）

---

### 💡 扩展建议

- 可通过更完善的订单状态管理（如：草稿、待付款、已付款、待发货/待服务、已发货/服务中、已完成、已关闭、退款申请中、已退款等）实现复杂的订单生命周期和业务流程控制。
- 可扩展支持优惠券、折扣、积分抵扣等促销功能。
- 可基于订单数据生成发票或对接电子发票系统。
