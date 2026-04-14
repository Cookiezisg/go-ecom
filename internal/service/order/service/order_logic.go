package service

import (
	"context"
	"fmt"
	"time"

	"ecommerce-system/internal/pkg/cache"
	"ecommerce-system/internal/pkg/client"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/idgen"
	"ecommerce-system/internal/pkg/mq"
	"ecommerce-system/internal/service/order/model"
	"ecommerce-system/internal/service/order/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// OrderLogic 订单业务逻辑
type OrderLogic struct {
	db               *gorm.DB
	orderRepo        repository.OrderRepository
	orderItemRepo    repository.OrderItemRepository
	orderLogRepo     repository.OrderLogRepository
	cache            *cache.CacheOperations
	idGen            *idgen.Generator
	mqProducer       *mq.Producer
	userClient       *client.UserClient
	productClient    *client.ProductClient
	invClient        *client.InventoryClient
	logisticsClient  *client.LogisticsClient
	promotionClient  *client.PromotionClient
}

// NewOrderLogic 创建订单业务逻辑
func NewOrderLogic(
	db *gorm.DB,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	orderLogRepo repository.OrderLogRepository,
	cache *cache.CacheOperations,
	idGen *idgen.Generator,
	mqProducer *mq.Producer,
	userClient *client.UserClient,
	productClient *client.ProductClient,
	invClient *client.InventoryClient,
	logisticsClient *client.LogisticsClient,
	promotionClient *client.PromotionClient,
) *OrderLogic {
	return &OrderLogic{
		db:              db,
		orderRepo:       orderRepo,
		orderItemRepo:   orderItemRepo,
		orderLogRepo:    orderLogRepo,
		cache:           cache,
		idGen:           idGen,
		mqProducer:      mqProducer,
		userClient:      userClient,
		productClient:   productClient,
		invClient:       invClient,
		logisticsClient: logisticsClient,
		promotionClient: promotionClient,
	}
}

// -----------------------------------------------------------------------
// CreateOrder
// -----------------------------------------------------------------------

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID    uint64
	AddressID uint64
	Items     []OrderItemRequest
	OrderType int8
	CouponID  uint64
	Remark    string
}

// OrderItemRequest 订单商品项请求
type OrderItemRequest struct {
	SkuID    uint64
	Quantity int
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	Order *model.Order
	Items []*model.OrderItem
}

