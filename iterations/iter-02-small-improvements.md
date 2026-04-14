# Iteration 02 — 小修小补优化清单

> 分支：待建（基于 `main` 合并后的代码）
> 目标：不做大重构，专注细节质量——错误处理、魔法值、校验、日志、测试覆盖、接口完整性
> 执行方式：每个条目独立可落地，可单独提 PR，也可分批执行

状态说明：⬜ 待开始 / 🔄 进行中 / ✅ 已完成

---

## 分类索引

| 类别 | 条目数 |
|------|--------|
| [ERR] 错误处理 & 哨兵错误 | 8 |
| [MAGIC] 魔法数字 & 硬编码值 | 8 |
| [VALID] 入参校验缺失 | 6 |
| [LOG] 日志质量 | 6 |
| [NIL] Nil 指针风险 | 3 |
| [CACHE] 缓存键 & 缓存失效 | 3 |
| [KAFKA] Kafka 生产者健壮性 | 2 |
| [SIG] 函数签名 & 构造器 | 3 |
| [DUP] 重复代码 & 工具函数 | 4 |
| [IFACE] 仓库接口完整性 | 3 |
| [TEST] 测试覆盖 | 4 |
| **合计** | **50** |

---

## [ERR] 错误处理 & 哨兵错误

### ERR-01 OrderLog 写入错误被静默丢弃
**文件：** `internal/service/order/service/order_logic.go`（所有 `_ = l.orderLogRepo.Create(...)` 调用处）
**问题：** 订单操作日志写入失败时使用 `_ =` 忽略 error，出现数据库问题时完全无法感知。
**修复：** 替换为：
```go
if err := l.orderLogRepo.Create(ctx, &model.OrderLog{...}); err != nil {
    logx.Errorf("failed to create order log: %v", err)
}
```

---

### ERR-02 Kafka 发布失败被静默丢弃
**文件：** `internal/service/order/service/order_logic.go`（所有 `_ = l.mqProducer.PublishWithKey(...)` 调用处）
**问题：** Kafka 发布失败不记录日志，下游消费者永远收不到事件但上游无任何告警。
**修复：**
```go
if err := l.mqProducer.PublishWithKey(ctx, topic, key, payload); err != nil {
    logx.Errorf("kafka publish failed topic=%s key=%s: %v", topic, key, err)
}
```

---

### ERR-03 缓存删除错误被静默丢弃
**文件：** `internal/service/order/service/order_logic.go`（cache.Delete / DeletePattern 调用处）
**问题：** 缓存删除失败导致脏数据长期存在，但调用方完全不知情。
**修复：**
```go
if err := l.cache.Delete(ctx, key); err != nil {
    logx.Warnf("cache delete failed key=%s: %v", key, err)
}
```

---

### ERR-04 gRPC Client 返回 fmt.Errorf 丢失类型信息
**文件：** `internal/pkg/client/user.go:44,58`；`internal/pkg/client/product.go:46`；以及其他 client 文件
**问题：** 客户端层返回 `fmt.Errorf("...")` 而非 `fmt.Errorf("...: %w", sentinel)`，调用方无法用 `errors.Is()` 区分错误类型（如 not-found vs. internal error）。
**修复：**
```go
// 在 internal/pkg/client/ 中定义哨兵错误
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)
// 返回时包装：
return nil, fmt.Errorf("get user address: %w", ErrNotFound)
```

---

### ERR-05 Promotion 优惠计算出错时静默降级无日志
**文件：** `internal/service/promotion/service/promotion_logic.go:243-244`
**问题：** `if err == nil { coupon = ... }` 写法——当 GetByID 出错时，coupon 为零值，折扣计算结果为 0 且调用方不知道发生了错误。
**修复：**
```go
coupon, err := l.couponRepo.GetByID(ctx, id)
if err != nil {
    logx.Warnf("get coupon %d failed: %v, skipping discount", id, err)
    discountAmount = 0
} else {
    // 正常计算
}
```

---

### ERR-06 Inventory Consumer 行未更新没有明确错误路径
**文件：** `internal/service/inventory/service/inventory_consumer.go:89-90`
**问题：** `RowsAffected == 0` 时只打日志后 return，但没有返回 error 给 Kafka 消费框架，导致 offset 被提交，消息实际上被静默丢弃而非重试。
**修复：** 返回 `error` 而非 `nil`，让消费框架决定是否重试：
```go
if result.RowsAffected == 0 {
    return fmt.Errorf("inventory deduct no rows affected sku_id=%d qty=%d", msg.SkuID, msg.Quantity)
}
```

