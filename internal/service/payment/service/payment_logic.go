package service

import (
	"context"
	"errors"
	"time"

	"ecommerce-system/internal/pkg/client"
	apperrors "ecommerce-system/internal/pkg/errors"
	"ecommerce-system/internal/pkg/idgen"
	"ecommerce-system/internal/service/payment/model"
	"ecommerce-system/internal/service/payment/repository"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// PaymentLogic 支付业务逻辑
type PaymentLogic struct {
	idGen          *idgen.Generator
	paymentRepo    repository.PaymentRepository
	paymentLogRepo repository.PaymentLogRepository
	orderClient    *client.OrderClient
	invClient      *client.InventoryClient
}

// NewPaymentLogic 创建支付业务逻辑
func NewPaymentLogic(
	idGen *idgen.Generator,
	paymentRepo repository.PaymentRepository,
	paymentLogRepo repository.PaymentLogRepository,
	orderClient *client.OrderClient,
	invClient *client.InventoryClient,
) *PaymentLogic {
	return &PaymentLogic{
		idGen:          idGen,
		paymentRepo:    paymentRepo,
		paymentLogRepo: paymentLogRepo,
		orderClient:    orderClient,
		invClient:      invClient,
	}
}

// -----------------------------------------------------------------------
// CreatePayment
// -----------------------------------------------------------------------

// CreatePaymentRequest 创建支付单请求
type CreatePaymentRequest struct {
	OrderID       uint64
	OrderNo       string
	UserID        uint64
	Amount        float64
	PaymentMethod int8
}

// CreatePaymentResponse 创建支付单响应
type CreatePaymentResponse struct {
	Payment *model.Payment
	PayURL  string
}

// CreatePayment 创建支付单
func (l *PaymentLogic) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	// 幂等检查：同一订单是否已有待支付的支付单
	existing, err := l.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if existing != nil && existing.Status == 0 {
		// 返回已有的待支付单
		return &CreatePaymentResponse{
			Payment: existing,
			PayURL:  "/payment/pay?payment_no=" + existing.PaymentNo,
		}, nil
	}

	paymentNo := l.idGen.PaymentNo(ctx)
	expireAt := time.Now().Add(30 * time.Minute)

	payment := &model.Payment{
		PaymentNo:     paymentNo,
		OrderID:       req.OrderID,
		OrderNo:       req.OrderNo,
		UserID:        req.UserID,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		Status:        0, // 待支付
		ExpireAt:      &expireAt,
	}

	if err := l.paymentRepo.Create(ctx, payment); err != nil {
		return nil, apperrors.NewInternalError("创建支付单失败: " + err.Error())
	}

	_ = l.paymentLogRepo.Create(ctx, &model.PaymentLog{
		PaymentID:   payment.ID,
		PaymentNo:   paymentNo,
		Action:      "create",
		Amount:      req.Amount,
		AfterStatus: &payment.Status,
	})

	// 实际项目中此处调用微信/支付宝 SDK 生成支付链接
	payURL := "/payment/pay?payment_no=" + paymentNo

	return &CreatePaymentResponse{Payment: payment, PayURL: payURL}, nil
}

// -----------------------------------------------------------------------
// GetPayment
// -----------------------------------------------------------------------

// GetPaymentRequest 获取支付单请求
type GetPaymentRequest struct {
	PaymentNo string
}

// GetPaymentResponse 获取支付单响应
type GetPaymentResponse struct {
	Payment *model.Payment
}

// GetPayment 获取支付单
func (l *PaymentLogic) GetPayment(ctx context.Context, req *GetPaymentRequest) (*GetPaymentResponse, error) {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodePaymentNotFound)
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewError(apperrors.CodePaymentNotFound)
	}
	return &GetPaymentResponse{Payment: payment}, nil
}

// -----------------------------------------------------------------------
// PaymentCallback
// -----------------------------------------------------------------------

// PaymentCallbackRequest 支付回调请求
type PaymentCallbackRequest struct {
	PaymentNo    string
	ThirdPartyNo string
	Status       int8 // 1-成功, 2-失败
	CallbackData string
}