// CreateOrder 创建订单
// 流程：获取地址 → 查 SKU 信息 → 计算金额 → 事务写库 → 锁定库存 → 发 Kafka
func (l *OrderLogic) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 1. 参数校验
	if req.UserID == 0 {
		return nil, apperrors.NewInvalidParamError("用户ID不能为空")
	}
	if req.AddressID == 0 {
		return nil, apperrors.NewInvalidParamError("收货地址ID不能为空")
	}
	if len(req.Items) == 0 {
		return nil, apperrors.NewInvalidParamError("订单商品项不能为空")
	}

	// 2. 获取收货地址
	var receiverName, receiverPhone, receiverAddress string
	if l.userClient != nil {
		addr, err := l.userClient.GetAddressByID(ctx, int64(req.UserID), int64(req.AddressID))
		if err != nil {
			return nil, apperrors.NewError(apperrors.CodeAddressNotFound, "获取收货地址失败: "+err.Error())
		}
		receiverName = addr.ReceiverName
		receiverPhone = addr.ReceiverPhone
		receiverAddress = fmt.Sprintf("%s%s%s%s", addr.Province, addr.City, addr.District, addr.Detail)
	}

	// 3. 查询 SKU 信息，构建订单项，计算总金额
	items := make([]*model.OrderItem, 0, len(req.Items))
	var totalAmount float64

	for _, itemReq := range req.Items {
		if itemReq.Quantity <= 0 {
			return nil, apperrors.NewInvalidParamError("商品数量必须大于0")
		}

		var (
			productID   uint64
			productName string
			skuCode     string
			skuName     string
			skuImage    *string
			skuSpecs    model.JSONMap
			price       float64
		)

		if l.productClient != nil {
			sku, err := l.productClient.GetSku(ctx, int64(itemReq.SkuID))
			if err != nil {
				return nil, apperrors.NewError(apperrors.CodeSkuNotFound, "查询SKU信息失败: "+err.Error())
			}
			if sku.Status != 1 {
				return nil, apperrors.NewError(apperrors.CodeSkuOffline, fmt.Sprintf("SKU %d 已下架", itemReq.SkuID))
			}

			productID = uint64(sku.ProductId)
			skuCode = sku.SkuCode
			skuName = sku.Name
			if sku.Image != "" {
				img := sku.Image
				skuImage = &img
			}
			price = sku.Price

			// 规格转换
			if len(sku.Specs) > 0 {
				skuSpecs = make(model.JSONMap)
				for k, v := range sku.Specs {
					skuSpecs[k] = v
				}
			}

			// 获取商品名
			if productID > 0 {
				product, err := l.productClient.GetProduct(ctx, int64(productID))
				if err == nil && product != nil {
					productName = product.Name
				}
			}
		}

		itemAmount := price * float64(itemReq.Quantity)
		totalAmount += itemAmount

		items = append(items, &model.OrderItem{
			ProductID:   productID,
			ProductName: productName,
			SkuID:       itemReq.SkuID,
			SkuCode:     skuCode,
			SkuName:     skuName,
			SkuImage:    skuImage,
			SkuSpecs:    skuSpecs,
			Price:       price,
			Quantity:    itemReq.Quantity,
			TotalAmount: itemAmount,
		})
	}

	// 4. 生成订单号
	orderNo := l.idGen.OrderNo(ctx)

	// 4.5 计算优惠金额（promotion service 可选）
	discountAmount := 0.0
	payAmount := totalAmount
	if req.CouponID > 0 && l.promotionClient != nil {
		productIDs := make([]int64, 0, len(items))
		quantities := make([]int32, 0, len(items))
		for _, item := range items {
			productIDs = append(productIDs, int64(item.ProductID))
			quantities = append(quantities, int32(item.Quantity))
		}
		disc, final, err := l.promotionClient.CalculateDiscount(ctx, int64(req.UserID), productIDs, quantities, int64(req.CouponID), totalAmount)
		if err != nil {
			logx.Errorf("计算优惠金额失败 coupon_id=%d: %v，忽略优惠继续下单", req.CouponID, err)
		} else {
			discountAmount = disc
			payAmount = final
		}
	}

	// 5. 构建订单
	order := &model.Order{
		OrderNo:         orderNo,
		UserID:          req.UserID,
		OrderType:       req.OrderType,
		Status:          model.OrderStatusPending,
		TotalAmount:     totalAmount,
		PayAmount:       payAmount,
		DiscountAmount:  discountAmount,
		FreightAmount:   0,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ReceiverAddress: receiverAddress,
	}
	if req.Remark != "" {
		order.Remark = &req.Remark
	}

	// 6. 事务写库：订单 + 订单项
	err := l.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for _, item := range items {
			item.OrderID = order.ID
			item.OrderNo = orderNo
		}
		return tx.Create(&items).Error
	})
	if err != nil {
		return nil, apperrors.NewInternalError("创建订单失败: " + err.Error())
	}

	// 7. 锁定库存（下单成功后逐个锁库存，有任一失败则回退已锁的）
	if l.invClient != nil {
		var lockedItems []OrderItemRequest
		for _, item := range items {
			lockErr := l.invClient.LockStock(ctx, int64(item.SkuID), int32(item.Quantity), int64(order.ID), "下单锁库存")
			if lockErr != nil {
				// 回退已锁库存
				for _, locked := range lockedItems {
					_ = l.invClient.UnlockStock(ctx, int64(locked.SkuID), int32(locked.Quantity), int64(order.ID), "锁库存失败回退")
				}
				// 取消订单
				_ = l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCancelled, strPtr("库存不足"))
				return nil, apperrors.NewError(apperrors.CodeStockInsufficient, "库存不足: "+lockErr.Error())
			}
			lockedItems = append(lockedItems, OrderItemRequest{SkuID: item.SkuID, Quantity: item.Quantity})
		}
	}

	// 8. 清除用户订单列表缓存
	if l.cache != nil {
		_ = l.cache.DeletePattern(ctx, fmt.Sprintf("%s%d:*", cache.KeyPrefixOrderList, order.UserID))
	}

	// 9. 记录订单日志
	operatorID := order.UserID
	afterStatus := order.Status
	_ = l.orderLogRepo.Create(ctx, &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      orderNo,
		OperatorType: 1,
		OperatorID:   &operatorID,
		Action:       "create",
		AfterStatus:  &afterStatus,
	})

	// 10. 发 Kafka 事件
	if l.mqProducer != nil {
		msg := mq.NewMessage(mq.TopicOrderCreated, map[string]interface{}{
			"order_id":     order.ID,
			"order_no":     orderNo,
			"user_id":      req.UserID,
			"total_amount": order.TotalAmount,
			"created_at":   time.Now().Format(time.RFC3339),
		})
		_ = l.mqProducer.PublishWithKey(ctx, mq.TopicOrderCreated, orderNo, msg)
	}

	return &CreateOrderResponse{Order: order, Items: items}, nil
}

