#!/bin/bash

# 启动用户服务脚本

set -e

echo "============================================"
echo "启动用户服务"
echo "============================================"

# 检查配置文件
CONFIG_FILE="configs/dev/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    echo "请先复制 configs/dev/config.yaml.example 为 configs/dev/config.yaml"
    exit 1
fi

# 检查是否已编译
if [ ! -f "bin/user-service" ]; then
    echo "正在编译用户服务..."
    make build-service SERVICE=user-service
fi

# 启动服务
echo "启动用户服务..."
./bin/user-service -f "$CONFIG_FILE"

