#!/bin/bash

# 配置中心配置管理脚本
# 用于将配置文件同步到 etcd

set -e

# 配置变量
ETCD_HOST="${ETCD_HOST:-localhost:2379}"
ENV="${ENV:-dev}"
CONFIG_FILE="${CONFIG_FILE:-configs/${ENV}/business.json}"
ETCD_KEY="ecommerce/${ENV}/business/config"

echo "============================================"
echo "配置中心配置同步工具"
echo "============================================"
echo "etcd 地址: $ETCD_HOST"
echo "环境: $ENV"
echo "配置文件: $CONFIG_FILE"
echo "etcd key: $ETCD_KEY"
echo "============================================"

# 检查 etcd 是否可用
if ! command -v etcdctl &> /dev/null; then
    echo "错误: 未找到 etcdctl 命令，请先安装 etcd 客户端"
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 验证 JSON 格式
if ! python3 -m json.tool "$CONFIG_FILE" > /dev/null 2>&1; then
    echo "错误: 配置文件 JSON 格式无效"
    exit 1
fi

# 同步配置到 etcd
echo "正在同步配置到 etcd..."
cat "$CONFIG_FILE" | etcdctl --endpoints="$ETCD_HOST" put "$ETCD_KEY"

if [ $? -eq 0 ]; then
    echo "✓ 配置同步成功"
    
    # 验证配置
    echo "验证配置..."
    etcdctl --endpoints="$ETCD_HOST" get "$ETCD_KEY" --print-value-only | python3 -m json.tool > /dev/null
    if [ $? -eq 0 ]; then
        echo "✓ 配置验证通过"
    else
        echo "警告: 配置验证失败"
    fi
else
    echo "错误: 配置同步失败"
    exit 1
fi

echo "============================================"
echo "配置同步完成！"
echo "============================================"