---

### ERR-07 Job Service CancelOrders 不检查部分失败
**文件：** `internal/service/job/service/job_logic.go`（CancelExpiredOrders 函数）
**问题：** 批量 cancel 只检查最终 error，不检查实际影响行数。若订单已被用户手动取消，批量 SQL 影响行数为 0 也不会报错，但 UnlockStock 已被调用。
**修复：** 在 CancelOrders 实现中检查 `RowsAffected` 并与传入 ID 数量对比，记录差异：
```go
if result.RowsAffected != int64(len(ids)) {
    logx.Warnf("cancel orders: expected %d, affected %d", len(ids), result.RowsAffected)
}
```

---

### ERR-08 Review CreateReview 没有包装 OrderClient 错误
**文件：** `internal/service/review/service/review_logic.go`（GetOrder 调用后）
**问题：** `orderClient.GetOrder` 失败直接 return err，原始 gRPC error 没有被包装成业务错误，调用方收到裸 gRPC status 码。
**修复：**
```go
if err != nil {
    return nil, apperrors.ConvertToGRPCError(fmt.Errorf("verify order status: %w", err))
}
```

---

## [MAGIC] 魔法数字 & 硬编码值

### MAGIC-01 Order/Payment/Inventory 中散落的状态码整数
**文件：** `internal/service/order/service/order_logic.go`；`internal/service/payment/service/payment_logic.go`；`internal/service/review/service/review_logic.go:14`
**问题：** 订单状态 `4`（已完成）、`5`（已取消）、运营商类型 `3`（系统），支付状态 `2`（已支付）分散在各服务中作为 magic number。
**修复：** 在 `internal/service/order/model/` 下新建 `constants.go`：
```go
package model

const (
    OrderStatusPending   = 1
    OrderStatusPaid      = 2
    OrderStatusShipping  = 3
    OrderStatusCompleted = 4
    OrderStatusCancelled = 5

    OperatorTypeUser   = 1
    OperatorTypeShop   = 2
    OperatorTypeSystem = 3
)
```
各服务 import 该包而非自己定义 `const orderStatusCompleted = 4`。

---

### MAGIC-02 Coupon 状态码魔法数字
**文件：** `internal/service/promotion/service/promotion_logic.go:75,112,174,180,187,245-251`
**问题：** 优惠券状态 `0`=未使用 `1`=已使用 `2`=已过期；折扣类型 `1`=固定金额 `2`=折扣百分比，全部硬编码。
**修复：** 在 `internal/service/promotion/model/constants.go` 中定义：
```go
const (
    CouponStatusUnused  = 0
    CouponStatusUsed    = 1
    CouponStatusExpired = 2

    DiscountTypeFixed      = 1
    DiscountTypePercentage = 2
)
```

---

### MAGIC-03 Cache TTL 硬编码在业务逻辑中
**文件：** `internal/service/order/service/order_logic.go`（多处 `10*time.Minute`、`5*time.Minute`）
**问题：** Cache TTL 写死在业务代码中，调整需要全局 grep，容易遗漏。
**修复：** 在 `order_logic.go` 顶部定义常量：
```go
const (
    cacheOrderDetailTTL = 10 * time.Minute
    cacheOrderListTTL   = 5 * time.Minute
)
```

---

### MAGIC-04 Inventory 低库存阈值硬编码
**文件：** `internal/service/inventory/service/inventory_logic.go`（LowStockThreshold: 10）
**问题：** 10 件的低库存警戒线写死，不同 SKU 品类可能需要不同阈值，且无法通过配置调整。
**修复：** 加到 Config：
```go
// inventory config
DefaultLowStockThreshold int `yaml:"DefaultLowStockThreshold"` // default: 10
```
NewInventoryLogic 接收 cfg 并使用 `cfg.DefaultLowStockThreshold`。

---

