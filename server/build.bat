@echo off
REM 构建脚本 - 在Windows环境下为Debian 12编译二进制文件
setlocal enabledelayedexpansion

echo 开始构建 alist2strm 项目...

REM 设置构建参数
set APP_NAME=alist2strm
set BUILD_DIR=build

REM 获取版本信息（如果有git的话）
for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=dev

REM 获取构建时间
for /f "tokens=2 delims==" %%i in ('wmic os get localdatetime /value') do set datetime=%%i
set BUILD_TIME=%datetime:~0,4%-%datetime:~4,2%-%datetime:~6,2%_%datetime:~8,2%:%datetime:~10,2%:%datetime:~12,2%

REM 获取Go版本
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i

REM 设置目标平台为Linux AMD64 (Debian 12)
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=1

echo 构建信息:
echo   应用名称: %APP_NAME%
echo   版本: %VERSION%
echo   构建时间: %BUILD_TIME%
echo   Go版本: %GO_VERSION%
echo   目标平台: %GOOS%/%GOARCH%

REM 创建构建目录
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

REM 设置编译标志
set LDFLAGS=-w -s
set LDFLAGS=%LDFLAGS% -X main.Version=%VERSION%
set LDFLAGS=%LDFLAGS% -X main.BuildTime=%BUILD_TIME%
set LDFLAGS=%LDFLAGS% -X main.GoVersion=%GO_VERSION%

echo 开始编译...

REM 编译二进制文件
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%\%APP_NAME% .

REM 检查编译结果
if exist "%BUILD_DIR%\%APP_NAME%" (
    echo 编译成功!
    echo 二进制文件位置: %BUILD_DIR%\%APP_NAME%
    
    REM 显示文件信息
    dir %BUILD_DIR%\%APP_NAME%
    
    REM 复制必要的配置文件
    if exist ".env.example" (
        copy .env.example %BUILD_DIR%\
        echo 已复制 .env.example 到构建目录
    )
    
    REM 创建启动脚本
    echo #!/bin/bash > %BUILD_DIR%\start.sh
    echo # 启动脚本 >> %BUILD_DIR%\start.sh
    echo. >> %BUILD_DIR%\start.sh
    echo # 检查配置文件 >> %BUILD_DIR%\start.sh
    echo if [ ! -f ".env" ]; then >> %BUILD_DIR%\start.sh
    echo     if [ -f ".env.example" ]; then >> %BUILD_DIR%\start.sh
    echo         echo "未找到 .env 文件，正在从 .env.example 创建..." >> %BUILD_DIR%\start.sh
    echo         cp .env.example .env >> %BUILD_DIR%\start.sh
    echo         echo "请编辑 .env 文件配置您的设置" >> %BUILD_DIR%\start.sh
    echo     else >> %BUILD_DIR%\start.sh
    echo         echo "警告: 未找到配置文件" >> %BUILD_DIR%\start.sh
    echo     fi >> %BUILD_DIR%\start.sh
    echo fi >> %BUILD_DIR%\start.sh
    echo. >> %BUILD_DIR%\start.sh
    echo # 启动应用 >> %BUILD_DIR%\start.sh
    echo echo "启动 alist2strm..." >> %BUILD_DIR%\start.sh
    echo ./alist2strm >> %BUILD_DIR%\start.sh
    
    echo 已创建启动脚本: %BUILD_DIR%\start.sh
    
    echo.
    echo 构建完成! 使用方法:
    echo 1. 将 %BUILD_DIR% 目录复制到 Debian 12 系统
    echo 2. 在Linux系统中运行: chmod +x alist2strm start.sh
    echo 3. 编辑 .env 文件配置应用设置
    echo 4. 运行 ./start.sh 启动应用
    echo.
    echo 或者直接运行: ./alist2strm
    
) else (
    echo 编译失败!
    exit /b 1
)

pause