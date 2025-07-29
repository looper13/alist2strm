# Makefile for alist2strm

APP_NAME := alist2strm
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')

# 编译标志
LDFLAGS := -w -s
LDFLAGS += -X main.Version=$(VERSION)
LDFLAGS += -X main.BuildTime=$(BUILD_TIME)
LDFLAGS += -X main.GoVersion=$(GO_VERSION)

# 默认目标
.PHONY: all
all: clean build

# 清理构建目录
.PHONY: clean
clean:
	@echo "清理构建目录..."
	@rm -rf $(BUILD_DIR)

# 创建构建目录
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# 构建Linux AMD64版本 (Debian 12)
.PHONY: build
build: $(BUILD_DIR)
	@echo "构建 $(APP_NAME) for Linux AMD64..."
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) .
	@if [ -f ".env.example" ]; then cp .env.example $(BUILD_DIR)/; fi
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 构建Windows版本
.PHONY: build-windows
build-windows: $(BUILD_DIR)
	@echo "构建 $(APP_NAME) for Windows AMD64..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME).exe .
	@if [ -f ".env.example" ]; then cp .env.example $(BUILD_DIR)/; fi
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME).exe"

# 构建macOS版本
.PHONY: build-darwin
build-darwin: $(BUILD_DIR)
	@echo "构建 $(APP_NAME) for macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin .
	@if [ -f ".env.example" ]; then cp .env.example $(BUILD_DIR)/; fi
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)-darwin"

# 构建所有平台版本
.PHONY: build-all
build-all: build build-windows build-darwin

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	@go test -v ./...

# 代码格式化
.PHONY: fmt
fmt:
	@echo "格式化代码..."
	@go fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "代码检查..."
	@go vet ./...

# 更新依赖
.PHONY: mod-tidy
mod-tidy:
	@echo "整理依赖..."
	@go mod tidy

# 安装到Debian系统
.PHONY: install-debian
install-debian: build
	@echo "安装到Debian系统..."
	@chmod +x deploy-debian.sh
	@sudo ./deploy-debian.sh

# 创建发布包
.PHONY: package
package: build
	@echo "创建发布包..."
	@tar -czf $(BUILD_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(APP_NAME) .env.example
	@echo "发布包已创建: $(BUILD_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz"

# 显示帮助信息
.PHONY: help
help:
	@echo "可用的构建目标:"
	@echo "  build          - 构建Linux AMD64版本 (Debian 12)"
	@echo "  build-windows  - 构建Windows版本"
	@echo "  build-darwin   - 构建macOS版本"
	@echo "  build-all      - 构建所有平台版本"
	@echo "  test           - 运行测试"
	@echo "  fmt            - 格式化代码"
	@echo "  vet            - 代码检查"
	@echo "  mod-tidy       - 整理依赖"
	@echo "  package        - 创建发布包"
	@echo "  install-debian - 安装到Debian系统"
	@echo "  clean          - 清理构建目录"
	@echo "  help           - 显示此帮助信息"