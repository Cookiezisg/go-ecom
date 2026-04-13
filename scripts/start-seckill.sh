#!/bin/bash

# 秒杀功能启动脚本

echo "=========================================="
echo "秒杀功能启动脚本"
echo "=========================================="

# 1. 启动基础设施（Redis、Kafka等）
echo ""
echo "1. 启动基础设施服务..."
docker-compose -f docker-compose-infra.yml up -d

# 等待 Kafka 启动
echo "等待 Kafka 启动..."
sleep 10

# 2. 初始化 Redis 库存（示例）
echo ""
echo "2. 初始化 Redis 库存（示例）..."
echo "请手动执行以下命令设置库存："
echo "  redis-cli SET seckill:stock:1 100"
echo "  redis-cli SET seckill:stock:2 50"
echo ""

# 3. 生成 Proto 文件
echo ""
echo "3. 生成 Proto 文件..."
cd api/seckill/v1
protoc --go_out=. --go-grpc_out=. seckill.proto
cd ../../..

# 4. 启动秒杀服务
echo ""
echo "4. 启动秒杀服务..."
echo "执行: go run cmd/seckill-service/main.go -f configs/dev/seckill-config.yaml"
echo ""

# 5. 启动订单服务消费者
echo ""
echo "5. 启动订单服务消费者..."
echo "执行: go run cmd/order-service-consumer/main.go -f configs/dev/order-config.yaml"
echo ""

echo "=========================================="
echo "启动完成！"
echo "=========================================="
echo ""
echo "测试秒杀接口："
echo "grpcurl -plaintext -d '{\"user_id\": 1, \"sku_id\": 1, \"quantity\": 1}' localhost:8090 seckill.v1.SeckillService/Seckill"
echo ""