### MAGIC-05 Job Service 超时订单分钟数硬编码
**文件：** `internal/service/job/service/job_logic.go:48`（`TimeoutMinutes: 30`）
**问题：** 30 分钟超时写死，无法通过配置调整。促销活动时可能需要更短的超时。
**修复：**
```yaml
# job-config.yaml
OrderTimeoutMinutes: 30
```
```go
// job config struct
OrderTimeoutMinutes int `yaml:"OrderTimeoutMinutes"`
```

---

### MAGIC-06 Recommend Service Redis ZSet Score 写死
**文件：** `internal/service/recommend/repository/recommend_repo.go`（各 Get 方法中）
**问题：** `Limit: 20` 在多个 Get 方法中重复出现，且不来自调用方参数。若调用方传 limit=10，仍会从 Redis 取 20 条。
**修复：** 确保所有 Get 方法的 `ZRevRangeWithScores` 使用传入的 `limit` 参数而非固定值。

---

### MAGIC-07 Logistics 运单状态码魔法数字
**文件：** `internal/service/logistics/service/logistics_logic.go`
**问题：** 运单状态（1=已揽件、2=运输中、3=已派送、4=已签收）等以整数硬编码。
**修复：** 在 `internal/service/logistics/model/constants.go` 定义：
```go
const (
    LogisticsStatusPickedUp  = 1
    LogisticsStatusInTransit = 2
    LogisticsStatusDelivered = 3
    LogisticsStatusSigned    = 4
)
```

---

### MAGIC-08 Message Service Kafka Topic 字符串重复
**文件：** `internal/service/message/service/message_consumer.go`
**问题：** `"order.created"` 等 topic 字符串在 message consumer 中重复出现，而 `internal/pkg/mq` 中已定义了 Topic 常量。
**修复：** 统一使用 `mq.TopicOrderCreated` 等常量，不在 consumer 里重新写字符串字面量。

---

## [VALID] 入参校验缺失

### VALID-01 Order ID/No 两者皆空时直接查库
**文件：** `internal/service/order/service/order_logic.go`（GetOrder 函数头部）
**问题：** 若 `req.OrderNo == ""` 且 `req.ID == 0`，查询仍会执行并返回第一条记录（GORM WHERE 0 = id 行为取决于实现）。
**修复：**
```go
if req.OrderNo == "" && req.ID == 0 {
    return nil, apperrors.NewInvalidParamError("order_no or id is required")
}
```

---

### VALID-02 分页参数无下界校验
**文件：** `internal/service/order/service/order_logic.go`（ListOrders 函数）；同类问题存在于 review、promotion 等多个 List 方法
**问题：** `Page == 0` 时 offset 变为 `-pageSize`，可能触发 DB driver 错误；`PageSize == 0` 时 LIMIT 0 返回空结果集，调用方误以为无数据。
**修复：**
```go
if req.Page < 1 {
    req.Page = 1
}
if req.PageSize < 1 || req.PageSize > 100 {
    req.PageSize = 20
}
```
并在 Review、Promotion、Recommend 的 List 类方法中统一应用。

---

### VALID-03 Review Rating 边界未校验
**文件：** `internal/service/review/service/review_logic.go`（CreateReview 头部）
**问题：** Rating 字段只有隐式范围约定（1–5），若调用方传入 0 或 100，数据会直接写库，破坏统计汇总。
**修复：**
```go
if req.Rating < 1 || req.Rating > 5 {
    return nil, apperrors.NewInvalidParamError("rating must be between 1 and 5")
}
```

---

### VALID-04 Cart AddItem Quantity 未校验正整数
**文件：** `internal/service/cart/service/cart_logic.go`（AddItem 头部）
**问题：** Quantity ≤ 0 时会通过库存校验（`availableStock >= 0` 总为真），并写入负数数量到购物车。
**修复：**
```go
if req.Quantity <= 0 {
    return nil, apperrors.NewInvalidParamError("quantity must be >= 1")
}
```

---

### VALID-05 Payment Callback 未校验签名/来源
**文件：** `internal/service/payment/service/payment_logic.go`（PaymentCallback 函数）
**问题：** Callback 接口接收第三方支付回调，目前无签名验证（HMAC / RSA），任何人发假 callback 均可触发订单状态变更。
**修复：** 增加签名校验步骤（具体算法依支付渠道而定），至少添加注释标注 TODO 和安全风险：
```go
// TODO: verify payment provider signature before processing callback
// e.g., HMAC-SHA256 with shared secret or RSA public key
```

