#!/bin/bash

# 秒杀服务诊断脚本

echo "=========================================="
echo "秒杀服务诊断"
echo "=========================================="
echo ""

# 1. 检查配置文件
echo "1. 检查配置文件..."
if [ -f "configs/dev/seckill-config.yaml" ]; then
    echo "  ✓ seckill-config.yaml 存在"
else
    echo "  ✗ seckill-config.yaml 不存在"
    echo "    请创建: configs/dev/seckill-config.yaml"
fi

if [ -f "configs/dev/order-config.yaml" ]; then
    echo "  ✓ order-config.yaml 存在"
else
    echo "  ✗ order-config.yaml 不存在"
fi
echo ""

# 2. 检查 Proto 文件
echo "2. 检查 Proto 文件..."
if [ -f "api/seckill/v1/seckill.pb.go" ] && [ -f "api/seckill/v1/seckill_grpc.pb.go" ]; then
    echo "  ✓ Proto 文件已生成"
else
    echo "  ✗ Proto 文件未生成"
    echo "    执行: protoc --go_out=. --go-grpc_out=. api/seckill/v1/seckill.proto"
fi
echo ""

# 3. 检查 Docker 容器
echo "3. 检查 Docker 容器..."
if docker ps | grep -q "infra-redis"; then
    echo "  ✓ Redis 容器正在运行"
else
    echo "  ✗ Redis 容器未运行"
    echo "    执行: make start-infra"
fi

if docker ps | grep -q "infra-kafka"; then
    echo "  ✓ Kafka 容器正在运行"
else
    echo "  ✗ Kafka 容器未运行"
    echo "    执行: make start-infra"
fi

if docker ps | grep -q "infra-zookeeper"; then
    echo "  ✓ Zookeeper 容器正在运行"
else
    echo "  ✗ Zookeeper 容器未运行"
    echo "    执行: make start-infra"
fi
echo ""

# 4. 检查 Redis 连接
echo "4. 检查 Redis 连接..."
if docker exec infra-redis redis-cli PING 2>/dev/null | grep -q "PONG"; then
    echo "  ✓ Redis 连接正常"
    
    # 检查库存
    echo "  当前库存:"
    docker exec infra-redis redis-cli KEYS "seckill:stock:*" 2>/dev/null | while read key; do
        if [ -n "$key" ]; then
            value=$(docker exec infra-redis redis-cli GET "$key" 2>/dev/null)
            echo "    $key = $value"
        fi
    done
else
    echo "  ✗ Redis 连接失败"
fi
echo ""

# 5. 检查端口占用
echo "5. 检查端口占用..."
if lsof -Pi :8090 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "  ⚠ 端口 8090 已被占用（秒杀服务端口）"
    lsof -Pi :8090 -sTCP:LISTEN
else
    echo "  ✓ 端口 8090 可用"
fi
echo ""

# 6. 检查日志
echo "6. 检查服务日志..."
if [ -f "logs/seckill-service.log" ]; then
    echo "  秒杀服务日志（最后 10 行）:"
    tail -10 logs/seckill-service.log | sed 's/^/    /'
else
    echo "  ⚠ 秒杀服务日志文件不存在"
fi
echo ""

if [ -f "logs/order-consumer.log" ]; then
    echo "  订单消费者日志（最后 10 行）:"
    tail -10 logs/order-consumer.log | sed 's/^/    /'
else
    echo "  ⚠ 订单消费者日志文件不存在"
fi
echo ""

# 7. 检查代码编译
echo "7. 检查代码编译..."
if go build -o /tmp/test-seckill ./cmd/seckill-service 2>&1 | head -5; then
    echo "  ✓ 秒杀服务编译成功"
    rm -f /tmp/test-seckill
else
    echo "  ✗ 秒杀服务编译失败"
fi

if go build -o /tmp/test-consumer ./cmd/order-service-consumer 2>&1 | head -5; then
    echo "  ✓ 订单消费者编译成功"
    rm -f /tmp/test-consumer
else
    echo "  ✗ 订单消费者编译失败"
fi
echo ""

echo "=========================================="
echo "诊断完成"
echo "=========================================="

