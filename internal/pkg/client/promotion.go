package client

import (
	"context"
	"fmt"
	"strconv"

	promotionv1 "ecommerce-system/api/promotion/v1"

	"google.golang.org/grpc"
)

// PromotionClient 营销服务客户端
type PromotionClient struct {
	conn    *grpc.ClientConn
	client  promotionv1.PromotionServiceClient
	timeout RpcConf
}

// NewPromotionClient 创建营销服务客户端
func NewPromotionClient(conf RpcConf) (*PromotionClient, error) {
	conn, err := newConn(conf)
	if err != nil {
		return nil, fmt.Errorf("dial promotion service %s: %w", conf.Endpoint, err)
	}
	return &PromotionClient{
		conn:    conn,
		client:  promotionv1.NewPromotionServiceClient(conn),
		timeout: conf,
	}, nil
}

// Close 关闭连接
func (c *PromotionClient) Close() error {
	return c.conn.Close()
}

// CalculateDiscount 计算优惠金额，返回 (discountAmount, finalAmount, error)
func (c *PromotionClient) CalculateDiscount(ctx context.Context, userID int64, productIDs []int64, quantities []int32, couponID int64, totalAmount float64) (float64, float64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.CalculateDiscount(ctx, &promotionv1.CalculateDiscountRequest{
		UserId:      userID,
		ProductIds:  productIDs,
		Quantities:  quantities,
		CouponId:    couponID,
		TotalAmount: strconv.FormatFloat(totalAmount, 'f', 2, 64),
	})
	if err != nil {
		return 0, totalAmount, fmt.Errorf("calculate discount: %w", err)
	}
	if resp.Code != 0 {
		return 0, totalAmount, fmt.Errorf("calculate discount: %s", resp.Message)
	}

	discount, _ := strconv.ParseFloat(resp.DiscountAmount, 64)
	final, _ := strconv.ParseFloat(resp.FinalAmount, 64)
	return discount, final, nil
}

// UseCoupon 核销优惠券（下单后调用）
func (c *PromotionClient) UseCoupon(ctx context.Context, userID, userCouponID, orderID int64) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout.timeout())
	defer cancel()

	resp, err := c.client.UseCoupon(ctx, &promotionv1.UseCouponRequest{
		UserId:       userID,
		UserCouponId: userCouponID,
		OrderId:      orderID,
	})
	if err != nil {
		return fmt.Errorf("use coupon: %w", err)
	}
	if resp.Code != 0 {
		return fmt.Errorf("use coupon: %s", resp.Message)
	}
	return nil
}
