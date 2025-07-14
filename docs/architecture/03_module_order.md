# 🛒 订单模块设计文档 (Order Module) - CRM Lite

## 📌 模块概述

订单模块是 CRM 系统的业务核心，用于记录客户的购买行为。它与客户模块和产品模块紧密相连，负责创建、查询和管理订单及其包含的订单项。

---

## 🧱 数据模型设计

### 订单主表 (`orders`)

| 字段名        | 类型          | 说明                             |
|---------------|---------------|----------------------------------|
| `id`          | bigint        | 订单唯一标识 (自增主键)         |
| `order_no`    | varchar(100)  | 唯一订单号 (业务ID)            |
| `customer_id` | bigint        | 关联的客户 ID                   |
| `order_date`  | datetime      | 下单日期                         |
| `status`      | varchar(50)   | 订单状态 (e.g., draft, pending, confirmed) |
| `total_amount`| decimal(10,2) | 订单总金额 (各订单项最终价格之和) |
| `final_amount`| decimal(10,2) | 最终支付金额 (可包含折扣、运费等) |
| `remark`      | text          | 订单备注                         |
| `created_at`  | datetime      | 创建时间                         |
| `updated_at`  | datetime      | 更新时间                         |
| `deleted_at`  | datetime      | 软删除字段                       |

### 订单项表 (`order_items`)

| 字段名         | 类型         | 说明                         |
|----------------|--------------|------------------------------|
| `id`           | bigint       | 订单项唯一标识 (自增主键)   |
| `order_id`     | bigint       | 关联的订单 ID                |
| `product_id`   | bigint       | 关联的产品 ID                |
| `product_name` | varchar(255) | 产品名称快照                 |
| `quantity`     | int          | 购买数量                     |
| `unit_price`   | decimal(10,2)| 成交单价 (快照)              |
| `final_price`  | decimal(10,2)| 最终价格 (quantity * unit_price) |

**设计考量**:
*   `order_items` 中的 `product_name` 和 `unit_price` 是**数据快照**，确保即使未来产品信息变更，历史订单的记录也不会改变。
*   订单的创建是**事务性**的，`orders` 和 `order_items` 的写入必须同时成功或失败。

---

## 📂 模块目录结构

```
internal/
├── controller/
│   └── order_controller.go    # 处理订单相关的 HTTP 请求
├── service/
│   └── order_service.go       # 封装订单业务逻辑，包含事务处理
├── dto/
│   └── order.go               # 订单模块的数据传输对象 (DTO)
└── routes/
    └── order.go               # 注册订单模块的 API 路由
```
*GORM Gen 生成的 `dao` 层文件未在此列出。*

---

## 🔌 接口设计 (RESTful API)

### `POST /orders`

*   **描述**: 创建一个新订单。此操作是事务性的。
*   **请求体** (`dto.OrderCreateRequest`):
    ```json
    {
      "customer_id": 1,
      "order_date": "2024-07-15T10:00:00Z",
      "status": "pending",
      "remark": "请尽快发货",
      "items": [
        {
          "product_id": 101,
          "quantity": 2,
          "unit_price": 499.50
        },
        {
          "product_id": 102,
          "quantity": 1,
          "unit_price": 899.00
        }
      ]
    }
    ```
*   **成功响应** (200 OK): 返回新创建的订单完整信息，包括其订单项 (`dto.OrderResponse`)。
*   **错误响应**:
    *   `400 Bad Request`: 客户不存在或某个产品不存在。
    *   `500 Internal Server Error`: 数据库事务失败。

### `GET /orders/:id`

*   **描述**: 根据 ID 获取单个订单的详细信息，会预加载 (Preload) 关联的订单项。
*   **成功响应** (200 OK): 返回订单及其订单项的完整信息 (`dto.OrderResponse`)。
*   **错误响应**:
    *   `404 Not Found`: 订单不存在。

### `GET /orders`

*   **描述**: 获取订单列表，支持分页、筛选和排序，同样会预加载订单项。
*   **查询参数** (`dto.OrderListRequest`):
    *   `page`, `page_size` (可选): 分页参数。
    *   `customer_id` (可选): 按客户 ID 筛选。
    *   `status` (可选): 按订单状态筛选。
    *   `order_by` (可选): 排序，例如 `order_date_desc`。
    *   `ids` (可选): 按订单ID列表批量查询。
*   **成功响应** (200 OK): 返回订单列表和总数 (`dto.OrderListResponse`)。

---

### 🔐 权限设计建议

*   **销售 (`sales`)**: 可以为自己名下的客户创建订单，查看自己创建的订单。
*   **订单管理员 (`order_admin`)**: 可以查看和管理所有订单，但可能无法修改已完成或已取消的订单。
*   **客户**: (如果未来有客户门户) 可以查看自己的历史订单。

---

### 🧪 测试建议

*   **核心逻辑**: 重点测试 `OrderService` 的 `CreateOrder` 方法，确保数据库事务在各种情况下（如客户不存在、产品不存在、库存不足等）都能正确回滚。
*   **数据一致性**: 验证创建订单后，产品快照信息（名称、单价）是否正确写入 `order_items` 表。
*   **查询性能**: 验证 `GetOrder` 和 `ListOrders` 是否能通过 `Preload` 有效避免 N+1 查询问题。
*   **边界情况**: 测试创建包含一个或多个订单项的订单。

---

### 📎 依赖说明

*   **GORM**: 用于事务处理和数据预加载。
*   **Gin**: Web 框架。
*   **`pkg/utils/randx.go`**: 用于生成唯一的订单号 `order_no`。

---

### 💡 扩展建议

*   **订单状态流转**: 引入状态机来管理订单状态的变更（例如，`pending` -> `confirmed` -> `shipped`），确保状态变更的合规性。
*   **更新与取消**: 实现 `PUT /orders/:id` 和 `DELETE /orders/:id` 接口，并加入严格的业务规则（例如，只有处于 `draft` 状态的订单才能被修改或取消）。
*   **关联客户与产品信息**: 在 `OrderResponse` 中可以进一步丰富 `CustomerName` 和每个 `OrderItem` 的 `ProductName` 等信息，提升 API 的可用性。
