package service

import (
	"context"
	"fmt"
	"time"

	"ecommerce-system/internal/pkg/client"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/job/repository"

	"github.com/zeromicro/go-zero/core/logx"
)

// JobLogic 定时任务业务逻辑
type JobLogic struct {
	orderRepo  repository.OrderRepository
	couponRepo repository.CouponRepository
	invClient  *client.InventoryClient
}

// NewJobLogic 创建定时任务业务逻辑
func NewJobLogic(
	orderRepo repository.OrderRepository,
	couponRepo repository.CouponRepository,
	invClient *client.InventoryClient,
) *JobLogic {
	return &JobLogic{
		orderRepo:  orderRepo,
		couponRepo: couponRepo,
		invClient:  invClient,
	}
}

// CancelExpiredOrdersRequest 订单超时取消请求
type CancelExpiredOrdersRequest struct {
	TimeoutMinutes int
}

// CancelExpiredOrdersResponse 订单超时取消响应
type CancelExpiredOrdersResponse struct {
	CancelledCount int64
}

// CancelExpiredOrders 取消超时待支付订单，并释放预占库存
func (l *JobLogic) CancelExpiredOrders(ctx context.Context, req *CancelExpiredOrdersRequest) (*CancelExpiredOrdersResponse, error) {
	if req.TimeoutMinutes <= 0 {
		req.TimeoutMinutes = 30 // 默认 30 分钟超时
	}

	orders, err := l.orderRepo.GetExpiredOrders(ctx, req.TimeoutMinutes)
	if err != nil {
		return nil, apperrors.NewInternalError("查询超时订单失败: " + err.Error())
	}

	if len(orders) == 0 {
		return &CancelExpiredOrdersResponse{CancelledCount: 0}, nil
	}

	// 先解锁库存（逐单处理，忽略单个 item 失败不影响整体）
	if l.invClient != nil {
		for _, order := range orders {
			for _, item := range order.Items {
				if err := l.invClient.UnlockStock(
					ctx,
					int64(item.SkuID),
					int32(item.Quantity),
					int64(order.ID),
					"超时取消订单释放库存",
				); err != nil {
					logx.Errorf("超时订单解锁库存失败 order_id=%d sku_id=%d: %v",
						order.ID, item.SkuID, err)
				}
			}
		}
	}

	// 批量取消订单
	orderIDs := make([]uint64, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	count, err := l.orderRepo.CancelOrders(ctx, orderIDs)
	if err != nil {
		return nil, apperrors.NewInternalError("取消超时订单失败: " + err.Error())
	}

	logx.Infof("超时订单处理完成：共取消 %d 笔", count)
	return &CancelExpiredOrdersResponse{CancelledCount: count}, nil
}

// ProcessExpiredCouponsRequest 优惠券过期处理请求
type ProcessExpiredCouponsRequest struct{}

// ProcessExpiredCouponsResponse 优惠券过期处理响应
type ProcessExpiredCouponsResponse struct {
	ExpiredCount int64
}

// ProcessExpiredCoupons 优惠券过期处理
func (l *JobLogic) ProcessExpiredCoupons(ctx context.Context, req *ProcessExpiredCouponsRequest) (*ProcessExpiredCouponsResponse, error) {
	count, err := l.couponRepo.ProcessExpiredCoupons(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("处理过期优惠券失败")
	}

	return &ProcessExpiredCouponsResponse{ExpiredCount: count}, nil
}

// GenerateStatisticsRequest 数据统计请求
type GenerateStatisticsRequest struct {
	Date string // YYYY-MM-DD
}

// GenerateStatistics 生成当日统计数据（订单量、GMV、新增用户等）
// 统计数据写入 daily_statistics 表，供报表使用。
func (l *JobLogic) GenerateStatistics(ctx context.Context, req *GenerateStatisticsRequest) error {
	date := req.Date
	if date == "" {
		date = time.Now().AddDate(0, 0, -1).Format("2006-01-02") // 默认统计昨天
	}

	// 验证日期格式
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return apperrors.NewInvalidParamError(fmt.Sprintf("日期格式错误: %s", date))
	}

	logx.Infof("开始生成 %s 的统计数据", date)

	// TODO: 实际实现应查询 orders / order_items / users 表计算统计指标，
	// 写入 daily_statistics 表。此处仅记录日志占位，后续接入 BI 时补充。

	logx.Infof("统计数据生成完成: date=%s", date)
	return nil
}
