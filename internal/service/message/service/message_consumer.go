package service

import (
	"context"
	"encoding/json"
	"fmt"

	"ecommerce-system/internal/pkg/mq"

	"github.com/zeromicro/go-zero/core/logx"
)

// MessageConsumer 消息服务 Kafka 消费者
// 监听订单、支付相关事件，自动给用户推送站内消息。
type MessageConsumer struct {
	logic *MessageLogic
}

// NewMessageConsumer 创建消息消费者
func NewMessageConsumer(logic *MessageLogic) *MessageConsumer {
	return &MessageConsumer{logic: logic}
}

// HandleOrderCreated 处理订单创建事件 → 发送「下单成功」通知
func (c *MessageConsumer) HandleOrderCreated(ctx context.Context, msg *mq.Message) error {
	type payload struct {
		OrderID     uint64  `json:"order_id"`
		OrderNo     string  `json:"order_no"`
		UserID      uint64  `json:"user_id"`
		TotalAmount float64 `json:"total_amount"`
	}
	var p payload
	if err := decodePayload(msg.Data, &p); err != nil {
		logx.Errorf("解析订单创建消息失败: %v", err)
		return nil // 不返回 error，避免无限重试
	}
	if p.UserID == 0 {
		return nil
	}

	return c.logic.SendMessage(ctx, &SendMessageRequest{
		UserID:  p.UserID,
		Type:    1, // 订单消息
		Title:   "下单成功",
		Content: fmt.Sprintf("您的订单 %s 已成功提交，请尽快完成支付。", p.OrderNo),
		Link:    fmt.Sprintf("/orders/%s", p.OrderNo),
	})
}

// HandleOrderCancelled 处理订单取消事件 → 发送「订单取消」通知
func (c *MessageConsumer) HandleOrderCancelled(ctx context.Context, msg *mq.Message) error {
	type payload struct {
		OrderID uint64 `json:"order_id"`
		OrderNo string `json:"order_no"`
		UserID  uint64 `json:"user_id"`
		Reason  string `json:"reason"`
	}
	var p payload
	if err := decodePayload(msg.Data, &p); err != nil {
		logx.Errorf("解析订单取消消息失败: %v", err)
		return nil
	}
	if p.UserID == 0 {
		return nil
	}

	content := fmt.Sprintf("您的订单 %s 已取消。", p.OrderNo)
	if p.Reason != "" {
		content = fmt.Sprintf("您的订单 %s 已取消，原因：%s", p.OrderNo, p.Reason)
	}

	return c.logic.SendMessage(ctx, &SendMessageRequest{
		UserID:  p.UserID,
		Type:    1,
		Title:   "订单已取消",
		Content: content,
		Link:    fmt.Sprintf("/orders/%s", p.OrderNo),
	})
}

// HandlePaymentSuccess 处理支付成功事件 → 发送「支付成功」通知
func (c *MessageConsumer) HandlePaymentSuccess(ctx context.Context, msg *mq.Message) error {
	type payload struct {
		OrderID   uint64  `json:"order_id"`
		OrderNo   string  `json:"order_no"`
		UserID    uint64  `json:"user_id"`
		PayAmount float64 `json:"pay_amount"`
	}
	var p payload
	if err := decodePayload(msg.Data, &p); err != nil {
		logx.Errorf("解析支付成功消息失败: %v", err)
		return nil
	}
	if p.UserID == 0 {
		return nil
	}

	return c.logic.SendMessage(ctx, &SendMessageRequest{
		UserID:  p.UserID,
		Type:    2, // 支付消息
		Title:   "支付成功",
		Content: fmt.Sprintf("订单 %s 支付成功，实付金额 %.2f 元，商家正在处理您的订单。", p.OrderNo, p.PayAmount),
		Link:    fmt.Sprintf("/orders/%s", p.OrderNo),
	})
}

// HandlePaymentRefunded 处理退款成功事件 → 发送「退款成功」通知
func (c *MessageConsumer) HandlePaymentRefunded(ctx context.Context, msg *mq.Message) error {
	type payload struct {
		OrderID      uint64  `json:"order_id"`
		OrderNo      string  `json:"order_no"`
		UserID       uint64  `json:"user_id"`
		RefundAmount float64 `json:"refund_amount"`
	}
	var p payload
	if err := decodePayload(msg.Data, &p); err != nil {
		logx.Errorf("解析退款消息失败: %v", err)
		return nil
	}
	if p.UserID == 0 {
		return nil
	}

	return c.logic.SendMessage(ctx, &SendMessageRequest{
		UserID:  p.UserID,
		Type:    2,
		Title:   "退款成功",
		Content: fmt.Sprintf("订单 %s 退款已处理，退款金额 %.2f 元将原路退回，预计 3-5 个工作日到账。", p.OrderNo, p.RefundAmount),
		Link:    fmt.Sprintf("/orders/%s", p.OrderNo),
	})
}

// decodePayload 把 map[string]interface{} 转换为目标结构体（通过 JSON 中转）
func decodePayload(data map[string]interface{}, dst interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	return json.Unmarshal(b, dst)
}
