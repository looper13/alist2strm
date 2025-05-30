#!/bin/sh

# 设置时区
if [ -n "$TZ" ]; then
    cp /usr/share/zoneinfo/$TZ /etc/localtime
    echo $TZ > /etc/timezone
    echo "Timezone set to $TZ"
fi

# 确保目录存在
mkdir -p /app/data/logs/nginx
mkdir -p /app/data/db

# 修改用户和组ID
groupmod -o -g ${PGID} node
usermod -o -u ${PUID} node

# 设置目录权限
chown -R node:node /app/data
chown -R node:node /app/server
chmod -R 755 /app/data

# 设置 umask
umask ${UMASK:-022}

# 启动 Nginx
nginx

# 使用 node 用户启动后端服务
exec su-exec node sh -c "cd /app/server && node dist/index.js"
