#!/bin/bash

# 启动所有微服务脚本 (Linux/Mac)
# 使用方法: ./scripts/start-all-services.sh

set -e

BUILD=false
GATEWAY=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --build)
            BUILD=true
            shift
            ;;
        --gateway)
            GATEWAY=true
            shift
            ;;
        *)
            echo "未知参数: $1"
            exit 1
            ;;
    esac
done

# 服务配置列表（使用普通数组，兼容 macOS bash 3.2）
SERVICES=(
    "user-service:configs/dev/config.yaml:8080"
    "product-service:configs/dev/product-config.yaml:8081"
    "inventory-service:configs/dev/inventory-config.yaml:8084"
    "cart-service:configs/dev/cart-config.yaml:8085"
    "order-service:configs/dev/order-config.yaml:8082"
    "payment-service:configs/dev/payment-config.yaml:8083"
)

echo "============================================"
echo "启动所有微服务"
echo "============================================"
echo ""

# 检查配置文件
echo "检查配置文件..."
for service_config in "${SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    config=$(echo "$service_config" | cut -d: -f2)
    if [ ! -f "$config" ]; then
        echo "错误: 配置文件不存在: $config"
        echo "请先复制示例文件: ${config}.example -> $config"
        exit 1
    fi
done
echo "配置文件检查通过"
echo ""

# 编译服务（如果需要）
if [ "$BUILD" = true ]; then
    echo "编译所有服务..."
    for service_config in "${SERVICES[@]}"; do
        service=$(echo "$service_config" | cut -d: -f1)
        echo "编译 $service..."
        if [ ! -f "bin/$service" ]; then
            go build -o "bin/$service" "./cmd/$service"
        fi
    done
    echo "编译完成"
    echo ""
fi

# 创建日志目录
mkdir -p logs

# 启动服务
echo "启动服务..."
PIDS=()

for service_config in "${SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    config=$(echo "$service_config" | cut -d: -f2)
    port=$(echo "$service_config" | cut -d: -f3)
    
    echo "启动 $service (端口: $port)..."
    
    if [ -f "bin/$service" ]; then
        # 使用编译后的可执行文件
        ./bin/$service -f "$config" &
    else
        # 直接运行 Go 程序
        go run "cmd/$service/main.go" -f "$config" &
    fi
    
    PIDS+=($!)
    sleep 2  # 等待服务启动
done

# 启动 API 网关（如果指定）
if [ "$GATEWAY" = true ]; then
    echo ""
    echo "启动 API 网关..."
    gateway_config="configs/dev/gateway.yaml"
    if [ -f "$gateway_config" ]; then
        if [ -f "bin/api-gateway" ]; then
            ./bin/api-gateway -f "$gateway_config" &
        else
            go run cmd/api-gateway/main.go -f "$gateway_config" &
        fi
        PIDS+=($!)
    else
        echo "警告: 网关配置文件不存在: $gateway_config"
    fi
fi

echo ""
echo "============================================"
echo "所有服务已启动"
echo "============================================"
echo ""
echo "服务列表:"
for service_config in "${SERVICES[@]}"; do
    service=$(echo "$service_config" | cut -d: -f1)
    port=$(echo "$service_config" | cut -d: -f3)
    echo "  - $service: http://localhost:$port"
done
if [ "$GATEWAY" = true ]; then
    echo "  - api-gateway: http://localhost:8080"
fi
echo ""
echo "按 Ctrl+C 停止所有服务"

# 等待用户中断
trap "echo ''; echo '正在停止所有服务...'; kill ${PIDS[@]} 2>/dev/null; echo '所有服务已停止'; exit" INT TERM

wait

