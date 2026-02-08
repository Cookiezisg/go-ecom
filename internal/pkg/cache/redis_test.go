package cache

import (
	"context"
	"testing"
)

func TestNewRedis(t *testing.T) {

	redisConfig := &Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MinIdleConns: 5,
	}
	if _, err := NewRedis(redisConfig); err != nil {
		t.Errorf("Failed to create Redis client: %v", err)
	}

}

func TestRedisOperations(t *testing.T) {
	conf := &Config{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		Database:     0,
		PoolSize:     10,
		MinIdleConns: 5,
	}
	client, err := NewRedis(conf)
	if err != nil {
		t.Fatalf("连接都连不上，还测个毛: %v", err)
	}

	ctx := context.Background()
	key := "test_key"
	val := "hello_shopee"

	err = client.Set(ctx, key, val, 0).Err()
	if err != nil {
		t.Errorf("存数据失败: %v", err)
	}

	res, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Errorf("取数据失败: %v", err)
	}

	if res != val {
		t.Errorf("数据不一致！存的是 %s，取出来成了 %s", val, res)
	}

	client.Del(ctx, key)
}
