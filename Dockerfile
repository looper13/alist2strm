# 使用 node 作为基础镜像
FROM node:24-slim

# 安装必要的依赖
RUN apt-get update && apt-get install -y \
    nginx \
    python3 \
    make \
    g++ \
    sqlite3 \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# 创建 nginx 所需的目录并设置权限
RUN mkdir -p /run/nginx && \
    mkdir -p /var/lib/nginx && \
    mkdir -p /var/lib/nginx/tmp && \
    mkdir -p /usr/share/nginx/html && \
    chown -R www-data:www-data /var/lib/nginx && \
    chown -R www-data:www-data /var/log/nginx && \
    chown -R www-data:www-data /run/nginx && \
    chown -R www-data:www-data /usr/share/nginx/html

# 设置工作目录
WORKDIR /app

# 复制后端构建产物和必要文件
COPY packages/server/dist ./server
COPY packages/server/package.json ./server/
COPY packages/server/pnpm-lock.yaml ./server/

# 复制前端构建产物
COPY packages/web/dist /usr/share/nginx/html
RUN chown -R www-data:www-data /usr/share/nginx/html

# 删除默认的 nginx 配置
RUN rm -f /etc/nginx/sites-enabled/default

# 配置 nginx
COPY buider/default.conf /etc/nginx/conf.d/default.conf
RUN chmod 644 /etc/nginx/conf.d/default.conf

# 安装 pnpm
RUN npm install -g pnpm

# 安装后端依赖
WORKDIR /app/server

# 设置 node-gyp 和 sqlite3 编译相关的环境变量
ENV PYTHON=/usr/bin/python3
ENV NODE_GYP_FORCE_PYTHON=/usr/bin/python3
ENV npm_config_build_from_source=true

# 安装依赖并重新编译 sqlite3
RUN pnpm install --prod && \
    cd node_modules/sqlite3 && \
    pnpm rebuild

# 创建数据目录
RUN mkdir -p /data/logs /data/db && \
    chown -R node:node /data

# 创建启动脚本
RUN echo '#!/bin/sh\nnginx -g "daemon off;" &\ncd /app/server && node app.js' > /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# 暴露端口
EXPOSE 3000 80

# 启动应用
CMD ["/app/entrypoint.sh"] 