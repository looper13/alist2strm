# Docker构建脚本 - 使用Docker容器进行交叉编译

Write-Host "使用Docker进行交叉编译..." -ForegroundColor Green

# 检查Docker是否可用
try {
    docker --version | Out-Null
} catch {
    Write-Host "错误: Docker未安装或不可用" -ForegroundColor Red
    Write-Host "请安装Docker Desktop: https://www.docker.com/products/docker-desktop"
    exit 1
}

# 设置构建参数
$APP_NAME = "alist2strm"
$BUILD_DIR = "build"

# 获取版本信息
try {
    $VERSION = git describe --tags --always --dirty 2>$null
    if (-not $VERSION) { $VERSION = "dev" }
} catch {
    $VERSION = "dev"
}

$BUILD_TIME = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"

Write-Host "构建信息:" -ForegroundColor Yellow
Write-Host "  应用名称: $APP_NAME"
Write-Host "  版本: $VERSION"
Write-Host "  构建时间: $BUILD_TIME"
Write-Host "  构建方式: Docker交叉编译"

# 创建构建目录
if (-not (Test-Path $BUILD_DIR)) {
    New-Item -ItemType Directory -Path $BUILD_DIR | Out-Null
}

# 创建Dockerfile用于构建
$dockerfile = @"
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 设置构建参数
ARG VERSION=dev
ARG BUILD_TIME
ARG GO_VERSION

# 编译应用
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-w -s -X main.Version=`$VERSION -X main.BuildTime=`$BUILD_TIME -X main.GoVersion=`$GO_VERSION" \
    -o alist2strm .

# 最终镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite

# 创建非root用户
RUN addgroup -g 1001 -S alist2strm && \
    adduser -u 1001 -S alist2strm -G alist2strm

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/alist2strm .
COPY --from=builder /app/.env.example .

# 创建数据目录
RUN mkdir -p data logs && \
    chown -R alist2strm:alist2strm /app

# 切换到非root用户
USER alist2strm

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./alist2strm"]
"@

$dockerfilePath = "Dockerfile.build"
$dockerfile | Out-File -FilePath $dockerfilePath -Encoding UTF8

Write-Host "开始Docker构建..." -ForegroundColor Yellow

# 构建Docker镜像并提取二进制文件
$containerName = "alist2strm-builder-$(Get-Random)"

try {
    # 构建镜像
    docker build -f $dockerfilePath -t alist2strm-builder --build-arg VERSION=$VERSION --build-arg BUILD_TIME=$BUILD_TIME --build-arg GO_VERSION="go1.24.4" .
    
    if ($LASTEXITCODE -ne 0) {
        throw "Docker构建失败"
    }
    
    # 创建容器并复制文件
    docker create --name $containerName alist2strm-builder
    docker cp "${containerName}:/app/alist2strm" "$BUILD_DIR/"
    docker cp "${containerName}:/app/.env.example" "$BUILD_DIR/" 2>$null
    
    Write-Host "构建成功!" -ForegroundColor Green
    
    # 显示文件信息
    $buildPath = Join-Path $BUILD_DIR $APP_NAME
    if (Test-Path $buildPath) {
        Get-ChildItem $buildPath | Format-Table Name, Length, LastWriteTime
        
        # 创建启动脚本
        $startScript = @"
#!/bin/bash
# 启动脚本

# 检查配置文件
if [ ! -f ".env" ]; then
    if [ -f ".env.example" ]; then
        echo "未找到 .env 文件，正在从 .env.example 创建..."
        cp .env.example .env
        echo "请编辑 .env 文件配置您的设置"
    else
        echo "警告: 未找到配置文件"
    fi
fi

# 创建必要的目录
mkdir -p data logs

# 启动应用
echo "启动 alist2strm..."
./alist2strm
"@
        
        $startScriptPath = Join-Path $BUILD_DIR "start.sh"
        $startScript | Out-File -FilePath $startScriptPath -Encoding UTF8
        
        # 创建Docker运行脚本
        $dockerRunScript = @"
#!/bin/bash
# Docker运行脚本

# 构建运行镜像
docker build -f Dockerfile.run -t alist2strm:latest .

# 运行容器
docker run -d \
  --name alist2strm \
  -p 8080:8080 \
  -v `$(pwd)/data:/app/data \
  -v `$(pwd)/logs:/app/logs \
  -v `$(pwd)/.env:/app/.env \
  --restart unless-stopped \
  alist2strm:latest

echo "容器已启动，访问 http://localhost:8080"
echo "查看日志: docker logs -f alist2strm"
echo "停止容器: docker stop alist2strm"
echo "删除容器: docker rm alist2strm"
"@
        
        $dockerRunScriptPath = Join-Path $BUILD_DIR "docker-run.sh"
        $dockerRunScript | Out-File -FilePath $dockerRunScriptPath -Encoding UTF8
        
        # 创建运行时Dockerfile
        $runDockerfile = @"
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite

# 创建非root用户
RUN addgroup -g 1001 -S alist2strm && \
    adduser -u 1001 -S alist2strm -G alist2strm

# 设置工作目录
WORKDIR /app

# 复制二进制文件和配置
COPY alist2strm .
COPY .env.example .

# 创建数据目录
RUN mkdir -p data logs && \
    chown -R alist2strm:alist2strm /app

# 切换到非root用户
USER alist2strm

# 暴露端口
EXPOSE 8080

# 启动命令
CMD ["./alist2strm"]
"@
        
        $runDockerfilePath = Join-Path $BUILD_DIR "Dockerfile.run"
        $runDockerfile | Out-File -FilePath $runDockerfilePath -Encoding UTF8
        
        Write-Host "已创建启动脚本: $startScriptPath"
        Write-Host "已创建Docker运行脚本: $dockerRunScriptPath"
        Write-Host "已创建运行时Dockerfile: $runDockerfilePath"
        
        Write-Host ""
        Write-Host "构建完成! 部署选项:" -ForegroundColor Green
        Write-Host "1. 直接部署到Linux:"
        Write-Host "   - 将 $BUILD_DIR 目录复制到 Debian 12 系统"
        Write-Host "   - 运行: chmod +x alist2strm start.sh"
        Write-Host "   - 编辑 .env 文件并运行 ./start.sh"
        Write-Host ""
        Write-Host "2. 使用Docker部署:"
        Write-Host "   - 将 $BUILD_DIR 目录复制到目标系统"
        Write-Host "   - 运行: chmod +x docker-run.sh && ./docker-run.sh"
    }
    
} catch {
    Write-Host "构建失败: $_" -ForegroundColor Red
    exit 1
} finally {
    # 清理
    docker rm $containerName 2>$null | Out-Null
    docker rmi alist2strm-builder 2>$null | Out-Null
    Remove-Item $dockerfilePath -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "按任意键继续..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")