---

### VALID-06 Promotion ReceiveCoupon 缺少 UserID 校验
**文件：** `internal/service/promotion/service/promotion_logic.go`（ReceiveCoupon 头部）
**问题：** 若 `req.UserID == 0`，会在 user_coupon 表插入 user_id=0 的记录，破坏后续按用户查询的结果。
**修复：**
```go
if req.UserID == 0 {
    return nil, apperrors.NewInvalidParamError("user_id is required")
}
```

---

## [LOG] 日志质量

### LOG-01 Order CreateOrder 各步骤无成功日志
**文件：** `internal/service/order/service/order_logic.go`（CreateOrder 主流程）
**问题：** 整个 CreateOrder 流程（验证地址→锁库存→计算折扣→创建订单）只在 error 路径有日志，成功路径无任何追踪信息，线上问题排查极为困难。
**修复：** 在关键节点加 Info 日志：
```go
logx.Infof("order created: order_no=%s user_id=%d amount=%.2f", order.OrderNo, req.UserID, order.TotalAmount)
```

---

### LOG-02 Product/User Client 调用失败无日志
**文件：** `internal/service/order/service/order_logic.go`（SKU 信息填充段）
**问题：** `if product, err := ...; err == nil && product != nil { ... }` 写法在 err != nil 时静默跳过，SKU 信息留空且无警告。
**修复：**
```go
product, err := l.productClient.GetProduct(ctx, itemReq.SkuID)
if err != nil {
    logx.Warnf("fetch product sku_id=%d failed: %v, order item name will be empty", itemReq.SkuID, err)
} else if product != nil {
    // fill name, image
}
```

---

### LOG-03 Inventory Consumer 错误日志缺乏结构化字段
**文件：** `internal/service/inventory/service/inventory_consumer.go:89`
**问题：** 日志 `"库存扣减 MySQL 行未更新"` 没有附带 sku_id、quantity，日志系统无法聚合和告警。
**修复：**
```go
logx.WithContext(ctx).Errorw("inventory deduct no rows affected",
    logx.Field("sku_id", msg.SkuID),
    logx.Field("quantity", msg.Quantity),
)
```

---

### LOG-04 Message Consumer 处理成功无日志
**文件：** `internal/service/message/service/message_consumer.go`（所有 Handle* 函数）
**问题：** 四个事件处理函数只有错误路径有日志，消息被成功处理时无记录，无法统计消息处理吞吐量。
**修复：** 在每个 handler 入口和出口加日志：
```go
logx.Infof("message consumer: handling order_created event order_no=%s", msg.OrderNo)
```

---

### LOG-05 Job Service CancelExpiredOrders 无汇总日志
**文件：** `internal/service/job/service/job_logic.go`（CancelExpiredOrders 结束处）
**问题：** 定时任务每次执行后没有汇总日志（共处理了多少订单、成功了几个、失败了几个）。
**修复：**
```go
logx.Infof("cancel expired orders: found=%d unlocked_stocks=%d cancelled=%d",
    len(orders), unlockedCount, cancelledCount)
```

---

### LOG-06 Recommend Service 无任何日志
**文件：** `internal/service/recommend/service/recommend_logic.go`
**问题：** Get* 方法在 Redis miss 或 DB fallback 时没有日志，无法判断推荐命中率。
**修复：** 在 Redis Miss 时打 Debug 日志，在 fallback 到 DB 时打 Info 日志。

---

## [NIL] Nil 指针风险

### NIL-01 Promotion GetByID 返回 nil 后未检查即解引用
**文件：** `internal/service/promotion/service/promotion_logic.go:69`（ReceiveCoupon 中）
**问题：** `coupon, err := l.couponRepo.GetByID(ctx, id)` 若 GORM 返回 `ErrRecordNotFound`，err 被处理但 coupon 为 nil，后续 `coupon.XXX` 调用 panic。
**修复：**
```go
if err != nil {
    return nil, apperrors.ConvertToGRPCError(err)
}
if coupon == nil {
    return nil, apperrors.NewNotFoundError("coupon not found")
}
```

---

