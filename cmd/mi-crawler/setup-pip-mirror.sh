#!/bin/bash
# 配置 pip 使用清华大学镜像源

echo "正在配置 pip 镜像源..."

# 创建 .pip 目录
mkdir -p ~/.pip

# 创建配置文件
cat > ~/.pip/pip.conf << EOF
[global]
index-url = https://pypi.tuna.tsinghua.edu.cn/simple
trusted-host = pypi.tuna.tsinghua.edu.cn
EOF

echo "✅ pip 镜像源配置完成！"
echo ""
echo "当前配置："
cat ~/.pip/pip.conf
echo ""
echo "现在可以使用以下命令安装依赖："
echo "  pip install -r requirements.txt"

