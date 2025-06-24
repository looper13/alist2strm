# 基础镜像 - 前端构建
FROM node:22.16.0-alpine3.22 AS frontend-build
WORKDIR /app/frontend

# 添加阿里云镜像源并安装依赖
RUN echo "https://mirrors.aliyun.com/alpine/v3.22/main/" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.22/community/" >> /etc/apk/repositories && \
    apk update && \
    corepack enable && corepack prepare pnpm@10.11.0 --activate && \
    pnpm config set registry https://registry.npmmirror.com && \
    apk add --no-cache python3 build-base

# 复制前端代码并构建
COPY frontend/ ./
RUN pnpm install --frozen-lockfile --force && pnpm run build 

# 后端构建镜像
FROM golang:1.24.4-alpine3.22 AS backend-build
WORKDIR /app/server

# 添加阿里云镜像源并安装构建依赖
RUN echo "https://mirrors.aliyun.com/alpine/v3.22/main/" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.22/community/" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache build-base git ca-certificates tzdata sqlite-dev

# 复制后端代码
COPY server/ ./

# 设置 Go 环境变量
ENV CGO_ENABLED=1 \
    GOOS=linux

# 下载依赖并构建，根据目标架构进行编译
ARG TARGETARCH
RUN go mod download && \
    GOARCH=$TARGETARCH go build -ldflags="-s -w" -o alist2strm .

# 最终运行镜像
FROM alpine:3.22.0

WORKDIR /app

# 添加阿里云镜像源并安装必要工具
RUN echo "https://mirrors.aliyun.com/alpine/v3.22/main/" > /etc/apk/repositories && \
    echo "https://mirrors.aliyun.com/alpine/v3.22/community/" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache nginx shadow su-exec tzdata sqlite-libs sqlite-dev ca-certificates && \
    rm -rf /var/cache/apk/*

# 默认环境变量
ENV TZ=Asia/Shanghai \
    LOG_BASE_DIR=/app/data/logs \
    LOG_LEVEL=info \
    LOG_APP_NAME=alist2strm \
    LOG_MAX_DAYS=30 \
    LOG_MAX_FILE_SIZE=10 \
    DB_BASE_DIR=/app/data/db \
    DB_NAME=database.sqlite \
    PUID=1000 \
    PGID=1000 \
    UMASK=022 \
    PORT=3210

# 创建必要目录
RUN mkdir -p /app/data/logs/nginx /app/data/db

# 添加配置和脚本
COPY builder/default.conf /etc/nginx/http.d/default.conf
COPY builder/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# 从构建阶段复制产物
COPY --from=backend-build /app/server/alist2strm /app/server/
COPY --from=frontend-build /app/frontend/dist /app/frontend/dist

# 开放端口
EXPOSE 80 3210

# 启动命令
ENTRYPOINT ["/entrypoint.sh"]
