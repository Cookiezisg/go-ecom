package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 错误码定义
const (
	// 通用错误码 1000-1999
	CodeSuccess          = 0
	CodeInternalError    = 1000
	CodeInvalidParam     = 1001
	CodeUnauthorized     = 1002
	CodeForbidden        = 1003
	CodeNotFound         = 1004
	CodeAlreadyExists    = 1005
	CodeDatabaseError    = 1006
	CodeCacheError       = 1007
	CodeExternalAPIError = 1008
	CodeTimeout          = 1009
	CodeTooManyRequests  = 1010

	// 用户服务错误码 2000-2999
	CodeUserNotFound      = 2000
	CodeUserAlreadyExists = 2001
	CodePasswordError     = 2002
	CodeTokenExpired      = 2003
	CodeTokenInvalid      = 2004
	CodeUserDisabled      = 2005
	CodeVerifyCodeError   = 2006

	// 商品服务错误码 3000-3999
	CodeProductNotFound  = 3000
	CodeProductOffline   = 3001
	CodeSkuNotFound      = 3002
	CodeCategoryNotFound = 3003
	CodeSkuOffline       = 3004

	// 订单服务错误码 4000-4999
	CodeOrderNotFound       = 4000
	CodeOrderStatusError    = 4001
	CodeOrderCanceled       = 4002
	CodeOrderPaid           = 4003
	CodeOrderNotPaid        = 4004
	CodeOrderNotShipped     = 4005
	CodeOrderAlreadyExists  = 4006
	CodeAddressNotFound     = 4007

	// 库存服务错误码 5000-5999
	CodeStockInsufficient = 5000
	CodeStockLocked       = 5001
	CodeStockNotFound     = 5002

	// 支付服务错误码 6000-6999
	CodePaymentFailed      = 6000
	CodePaymentExpired     = 6001
	CodeRefundFailed       = 6002
	CodePaymentNotFound    = 6003
	CodePaymentNotPaid     = 6004
	CodePaymentDuplicate   = 6005

	// 营销服务错误码 7000-7999
	CodeCouponNotFound     = 7000
	CodeCouponExpired      = 7001
	CodeCouponUsed         = 7002
	CodeCouponNotAvailable = 7003
	CodeCouponLimitReached = 7004

	// 秒杀服务错误码 8000-8999
	CodeSeckillNotStarted  = 8000
	CodeSeckillEnded       = 8001
	CodeSeckillSoldOut     = 8002
	CodeSeckillDuplicate   = 8003
	CodeSeckillNotInTime   = 8004

	// 物流服务错误码 9000-9999
	CodeLogisticsNotFound  = 9000
	CodeLogisticsError     = 9001
)

// BusinessError 业务错误
type BusinessError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建业务错误
func New(code int, message string) *BusinessError {
	return &BusinessError{Code: code, Message: message}
}

// NewInternalError 内部错误
func NewInternalError(message string) *BusinessError {
	return &BusinessError{Code: CodeInternalError, Message: message}
}

// NewInvalidParamError 参数错误
func NewInvalidParamError(message string) *BusinessError {
	return &BusinessError{Code: CodeInvalidParam, Message: message}
}

// NewNotFoundError 资源不存在
func NewNotFoundError(message string) *BusinessError {
	return &BusinessError{Code: CodeNotFound, Message: message}
}

// NewUnauthorizedError 未授权
func NewUnauthorizedError(message string) *BusinessError {
	return &BusinessError{Code: CodeUnauthorized, Message: message}
}

// NewForbiddenError 禁止访问
func NewForbiddenError(message string) *BusinessError {
	return &BusinessError{Code: CodeForbidden, Message: message}
}

// NewAlreadyExistsError 资源已存在
func NewAlreadyExistsError(message string) *BusinessError {
	return &BusinessError{Code: CodeAlreadyExists, Message: message}
}

// NewError 兼容旧调用（code + 可选 message）
func NewError(code int, message ...string) *BusinessError {
	msg := defaultMessage(code)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &BusinessError{Code: code, Message: msg}
}

