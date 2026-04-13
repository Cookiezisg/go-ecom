#!/bin/bash

# 启动单个微服务脚本 (Linux/Mac)
# 使用方法: ./scripts/start-service.sh user-service

set -e

if [ $# -eq 0 ]; then
    echo "用法: $0 <service-name> [config-file]"
    echo ""
    echo "可用服务:"
    echo "  user-service, product-service, order-service, payment-service"
    echo "  inventory-service, cart-service, promotion-service, review-service"
    echo "  logistics-service, message-service, search-service, recommend-service"
    echo "  file-service, job-service, api-gateway"
    exit 1
fi

SERVICE_NAME=$1
CONFIG_FILE=$2

# 服务配置映射
declare -A SERVICE_CONFIGS=(
    ["user-service"]="configs/dev/config.yaml"
    ["product-service"]="configs/dev/product-config.yaml"
    ["order-service"]="configs/dev/order-config.yaml"
    ["payment-service"]="configs/dev/payment-config.yaml"
    ["inventory-service"]="configs/dev/inventory-config.yaml"
    ["cart-service"]="configs/dev/cart-config.yaml"
    ["promotion-service"]="configs/dev/promotion-config.yaml"
    ["review-service"]="configs/dev/review-config.yaml"
    ["logistics-service"]="configs/dev/logistics-config.yaml"
    ["message-service"]="configs/dev/message-config.yaml"
    ["search-service"]="configs/dev/search-config.yaml"
    ["recommend-service"]="configs/dev/recommend-config.yaml"
    ["file-service"]="configs/dev/file-config.yaml"
    ["job-service"]="configs/dev/job-config.yaml"
    ["api-gateway"]="configs/dev/gateway.yaml"
)

# 确定配置文件
if [ -z "$CONFIG_FILE" ]; then
    if [ -z "${SERVICE_CONFIGS[$SERVICE_NAME]}" ]; then
        echo "错误: 未知的服务名称: $SERVICE_NAME"
        echo "可用服务: ${!SERVICE_CONFIGS[@]}"
        exit 1
    fi
    CONFIG_FILE="${SERVICE_CONFIGS[$SERVICE_NAME]}"
fi

echo "============================================"
echo "启动服务: $SERVICE_NAME"
echo "============================================"
echo ""

# 检查配置文件
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    if [ -f "${CONFIG_FILE}.example" ]; then
        echo "请先复制示例文件: ${CONFIG_FILE}.example -> $CONFIG_FILE"
    fi
    exit 1
fi

# 启动服务
echo "启动 $SERVICE_NAME..."
echo "配置文件: $CONFIG_FILE"
echo ""

if [ -f "bin/$SERVICE_NAME" ]; then
    # 使用编译后的可执行文件
    ./bin/$SERVICE_NAME -f "$CONFIG_FILE"
else
    # 直接运行 Go 程序
    go run "cmd/$SERVICE_NAME/main.go" -f "$CONFIG_FILE"
fi

