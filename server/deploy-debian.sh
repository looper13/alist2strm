#!/bin/bash

# Debian 12 部署脚本
set -e

APP_NAME="alist2strm"
SERVICE_USER="alist2strm"
INSTALL_DIR="/opt/$APP_NAME"
SERVICE_FILE="/etc/systemd/system/$APP_NAME.service"

echo "=== Debian 12 部署脚本 ==="
echo "应用名称: $APP_NAME"
echo "安装目录: $INSTALL_DIR"

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用 root 权限运行此脚本"
    echo "使用: sudo $0"
    exit 1
fi

# 检查二进制文件是否存在
if [ ! -f "build/$APP_NAME" ]; then
    echo "错误: 未找到编译后的二进制文件"
    echo "请先运行: ./build.sh"
    exit 1
fi

echo "开始部署..."

# 创建系统用户
if ! id "$SERVICE_USER" &>/dev/null; then
    echo "创建系统用户: $SERVICE_USER"
    useradd --system --shell /bin/false --home-dir $INSTALL_DIR --create-home $SERVICE_USER
else
    echo "用户 $SERVICE_USER 已存在"
fi

# 创建安装目录
echo "创建安装目录: $INSTALL_DIR"
mkdir -p $INSTALL_DIR
mkdir -p $INSTALL_DIR/logs
mkdir -p $INSTALL_DIR/data

# 复制文件
echo "复制应用文件..."
cp build/$APP_NAME $INSTALL_DIR/
chmod +x $INSTALL_DIR/$APP_NAME

# 复制配置文件
if [ -f "build/.env.example" ]; then
    cp build/.env.example $INSTALL_DIR/
fi

if [ ! -f "$INSTALL_DIR/.env" ]; then
    if [ -f "$INSTALL_DIR/.env.example" ]; then
        cp $INSTALL_DIR/.env.example $INSTALL_DIR/.env
        echo "已创建默认配置文件: $INSTALL_DIR/.env"
    fi
fi

# 设置文件权限
chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
chmod 755 $INSTALL_DIR
chmod 644 $INSTALL_DIR/.env* 2>/dev/null || true

# 创建systemd服务文件
echo "创建 systemd 服务..."
cat > $SERVICE_FILE << EOF
[Unit]
Description=AList2STRM Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$APP_NAME
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$APP_NAME

# 环境变量
Environment=GIN_MODE=release

# 安全设置
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF

# 重新加载systemd配置
echo "重新加载 systemd 配置..."
systemctl daemon-reload

# 启用服务
echo "启用服务..."
systemctl enable $APP_NAME

echo ""
echo "=== 部署完成! ==="
echo ""
echo "配置文件位置: $INSTALL_DIR/.env"
echo "日志目录: $INSTALL_DIR/logs"
echo "数据目录: $INSTALL_DIR/data"
echo ""
echo "服务管理命令:"
echo "  启动服务: sudo systemctl start $APP_NAME"
echo "  停止服务: sudo systemctl stop $APP_NAME"
echo "  重启服务: sudo systemctl restart $APP_NAME"
echo "  查看状态: sudo systemctl status $APP_NAME"
echo "  查看日志: sudo journalctl -u $APP_NAME -f"
echo ""
echo "请编辑配置文件后启动服务:"
echo "  sudo nano $INSTALL_DIR/.env"
echo "  sudo systemctl start $APP_NAME"