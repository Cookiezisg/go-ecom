#!/bin/bash

# 生成 proto descriptor 文件脚本
# 用于 Gateway 的 protoDescriptor 模式

set -e

echo "============================================"
echo "生成 Proto Descriptor 文件"
echo "============================================"

# 定义 proto 文件目录
PROTO_DIR="api"
OUTPUT_DIR="api"

# 查找所有 proto 文件
find "$PROTO_DIR" -name "*.proto" | while read proto_file; do
    # 获取相对路径
    rel_path="${proto_file#$PROTO_DIR/}"
    dir_path=$(dirname "$rel_path")
    file_name=$(basename "$proto_file" .proto)
    
    # 创建输出目录
    output_path="$OUTPUT_DIR/$dir_path"
    mkdir -p "$output_path"
    
    # 生成 descriptor 文件
    echo "生成: $proto_file -> $output_path/$file_name.pb"
    protoc --descriptor_set_out="$output_path/$file_name.pb" "$proto_file"
    
    if [ $? -eq 0 ]; then
        echo "✓ 成功: $file_name.pb"
    else
        echo "✗ 失败: $file_name.pb"
        exit 1
    fi
done

echo "============================================"
echo "Proto Descriptor 文件生成完成！"
echo "============================================"

