#!/bin/bash
# 为 Gateway 配置中的所有路由添加 OPTIONS 方法支持（用于 CORS 预检请求）

CONFIG_FILE="configs/dev/gateway.yaml"
TEMP_FILE=$(mktemp)

# 读取配置文件，为每个非 OPTIONS 的路由添加对应的 OPTIONS 路由
awk '
/^      - Method:/ {
    method = $3
    getline
    path = $2
    getline
    rpcpath = $2
    
    # 如果不是 OPTIONS 方法，先添加 OPTIONS 路由
    if (method != "options") {
        print "      - Method: options"
        print "        Path: " path
        print "        RpcPath: " rpcpath
    }
    
    # 打印原始路由
    print "      - Method: " method
    print "        Path: " path
    print "        RpcPath: " rpcpath
    next
}
{ print }
' "$CONFIG_FILE" > "$TEMP_FILE"

mv "$TEMP_FILE" "$CONFIG_FILE"
echo "已为所有路由添加 OPTIONS 方法支持"

