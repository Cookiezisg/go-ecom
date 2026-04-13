#!/bin/bash

# etcd 可视化查看脚本
# 提供多种方式查看 etcd 数据

ETCD_ENDPOINT="localhost:2379"

echo "=========================================="
echo "etcd 数据查看工具"
echo "=========================================="
echo ""

# 检查 etcd 是否运行
if ! docker ps | grep -q "infra-etcd"; then
    echo "❌ etcd 容器未运行，请先启动基础设施服务："
    echo "   docker-compose -f docker-compose-infra.yml up -d etcd"
    echo ""
    exit 1
fi

echo "✅ etcd 正在运行 (${ETCD_ENDPOINT})"
echo ""

# 方法 1: 使用 Docker 容器中的 etcdctl
echo "方法 1: 使用 etcdctl 命令行工具"
echo "----------------------------------------"
echo "查看所有键:"
docker exec infra-etcd etcdctl get --prefix "" --keys-only
echo ""
echo "查看所有键值对:"
docker exec infra-etcd etcdctl get --prefix ""
echo ""
echo "查看特定前缀的键 (例如: user-service):"
docker exec infra-etcd etcdctl get --prefix "user-service"
echo ""

# 方法 2: Web UI
echo "方法 2: Web UI 可视化界面"
echo "----------------------------------------"
if docker ps | grep -q "infra-etcd-keeper"; then
    echo "✅ etcd-keeper Web UI 正在运行"
    echo "   访问地址: http://localhost:8089"
    echo "   在浏览器中打开即可可视化查看和管理 etcd 数据"
else
    echo "⚠️  etcd-keeper Web UI 未运行"
    echo "   启动命令: docker-compose -f docker-compose-infra.yml up -d etcd-keeper"
    echo "   然后访问: http://localhost:8089"
fi
echo ""

# 方法 3: 使用 etcdctl 的常用命令
echo "方法 3: 常用 etcdctl 命令"
echo "----------------------------------------"
echo "查看所有键（仅键名）:"
echo "  docker exec infra-etcd etcdctl get --prefix '' --keys-only"
echo ""
echo "查看所有键值对:"
echo "  docker exec infra-etcd etcdctl get --prefix ''"
echo ""
echo "查看特定键的值:"
echo "  docker exec infra-etcd etcdctl get <key>"
echo ""
echo "查看特定前缀的所有键:"
echo "  docker exec infra-etcd etcdctl get --prefix <prefix>"
echo ""
echo "删除键:"
echo "  docker exec infra-etcd etcdctl del <key>"
echo ""
echo "设置键值:"
echo "  docker exec infra-etcd etcdctl put <key> <value>"
echo ""

# 显示当前注册的服务
echo "当前注册的服务:"
echo "----------------------------------------"
docker exec infra-etcd etcdctl get --prefix "" --keys-only | grep -E "\.rpc$" || echo "  暂无服务注册"
echo ""

