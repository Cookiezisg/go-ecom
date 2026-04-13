package service

import (
	"context"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/service/job/repository"
)

// JobLogic 定时任务业务逻辑
type JobLogic struct {
	orderRepo  repository.OrderRepository
	couponRepo repository.CouponRepository
}

// NewJobLogic 创建定时任务业务逻辑
func NewJobLogic(
	orderRepo repository.OrderRepository,
	couponRepo repository.CouponRepository,
) *JobLogic {
	return &JobLogic{
		orderRepo:  orderRepo,
		couponRepo: couponRepo,
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

// CancelExpiredOrders 订单超时取消
func (l *JobLogic) CancelExpiredOrders(ctx context.Context, req *CancelExpiredOrdersRequest) (*CancelExpiredOrdersResponse, error) {
	count, err := l.orderRepo.CancelExpiredOrders(ctx, req.TimeoutMinutes)
	if err != nil {
		return nil, apperrors.NewInternalError("取消超时订单失败")
	}

	return &CancelExpiredOrdersResponse{
		CancelledCount: count,
	}, nil
}

// ProcessExpiredCouponsRequest 优惠券过期处理请求
type ProcessExpiredCouponsRequest struct {
}

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

	return &ProcessExpiredCouponsResponse{
		ExpiredCount: count,
	}, nil
}

// GenerateStatisticsRequest 数据统计请求
type GenerateStatisticsRequest struct {
	Date string // YYYY-MM-DD
}

// GenerateStatistics 数据统计
func (l *JobLogic) GenerateStatistics(ctx context.Context, req *GenerateStatisticsRequest) error {
	// 实际实现应该生成各种统计数据
	// 这里简化处理
	return nil
}