// -----------------------------------------------------------------------
// GetOrder
// -----------------------------------------------------------------------

// GetOrderRequest 获取订单详情请求
type GetOrderRequest struct {
	ID      uint64
	OrderNo string
}

// GetOrderResponse 获取订单详情响应
type GetOrderResponse struct {
	Order      *model.Order
	OrderItems []*model.OrderItem
}

// GetOrder 获取订单详情（缓存优先）
func (l *OrderLogic) GetOrder(ctx context.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	var orderID uint64

	if req.ID > 0 {
		orderID = req.ID
	} else if req.OrderNo != "" {
		o, err := l.orderRepo.GetByOrderNo(ctx, req.OrderNo)
		if err != nil {
			return nil, apperrors.NewInternalError("查询订单失败: " + err.Error())
		}
		if o == nil {
			return nil, apperrors.NewError(apperrors.CodeOrderNotFound)
		}
		orderID = o.ID
	} else {
		return nil, apperrors.NewInvalidParamError("订单ID或订单号不能为空")
	}

	// 缓存读取
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixOrderDetail, orderID)
		var cached GetOrderResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	order, err := l.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单失败: " + err.Error())
	}
	if order == nil {
		return nil, apperrors.NewError(apperrors.CodeOrderNotFound)
	}

	orderItems, err := l.orderItemRepo.GetByOrderID(ctx, order.ID)
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单商品项失败: " + err.Error())
	}

	resp := &GetOrderResponse{Order: order, OrderItems: orderItems}

	// 写缓存（10 分钟）
	if l.cache != nil {
		cacheKey := cache.BuildKey(cache.KeyPrefixOrderDetail, orderID)
		_ = l.cache.Set(ctx, cacheKey, resp, 10*time.Minute)
	}

	return resp, nil
}

// -----------------------------------------------------------------------
// ListOrders
// -----------------------------------------------------------------------

// ListOrdersRequest 获取订单列表请求
type ListOrdersRequest struct {
	UserID   uint64
	Status   int8
	Page     int
	PageSize int
}

