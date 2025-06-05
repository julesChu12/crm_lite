# 📣 营销模块设计文档（Marketing Module） - CRM Lite

## 📌 模块目标

营销模块旨在支持基于客户画像进行主动触达、优惠活动管理、以及客户运营数据追踪。其核心能力包括营销活动创建、目标客户筛选、营销记录存储与结果追踪。

---

## 🧱 数据模型设计

### 营销活动表（marketing_campaign）

| 字段名       | 类型       | 描述                       |
|--------------|------------|----------------------------|
| `id`         | UUID       | 活动唯一标识               |
| `name`       | string     | 活动名称                   |
| `type`       | string     | 活动类型（如 `sms`, `email`, `push_notification`）   |
| `status`     | string     | 活动状态（如 `draft`, `scheduled`, `active`, `completed`, `archived`）|
| `target_tags`| []string   | 目标客户标签 (JSONB or array) |
| `target_segment_id` | UUID    | 目标客户分群ID (可选, 更复杂的目标群体) |
| `content_template_id` | UUID | 内容模板ID (可选, 用于个性化消息) |
| `content`    | text       | 活动具体内容或模板变量的JSON数据 |
| `start_time` | datetime   | 计划启动时间                   |
| `end_time`   | datetime   | 计划结束时间                   |
| `actual_start_time` | datetime | 实际启动时间 (可选) |
| `actual_end_time` | datetime | 实际结束时间 (可选) |
| `created_by` | UUID       | 创建人ID |
| `updated_by` | UUID       | 最后修改人ID |
| `created_at` | datetime   | 创建时间                   |
| `updated_at` | datetime   | 更新时间                   |

### 营销触达记录表（marketing_record）

| 字段名        | 类型     | 描述                         |
|---------------|----------|------------------------------|
| `id`          | UUID     | 唯一标识                     |
| `campaign_id` | UUID     | 关联的营销活动ID             |
| `customer_id` | UUID     | 被触达客户ID                 |
| `channel`     | string   | 触达渠道 (如 `sms`, `email`) |
| `status`      | string   | 发送状态（如 `pending`, `sent`, `delivered`, `failed`, `opened`, `clicked`）        |
| `error_message`| string  | 发送失败时的错误信息 (可选) |
| `response`    | jsonb    | 客户反馈或互动数据（如：`{"action": "click", "link_id": "promo123"}`，可使用 JSONB 类型存储灵活结构）      |
| `sent_at`     | datetime | 实际发送时间                     |
| `opened_at`   | datetime | 打开时间 (邮件/推送可选) |
| `clicked_at`  | datetime | 点击时间 (邮件/短信链接可选) |

---

## 📂 模块目录结构建议

internal/marketing/
├── controller/
│   └── marketing_controller.go
├── service/
│   └── marketing_service.go
├── model/
│   ├── marketing_campaign.go
│   └── marketing_record.go
├── router/
│   └── marketing_router.go

---

## 🔌 接口设计

### 营销活动 (Campaigns)

#### POST `/api/marketing/campaigns`

- 描述：新建营销活动。
- 请求体 (JSON):

```json
{
  "name": "六月会员关怀短信",
  "type": "sms",
  "target_tags": ["VIP", "新用户"],
  // "target_segment_id": "segment-uuid", // 可选
  "content": "亲爱的[客户姓名]，您有一份6月专属好礼待领取！详情请访问[链接]。退订请回复TD。",
  // "content_template_id": "template-uuid", // 可选，如果使用模板
  "start_time": "2024-06-10T09:00:00Z",
  "end_time": "2024-06-15T23:59:59Z"
}
```

- 返回：成功时返回新创建的营销活动详情。

```json
{
  "id": "campaign-uuid",
  "name": "六月会员关怀短信",
  "type": "sms",
  "status": "draft", // 或 "scheduled" 如果 start_time 是未来
  // ... 其他字段
  "created_at": "2024-06-01T10:00:00Z"
}
```

#### GET `/api/marketing/campaigns`

- 描述：分页获取营销活动列表。
- 请求参数：
  - `status` (string, 可选)：按活动状态筛选。
  - `type` (string, 可选)：按活动类型筛选。
  - `name` (string, 可选): 按活动名称模糊搜索。
  - `page` (int, 可选)：页码。
  - `page_size` (int, 可选)：每页数量。
- 返回：

```json
{
  "campaigns": [
    {
      "id": "campaign-uuid-1",
      "name": "六月会员关怀短信",
      "type": "sms",
      "status": "active",
      "start_time": "2024-06-10T09:00:00Z",
      "end_time": "2024-06-15T23:59:59Z"
    }
  ],
  "pagination": {
    "total": 25,
    "page": 1,
    "page_size": 10
  }
}
```

