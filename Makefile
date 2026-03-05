.PHONY: build clean package

APP_NAME := vpublish
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GO_VERSION := $(shell go version | awk '{print $$3}')

# 构建目录
BUILD_DIR := build
DIST_DIR := dist
PACKAGE_NAME := $(APP_NAME)-$(VERSION)-linux-amd64

# Go 编译参数
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

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
	cp $(BUILD_DIR)/vpublish-mcp $(DIST_DIR)/$(PACKAGE_NAME)/
	
	# 复制前端静态文件
	cp -r web/dist $(DIST_DIR)/$(PACKAGE_NAME)/web
	
	# 复制配置和迁移文件
	cp -r configs $(DIST_DIR)/$(PACKAGE_NAME)/
	cp -r migrations $(DIST_DIR)/$(PACKAGE_NAME)/
	
	# 复制说明文档
	cp README.md $(DIST_DIR)/$(PACKAGE_NAME)/ 2>/dev/null || true
	cp DEPLOYMENT.md $(DIST_DIR)/$(PACKAGE_NAME)/ 2>/dev/null || true
	cp MCP_README.md $(DIST_DIR)/$(PACKAGE_NAME)/ 2>/dev/null || true
	
	# 创建压缩包
	cd $(DIST_DIR) && tar -czf $(PACKAGE_NAME).tar.gz $(PACKAGE_NAME)
	
	@echo ""
	@echo "Package created: $(DIST_DIR)/$(PACKAGE_NAME).tar.gz"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"

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