// ListOrdersResponse 获取订单列表响应
type ListOrdersResponse struct {
	Orders     []*model.Order
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

// ListOrders 获取订单列表（缓存优先）
func (l *OrderLogic) ListOrders(ctx context.Context, req *ListOrdersRequest) (*ListOrdersResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	cacheKey := fmt.Sprintf("%s%d:%d:%d", cache.KeyPrefixOrderList, req.UserID, req.Status, req.Page)
	if l.cache != nil {
		var cached ListOrdersResponse
		if err := l.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	orders, total, err := l.orderRepo.List(ctx, &repository.ListOrdersRequest{
		UserID:   req.UserID,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, apperrors.NewInternalError("查询订单列表失败: " + err.Error())
	}

	for _, o := range orders {
		items, err := l.orderItemRepo.GetByOrderID(ctx, o.ID)
		if err != nil {
			logx.Errorf("加载订单项失败 order_id=%d: %v", o.ID, err)
			o.Items = []model.OrderItem{}
		} else {
			orderItems := make([]model.OrderItem, len(items))
			for i, item := range items {
				if item != nil {
					orderItems[i] = *item
				}
			}
			o.Items = orderItems
		}
	}

	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	resp := &ListOrdersResponse{
		Orders:     orders,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	if l.cache != nil {
		_ = l.cache.Set(ctx, cacheKey, resp, 5*time.Minute)
	}

	return resp, nil
}

// -----------------------------------------------------------------------
// CancelOrder
// -----------------------------------------------------------------------

// CancelOrderRequest 取消订单请求
type CancelOrderRequest struct {
	ID      uint64
	OrderNo string
	Reason  string
}

// CancelOrderResponse 取消订单响应
type CancelOrderResponse struct {
	Success bool
}

// CancelOrder 取消订单，同时解锁库存（适用于待支付状态）
func (l *OrderLogic) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*CancelOrderResponse, error) {
	getResp, err := l.GetOrder(ctx, &GetOrderRequest{ID: req.ID, OrderNo: req.OrderNo})
	if err != nil {
		return nil, err
	}
	order := getResp.Order

	if order.Status != model.OrderStatusPending {
		return nil, apperrors.NewError(apperrors.CodeOrderStatusError, "只能取消待支付订单")
	}

	reason := req.Reason
	if err := l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCancelled, &reason); err != nil {
		return nil, apperrors.NewInternalError("取消订单失败: " + err.Error())
	}

	// 解锁库存（最大努力，失败只记录日志不回滚）
	if l.invClient != nil {
		for _, item := range getResp.OrderItems {
			if unlockErr := l.invClient.UnlockStock(ctx, int64(item.SkuID), int32(item.Quantity), int64(order.ID), "取消订单释放库存"); unlockErr != nil {
				logx.Errorf("解锁库存失败 order_id=%d sku_id=%d: %v", order.ID, item.SkuID, unlockErr)
			}
		}
	}

	l.invalidateOrderCache(ctx, order)

	afterStatus := model.OrderStatusCancelled
	_ = l.orderLogRepo.Create(ctx, &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 1,
		Action:       "cancel",
		BeforeStatus: &order.Status,
		AfterStatus:  &afterStatus,
		Remark:       &reason,
	})

	if l.mqProducer != nil {
		msg := mq.NewMessage(mq.TopicOrderCancelled, map[string]interface{}{
			"order_id": order.ID,
			"order_no": order.OrderNo,
			"reason":   reason,
		})
		_ = l.mqProducer.PublishWithKey(ctx, mq.TopicOrderCancelled, order.OrderNo, msg)
	}

	return &CancelOrderResponse{Success: true}, nil
}

// -----------------------------------------------------------------------
// ConfirmReceive
// -----------------------------------------------------------------------

// ConfirmReceiveRequest 确认收货请求
type ConfirmReceiveRequest struct {
	ID      uint64
	OrderNo string
}

// ConfirmReceiveResponse 确认收货响应
type ConfirmReceiveResponse struct {
	Success bool
}

// ConfirmReceive 确认收货（待收货→已完成）
func (l *OrderLogic) ConfirmReceive(ctx context.Context, req *ConfirmReceiveRequest) (*ConfirmReceiveResponse, error) {
	getResp, err := l.GetOrder(ctx, &GetOrderRequest{ID: req.ID, OrderNo: req.OrderNo})
	if err != nil {
		return nil, err
	}
	order := getResp.Order

	if order.Status != model.OrderStatusShipped {
		return nil, apperrors.NewError(apperrors.CodeOrderStatusError, "只能确认待收货的订单")
	}

	if err := l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusCompleted, nil); err != nil {
		return nil, apperrors.NewInternalError("确认收货失败: " + err.Error())
	}

	l.invalidateOrderCache(ctx, order)

	afterStatus := model.OrderStatusCompleted
	_ = l.orderLogRepo.Create(ctx, &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 1,
		Action:       "confirm_receive",
		BeforeStatus: &order.Status,
		AfterStatus:  &afterStatus,
	})

	return &ConfirmReceiveResponse{Success: true}, nil
}

