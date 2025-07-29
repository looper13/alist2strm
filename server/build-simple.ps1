# 简化构建脚本 - 尝试使用纯Go构建

Write-Host "开始构建 alist2strm 项目 (纯Go版本)..." -ForegroundColor Green

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
$GO_VERSION = (go version).Split()[2]

Write-Host "构建信息:" -ForegroundColor Yellow
Write-Host "  应用名称: $APP_NAME"
Write-Host "  版本: $VERSION"
Write-Host "  构建时间: $BUILD_TIME"
Write-Host "  Go版本: $GO_VERSION"
Write-Host "  目标平台: linux/amd64"

# 创建构建目录
if (-not (Test-Path $BUILD_DIR)) {
    New-Item -ItemType Directory -Path $BUILD_DIR | Out-Null
}

# 设置环境变量 - 禁用CGO
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

# 设置编译标志
$LDFLAGS = "-w -s"
$LDFLAGS += " -X main.Version=$VERSION"
$LDFLAGS += " -X main.BuildTime=$BUILD_TIME"
$LDFLAGS += " -X main.GoVersion=$GO_VERSION"

Write-Host "开始编译 (CGO_ENABLED=1)..." -ForegroundColor Yellow

# 编译二进制文件
$buildPath = Join-Path $BUILD_DIR $APP_NAME
& go build -ldflags $LDFLAGS -o $buildPath .

# 检查编译结果
if (Test-Path $buildPath) {
    Write-Host "编译成功!" -ForegroundColor Green
    Write-Host "二进制文件位置: $buildPath"
    
    # 显示文件信息
    Get-ChildItem $buildPath | Format-Table Name, Length, LastWriteTime
    
    # 复制必要的配置文件
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" $BUILD_DIR
        Write-Host "已复制 .env.example 到构建目录"
    }
    
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
    Write-Host "已创建启动脚本: $startScriptPath"
    
    # 创建部署说明
    $readme = @"
# AList2STRM - Debian 12 部署包

## 部署步骤

1. 将此目录上传到 Debian 12 系统
2. 设置执行权限：
   ```bash
   chmod +x alist2strm start.sh
   ```
3. 配置应用：
   ```bash
   cp .env.example .env
   nano .env  # 编辑配置文件
   ```
4. 启动应用：
   ```bash
   ./start.sh
   ```

## 注意事项

此版本使用纯Go编译（CGO_ENABLED=0），如果遇到SQLite相关问题，
请使用Docker版本或在Linux系统上直接编译。

## 构建信息

- 版本: $VERSION
- 构建时间: $BUILD_TIME
- Go版本: $GO_VERSION
- 目标平台: linux/amd64
- CGO: 禁用
"@
    
    $readmePath = Join-Path $BUILD_DIR "README.md"
    $readme | Out-File -FilePath $readmePath -Encoding UTF8
    Write-Host "已创建部署说明: $readmePath"
    
    Write-Host ""
    Write-Host "构建完成!" -ForegroundColor Green
    Write-Host ""
    Write-Host "注意: 此版本禁用了CGO，如果应用使用了需要CGO的SQLite驱动，"
    Write-Host "可能会在运行时出现问题。建议使用以下替代方案："
    Write-Host ""
    Write-Host "1. 使用Docker构建: powershell -ExecutionPolicy Bypass -File build-docker.ps1"
    Write-Host "2. 在Linux系统上直接编译"
    Write-Host "3. 使用GitHub Actions等CI/CD服务进行交叉编译"
    
} else {
    Write-Host "编译失败!" -ForegroundColor Red
    Write-Host ""
    Write-Host "可能的原因:"
    Write-Host "1. 项目依赖需要CGO支持的库（如SQLite驱动）"
    Write-Host "2. 缺少必要的依赖"
    Write-Host ""
    Write-Host "建议使用Docker构建: powershell -ExecutionPolicy Bypass -File build-docker.ps1"
    exit 1
}

Write-Host ""
Write-Host "按任意键继续..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")