.PHONY: build clean package swag test coverage version help

APP_NAME := vpublish
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "v2.0.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y%m%d-%H%M")
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建目录
BUILD_DIR := build
DIST_DIR := dist
PACKAGE_NAME := $(APP_NAME)-$(VERSION)-linux-amd64

# Go 编译参数 - 注入版本信息
LDFLAGS := -ldflags "-s -w \
	-X 'github.com/taerc/vpublish/internal/version.Version=$(VERSION)' \
	-X 'github.com/taerc/vpublish/internal/version.GitCommit=$(GIT_COMMIT)' \
	-X 'github.com/taerc/vpublish/internal/version.BuildTime=$(BUILD_TIME)'"

all: clean build package

# 构建后端
build-backend:
	@echo "Building backend..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/vpublish-server ./cmd/server
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/vpublish-mcp ./cmd/mcp

# 构建前端
build-frontend:
	@echo "Building frontend..."
	cd web && npm install && npm run build

# 构建所有
build: build-backend build-frontend

# 打包发布
package: build
	@echo "Packaging..."
	@mkdir -p $(DIST_DIR)
	@rm -rf $(DIST_DIR)/$(PACKAGE_NAME)
	@mkdir -p $(DIST_DIR)/$(PACKAGE_NAME)
	
	# 复制二进制文件
	cp $(BUILD_DIR)/vpublish-server $(DIST_DIR)/$(PACKAGE_NAME)/
	
	# 复制前端静态文件
	cp -r web/dist $(DIST_DIR)/$(PACKAGE_NAME)/web
	
	# 复制配置文件模板
	mkdir -p $(DIST_DIR)/$(PACKAGE_NAME)/configs
	cp configs/config.yaml.example $(DIST_DIR)/$(PACKAGE_NAME)/configs/ 2>/dev/null || \
		cp configs/config.yaml $(DIST_DIR)/$(PACKAGE_NAME)/configs/config.yaml.example
	
	# 复制部署文件 (systemd service + 安装脚本 + nginx配置)
	cp -r deploy $(DIST_DIR)/$(PACKAGE_NAME)/
	chmod +x $(DIST_DIR)/$(PACKAGE_NAME)/deploy/install.sh
	
	# 复制说明文档
	cp README.md $(DIST_DIR)/$(PACKAGE_NAME)/ 2>/dev/null || true
	cp DEPLOYMENT.md $(DIST_DIR)/$(PACKAGE_NAME)/ 2>/dev/null || true
	
	# 创建压缩包
	cd $(DIST_DIR) && tar -czf $(PACKAGE_NAME).tar.gz $(PACKAGE_NAME)
	
	@echo ""
	@echo "=========================================="
	@echo "Package created: $(DIST_DIR)/$(PACKAGE_NAME).tar.gz"
	@echo "=========================================="
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo ""
	@echo "Deploy steps:"
	@echo "  1. Copy package to target server"
	@echo "  2. tar -xzf $(PACKAGE_NAME).tar.gz"
	@echo "  3. cd $(PACKAGE_NAME) && ./deploy/install.sh"

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)

# 仅构建后端（快速打包）
build-server:
	@echo "Building server only..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/vpublish-server ./cmd/server

# 生成 Swagger 文档
swag:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@v1.16.4
	swag init -g ./cmd/server/main.go -o ./docs

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 运行测试并生成覆盖率报告
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 打印版本信息
version:
	@echo "Version:    $(VERSION)"
	@echo "GitCommit:  $(GIT_COMMIT)"
	@echo "BuildTime:  $(BUILD_TIME)"
	@echo "GoVersion:  $(GO_VERSION)"

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
lint:
	@echo "Linting code..."
	go vet ./...

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go get github.com/swaggo/swag/cmd/swag@v1.16.4
	go get github.com/swaggo/gin-swagger@v1.6.0
	go get github.com/swaggo/files@v1.0.1

# 开发模式运行
run:
	go run ./cmd/server/main.go

# 帮助
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all           清理并构建完整发布包 (默认)"
	@echo "  build         构建前后端"
	@echo "  build-server  仅构建后端服务"
	@echo "  build-backend 构建后端二进制"
	@echo "  build-frontend 构建前端"
	@echo "  swag          生成 Swagger 文档"
	@echo "  package       打包发布"
	@echo "  clean         清理构建产物"
	@echo "  test          运行测试"
	@echo "  coverage      运行测试并生成覆盖率报告"
	@echo "  version       打印版本信息"
	@echo "  fmt           代码格式化"
	@echo "  lint          代码检查"
	@echo "  deps          安装依赖"
	@echo "  run           开发模式运行"