### NIL-02 Order ShipOrder LogisticsClient 为 nil 时 panic
**文件：** `internal/service/order/service/order_logic.go`（ShipOrder 函数）
**问题：** 若 LogisticsClient 初始化失败，ServiceContext 中为 nil，ShipOrder 中的 `l.logisticsClient.CreateLogistics(...)` 会 panic。
**修复：**
```go
if l.logisticsClient == nil {
    logx.Errorf("logistics client not initialized, shipping without creating logistics order")
} else {
    // create logistics
}
```

---

### NIL-03 Job CancelExpiredOrders Items 为空时调用 UnlockStock
**文件：** `internal/service/job/service/job_logic.go`（CancelExpiredOrders 内 order.Items 遍历）
**问题：** 若 order.Items 为 nil（查询时未预加载），遍历为空且 UnlockStock 不被调用，但订单仍被取消，导致库存泄漏。
**修复：**
```go
if len(order.Items) == 0 {
    logx.Warnf("cancel expired order %d: no items found, skipping stock unlock", order.ID)
}
```

---

## [CACHE] 缓存键 & 缓存失效

### CACHE-01 Redis Key 通过 fmt.Sprintf 内联构建
**文件：** `internal/service/order/service/order_logic.go`（多处 `fmt.Sprintf("%s%d", cache.KeyPrefixOrderDetail, id)`）；同类问题在 cart、recommend 等服务中也存在
**问题：** 缓存键构建逻辑分散在业务代码中，修改前缀需要全局 grep，且前缀和 ID 的拼接格式不统一（有的用冒号有的不用）。
**修复：** 在 `internal/pkg/cache/keys.go` 中集中定义键构建函数：
```go
func OrderDetailKey(id uint64) string  { return fmt.Sprintf("%s:%d", KeyPrefixOrderDetail, id) }
func OrderListKey(userID uint64) string { return fmt.Sprintf("%s:%d", KeyPrefixOrderList, userID) }
```

---

### CACHE-02 Coupon 更新后无缓存失效
**文件：** `internal/service/promotion/repository/coupon_repo.go`（Update 方法）
**问题：** Coupon 记录被 Update 后，若其他地方有按 ID 缓存优惠券详情，缓存不会被主动失效，读到旧数据。
**修复：** Update 方法执行后调用缓存删除，或将缓存依赖移到 service 层统一管理：
```go
func (r *couponRepo) Update(ctx context.Context, coupon *model.Coupon) error {
    if err := r.db.WithContext(ctx).Save(coupon).Error; err != nil {
        return err
    }
    return r.cache.Delete(ctx, couponDetailKey(coupon.ID))
}
```

---

### CACHE-03 分页查询缓存 Pattern Delete 可能误删
**文件：** `internal/service/order/service/order_logic.go`（`cache.DeletePattern(ctx, fmt.Sprintf("%s%d:*", KeyPrefixOrderList, userID))`）
**问题：** Pattern delete 在 Redis Cluster 模式下不支持（`KEYS` / `SCAN` 跨 slot 受限），且若两个不同 userID 的前缀刚好匹配同一 pattern，可能误删。
**修复：** 改为按确定性 key 删除（存储时记录具体 key），或使用 Redis tag/hash-tag 方案隔离。短期内至少添加注释说明限制。

---

## [KAFKA] Kafka 生产者健壮性

### KAFKA-01 Kafka Publish 无超时保护
**文件：** `internal/pkg/mq/kafka.go`（AsyncPublish / PublishWithKey 函数）
**问题：** `select { case p.producer.Input() <- msg: ... }` 没有超时分支，若 Kafka broker 不可达且 channel 已满，调用方 goroutine 会永久阻塞。
**修复：**
```go
select {
case p.producer.Input() <- msg:
    return nil
case <-ctx.Done():
    return fmt.Errorf("kafka publish timeout: %w", ctx.Err())
}
```

---

### KAFKA-02 Consumer 启动失败仅打印日志不终止服务
**文件：** `internal/service/message/message.go`（后台 goroutine 启动处）；`internal/service/inventory/inventory.go` 同类问题
**问题：** Consumer 以 goroutine 启动，若 Kafka 连接失败，goroutine 静默退出，服务继续运行但消息完全不被消费，无告警。
**修复：** 使用 health check 或 panic：
```go
go func() {
    if err := consumer.Start(ctx); err != nil {
        logx.Errorf("message consumer exited: %v", err)
        // 考虑通知主进程退出
    }
}()
```