// PaymentCallback 处理第三方支付回调
// 成功：更新支付单 → 通知 order.PayOrder → (inventory.DeductStock 由 order service 负责)
// 失败：更新支付单 → 通知 order.CancelOrder
func (l *PaymentLogic) PaymentCallback(ctx context.Context, req *PaymentCallbackRequest) error {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewError(apperrors.CodePaymentNotFound)
		}
		return apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return apperrors.NewError(apperrors.CodePaymentNotFound)
	}

	// 幂等：已处理过直接返回
	if payment.Status != 0 {
		return nil
	}

	now := time.Now()
	beforeStatus := payment.Status
	payment.Status = req.Status
	payment.ThirdPartyNo = &req.ThirdPartyNo
	if req.Status == 1 {
		payment.PaidAt = &now
	}

	if err := l.paymentRepo.Update(ctx, payment); err != nil {
		return apperrors.NewInternalError("更新支付状态失败: " + err.Error())
	}

	_ = l.paymentLogRepo.Create(ctx, &model.PaymentLog{
		PaymentID:    payment.ID,
		PaymentNo:    req.PaymentNo,
		Action:       "callback",
		Amount:       payment.Amount,
		BeforeStatus: &beforeStatus,
		AfterStatus:  &req.Status,
	})

	// 回调下游订单服务
	if l.orderClient != nil {
		if req.Status == 1 {
			// 支付成功 → 通知订单服务（订单状态 待支付→待发货，并扣减库存）
			if callErr := l.orderClient.PayOrder(ctx, int64(payment.OrderID), payment.OrderNo, req.PaymentNo, int32(payment.PaymentMethod)); callErr != nil {
				logx.Errorf("通知订单服务支付成功失败 order_no=%s: %v", payment.OrderNo, callErr)
				// 不回滚支付状态，由运营人工处理
			}
		} else {
			// 支付失败 → 取消订单，解锁库存
			if cancelErr := l.orderClient.CancelOrder(ctx, int64(payment.OrderID), payment.OrderNo, "支付失败"); cancelErr != nil {
				logx.Errorf("通知订单服务取消订单失败 order_no=%s: %v", payment.OrderNo, cancelErr)
			}
		}
	}

	return nil
}

// -----------------------------------------------------------------------
// Refund
// -----------------------------------------------------------------------

// RefundRequest 退款请求
type RefundRequest struct {
	PaymentNo    string
	RefundAmount float64
	Reason       string
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo string
}

// Refund 申请退款（支付成功才能退款）
func (l *PaymentLogic) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodePaymentNotFound)
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewError(apperrors.CodePaymentNotFound)
	}
	if payment.Status != 1 {
		return nil, apperrors.NewError(apperrors.CodeRefundFailed, "只有支付成功的订单才能退款")
	}

	refundNo := l.idGen.RefundNo(ctx)
	beforeStatus := payment.Status
	payment.Status = 3 // 已退款

	if err := l.paymentRepo.Update(ctx, payment); err != nil {
		return nil, apperrors.NewInternalError("退款失败: " + err.Error())
	}

	afterStatus := int8(3)
	_ = l.paymentLogRepo.Create(ctx, &model.PaymentLog{
		PaymentID:    payment.ID,
		PaymentNo:    req.PaymentNo,
		Action:       "refund",
		Amount:       req.RefundAmount,
		BeforeStatus: &beforeStatus,
		AfterStatus:  &afterStatus,
		Remark:       req.Reason,
	})

	// 通知订单服务退款完成（更新状态 + 回退库存）
	if l.orderClient != nil {
		if refundErr := l.orderClient.RefundOrder(ctx, int64(payment.OrderID), payment.OrderNo, req.Reason); refundErr != nil {
			logx.Errorf("通知订单服务退款失败 order_no=%s: %v", payment.OrderNo, refundErr)
		}
	}

	return &RefundResponse{RefundNo: refundNo}, nil
}

// -----------------------------------------------------------------------
// QueryPaymentStatus
// -----------------------------------------------------------------------

// QueryPaymentStatusRequest 查询支付状态请求
type QueryPaymentStatusRequest struct {
	PaymentNo string
}

// QueryPaymentStatusResponse 查询支付状态响应
type QueryPaymentStatusResponse struct {
	Status int8
}

// QueryPaymentStatus 查询支付状态
func (l *PaymentLogic) QueryPaymentStatus(ctx context.Context, req *QueryPaymentStatusRequest) (*QueryPaymentStatusResponse, error) {
	payment, err := l.paymentRepo.GetByPaymentNo(ctx, req.PaymentNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewError(apperrors.CodePaymentNotFound)
		}
		return nil, apperrors.NewInternalError("查询支付单失败: " + err.Error())
	}
	if payment == nil {
		return nil, apperrors.NewError(apperrors.CodePaymentNotFound)
	}
	return &QueryPaymentStatusResponse{Status: payment.Status}, nil
}