#### GET `/api/marketing/campaigns/:id`

- 描述：获取指定营销活动的详细信息。
- 请求参数：无
- 返回：单个营销活动详情，结构类似 `POST /api/marketing/campaigns` 的成功响应。

#### PUT `/api/marketing/campaigns/:id`

- 描述：更新营销活动信息（通常在活动未开始前允许修改）。
- 请求体 (JSON)：包含需要更新的字段，结构类似创建请求。
- 返回：成功时返回更新后的营销活动详情。

#### POST `/api/marketing/campaigns/:id/trigger`

- 描述：手动触发营销活动执行（例如，对于非定时活动，或需要立即执行的场景）。对于已设定 `start_time` 的活动，通常由调度器自动触发。
- 请求体 (JSON, 可选)：

```json
{
  "execution_type": "actual" // "actual" 或 "simulation"
}
```

- 返回：

```json
{
  "status": "triggered", // 或 "simulation_started"
  "message": "营销活动已成功触发执行。",
  "execution_id": "exec-uuid" // 可选，用于追踪本次执行
}
```

#### POST `/api/marketing/campaigns/:id/archive`

- 描述：归档营销活动。
- 请求参数：无
- 返回：成功信息或更新后的活动状态。

### 营销触达记录 (Records)

#### GET `/api/marketing/records`

- 描述：查询营销活动的客户触达记录与反馈情况。
- 请求参数：
  - `campaign_id` (string, 必选)：关联的营销活动ID。
  - `customer_id` (string, 可选)：按客户ID筛选。
  - `status` (string, 可选)：按发送状态筛选。
  - `channel` (string, 可选): 按触达渠道筛选。
  - `page` (int, 可选)：页码。
  - `page_size` (int, 可选)：每页数量。
- 返回：

```json
{
  "records": [
    {
      "id": "record-uuid-1",
      "campaign_id": "campaign-uuid-1",
      "customer_id": "customer-uuid-123",
      "channel": "sms",
      "status": "delivered",
      "sent_at": "2024-06-10T09:05:00Z",
      "response": null
    },
    {
      "id": "record-uuid-2",
      "campaign_id": "campaign-uuid-1",
      "customer_id": "customer-uuid-456",
      "channel": "sms",
      "status": "clicked",
      "sent_at": "2024-06-10T09:06:00Z",
      "response": {"action": "click", "link_id": "promo_link"},
      "clicked_at": "2024-06-10T10:15:00Z"
    }
  ],
  "pagination": {
    "total": 1500,
    "page": 1,
    "page_size": 50
  }
}
```

---

### 📈 成效评估建议

- **触达率指标**：
  - 计划触达数 vs 实际尝试发送数
  - 发送成功率（`sent` / 尝试发送数）
  - 送达成功率（`delivered` / `sent`）
- **互动率指标** (根据渠道和内容设计)：
  - 打开率（`opened` / `delivered`），主要用于邮件、推送
  - 点击率（`clicked` / `delivered` 或 `opened`）
  - 转化率（完成目标行为数 / `delivered` 或 `clicked`），需结合业务目标定义
- **成本效益分析**：活动成本 vs 带来的收益（如订单增加、客户活跃度提升）。
- 可通过 A/B 测试不同营销内容、目标群体、发送时间等，优化营销效果。

---

### 💡 扩展建议

- **模板化与个性化**：
  - 支持消息模板管理，允许动态插入客户变量（如姓名、最近购买商品等）。
  - 利用客户画像和行为数据，实现千人千面的个性化营销内容推荐。
- **多渠道整合**：
  - 无缝对接短信、邮件、App Push、社交媒体（如微信公众号/小程序消息）等多种营销渠道。
  - 统一管理客户在各渠道的营销偏好和禁止打扰设置 (Opt-out)。
- **自动化与智能化**：
  - 结合任务调度系统（如 Cron Job）实现定时、周期性或事件触发式营销活动（例如，客户生日祝福、流失预警关怀）。
  - 利用客户分群 (Segmentation) 和客户旅程 (Customer Journey) 设计，实现自动化营销流程。
  - 引入简单的规则引擎或 AI 算法，进行营销时机和内容推荐。
- **合规性与用户体验**：
  - 严格遵守各国关于营销消息发送的法规（如 GDPR, CCPA），提供明确的退订机制。
  - 控制营销频率，避免过度打扰用户，提升用户体验。
