# Iteration 01 — 全量基础重构

> 来源：原 `REFACTOR_PLAN.md`
> 分支：`feature/refactor-all-services`
> 完成时间：2026-04-14
> 合并目标：`main`

---

分支：`feature/refactor-all-services`
合并目标：`main`

状态说明：⬜ 待开始 / 🔄 进行中 / ✅ 已完成

---

## Step 1 — 基础设施层 & 公共层修复

> 目标：把地基打稳，后续所有服务都依赖这层。

| # | 任务 | 状态 | 备注 |
|---|------|------|------|
| 1.1 | `internal/pkg/errors` — 扩展所有业务错误码，统一 `ConvertToGRPCError` 包级函数 | ✅ | 删除各服务重复的 convertError |
| 1.2 | `internal/pkg/cache` — 修复 `redis.Nil` 比较方式，补充库存原子扣减 Lua 脚本 | ✅ | 新增 `AtomicDeductStock` / `AtomicRollbackStock` / `IsNil` / `SetNX` |
| 1.3 | `internal/pkg/idgen`（新建）— 统一单号生成器（订单号/支付单号/退款单号） | ✅ | 新建 `idgen.Generator`，基于 Redis INCR + 日期前缀，降级用纳秒时间戳 |
| 1.4 | 所有服务 `NewServiceContext` — DB/Redis 初始化失败改为 `log.Fatal`，不静默放行 | ✅ | 新增 `MustNewMySQL` / `MustNewRedis`，11 个服务全部更新 |
| 1.5 | `internal/pkg/middleware` — 统一 gRPC auth interceptor，各服务复用 | ✅ | 新增 `AuthInterceptor` + `RequireAuthInterceptor`，删除 user/cart 各自的 interceptor 目录 |

---

## Step 2 — 服务间通信层 + 核心业务主链路

> 目标：打通「下单 → 支付 → 库存」这条最核心的链路。

| # | 任务 | 状态 | 备注 |
|---|------|------|------|
| 2.1 | `internal/pkg/client`（新建）— 封装各下游服务 gRPC client，带超时/重试 | ✅ | user / product / inventory / order client，统一 RpcConf，默认 5s 超时 |
| 2.2 | **Order service 完整重写** — CreateOrder 调用 user/product/inventory 服务，事务包裹，补全 PayOrder / ShipOrder / RefundOrder | ✅ | 地址/SKU/金额全部真实填充；CancelOrder 联动解锁库存；PayOrder/ShipOrder/RefundOrder 新增 |
| 2.3 | **Payment service 完整重写** — 单号走 idgen，Callback 联动更新订单状态 + 库存 | ✅ | PaymentNo/RefundNo 换 idgen；Callback 成功→order.PayOrder，失败→order.CancelOrder；Refund→order.RefundOrder |
| 2.4 | **Inventory service 修复** — DeductStock 改 Lua 原子扣减，补全 Kafka 消费者同步 MySQL | ✅ | AtomicDeductStock 替换旧 TOCTOU；LockStock/UnlockStock/RollbackStock 同步 Redis；新增 InventoryConsumer |

---

## Step 3 — 剩余业务服务补全

> 目标：把所有半成品服务做到可用。

| # | 任务 | 状态 | 备注 |
|---|------|------|------|
| 3.1 | **Cart service** — AddItem 改 upsert，加库存校验，返回带价格的商品信息 | ✅ | SKU 状态/库存校验；AddItem/GetCart 返回 price/sku_name/stock_status；proto 新增字段并重新生成 |
| 3.2 | **Promotion service** — 补全优惠券领取/核销/过期逻辑，接入 CreateOrder 抵扣计算 | ✅ | ReceiveCoupon 改原子 SQL 防超发；CreateOrder 接 PromotionClient 计算折扣；新增 PromotionClient |
| 3.3 | **Logistics service** — 补全运单创建/轨迹查询/签收流程，ShipOrder 时联动建运单 | ✅ | 文件名 typo 修复；物流单号改 idgen.LogisticsNo；新增 LogisticsClient；ShipOrder 联动建运单 |
| 3.4 | **Message service** — 补全消息推送，Kafka consumer 消费订单/支付事件后发通知 | ✅ | 新增 MessageConsumer；消费 order.created/cancelled/payment.success/refunded 事件；后台 goroutine 启动 |
| 3.5 | **Review service** — 校验订单已完成才能评价，补全评分统计汇总 | ✅ | CreateReview 接 OrderClient 校验状态=4（已完成）及订单归属 |
| 3.6 | **Recommend service** — 全部换强类型结构体，补全基于历史行为的推荐逻辑 | ✅ | 仓库接口改 []*RecommendItem 强类型；Redis ZSet 存储热门/个性化/相似/实时推荐 |
| 3.7 | **Job service** — 补全定时任务：超时订单自动取消、优惠券过期处理、低库存告警 | ✅ | CancelExpiredOrders 先调 InvClient.UnlockStock 再批量取消；GenerateStatistics 补逻辑骨架 |

---

## 执行原则

- 每个 Step 结束后代码必须能 `go build ./...` 通过
- 不保留向后兼容 shim，直接改干净
- proto 文件缺字段直接加，不绕路
- 改完每个大任务后更新本文档状态

---

## 变更日志

| 时间 | 变更内容 |
|------|---------|
| 2026-04-14 | 创建计划文档，建立 feature 分支 |
| 2026-04-14 | Step 1 完成：pkg/errors、cache、idgen、middleware 重写；所有 ServiceContext 修复；`go build ./...` 通过 |
| 2026-04-14 | Step 2 完成：pkg/client 新建；Order/Payment/Inventory service 完整重写；`go build ./...` 通过 |
| 2026-04-14 | Step 3 完成：Cart/Promotion/Logistics/Message/Review/Recommend/Job service 全部补全；`go build ./...` 通过 |