// defaultMessage 错误码默认文案
func defaultMessage(code int) string {
	msgs := map[int]string{
		CodeSuccess:          "成功",
		CodeInternalError:    "内部服务器错误",
		CodeInvalidParam:     "参数错误",
		CodeUnauthorized:     "未授权",
		CodeForbidden:        "禁止访问",
		CodeNotFound:         "资源不存在",
		CodeAlreadyExists:    "资源已存在",
		CodeDatabaseError:    "数据库错误",
		CodeCacheError:       "缓存错误",
		CodeExternalAPIError: "外部API调用失败",
		CodeTimeout:          "请求超时",
		CodeTooManyRequests:  "请求过于频繁",

		CodeUserNotFound:      "用户不存在",
		CodeUserAlreadyExists: "用户已存在",
		CodePasswordError:     "密码错误",
		CodeTokenExpired:      "Token已过期",
		CodeTokenInvalid:      "Token无效",
		CodeUserDisabled:      "用户已被禁用",
		CodeVerifyCodeError:   "验证码错误",

		CodeProductNotFound:  "商品不存在",
		CodeProductOffline:   "商品已下架",
		CodeSkuNotFound:      "SKU不存在",
		CodeCategoryNotFound: "类目不存在",
		CodeSkuOffline:       "SKU已下架",

		CodeOrderNotFound:      "订单不存在",
		CodeOrderStatusError:   "订单状态错误",
		CodeOrderCanceled:      "订单已取消",
		CodeOrderPaid:          "订单已支付",
		CodeOrderNotPaid:       "订单未支付",
		CodeOrderNotShipped:    "订单未发货",
		CodeOrderAlreadyExists: "订单已存在",
		CodeAddressNotFound:    "收货地址不存在",

		CodeStockInsufficient: "库存不足",
		CodeStockLocked:       "库存已锁定",
		CodeStockNotFound:     "库存记录不存在",

		CodePaymentFailed:    "支付失败",
		CodePaymentExpired:   "支付已过期",
		CodeRefundFailed:     "退款失败",
		CodePaymentNotFound:  "支付单不存在",
		CodePaymentNotPaid:   "支付单未支付",
		CodePaymentDuplicate: "重复支付",

		CodeCouponNotFound:     "优惠券不存在",
		CodeCouponExpired:      "优惠券已过期",
		CodeCouponUsed:         "优惠券已使用",
		CodeCouponNotAvailable: "优惠券不可用",
		CodeCouponLimitReached: "优惠券领取已达上限",

		CodeSeckillNotStarted: "秒杀活动未开始",
		CodeSeckillEnded:      "秒杀活动已结束",
		CodeSeckillSoldOut:    "商品已抢光",
		CodeSeckillDuplicate:  "请勿重复抢购",
		CodeSeckillNotInTime:  "不在秒杀时间内",

		CodeLogisticsNotFound: "物流信息不存在",
		CodeLogisticsError:    "物流操作失败",
	}
	if msg, ok := msgs[code]; ok {
		return msg
	}
	return "未知错误"
}

// ConvertToGRPCError 将 BusinessError 转换为 gRPC status error。
// 所有服务统一调用此函数，不再各自实现 convertError。
func ConvertToGRPCError(err error) error {
	if err == nil {
		return nil
	}
	bizErr, ok := err.(*BusinessError)
	if !ok {
		return status.Error(codes.Internal, err.Error())
	}
	return status.Error(bizCodeToGRPC(bizErr.Code), bizErr.Error())
}

// bizCodeToGRPC 业务码 → gRPC codes
func bizCodeToGRPC(code int) codes.Code {
	switch code {
	case CodeNotFound,
		CodeUserNotFound,
		CodeProductNotFound, CodeSkuNotFound, CodeCategoryNotFound,
		CodeOrderNotFound,
		CodeStockNotFound,
		CodePaymentNotFound,
		CodeCouponNotFound,
		CodeLogisticsNotFound,
		CodeAddressNotFound:
		return codes.NotFound

	case CodeInvalidParam:
		return codes.InvalidArgument

	case CodeUnauthorized, CodeTokenExpired, CodeTokenInvalid:
		return codes.Unauthenticated

	case CodeForbidden, CodeUserDisabled:
		return codes.PermissionDenied

	case CodeAlreadyExists,
		CodeUserAlreadyExists,
		CodeOrderAlreadyExists,
		CodePaymentDuplicate,
		CodeSeckillDuplicate:
		return codes.AlreadyExists

	case CodeTooManyRequests:
		return codes.ResourceExhausted

	case CodeTimeout:
		return codes.DeadlineExceeded

	case CodeStockInsufficient, CodeSeckillSoldOut:
		return codes.FailedPrecondition

	default:
		return codes.Internal
	}
}