---

## [SIG] 函数签名 & 构造器

### SIG-01 NewOrderLogic 11 个位置参数
**文件：** `internal/service/order/service/order_logic.go:37-50`
**问题：** `NewOrderLogic(db, orderRepo, orderItemRepo, orderLogRepo, cache, idGen, mqProducer, userClient, productClient, inventoryClient, logisticsClient, promotionClient)` — 12 个参数，可读性极差，增加新依赖需修改所有调用点。
**修复：** 引入 Deps 结构体：
```go
type OrderLogicDeps struct {
    DB             *gorm.DB
    OrderRepo      repository.OrderRepository
    OrderItemRepo  repository.OrderItemRepository
    OrderLogRepo   repository.OrderLogRepository
    Cache          *cache.CacheOperations
    IDGen          *idgen.Generator
    MQProducer     mq.Producer
    UserClient     *client.UserClient
    ProductClient  *client.ProductClient
    InvClient      *client.InventoryClient
    LogisticsClient *client.LogisticsClient
    PromotionClient *client.PromotionClient
}

func NewOrderLogic(deps OrderLogicDeps) *OrderLogic
```

---

### SIG-02 Repository 方法中分页参数裸露
**文件：** `internal/service/inventory/repository/inventory_log_repo.go`；`internal/service/review/repository/review_repo.go`；`internal/service/promotion/repository/coupon_repo.go`
**问题：** `GetBySkuID(ctx, skuID, page, pageSize)` 等方法把分页参数作为位置参数，多个仓库接口签名不统一。
**修复：** 定义统一分页选项结构：
```go
// internal/pkg/utils/pagination.go
type PageQuery struct {
    Page     int
    PageSize int
}
func (p *PageQuery) Offset() int { return (p.Page - 1) * p.PageSize }
func (p *PageQuery) Limit() int  { return p.PageSize }
```

---

### SIG-03 CartLogic 三参数构造器 nil 容错不明确
**文件：** `internal/service/cart/service/cart_logic.go`（NewCartLogic 和测试中的 nil, nil）
**问题：** `NewCartLogic(repo, nil, nil)` 在测试中可用，但若 productClient 或 invClient 为 nil 时调用 AddItem 会 panic，没有保护。
**修复：** 在 AddItem 入口添加 nil guard：
```go
if l.productClient == nil || l.invClient == nil {
    return nil, apperrors.NewInternalError("cart service not fully initialized")
}
```

---

## [DUP] 重复代码 & 工具函数

### DUP-01 总页数计算双重逻辑 Bug
**文件：** `internal/pkg/utils/pagination.go:35-39`
**问题：** 第 36 行用 `(total + pageSize - 1) / pageSize` 已正确计算向上取整，第 37-38 行又再次 `if total%pageSize > 0 { totalPages++ }`，导致 off-by-one 错误（结果多 1 页）。
**修复：**
```go
func CalcTotalPages(total int64, pageSize int) int {
    if pageSize <= 0 {
        return 0
    }
    return int((total + int64(pageSize) - 1) / int64(pageSize))
}
```

---

### DUP-02 Offset 计算在各 Repo 中重复
**文件：** `internal/service/promotion/repository/coupon_repo.go:45-46`；`internal/service/review/repository/review_repo.go:87-88`；以及其他多个仓库文件
**问题：** `offset := (page - 1) * pageSize` 在至少 6 个仓库文件中重复出现，且 `PageQuery.Offset()` 方法（SIG-02 中建议的）可以统一替代。
**修复：** 应用 SIG-02 中的 `PageQuery` 后，删除所有内联 offset 计算。

---

### DUP-03 StatusToString 转换在多个服务中重复
**文件：** `internal/service/order/service/order_logic.go`；`internal/service/payment/service/payment_logic.go`；`internal/service/logistics/service/logistics_logic.go`
**问题：** 各服务各自维护状态整数→字符串的转换（switch-case），没有共享。
**修复：** 在各服务的 model/constants.go 中添加 `StatusText(status int) string` 方法。

---

