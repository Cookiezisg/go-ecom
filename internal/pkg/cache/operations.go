package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheOperations 缓存操作封装
type CacheOperations struct {
	client *redis.Client
}

// NewCacheOperations 创建缓存操作实例
func NewCacheOperations(client *redis.Client) *CacheOperations {
	return &CacheOperations{client: client}
}

// IsNil 判断是否为 redis key 不存在错误
func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}

// Set 设置缓存
func (c *CacheOperations) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var val string
	switch v := value.(type) {
	case string:
		val = v
	case []byte:
		val = string(v)
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		val = string(data)
	}
	return c.client.Set(ctx, key, val, expiration).Err()
}

// Get 获取缓存字符串值，key 不存在时返回 redis.Nil 错误（用 IsNil 判断）
func (c *CacheOperations) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// GetJSON 获取缓存并反序列化
func (c *CacheOperations) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete 删除一个或多个 key
func (c *CacheOperations) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查 key 是否存在
func (c *CacheOperations) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	return result > 0, err
}

// Expire 设置过期时间
func (c *CacheOperations) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余 TTL
func (c *CacheOperations) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Increment 自增 1
func (c *CacheOperations) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrementBy 自增指定值
func (c *CacheOperations) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decrement 自减 1
func (c *CacheOperations) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrementBy 自减指定值
func (c *CacheOperations) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// SetNX 仅在 key 不存在时设置（用于分布式锁）
func (c *CacheOperations) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// GetClient 返回原始 redis.Client（给需要直接操作的场景使用）
func (c *CacheOperations) GetClient() *redis.Client {
	return c.client
}

// HashSet 设置 Hash 字段
func (c *CacheOperations) HashSet(ctx context.Context, key string, values map[string]interface{}) error {
	return c.client.HSet(ctx, key, values).Err()
}

// HashGet 获取 Hash 字段
func (c *CacheOperations) HashGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HashGetAll 获取 Hash 所有字段
func (c *CacheOperations) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HashDelete 删除 Hash 字段
func (c *CacheOperations) HashDelete(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HashExists 检查 Hash 字段是否存在
func (c *CacheOperations) HashExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// ListPush 从左侧推入
func (c *CacheOperations) ListPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// ListPop 从左侧弹出
func (c *CacheOperations) ListPop(ctx context.Context, key string) (string, error) {
	return c.client.LPop(ctx, key).Result()
}

// ListRange 获取列表范围
func (c *CacheOperations) ListRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// ListLength 获取列表长度
func (c *CacheOperations) ListLength(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// SetAdd 添加 Set 成员
func (c *CacheOperations) SetAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SetRemove 移除 Set 成员
func (c *CacheOperations) SetRemove(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SetMembers 获取 Set 所有成员
func (c *CacheOperations) SetMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SetIsMember 检查是否为 Set 成员
func (c *CacheOperations) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// DeletePattern 用 SCAN 批量删除匹配的 key（避免 KEYS 阻塞）
func (c *CacheOperations) DeletePattern(ctx context.Context, pattern string) error {
	const batchSize = 100
	var cursor uint64
	var pending []string

	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			return err
		}
		pending = append(pending, keys...)
		cursor = nextCursor

		if len(pending) >= batchSize {
			if err := c.client.Del(ctx, pending...).Err(); err != nil {
				return err
			}
			pending = pending[:0]
		}

		if cursor == 0 {
			break
		}
	}

	if len(pending) > 0 {
		return c.client.Del(ctx, pending...).Err()
	}
	return nil
}

// AtomicDeductStock 原子扣减库存（Lua 脚本），返回扣减后剩余库存。
// 返回 -1 表示库存不足，返回 -2 表示 key 不存在。
func (c *CacheOperations) AtomicDeductStock(ctx context.Context, key string, quantity int64) (int64, error) {
	result, err := ExecuteLuaScript(ctx, c.client, LuaScriptInventoryDeduct, []string{key}, quantity)
	if err != nil {
		return 0, err
	}
	val, ok := result.(int64)
	if !ok {
		return 0, errors.New("lua script returned unexpected type")
	}
	return val, nil
}

// AtomicRollbackStock 原子回退库存（Lua 脚本），返回回退后剩余库存。
func (c *CacheOperations) AtomicRollbackStock(ctx context.Context, key string, quantity int64) (int64, error) {
	result, err := ExecuteLuaScript(ctx, c.client, LuaScriptInventoryRollback, []string{key}, quantity)
	if err != nil {
		return 0, err
	}
	val, ok := result.(int64)
	if !ok {
		return 0, errors.New("lua script returned unexpected type")
	}
	return val, nil
}
