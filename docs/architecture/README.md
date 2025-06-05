# CRM Lite 架构设计文档

本文档目录包含了 CRM Lite 项目核心模块的架构设计和技术说明。

## 目录索引

- [**01_project-structure.md**](./01_project-structure.md)
  - 项目整体顶层结构、技术栈选型以及 `internal/` 模块的基本组织方式。
- [**02_module_customer.md**](./02_module_customer.md)
  - 客户模块的核心功能、数据模型、接口设计和权限建议。
- [**03_module_order.md**](./03_module_order.md)
  - 订单模块的业务流程、数据模型（订单主表、明细表）、接口设计和相关建议。
- [**04_module_wallet.md**](./04_module_wallet.md)
  - 钱包模块（余额、积分）的数据模型、核心接口（充值、消费、流水查询）设计。
- [**05_module_marketing.md**](./05_module_marketing.md)
  - 营销活动模块的数据模型、接口设计（活动创建、触发、记录查询）以及成效评估建议。
- [**06_module_common.md**](./06_module_common.md)
  - 通用基础模块的设计，包括统一响应、错误码、日志、分页、时间工具等，并附带了部分Go代码实现参考。

---

请根据需要查阅对应模块的详细设计文档。