### DUP-04 Proto 响应构建的 ToProto 方法缺失
**文件：** `internal/service/cart/cart_service.go:50-53`；`internal/service/order/service/order_logic.go`（BuildOrderDetailResponse）；`internal/service/review/review_service.go`
**问题：** `for _, item := range resp.Items { items = append(items, convertDetailToProto(item)) }` 类模式在多处重复，且转换函数是私有的，无法跨层复用。
**修复：** 为各 domain model 实现 `ToProto()` 方法，收拢转换逻辑：
```go
func (d *CartItemDetail) ToProto() *cartv1.CartItem { ... }
func (o *OrderDetail) ToProto() *orderv1.Order { ... }
```

---

## [IFACE] 仓库接口完整性

### IFACE-01 Review Repository 缺少 Delete
**文件：** `internal/service/review/repository/review_repo.go:12-26`
**问题：** ReviewRepository 接口有 Create、GetByID、GetByProductID、Update、GetStats，但没有 Delete。线上下架商品的评价无法通过仓库层删除，只能绕过接口直接查库。
**修复：**
```go
Delete(ctx context.Context, id uint64) error
```

---

### IFACE-02 UserCoupon Repository 缺少批量过期
**文件：** `internal/service/promotion/repository/user_coupon_repo.go`
**问题：** 没有 `ExpireByUserID` 或 `ExpireAll` 方法，Job service 在执行优惠券过期时只能逐条更新，效率低下。
**修复：**
```go
ExpireBefore(ctx context.Context, expiredAt time.Time) (int64, error)
```
使用单条 `UPDATE user_coupons SET status=2 WHERE expire_at < ? AND status=0 LIMIT 1000`。

---

### IFACE-03 Inventory Repository 缺少批量查询
**文件：** `internal/service/inventory/repository/inventory_repo.go`
**问题：** 只有 `GetBySkuID(skuID uint64)` 单查，Order service 在 CreateOrder 时需要校验多个 SKU 库存，当前只能循环单查，N 次 DB 查询。
**修复：**
```go
GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*model.Inventory, error)
```

---

## [TEST] 测试覆盖

### TEST-01 Cart Logic 业务逻辑无测试
**文件：** `internal/service/cart/service/cart_logic_test.go`
**问题：** 现有测试只测试 `NewCartLogic(repo, nil, nil)` 构造器，没有测试 AddItem、GetCart、UpdateQuantity 的业务逻辑（库存校验、upsert 行为）。
**修复：** 使用 mock repository 补充：
- AddItem 成功路径
- AddItem 库存不足返回错误
- AddItem SKU 状态=0（下架）返回错误
- GetCart 空购物车返回空列表

---

### TEST-02 Promotion Logic 无任何测试
**文件：** `internal/service/promotion/service/`（无 _test.go）
**问题：** ReceiveCoupon 的原子防超发逻辑、CalculateDiscount 的折扣计算是最容易出 bug 的地方，却完全没有测试。
**修复：** 至少添加：
- ReceiveCoupon 正常领取
- ReceiveCoupon 超发（RowsAffected == 0）返回错误
- CalculateDiscount 固定金额折扣
- CalculateDiscount 百分比折扣

---

### TEST-03 Pagination 工具函数无测试
**文件：** `internal/pkg/utils/pagination.go`
**问题：** CalcTotalPages 存在 off-by-one bug（见 DUP-01），且没有任何测试覆盖，导致 bug 无法被自动发现。
**修复：** 添加 table-driven test：
```go
cases := []struct{ total, pageSize, want int }{ {0,10,0},{10,10,1},{11,10,2},{100,10,10} }
```

---

### TEST-04 idgen 降级路径无测试
**文件：** `internal/pkg/idgen/`
**问题：** Redis 不可用时的纳秒时间戳降级路径没有测试，不确定降级后生成的 ID 是否符合业务长度/格式要求。
**修复：** 添加测试：mock Redis 返回错误，断言 fallback 结果非空且格式合法。

---

## 执行原则

- 每个条目独立执行，改完即可 `go build ./...` 验证
- MAGIC 类改动优先做（无风险，改完就更好读）
- VALID 类改动需要同步检查现有集成测试是否依赖宽松校验
- TEST 类用 mock，不依赖真实 DB/Redis/Kafka
- 每批改动后更新本文档状态列

---

## 变更日志

| 时间 | 变更内容 |
|------|---------|
| 2026-04-14 | 扫描全量代码库，生成 50 个改进条目 |
