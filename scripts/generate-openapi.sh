#!/bin/bash

# 生成 OpenAPI (Swagger) 文档脚本
# 依赖插件：github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

set -e

echo "============================================"
echo "生成 OpenAPI (Swagger) 文档"
echo "============================================"

PROTO_DIR="api"
OUTPUT_DIR="docs/swagger"

mkdir -p "${OUTPUT_DIR}"

# 检查 protoc 是否安装
if ! command -v protoc >/dev/null 2>&1; then
  echo "错误: 未找到 protoc 命令，请先安装 Protocol Buffers"
  exit 1
fi

# 检查 protoc-gen-openapiv2 是否安装
if ! command -v protoc-gen-openapiv2 >/dev/null 2>&1; then
  echo "错误: 未找到 protoc-gen-openapiv2 插件"
  echo "请先安装："
  echo "  go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest"
  exit 1
fi

find "${PROTO_DIR}" -name "*.proto" | while read -r proto_file; do
  rel_path="${proto_file#${PROTO_DIR}/}"
  dir_path=$(dirname "${rel_path}")
  file_name=$(basename "${proto_file}" .proto)

  service_output_dir="${OUTPUT_DIR}/${dir_path}"
  mkdir -p "${service_output_dir}"

  echo "生成 OpenAPI: ${proto_file} -> ${service_output_dir}/${file_name}.swagger.json"
  protoc \
    -I "${PROTO_DIR}" \
    --openapiv2_out "${service_output_dir}" \
    --openapiv2_opt logtostderr=true \
    "${proto_file}"
done

echo "============================================"
echo "OpenAPI (Swagger) 文档生成完成！"
echo "输出目录: ${OUTPUT_DIR}"
echo "============================================"