// -----------------------------------------------------------------------
// PayOrder（支付回调通知）
// -----------------------------------------------------------------------

// PayOrderRequest 支付通知请求
type PayOrderRequest struct {
	OrderID       uint64
	OrderNo       string
	PaymentMethod int8
	PaymentNo     string
}

// PayOrder 支付成功，更新订单状态（待支付→待发货），扣减库存
func (l *OrderLogic) PayOrder(ctx context.Context, req *PayOrderRequest) error {
	getResp, err := l.GetOrder(ctx, &GetOrderRequest{ID: req.OrderID, OrderNo: req.OrderNo})
	if err != nil {
		return err
	}
	order := getResp.Order

	if order.Status != model.OrderStatusPending {
		// 幂等：已支付不报错
		if order.Status == model.OrderStatusPaid {
			return nil
		}
		return apperrors.NewError(apperrors.CodeOrderStatusError, "订单状态不允许支付操作")
	}

	now := time.Now()
	order.Status = model.OrderStatusPaid
	order.PaymentMethod = &req.PaymentMethod
	order.PaymentTime = &now
	if err := l.orderRepo.Update(ctx, order); err != nil {
		return apperrors.NewInternalError("更新订单状态失败: " + err.Error())
	}

	// 扣减库存（锁定库存→已售，最大努力，失败记录日志）
	if l.invClient != nil {
		for _, item := range getResp.OrderItems {
			if deductErr := l.invClient.DeductStock(ctx, int64(item.SkuID), int32(item.Quantity), int64(order.ID), "支付成功扣减库存"); deductErr != nil {
				logx.Errorf("扣减库存失败 order_id=%d sku_id=%d: %v", order.ID, item.SkuID, deductErr)
			}
		}
	}

	l.invalidateOrderCache(ctx, order)

	afterStatus := model.OrderStatusPaid
	_ = l.orderLogRepo.Create(ctx, &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 3, // 系统
		Action:       "pay",
		BeforeStatus: func() *int8 { s := model.OrderStatusPending; return &s }(),
		AfterStatus:  &afterStatus,
		Remark:       strPtr("支付单号: " + req.PaymentNo),
	})

	return nil
}

// -----------------------------------------------------------------------
// ShipOrder（发货通知）
// -----------------------------------------------------------------------

// ShipOrderRequest 发货请求
type ShipOrderRequest struct {
	OrderID    uint64
	OrderNo    string
	TrackingNo string
	Carrier    string
}

