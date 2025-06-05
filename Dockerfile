# 基础镜像
FROM node:22.16.0-alpine AS base

RUN corepack enable && corepack prepare pnpm@10.11.0 --activate && \
    pnpm config set registry https://registry.npmmirror.com && \
    apk add --no-cache \
    python3 \
    build-base \
    sqlite-dev \
    musl-dev \
    tzdata
WORKDIR /app

# ---------- 后端构建 ----------
FROM base AS backend-build-dev
WORKDIR /app/server
COPY packages/server/ ./
RUN pnpm install --frozen-lockfile --force && \
    pnpm run build

FROM base AS backend-build
WORKDIR /app/server
COPY packages/server/package.json packages/server/pnpm-lock.yaml ./
RUN pnpm install --production --frozen-lockfile --force


# ---------- 前端构建 ----------
FROM base AS frontend-build
WORKDIR /app/frontend
COPY packages/frontend/ ./
RUN pnpm install --frozen-lockfile --force && pnpm run build 

# ---------- 最终运行镜像 ----------
FROM node:22.16.0-alpine
WORKDIR /app

# nginx 和必要工具
RUN apk update && \
    apk add --no-cache nginx shadow su-exec tzdata && \
    rm -rf /var/cache/apk/*

# 默认环境
ENV PORT=3210 \
    TZ=Asia/Shanghai \
    LOG_BASE_DIR=/app/data/logs \
    LOG_LEVEL=info \
    LOG_APP_NAME=alist2strm \
    LOG_MAX_DAYS=30 \
    LOG_MAX_FILE_SIZE=10 \
    DB_BASE_DIR=/app/data/db \
    DB_NAME=database.sqlite \
    PUID=1000 \
    PGID=1000 \
    UMASK=022

# 创建必要的目录
RUN mkdir -p /app/data/logs/nginx /app/data/db

# 脚本
COPY builder/default.conf /etc/nginx/http.d/default.conf
COPY builder/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# 产物
COPY --from=backend-build-dev /app/server/dist /app/server/dist
COPY --from=backend-build /app/server/node_modules /app/server/node_modules
COPY --from=frontend-build /app/frontend/dist /app/frontend/dist

EXPOSE 80 3210

# entrypoint
ENTRYPOINT ["/entrypoint.sh"]