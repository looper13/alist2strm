#!/bin/sh

# 创建必要的目录
mkdir -p /app/data/logs/nginx
mkdir -p /app/data/db

# 启动 Nginx
nginx

# 启动后端服务
cd /app/server && node dist/index.js