// ShipOrder 发货，更新订单状态（待发货→待收货），并在物流服务中创建运单
func (l *OrderLogic) ShipOrder(ctx context.Context, req *ShipOrderRequest) error {
	getResp, err := l.GetOrder(ctx, &GetOrderRequest{ID: req.OrderID, OrderNo: req.OrderNo})
	if err != nil {
		return err
	}
	order := getResp.Order

	if order.Status != model.OrderStatusPaid {
		return apperrors.NewError(apperrors.CodeOrderStatusError, "只能对待发货订单执行发货操作")
	}

	// 在物流服务中创建运单（可选，失败不阻断发货主流程）
	if l.logisticsClient != nil {
		logisticsNo, logErr := l.logisticsClient.CreateLogistics(
			ctx,
			int64(order.ID),
			order.OrderNo,
			req.Carrier,
			order.ReceiverName,
			order.ReceiverPhone,
			order.ReceiverAddress,
		)
		if logErr != nil {
			logx.Errorf("创建物流运单失败 order_id=%d: %v，继续发货", order.ID, logErr)
		} else {
			logx.Infof("物流运单已创建 order_id=%d logistics_no=%s", order.ID, logisticsNo)
		}
	}

	now := time.Now()
	order.Status = model.OrderStatusShipped
	order.DeliveryTime = &now
	if err := l.orderRepo.Update(ctx, order); err != nil {
		return apperrors.NewInternalError("更新发货状态失败: " + err.Error())
	}

	l.invalidateOrderCache(ctx, order)

	remark := fmt.Sprintf("快递: %s 单号: %s", req.Carrier, req.TrackingNo)
	afterStatus := model.OrderStatusShipped
	_ = l.orderLogRepo.Create(ctx, &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 2, // 商家
		Action:       "ship",
		BeforeStatus: func() *int8 { s := model.OrderStatusPaid; return &s }(),
		AfterStatus:  &afterStatus,
		Remark:       &remark,
	})

	return nil
}

// -----------------------------------------------------------------------
// RefundOrder（退款通知）
// -----------------------------------------------------------------------

// RefundOrderRequest 退款通知请求
type RefundOrderRequest struct {
	OrderID uint64
	OrderNo string
	Reason  string
}

// RefundOrder 退款完成，更新订单状态→已退款，回退库存
func (l *OrderLogic) RefundOrder(ctx context.Context, req *RefundOrderRequest) error {
	getResp, err := l.GetOrder(ctx, &GetOrderRequest{ID: req.OrderID, OrderNo: req.OrderNo})
	if err != nil {
		return err
	}
	order := getResp.Order

	// 已退款幂等
	if order.Status == model.OrderStatusRefunded {
		return nil
	}
	if order.Status != model.OrderStatusPaid && order.Status != model.OrderStatusShipped {
		return apperrors.NewError(apperrors.CodeOrderStatusError, "当前订单状态不可退款")
	}

	beforeStatus := order.Status
	if err := l.orderRepo.UpdateStatus(ctx, order.ID, model.OrderStatusRefunded, strPtr(req.Reason)); err != nil {
		return apperrors.NewInternalError("更新退款状态失败: " + err.Error())
	}

	// 回退库存（最大努力）
	if l.invClient != nil {
		for _, item := range getResp.OrderItems {
			if rbErr := l.invClient.RollbackStock(ctx, int64(item.SkuID), int32(item.Quantity), int64(order.ID), "退款回退库存"); rbErr != nil {
				logx.Errorf("回退库存失败 order_id=%d sku_id=%d: %v", order.ID, item.SkuID, rbErr)
			}
		}
	}

	l.invalidateOrderCache(ctx, order)

	afterStatus := model.OrderStatusRefunded
	_ = l.orderLogRepo.Create(ctx, &model.OrderLog{
		OrderID:      order.ID,
		OrderNo:      order.OrderNo,
		OperatorType: 3,
		Action:       "refund",
		BeforeStatus: &beforeStatus,
		AfterStatus:  &afterStatus,
		Remark:       strPtr(req.Reason),
	})

	return nil
}

// -----------------------------------------------------------------------
// 内部辅助方法
// -----------------------------------------------------------------------

// invalidateOrderCache 清除订单相关缓存
func (l *OrderLogic) invalidateOrderCache(ctx context.Context, order *model.Order) {
	if l.cache == nil {
		return
	}
	_ = l.cache.Delete(ctx, cache.BuildKey(cache.KeyPrefixOrderDetail, order.ID))
	_ = l.cache.DeletePattern(ctx, fmt.Sprintf("%s%d:*", cache.KeyPrefixOrderList, order.UserID))
}

func strPtr(s string) *string {
	return &s
}
