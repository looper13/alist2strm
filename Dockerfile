# 使用 node 作为基础镜像
FROM node:23.11.1-alpine AS base
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app

# ---------- 后端构建阶段（包括 devDependencies） ----------
FROM base AS backend-build
WORKDIR /app/server

# 配置 Alpine 镜像源并安装编译工具链
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
    python3 \
    build-base \
    sqlite-dev \
    musl-dev

# 复制后端 package.json 和锁文件
COPY packages/server/package.json packages/server/pnpm-lock.yaml ./
# 使用国内镜像，加速安装
RUN pnpm config set registry https://registry.npmmirror.com
# 安装所有依赖（包括 devDependencies）
RUN pnpm install --frozen-lockfile
# 复制后端源代码并构建
COPY packages/server/ ./
RUN pnpm run build

# ---------- 前端构建阶段 ----------
FROM base AS frontend-build
WORKDIR /app/frontend
# 复制前端 package.json 和锁文件
COPY packages/frontend/package.json packages/frontend/pnpm-*.yaml ./
RUN pnpm config set registry https://registry.npmmirror.com
# 安装依赖并构建前端
RUN pnpm install --frozen-lockfile
COPY packages/frontend/ ./
RUN pnpm run build 

# ---------- 最终运行镜像 ----------
FROM node:23.11.1-alpine AS final
WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate

# 安装运行时环境：nginx 和 sqlite
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
    nginx \
    sqlite \
    sqlite-dev \
    python3 \
    build-base \
    musl-dev

# 安装 node-gyp
RUN npm install -g node-gyp

# 复制 Nginx 配置和 entrypoint 脚本
COPY builder/default.conf /etc/nginx/http.d/default.conf
COPY builder/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# 复制后端运行时产物（dist 和 node_modules）
COPY --from=backend-build /app/server /app/server

# 重新编译 sqlite3
WORKDIR /app/server
RUN cd node_modules/sqlite3 && \
    node-gyp rebuild

# 复制前端静态产物
COPY --from=frontend-build /app/frontend/dist /app/frontend/dist

# 设置默认环境变量
ENV PORT=3210 \
    LOG_BASE_DIR=/app/data/logs \
    LOG_LEVEL=info \
    LOG_APP_NAME=alist2strm \
    LOG_MAX_DAYS=30 \
    LOG_MAX_FILE_SIZE=10 \
    DB_BASE_DIR=/app/data/db \
    DB_NAME=database.sqlite

EXPOSE 80 3210

# 设置 entrypoint
ENTRYPOINT ["/entrypoint.sh"]