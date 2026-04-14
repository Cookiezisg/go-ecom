// Package idgen 提供分布式唯一 ID / 业务单号生成能力。
// 基于 Redis INCR + 日期前缀，保证同一天内序号单调递增且唯一。
// 降级策略：Redis 不可用时使用纳秒时间戳（低概率碰撞，仅用于开发/测试场景）。
package idgen

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const seqTTL = 25 * time.Hour // 序号 key 保留 25 小时，跨零点不丢失

// Generator 单号生成器
type Generator struct {
	redis *redis.Client
}

// New 创建生成器
func New(rdb *redis.Client) *Generator {
	return &Generator{redis: rdb}
}

// OrderNo 生成订单号，格式：ORD + yyyyMMdd + 8位序号，例如 ORD2026041400000001
func (g *Generator) OrderNo(ctx context.Context) string {
	return g.generate(ctx, "idgen:order:", "ORD")
}

// PaymentNo 生成支付单号，格式：PAY + yyyyMMdd + 8位序号
func (g *Generator) PaymentNo(ctx context.Context) string {
	return g.generate(ctx, "idgen:payment:", "PAY")
}

// RefundNo 生成退款单号，格式：REF + yyyyMMdd + 8位序号
func (g *Generator) RefundNo(ctx context.Context) string {
	return g.generate(ctx, "idgen:refund:", "REF")
}

// LogisticsNo 生成物流单号，格式：LGS + yyyyMMdd + 8位序号
func (g *Generator) LogisticsNo(ctx context.Context) string {
	return g.generate(ctx, "idgen:logistics:", "LGS")
}

// generate 通用生成逻辑
func (g *Generator) generate(ctx context.Context, keyPrefix, bizPrefix string) string {
	date := time.Now().Format("20060102")
	key := keyPrefix + date

	seq, err := g.redis.Incr(ctx, key).Result()
	if err != nil {
		// Redis 不可用时降级到纳秒时间戳（仅供开发测试）
		return fmt.Sprintf("%s%s%09d", bizPrefix, date, time.Now().Nanosecond()%1_000_000_000)
	}

	// 首次生成时设置 TTL，后续 INCR 不重置 TTL，所以只在 seq==1 时设置
	if seq == 1 {
		_ = g.redis.Expire(ctx, key, seqTTL).Err()
	}

	return fmt.Sprintf("%s%s%08d", bizPrefix, date, seq)
}
