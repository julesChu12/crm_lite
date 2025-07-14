# 📦 产品模块设计文档 (Product Module) - CRM Lite

## 📌 模块概述

产品模块是 CRM Lite 系统的基础模块之一，负责管理系统中所有可供销售的产品信息。这包括产品的基本属性（如名称、描述、价格）、库存单位 (SKU) 以及库存数量。订单模块将直接依赖于本模块提供的数据。

---

## 🧱 数据模型设计

### 产品数据表 (`products`)

| 字段名          | 类型         | 说明                         |
|-----------------|--------------|------------------------------|
| `id`            | bigint       | 产品唯一标识 (自增主键)     |
| `name`          | varchar(255) | 产品名称                     |
| `description`   | text         | 产品详细描述                 |
| `price`         | decimal(10,2)| 建议零售价                   |
| `category`      | varchar(100) | 库存单位 (SKU)               |
| `stock_quantity`| int          | 库存数量                     |
| `created_at`    | datetime     | 创建时间                     |
| `updated_at`    | datetime     | 更新时间                     |
| `deleted_at`    | datetime     | 软删除字段 (gorm 支持)      |

**注意**: 在 DTO 和服务层，`category` 字段被映射为 `SKU`，`stock_quantity` 被映射为 `Stock`，以提供更清晰的业务含义。

---

## 📂 模块目录结构

```
internal/
├── controller/
│   └── product_controller.go    # 处理产品相关的 HTTP 请求
├── service/
│   └── product_service.go       # 封装产品业务逻辑
├── dto/
│   └── product.go               # 产品模块的数据传输对象 (DTO)
└── routes/
    └── product.go               # 注册产品模块的 API 路由
```
*`dao/model/products.gen.go` 和 `dao/query/products.gen.go` 是通过 GORM Gen 自动生成的，未在此列出。*

---

## 🔌 接口设计 (RESTful API)

### `POST /products`

*   **描述**: 创建一个新产品。
*   **请求体** (`dto.ProductCreateRequest`):
    ```json
    {
      "name": "高效能笔记本电脑",
      "description": "最新款，性能强劲",
      "price": 7999.99,
      "sku": "LAPTOP-X1-2024",
      "stock": 100
    }
    ```
*   **成功响应** (200 OK): 返回新创建的产品信息 (`dto.ProductResponse`)。
*   **错误响应**:
    *   `400 Bad Request`: 请求参数验证失败。
    *   `409 Conflict`: SKU 已存在。

### `GET /products/:id`

*   **描述**: 根据 ID 获取单个产品的详细信息。
*   **成功响应** (200 OK): 返回产品信息 (`dto.ProductResponse`)。
*   **错误响应**:
    *   `404 Not Found`: 产品不存在。

### `GET /products`

*   **描述**: 获取产品列表，支持分页、筛选和排序。
*   **查询参数** (`dto.ProductListRequest`):
    *   `page` (可选): 页码, 默认为 1。
    *   `page_size` (可选): 每页数量, 默认为 10。
    *   `name` (可选): 按产品名称进行模糊搜索。
    *   `sku` (可选): 按 SKU 进行精确搜索。
    *   `order_by` (可选): 排序，格式为 `字段名_asc` 或 `字段名_desc` (例如, `created_at_desc`)。
*   **成功响应** (200 OK): 返回产品列表和总数 (`dto.ProductListResponse`)。

### `POST /products/batch-get`

*   **描述**: 根据产品 ID 列表批量获取产品信息。
*   **请求体** (`dto.ProductBatchGetRequest`):
    ```json
    {
      "ids": [1, 2, 5]
    }
    ```
*   **成功响应** (200 OK): 返回产品列表 (`dto.ProductListResponse`)，`total` 字段可能与列表长度不完全匹配（如果某些ID不存在）。

### `PUT /products/:id`

*   **描述**: 更新一个现有产品的信息。请求体中只包含需要更新的字段。
*   **请求体** (`dto.ProductUpdateRequest`):
    ```json
    {
      "price": 7899.00,
      "stock": 95
    }
    ```
*   **成功响应** (200 OK): 返回更新后的完整产品信息 (`dto.ProductResponse`)。
*   **错误响应**:
    *   `404 Not Found`: 产品不存在。

### `DELETE /products/:id`

*   **描述**: 软删除一个产品。
*   **成功响应** (204 No Content)。
*   **错误响应**:
    *   `404 Not Found`: 产品不存在。

---

### 🔐 权限设计建议

产品信息的管理权限通常比较集中，可以设计如下角色：

*   **产品管理员 (`product_admin`)**: 拥有对产品模块所有接口 (CRUD) 的访问权限。
*   **销售 (`sales`)**: 通常拥有对产品列表和详情的只读权限 (`GET`)。
*   **系统管理员 (`admin`)**: 默认拥有所有权限。

---

### 🧪 测试建议

*   **单元测试**:
    *   `ProductService` 中对 SKU 唯一性的检查。
    *   `UpdateProduct` 的零值更新逻辑。
    *   列表查询中各种筛选和排序条件的组合。
*   **集成测试**:
    *   验证 `DELETE` 接口是否为软删除。
    *   创建产品后，立即通过 `batch-get` 接口查询，验证数据一致性。

---

### 📎 依赖说明

*   **GORM & GORM Gen**: 用于数据库操作和模型生成。
*   **Gin**: Web 框架，用于处理 HTTP 请求和路由。
*   **Casbin**: (可选) 用于实现基于角色的访问控制。

---

### 💡 扩展建议

*   **产品分类**: 可以增加一个 `categories` 表，并通过外键或中间表与 `products` 关联，实现多级分类。
*   **产品图片**: 可以增加一个 `product_images` 表，用于存储产品的主图和详情图。
*   **规格参数**: 对于复杂产品（如电子产品），可以引入一个 `product_specs` 表（或使用 JSONB 字段）来存储键值对形式的规格参数。 