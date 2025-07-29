# PowerShell构建脚本 - 在Windows环境下为Debian 12编译二进制文件

Write-Host "开始构建 alist2strm 项目..." -ForegroundColor Green

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

# 获取构建时间
$BUILD_TIME = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"

# 获取Go版本
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

# 设置环境变量
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# 设置编译标志
$LDFLAGS = "-w -s"
$LDFLAGS += " -X main.Version=$VERSION"
$LDFLAGS += " -X main.BuildTime=$BUILD_TIME"
$LDFLAGS += " -X main.GoVersion=$GO_VERSION"

Write-Host "开始编译..." -ForegroundColor Yellow

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

# 启动应用
echo "启动 alist2strm..."
./alist2strm
"@
    
    $startScriptPath = Join-Path $BUILD_DIR "start.sh"
    $startScript | Out-File -FilePath $startScriptPath -Encoding UTF8
    Write-Host "已创建启动脚本: $startScriptPath"
    
    # 创建README文件
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
   或直接运行：
   ```bash
   ./alist2strm
   ```

## 系统服务部署

如需作为系统服务运行，可以创建 systemd 服务文件：

```bash
sudo nano /etc/systemd/system/alist2strm.service
```

服务文件内容：
```ini
[Unit]
Description=AList2STRM Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/alist2strm
ExecStart=/path/to/alist2strm/alist2strm
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用并启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable alist2strm
sudo systemctl start alist2strm
```

## 构建信息

- 版本: $VERSION
- 构建时间: $BUILD_TIME
- Go版本: $GO_VERSION
- 目标平台: linux/amd64
"@
    
    $readmePath = Join-Path $BUILD_DIR "README.md"
    $readme | Out-File -FilePath $readmePath -Encoding UTF8
    Write-Host "已创建部署说明: $readmePath"
    
    Write-Host ""
    Write-Host "构建完成! 使用方法:" -ForegroundColor Green
    Write-Host "1. 将 $BUILD_DIR 目录复制到 Debian 12 系统"
    Write-Host "2. 在Linux系统中运行: chmod +x alist2strm start.sh"
    Write-Host "3. 编辑 .env 文件配置应用设置"
    Write-Host "4. 运行 ./start.sh 启动应用"
    Write-Host ""
    Write-Host "或者直接运行: ./alist2strm"
    
} else {
    Write-Host "编译失败!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "按任意键继续..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")