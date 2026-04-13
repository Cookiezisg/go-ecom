#!/bin/bash

# 生成 Protobuf 代码脚本
# 用于从 .proto 文件生成 Go 代码和 gRPC 代码

set -e

echo "============================================"
echo "生成 Protobuf 代码"
echo "============================================"

# 检查 protoc 是否安装
if ! command -v protoc >/dev/null 2>&1; then
    echo "错误: 未找到 protoc 命令，请先安装 Protocol Buffers"
    echo "下载地址: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go >/dev/null 2>&1; then
    echo "错误: 未找到 protoc-gen-go 插件"
    echo "请先安装："
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# 检查 protoc-gen-go-grpc 是否安装
if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
    echo "错误: 未找到 protoc-gen-go-grpc 插件"
    echo "请先安装："
    echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# 定义 proto 文件目录
PROTO_DIR="api"

# 检查目录是否存在
if [ ! -d "$PROTO_DIR" ]; then
    echo "错误: 未找到 $PROTO_DIR 目录"
    exit 1
fi

# 查找所有 .proto 文件
proto_files=$(find "$PROTO_DIR" -name "*.proto" | sort)

if [ -z "$proto_files" ]; then
    echo "警告: 未找到任何 .proto 文件"
    exit 0
fi

# 统计文件数量
file_count=$(echo "$proto_files" | wc -l | tr -d ' ')
echo "找到 $file_count 个 .proto 文件"
echo ""

# 处理每个 .proto 文件
success_count=0
fail_count=0

while IFS= read -r proto_file; do
    # 获取相对路径（相对于项目根目录）
    rel_path="${proto_file#$PROTO_DIR/}"
    
    echo "处理: $rel_path"
    
    # 执行 protoc 命令
    # --proto_path=. 指定导入路径的根目录
    # --go_out=. 生成 Go 代码输出目录
    # --go_opt=paths=source_relative 使用相对路径
    # --go-grpc_out=. 生成 gRPC Go 代码输出目录
    # --go-grpc_opt=paths=source_relative 使用相对路径
    if protoc \
        --proto_path=. \
        --go_out=. \
        --go_opt=paths=source_relative \
        --go-grpc_out=. \
        --go-grpc_opt=paths=source_relative \
        "$proto_file" 2>&1; then
        echo "  ✓ 成功"
        ((success_count++))
    else
        echo "  ✗ 失败"
        ((fail_count++))
    fi
    echo ""
done <<< "$proto_files"

echo "============================================"
echo "Protobuf 代码生成完成！"
echo "成功: $success_count 个文件"
if [ $fail_count -gt 0 ]; then
    echo "失败: $fail_count 个文件"
    exit 1
fi
echo "============================================"

