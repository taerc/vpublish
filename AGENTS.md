# AGENTS.md - vpublish 项目指南

## 项目概述

vpublish 是一个低空智能平台软件包管理系统，提供软件版本管理、下载统计、APP端API等功能。

## 构建与运行命令

### 后端 (Go)

```bash
# 安装依赖
go mod tidy

# 开发模式运行
go run cmd/server/main.go

# 构建生产二进制
go build -o vpublish-server ./cmd/server

# 构建 MCP 服务
go build -o vpublish-mcp ./cmd/mcp

# 类型检查 / 静态分析
go vet ./...

# 代码格式化
go fmt ./...
```

### 前端 (Vue 3)

```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 类型检查
vue-tsc --noEmit

# 代码检查
npm run lint
```

### 数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE vpublish CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 自动迁移（程序启动时自动执行）
go run cmd/server/main.go
```

## API 文档规范

### Swagger API文档规范

为了保证API接口的一致性和可维护性，VPublish项目采用Swagger进行API文档管理，规范如下:

#### 1. Swagger 依赖安装

```bash
# 安装 Swag CLI 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 项目包依赖
go get -u github.com/swaggo/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

#### 2. 项目根目录添加文档注释 (main.go)

在 `cmd/server/main.go` 的 `main` 包之前添加全局文档配置:

```go
// vpublish - Go语言Gin框架实现的低空智能平台软件包管理系统
//
// 本API文档描述了vpublish系统的各个接口信息，包含管理员API和APP端API两大类
//
//     Schemes: http, https
//     Host: localhost:8080
//     BasePath: /api/v1
//     Version: 2.0.0
//     Contact: {name: vpublish团队, email: support@example.com}
//     License: MIT {url: http://opensource.org/licenses/MIT}
//
//     SecurityDefinitions:
//     - BearerAuth: apiKey header 仅用于管理员端 API
//      说明: Bearer <JWT-token>
//     - SignatureAuth: apiKey header 用于APP端 API
//      说明: 在X-App-Key, X-Timestamp, X-Signature中传递认证信息
//    Security:
//     - BearerAuth: []
//     - SignatureAuth: []
//
// swagger:info
package main
```

#### 3. Handler 方法的 Swagger 注解规范

每一个Handler方法都应该按照以下格式添加Swagger注释:

```go
// Login 登录
//
// @Summary 用户登录
// @Description 用户通过用户名和密码完成身份认证，成功后返回JWT token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body handler.LoginRequest true "登录请求参数"
// @Success 200 {object} handler.LoginResponse "登录成功，返回token和用户信息"
// @Failure 400 {object} pkg/response.Response "请求参数错误"
// @Failure 401 {object} pkg/response.Response "认证失败"
// @Failure 500 {object} pkg/response.Response "服务器内部错误"
// @Router /admin/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    // 实现代码保持不变
}
```

#### 4. Request/Response DTO 结构体注释规范

DTO结构体需要添加Swag标签，并配合结构体注释来完整描述数据结构:

```go
// LoginRequest 登录请求数据结构
type LoginRequest struct {
    // 用户名
    Username string `json:"username" binding:"required" example:"admin"`
    // 密码
    Password string `json:"password" binding:"required" example:"123456" swaggertype:"string" format:"password"`
}

// LoginResponse 登录响应数据结构
type LoginResponse struct {
    // JWT Token
    Token        string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    // 刷新Token
    RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    // Token过期时间(秒)
    ExpiresIn    int64  `json:"expires_in" example:"86400"`
    // 用户信息
    User         struct {
        // 用户ID
        ID       uint   `json:"id" example:"1"`
        // 用户名
        Username string `json:"username" example:"admin"`
        // 昵称
        Nickname string `json:"nickname" example:"管理员"`
        // 角色
        Role     string `json:"role" example:"admin"`
    } `json:"user"`
}
```

对于数组类型的响应包装在Response中时：

```go
// @Success 200 {object} pkg/response.Response "成功响应包装，Data字段包含具体的列表" 
// @Success 200 {object} pkg/response.Response{data=[]model.UserInfo} "详细响应格式，列表形式"
```

#### 5. 不同API类型的注释区别

##### 管理员端API (Admin API)

- 需要在`Tags`中标记为 `"Tags: 管理员/用户管理"` 等
- 安全方案设置为JWT认证
- 使用`@Security BearerAuth []`进行安全定义

```go
// GetUser 获取用户信息
//
// @Summary 获取指定用户详情
// @Description 通过ID获取单个用户的详细信息
// @Tags 管理员/用户管理  
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} pkg/response.Response{data=model.User} "用户信息"
// @Failure 401 {object} pkg/response.Response "未授权"
// @Failure 404 {object} pkg/response.Response "用户不存在" 
// @Failure 500 {object} pkg/response.Response "服务器错误"
// @Security BearerAuth []
// @Router /admin/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
    // 实现代码
}
```

##### APP端API

- 在`Tags`中标记为 `"Tags: APP端"` 
- 使用`@Security SignatureAuth []`设置为签名认证
- 需要注意的是APP端使用签名而非JWT

```go
// getCategoryActive 获取激活类别列表(供APP使用)
//
// @Summary 获取所有启用状态的类别列表
// @Description 获取所有可用的软件类别，只返回启用状态的类别
// @Tags APP端
// @Accept json
// @Produce json
// @Success 200 {object} pkg/response.Response{data=[]model.Category} "类别列表"
// @Failure 401 {object} pkg/response.Response "认证失败"  
// @Security SignatureAuth []
// @Router /app/categories [get]
func (h *CategoryHandler) ListActive(c *gin.Context) {
    // 实现代码
}
```

#### 6. 参数定义注释规范

根据不同参数类型使用不同的注释格式:

```go
// Path参数示例
// @Param id path int true "用户ID" minimum(1)

// Query参数示例  
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(1000) default(10)

// Body参数示例
// @Param request body CreateCategoryRequest true "创建类别的请求参数"

// 表单上传示例（针对文件上传接口）
// @Param file formData file true "上传的文件"
```

#### 7. 分页API注释规范

对于分页列表API，建议返回PageData格式并特别标注:

```go
// ListUsers 用户列表
//
// @Summary 用户分页列表
// @Description 获取用户分页列表，支持搜索和过滤
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(1000) default(10)
// @Param keyword query string false "搜索关键词(用户名/昵称)"
// @Success 200 {object} pkg/response.Response{data=pkg/response.PageData{list=[]model.User}} "分页用户列表"
// @Failure 401 {object} pkg/response.Response "未授权" 
// @Failure 500 {object} pkg/response.Response "服务器错误"
// @Security BearerAuth []
// @Router /admin/users [get]
func (h *UserHandler) List(c *gin.Context) {
    // 实现代码
}
```

#### 8. 文件上传接口注释规范

对于文件上传等 multipart/form-data 类型接口：

```go
// UploadVersion 上传新版本
// 
// @Summary 上传软件包新版本
// @Description 上传软件包版本文件，同时可以设置版本信息
// @Tags 管理员/版本管理
// @Accept mpfd # 表示 multipart/form-data
// @Produce json
// @Param package_id path int true "软件包ID" minimum(1)
// @Param file formData file true "要上传的文件"
// @Param version formData string true "版本号" example:"1.0.0"
// @Param description formData string false "版本描述"
// @Success 200 {object} pkg/response.Response{data=model.Version} "版本上传成功"
// @Failure 400 {object} pkg/response.Response "参数错误"
// @Failure 401 {object} pkg/response.Response "未授权"
// @Failure 413 {object} pkg/response.Response "文件太大"
// @Security BearerAuth []
// @Router /admin/packages/{package_id}/versions [post]
func (h *PackageHandler) UploadVersion(c *gin.Context) {
    // 实现代码
}
```

#### 9. 模型(Model)注释规范

数据模型需要包含详细字段说明和示例：

```go
// Category 软件类别模型
type Category struct {
    // 主键ID
    ID          uint           `gorm:"primaryKey" json:"id" example:"1"`
    // 类别名称
    Name        string         `gorm:"size:100;uniqueIndex;not null" json:"name" example:"无人机应用"`
    // 唯一代码，根据名称自动生成（中文转拼音，英文数字保留）
    // 示例：无人机应用 -> TYPE_WU_REN_JI_YING_YONG，无人机V2 -> TYPE_WU_REN_JI_V2
    Code        string         `gorm:"size:50;uniqueIndex;not null" json:"code" example:"TYPE_WU_REN_JI_YING_YONG"`
    // 描述
    Description string         `gorm:"size:500" json:"description" example:"各种无人机相关的应用程序分类"`
    // 排序值
    SortOrder   int            `gorm:"default:0" json:"sort_order" example:"10"`
    // 是否启用 (1:启用, 0:禁用)
    IsActive    int8           `gorm:"default:1" json:"is_active" example:"1"`
    // 创建时间
    CreatedAt   time.Time      `json:"created_at" example:"2026-03-12T10:00:00Z"`
    // 更新时间
    UpdatedAt   time.Time      `json:"updated_at" example:"2026-03-12T15:30:00Z"`
    // 删除时间
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

#### 10. 验证注解和数据示例

在验证规则和数据样例方面，建议在注释中详细说明：

```go
// CreateCategoryRequest 创建分类请求数据
type CreateCategoryRequest struct {
    // 名称 (必填，长度3-50)
    Name        string `json:"name" binding:"required,min=3,max=50" example:"低空监测系统" validate:"required,min=3,max=50"`
    // 描述信息 (最大500字符)
    Description string `json:"description,omitempty" binding:"max=500" example:"专门用于低空环境监测的软件系统"` 
    // 排序数值 (整数，默认值为100)
    SortOrder   int    `json:"sort_order,omitempty" binding:"omitempty,min=0,max=9999" example:"100"`
}
```

#### 11. 错误响应定义注释

常见错误码应该在文档中统一说明:

```go
// HTTP状态码说明:
// 200 - 成功，包含成功数据
// 400 - 请求参数错误 
// 401 - 认证失败，需要验证权限
// 403 - 权限不足，无法执行操作
// 404 - 资源不存在
// 405 - HTTP方法不允许
// 429 - 请求过于频繁，被限制
// 500 - 服务器内部错误
```

### 2. API 规范标准

#### 2.1 基础响应格式

为了确保前后端接口的一致性，所有API统一返回`Response`格式：

```go
type Response struct {
    // 响应状态码，非负整数
    Code    int         `json:"code"`
    // 响应消息文本
    Message string      `json:"message"`
    // 业务数据，可选
    Data    interface{} `json:"data,omitempty"` 
}
```

**状态码定义：**

| Code | Message | 含义 | 说明 |
|------|---------|------|------|
| 0 | "success" | 成功 | 请求正常响应 |
| 400 | "bad request" | 请求不合法 | 参数验证错误 |
| 401 | "unauthorized" | 未认证 | 认证失败 |
| 403 | "forbidden" | 禁止访问 | 权限不足 |
| 404 | "not found" | 未找到资源 | 资源不存在 |
| 429 | "too many requests" | 请求频繁 | API 限流 |
| 500 | "internal server error" | 服务错误 | 服务器内部错误 |

#### 2.2 分页响应标准

对于列表数据提供 `PageData` 统一分页格式：

```go
type PageData struct {
    // 数据列表
    List     interface{} `json:"list"`
    // 总数据量  
    Total    int64       `json:"total"`
    // 当前页码
    Page     int         `json:"page"`
    // 每页条数
    PageSize int         `json:"page_size"`
}
```

**示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      // 数据数组
    ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

#### 2.3 请求参数标准

- **Query参数**：用于GET/DELETE请求的筛选、分页等参数
- **Body参数**：用于POST/PUT请求的数据提交
- **路径参数**：用于资源定位的ID等参数
- **表单参数**：用于文件上传和少量键值对提交

#### 2.4 路径规范

- **管理员端接口**：`/api/v1/admin/*`
- **APP端接口**：`/api/v1/app/*`  
- **健康检查**：`/health`
- **API文档**：`/swagger/*` （如果启用）

#### 2.5 认证Header标准

**管理员API认证（JWT）：**
- 请求头: `Authorization: Bearer <token>`
  
**APP端API认证（HMAC-SHA256签名）：**
- `X-App-Key: <appkey>` - 应用标识
- `X-Timestamp: <timestamp>` - Unix时间戳  
- `X-Signature: <signature>` - 签名字符串

#### 2.6 错误处理标准

错误处理遵循统一的Response结构，包含准确的状态码和用户友好的错误信息：

- 业务错误：使用有意义的错误消息，避免泄露系统内部信息
- 验证错误：明确标识哪个参数出现问题，以及正确格式
- 系统错误：统一返回500错误，不暴露内部错误详情

#### 2.7 数据类型和格式规范

- **时间格式**：UTC时间，RFC3339格式 `"YYYY-MM-DDTHH:mm:ssZ"`
- **ID类型**：使用 `uint` 或 `string` 作为唯一标识符
- **枚举值**：优先使用数字枚举，减少存储空间，但文档中提供文字说明
- **文件大小**：使用字节(byte)为单位，使用int64防止溢出

#### 2.8 API版本控制

- **URL方式**：`/api/v1/*` 
- **向后兼容**：新版本API必须兼容老版本核心功能
- **弃用策略**：弃用的API版本至少保留3个月再移除

### 3. 文档生成流程

#### 3.1 依赖安装

首先需要安装Swag CLI工具及相应的Go包依赖：

```bash
# 安装Swag命令行工具
go install github.com/swaggo/swag/cmd/swag@latest

# 安装项目依赖
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files

# 打开模块模式（确保能够下载包）
go mod tidy
```

#### 3.2 预设注释规范

在项目的入口文件(main.go)顶部添加API文档初始化信息：

```go
// vpublish - Go语言Gin框架实现的低空智能平台软件包管理系统  
//
// 本API文档描述了vpublish系统的各个接口信息，包含管理员API和APP端API两大类
//
//     Schemes: http, https
//     Host: localhost:8080
//     BasePath: /api/v1
//     Version: 2.0.0
//     Contact: {name: vpublish团队, email: support@example.com}
//     License: MIT {url: http://opensource.org/licenses/MIT}
//
//     SecurityDefinitions:
//     - BearerAuth: apiKey header 仅用于管理员端 API  
//      说明: Bearer <JWT-token>
//     - SignatureAuth: apiKey header 用于APP端 API  
//      说明: 在X-App-Key, X-Timestamp, X-Signature中传递认证信息
//    Security:
//     - BearerAuth: []
//     - SignatureAuth: []
//
// swagger:info
package main
```

#### 3.3 生成文档命令

执行以下命令生成Swagger文档：

```bash
# 从项目根目录生成 (确保当前在GOPATH或启用了Go modules)
swag init -g ./cmd/server/main.go -o ./docs

# 也可以指定API描述输出目录  
swag init --generalInfo ./cmd/server/main.go --output ./docs
```

这将在`docs`目录下生成 `docs.go`, `swagger.json`, `swagger.yaml` 等文件。

#### 3.4 集成到Gin路由

在主要的路由配置文件中引入并集成Swagger文档：

```go
import (
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"        
    ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRoutes(r *gin.Engine, /* 参数 */) {
    // 省略已有路由配置... 

    // 添加Swagger文档路由（仅在开发环境中开启）
    if gin.Mode() == gin.DebugMode {
        r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }
}
```

#### 3.5 配置开发环境支持

在开发过程中推荐设置自动重新生成文档：

```bash
# 安装air热重载工具
go install github.com/cosmtrek/air@latest

# 确保air配置中包含docs目录的生成指令
```

可以创建`.air.toml`配置文件，以便自动执行 `swag init`：

```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "swag init && go build -o ./tmp/main cmd/server/main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "node_modules"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  kill_delay = 500
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true
```

#### 3.6 API文档访问路径

生成文档后，可以通过以下URL访问API文档:

- 文档页面：`http://localhost:8080/swagger/index.html`
- JSON格式：`http://localhost:8080/swagger/doc.json`

#### 3.7 常见问题和注意事项

1. **错误："no package was found in..."**
   - 确保在项目根目录运行命令，且项目结构正确
   
2. **错误："Could not find swagger version in any of vendor/github.com/swaggo/"
   - 需要手动下载依赖包: `go get -u github.com/swaggo/swag`
   
3. **struct tag语法错误**
   - 严格按照 `json:"field_name" binding:"rules"` 的格式书写
   - 注意在example中的实际值格式要与字段类型匹配

4. **文档不更新问题**
   - 每次修改完API注释后需重新执行 `swag init` 命令
   - 如果使用热重载工具需配置好文档重新生成事件

### 4. 发布流程考虑

#### 4.1 服务版本号规范

服务版本号用于标识当前运行的服务版本，便于问题追踪、版本对比和运维管理。

**版本号构成**

```
{语义化版本}-{Git短提交哈希}-{发布时间}
```

| 组成部分 | 格式 | 示例 | 说明 |
|---------|------|------|------|
| 语义化版本 | `v{major}.{minor}.{patch}` | `v2.1.0` | 遵循 SemVer 规范 |
| Git短提交哈希 | 7位字符 | `a1b2c3d` | 发布时的 commit 标识 |
| 发布时间 | `YYYYMMDD-HHMM` | `20260312-1530` | UTC 时间 |

**完整版本号示例**

```
v2.1.0-a1b2c3d-20260312-1530
```

**代码实现**

在 `internal/version/version.go` 中定义版本信息：

```go
package version

import (
    "fmt"
    "runtime"
)

var (
    // 以下变量通过 -ldflags 在编译时注入
    Version   = "dev"           // 语义化版本，如 v2.1.0
    GitCommit = "unknown"       // Git 提交哈希（短）
    BuildTime = "unknown"       // 构建时间
)

// Info 版本信息结构
type Info struct {
    Version   string `json:"version"`    // 完整版本号
    GitCommit string `json:"git_commit"` // Git 提交哈希
    BuildTime string `json:"build_time"` // 构建时间
    GoVersion string `json:"go_version"` // Go 版本
    Platform  string `json:"platform"`   // 运行平台
}

// Get 获取版本信息
func Get() Info {
    return Info{
        Version:   buildVersion(),
        GitCommit: GitCommit,
        BuildTime: BuildTime,
        GoVersion: runtime.Version(),
        Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
    }
}

// buildVersion 构建完整版本号
func buildVersion() string {
    if GitCommit == "unknown" {
        return Version
    }
    return fmt.Sprintf("%s-%s-%s", Version, GitCommit, BuildTime)
}

// String 返回版本号字符串
func String() string {
    return buildVersion()
}
```

**健康检查接口返回版本信息**

在 `cmd/server/main.go` 中返回版本信息：

```go
package main

import (
    "github.com/taerc/vpublish/internal/version"
)

func setupRoutes(r *gin.Engine, /* ... */) {
    // 健康检查 - 包含版本信息
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status":  "ok",
            "version": version.String(),
            "info":    version.Get(),
        })
    })
    
    // 版本信息接口
    r.GET("/version", func(c *gin.Context) {
        c.JSON(200, version.Get())
    })
}
```

**构建时注入版本信息**

在构建命令中使用 `-ldflags` 注入版本信息：

```bash
#!/bin/bash
# scripts/build.sh

VERSION="v2.1.0"
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_TIME=$(date -u +"%Y%m%d-%H%M")

go build -ldflags "\
    -X 'github.com/taerc/vpublish/internal/version.Version=${VERSION}' \
    -X 'github.com/taerc/vpublish/internal/version.GitCommit=${GIT_COMMIT}' \
    -X 'github.com/taerc/vpublish/internal/version.BuildTime=${BUILD_TIME}'" \
    -o vpublish-server ./cmd/server

echo "Build complete: ${VERSION}-${GIT_COMMIT}-${BUILD_TIME}"
```

**Makefile 示例**

```makefile
# Makefile

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y%m%d-%H%M")
LDFLAGS := -ldflags "\
    -X 'github.com/taerc/vpublish/internal/version.Version=$(VERSION)' \
    -X 'github.com/taerc/vpublish/internal/version.GitCommit=$(GIT_COMMIT)' \
    -X 'github.com/taerc/vpublish/internal/version.BuildTime=$(BUILD_TIME)'"

.PHONY: build
build:
	go build $(LDFLAGS) -o vpublish-server ./cmd/server

.PHONY: build-all
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/vpublish-server-linux-amd64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/vpublish-server-darwin-amd64 ./cmd/server
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/vpublish-server-windows-amd64.exe ./cmd/server

.PHONY: version
version:
	@echo "Version:    $(VERSION)"
	@echo "GitCommit:  $(GIT_COMMIT)"
	@echo "BuildTime:  $(BUILD_TIME)"
```

**健康检查响应示例**

```json
// GET /health
{
    "status": "ok",
    "version": "v2.1.0-a1b2c3d-20260312-1530",
    "info": {
        "version": "v2.1.0-a1b2c3d-20260312-1530",
        "git_commit": "a1b2c3d",
        "build_time": "20260312-1530",
        "go_version": "go1.24.0",
        "platform": "linux/amd64"
    }
}
```

**版本号与 Swagger 文档关联**

在 Swagger 文档注释中引用服务版本号：

```go
// vpublish - Go语言Gin框架实现的低空智能平台软件包管理系统
//
//     Version: v2.1.0  // 与服务版本号保持一致
//
// swagger:info
package main
```

**发布流程中的版本号更新**

1. 创建发布分支时确定语义化版本号
2. 更新 `Makefile` 或 `scripts/build.sh` 中的默认版本号
3. 更新 Swagger 文档注释中的版本号
4. 执行 `swag init` 重新生成文档
5. 构建时自动注入 Git 提交哈希和构建时间

#### 4.2 版本管理与文档关联

在每次版本发布前，应当同步更新API文档以确保其准确反映当前版本的特性。

**版本迭代流程**

- 在release分支创建后，更新Swagger注释中的版本号信息；
- 执行 `swag init` 生成最新的文档；
- 提交包含最新文档的代码至release分支；
- 部署完成后验证线上文档是否更新成功。

例如在 `main.go` 上方文档注释中保持与当前版本一致：

```go
//     Version: 2.1.0  // 应跟发布版本号一致
```

#### 4.3 CHANGELOG更新

当API有变更（新增endpoint、修改参数或响应结构）时，应在CHANGELOG中相应记录:

**变动分类**

- Added: 新增API endpoint或功能
- Changed: 修改现有接口的数据结构或行为  
- Removed: 移除已过时接口
- Fixed: 补强或修复已知问题

**样例**

```markdown
# v2.1.0 (2026-03-12)

## Added
- 新增 `/api/v1/admin/stats/category` 用于获取类别统计信息
- 添加类别详情接口支持

## Changed  
- 更新 `/api/v1/admin/versions` POST请求的参数验证规则
- 将 `/api/v1/admin/downloads` 移至 `/api/v1/admin/stats/counts`

## Fixed
- 修正版本上传接口文件扩展名校验
- 优化用户列表分页性能
```

#### 4.4 文档审查机制

API文档被视为与代码同等重要的资产，因此应纳入代码审查(CR)流程：

- 每次PR都需确认涉及的API变更对应的Swagger注释是否已更新；
- 新增的参数类型须添加适当的example值；
- 确认response schema定义的准确性，特别是嵌套对象的描述；

#### 4.5 生产环境文档部署

**开发/测试环境文档**

在开发与测试环境，应始终保持Swagger UI的可用性，便于联调和测试。

**生产环境文档**

根据安全要求决定是否在生产环境提供API文档：

- 内部系统或已做访问控制的API: 可允许通过内网访问
- 公开接口的文档: 须通过公司官网提供，不可直接部署至服务节点
- 简化版文档: 若必须随服务部署，可提供不展示完整数据schema的精简版

#### 4.6 回退机制

若新发布的API版本出现重大兼容性问题，除了需要回滚代码外，还应注意API文档同步回撤，避免误导外部使用者：

- 切换至旧版服务的同时，应切换对应文档的版本；
- 通知相关调用者注意版本变化。

### 5. 完整代码示例

#### API文档入口(main.go)示例

```go
// vpublish - Go语言Gin框架实现的低空智能平台软件包管理系统
//
// 本API文档描述了vpublish系统的各个接口信息，包含管理员API和APP端API两大类
//
//     Schemes: http, https
//     Host: localhost:8080
//     BasePath: /api/v1
//     Version: 2.1.0
//     Contact: {name: vpublish团队, email: support@example.com, url: https://example.com/support}
//     License: MIT {url: http://opensource.org/licenses/MIT}
//
//     SecurityDefinitions:
//     - BearerAuth: 
//        type: apiKey
//        in: header
//        name: Authorization
//        description: 仅用于管理员端 API, 格式: Bearer <JWT-token>
//     - SignatureAuth:
//        type: apiKey 
//        in: header
//        name: X-App-Key
//        description: 用于APP端 API 的X-App-Key值
//     - TimestampAuth:
//        type: apiKey
//        in: header
//        name: X-Timestamp
//        description: 用于APP端 API 的Unix时间戳
//     - SignatureValueAuth: 
//        type: apiKey
//        in: header
//        name: X-Signature
//        description: 用于APP端 API 的HMAC-SHA256签名
//    Security:
//     - BearerAuth: []
//     - SignatureAuth: []
//     - TimestampAuth: []
//     - SignatureValueAuth: []
//
// swagger:info
package main

import (
    "github.com/gin-gonic/gin"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
    // ... 应用初始化代码 ...
    
    // 创建路由
    gin.SetMode(cfg.Server.Mode)
    r := gin.New()

    // 设置CORS等中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS(&cfg.CORS))

    setupRoutes(r, /*注入的各种handler*/)

    // 只在非生产环境启用Swagger UI
    if gin.Mode() != gin.ReleaseMode {
        r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }
    
    // 启动服务器...
}

func setupRoutes( r *gin.Engine, /* 参数 */) {
    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok", "version": "2.1.0"})
    })

    // API v1
    v1 := r.Group("/api/v1")
    {
        // ============ 管理端 API ============
        admin := v1.Group("/admin")
        admin.Use(middleware.CORSMiddleware())
        
        // 认证相关
        admin.POST("/auth/login", authHandler.Login)
        
        // 需要登录的路由
        protected := admin.Group("")
        protected.Use(middleware.JWTAuth(jwtService))
        {
            // 在这里添加受保护的API
        }
    }
}
```

#### Handler方法注释示例

```go
package handler

import (
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/taerc/vpublish/internal/middleware"
    "github.com/taerc/vpublish/internal/service"
    "github.com/taerc/vpublish/pkg/response"
    "github.com/taerc/vpublish/internal/model"
)

// CreateCategoryRequest 创建分类请求数据
type CreateCategoryRequest struct {
    // 名称 (必填，长度3-50个字符)
    Name        string `json:"name" binding:"required,min=3,max=50" example:"低空监测系统"`
    // 描述信息 (最大500字符)
    Description string `json:"description,omitempty" binding:"max=500" example:"专门用于低空环境监测的软件系统" validate:"max=500"`
    // 排序数值 (整数，默认值为100)，数值越大越靠后
    SortOrder   int    `json:"sort_order,omitempty" binding:"omitempty,min=0,max=9999" example:"100"`
    // 是否启用 (1:启用, 0:禁用) 默认为启用
    IsActive    int8   `json:"is_active,omitempty" binding:"min=0,max=1" example:"1"`
}

// CreateCategoryResponse 创建分类响应数据
type CreateCategoryResponse struct {
    // 新建分类ID
    ID          uint   `json:"id" example:"101"`
    // 分类名称
    Name        string `json:"name" example:"低空监测系统"`
    // 分类代码 (基于名称自动生成的拼音码)
    Code        string `json:"code" example:"TYPE_DAO_KONG_JIAN_CE_XI_TONG"`
    // 描述信息
    Description string `json:"description" example:"专门用于低空环境监测的软件系统"`
    // 排序值
    SortOrder   int    `json:"sort_order" example:"100"`
    // 是否启用
    IsActive    int8   `json:"is_active" example:"1"`
    // 创建时间
    CreatedAt   string `json:"created_at" example:"2026-03-12T15:30:00Z"`
    // 更新时间
    UpdatedAt   string `json:"updated_at" example:"2026-03-12T15:30:00Z"`
}

// CreateCategory 创建软件类别
// @Summary 创建软件类别
// @Description 创建一个新的软件类别，如无人机类型、地面站系统等。
// @Description 系统会自动根据名称生成代码（中文转拼音，英文数字保留），便于后台识别。
// @Description 示例：无人机V2 -> TYPE_WU_REN_JI_V2，地面站Pro -> TYPE_DI_MIAN_ZHAN_PRO
// @Security BearerAuth [Bearer]
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param request body handler.CreateCategoryRequest true "创建分类的请求参数"
// @Success 200 {object} pkg/response.Response{data=handler.CreateCategoryResponse} "创建成功，返回新创建的类别信息"
// @Failure 400 {object} pkg/response.Response "请求参数错误，例如参数未满足约束条件"
// @Failure 401 {object} pkg/response.Response "未认证，需要有效的管理员Token"
// @Failure 403 {object} pkg/response.Response "权限不足，当前用户角色不允许此操作"
// @Failure 409 {object} pkg/response.Response "冲突，相同名称的类别已存在"
// @Failure 500 {object} pkg/response.Response "服务器内部错误"
// @Router /admin/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
    var req CreateCategoryRequest
    
    // 参数绑定与验证
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())
        return
    }

    // 构建service层的参数结构
    categoryReq := &service.CreateCategoryRequest{
        Name:        req.Name,
        Description: req.Description,
        SortOrder:   req.SortOrder,
        IsActive:    req.IsActive,
    }

    // 调用业务层创建方法
    category, err := h.categoryService.Create(c.Request.Context(), categoryReq)
    if err != nil {
        response.Error(c, 400, err.Error())
        return
    }

    // 转换为响应结构
    resp := CreateCategoryResponse{
        ID:          category.ID,
        Name:        category.Name,
        Code:        category.Code,
        Description: category.Description,
        SortOrder:   category.SortOrder,
        IsActive:    category.IsActive,
        CreatedAt:   category.CreatedAt.Format(time.RFC3339),
        UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
    }

    // 返回成功响应
    response.Success(c, resp)
}

// ListCategoriesResponse 列出分类的响应结构
type ListCategoriesResponse struct {
    // 分类列表 
    List []struct {
        // 分类ID
        ID          uint   `json:"id" example:"1"`
        // 分类名称
        Name        string `json:"name" example:"无人机应用"`
        // 分类代码
        Code        string `json:"code" example:"TYPE_WU_REN_JI_YING_YONG"`
        // 描述
        Description string `json:"description" example:"各类飞行器相关应用程序"`
        // 排序值
        SortOrder   int    `json:"sort_order" example:"10"`
        IsActive    int8   `json:"is_active" example:"1"`
        CreatedAt   string `json:"created_at" example:"2026-03-10T08:00:00Z"`
        UpdatedAt   string `json:"updated_at" example:"2026-03-12T10:30:00Z"`
    } `json:"list"`
    Total    int64 `json:"total" example:"25"`
    Page     int   `json:"page" example:"1"`
    PageSize int   `json:"page_size" example:"10"`
}

// ListCategories 分类列表
// @Summary 获取分类分页列表
// @Description 获取软件类别列表，支持按关键词搜索、按启用状态过滤和分页
// @Security BearerAuth []
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param keyword query string false "关键词搜索(分类名称或代码)"
// @Param is_active query int false "是否启用 (1:启用, 0:禁用)" Enums(0, 1)
// @Param page query int false "页码" minimum(1) default(1) maximum(200)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(10)
// @Success 200 {object} pkg/response.Response{data=pkg/response.PageData{list=[]model.Category}} "获取成功，返回分类分页结果"
// @Failure 401 {object} pkg/response.Response "未认证访问"
// @Failure 400 {object} pkg/response.Response "参数非法"
// @Failure 500 {object} pkg/response.Response "服务器内部错误"
// @Router /admin/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
    // 获取查询参数
    keyword := c.Query("keyword")
    isActiveStr := c.Query("is_active")
    pageStr := c.DefaultQuery("page", "1")
    pageSizeStr := c.DefaultQuery("page_size", "10")
    
    // 参数验证与转换省略...
    
    // 调用服务层获取结果
    result, total, err := h.categoryService.List(c.Request.Context(), keyword, isActive, page, pageSize)
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }
    
    // 构建响应数据
    response.Page(c, result, total, page, pageSize)
}

// GetCategory 获得分类详情
// @Summary 根据ID获取分类详情
// @Description 获取特定ID的软件分类详细信息
// @Security BearerAuth []
// @Tags 管理员/类别管理  
// @Accept json
// @Produce json
// @Param id path int true "类别ID" minimum(1) example(1) 
// @Success 200 {object} pkg/response.Response{data=model.Category} "返回类别详细信息"
// @Failure 401 {object} pkg/response.Response "未认证访问"
// @Failure 404 {object} pkg/response.Response "类别不存在"
// @Failure 500 {object} pkg/response.Response "服务器内部错误"
// @Router /admin/categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil || id == 0 {
        response.BadRequest(c, "无效的类别ID")
        return
    }

    category, err := h.categoryService.GetByID(c.Request.Context(), uint(id))
    if err != nil {
        response.NotFound(c, "类别不存在")
        return
    }

    response.Success(c, category)
}

// APP端接口示例
// ListActiveCategories APP获取活跃分类
// @Summary 获取所有激活的软件分类列表
// @Description APP端获取所有启用状态的软件分类，用于展示可用软件种类
// @Tags APP端
// @Accept json
// @Produce json
// @Success 200 {object} pkg/response.Response{data=[]model.Category} "返回启用的类别列表"
// @Failure 401 {object} pkg/response.Response "认证失败，AppKey或签名验证不正确"
// @Failure 500 {object} pkg/response.Response "服务器内部错误"
// @Security SignatureAuth []
// @Security TimestampAuth []
// @Security SignatureValueAuth []
// @Router /app/categories [get]
func (h *CategoryHandler) ListActive(c *gin.Context) {
    categories, err := h.categoryService.ListActive(c.Request.Context())
    if err != nil {
        response.InternalError(c, err.Error())
        return
    }

    response.Success(c, categories)
}
```

#### Model层结构体(Swagger模型定义)示例

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

// Category 软件类别模型
// swagger:model Category
type Category struct {
    // 主键ID
    ID          uint           `gorm:"primaryKey" json:"id" example:"1"`
    // 类别名称，唯一值，长度3-100字符
    Name        string         `gorm:"size:100;uniqueIndex;not null" json:"name" example:"无人机应用" validate:"min=3,max=100"`
    // 唯一代码，根据名称自动生成（中文转拼音，英文数字保留）
    // 示例：无人机应用 -> TYPE_WU_REN_JI_YING_YONG，无人机V2 -> TYPE_WU_REN_JI_V2
    Code        string         `gorm:"size:50;uniqueIndex;not null" json:"code" example:"TYPE_WU_REN_JI_YING_YONG" validate:"len=20,max=50"`
    // 可选的描述信息，最大500字符
    Description string         `gorm:"size:500" json:"description" example:"包含各类无人机飞行、遥控、管理相关应用软件" validate:"max=500"`
    // 排序权重，数值越大排列越靠后
    SortOrder   int            `gorm:"default:0" json:"sort_order" example:"100" validate:"min=0,max=9999"`
    // 状态标志位 (1:启用, 0:禁用)，默认启用
    IsActive    int8           `gorm:"default:1" json:"is_active" example:"1" validate:"min=0,max=1"`
    // 创建时间
    CreatedAt   time.Time      `json:"created_at" example:"2026-03-12T10:00:00Z"`
    // 更新时间  
    UpdatedAt   time.Time      `json:"updated_at" example:"2026-03-12T15:30:00Z"`
    // 软删除标记 (gorm自带软删除字段)
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户基础模型  
// swagger:model User
type User struct {
    // 主键ID
    ID          uint           `gorm:"primaryKey" json:"id" example:"123"`
    // 用户名，唯一，长度3-30字符，仅限字母数字下划线
    Username    string         `gorm:"size:30;uniqueIndex;not null" json:"username" example:"admin" validate:"min=3,max=30,alphanum"`
    // 昵称，长度不超过50字符
    Nickname    string         `gorm:"size:50" json:"nickname" example:"系统管理员" validate:"max=50"`
    // 邮箱，唯一，符合邮箱格式
    Email       string         `gorm:"size:100;uniqueIndex" json:"email" example:"admin@example.com" validate:"max=100,email"`
    // 用户角色标识
    Role        string         `gorm:"size:50" json:"role" example:"admin" validate:"max=50"`
    // 密码哈希值，不对外公开
    PasswordHash string        `gorm:"size:255;not null" json:"-" validate:"min=60"` // bcrypt哈希值
    // 激活状态 (1:活跃, 0:禁用)
    IsActive    int8           `gorm:"default:1" json:"is_active" example:"1"`
    // 最后登录IP地址
    LastIP      string         `gorm:"size:50" json:"last_ip,omitempty" example:"192.168.1.100"`
    // 最后登录时间
    LastLoginAt *time.Time      `json:"last_login_at,omitempty" example:"2026-03-12T15:30:00Z"`
    // 创建时间
    CreatedAt   time.Time      `json:"created_at" example:"2026-01-01T08:00:00Z"`  
    // 更新时间
    UpdatedAt   time.Time      `json:"updated_at" example:"2026-03-12T15:30:00Z"`
    // 软删除标记
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

#### 文件上传处理示例

```go
// UploadVersionResponse 上传版本返回结果
type UploadVersionResponse struct {
    // 版本ID
    ID          uint      `json:"id" example:"12"`
    // 关联软件包ID
    PackageID uint      `json:"package_id" example:"5"`
    // 版本号
    Version     string    `json:"version" example:"2.1.0"`
    // 文件名
    FileName    string    `json:"file_name" example:"drone-software-2.1.0.zip"`
    // 文件大小，单位Bytes
    FileSize    int64     `json:"file_size" example:"15728640"`
    // 文件SHA256校验码
    FileSHA256  string    `json:"file_sha256" example:"a1b2c3d4e5f6...xyz"`
    // 版本描述
    Description string    `json:"description" example:"修复了关键安全漏洞，提升性能并增加新功能"`
    // 更新日志
    Changelog   string    `json:"changelog" example:""- 修复重要安全漏洞\n- 性能优化\n- 新增功能XYZ"`
    // 是否推荐升级
    IsRecommend int8      `json:"is_recommend" example:"1"`
    // 最低兼容版本
    MinVersion  string    `json:"min_version" example:"1.8.0"`
    // 是否为稳定版(0:测试版, 1:稳定版)
    IsStable    int8      `json:"is_stable" example:"1"`
    // 发布时间
    ReleaseDate string    `json:"release_date" example:"2026-03-12T16:00:00Z"`
    // 创建时间
    CreatedAt   string    `json:"created_at" example:"2026-03-12T15:45:00Z"`
    // 是否强制更新
    IsForceUpgrade int8   `json:"is_force_upgrade" example:"0"`
}

// UploadVersion 上传版本文件
// @Summary 上传软件包的新版本
// @Description 上传新的版本文件并创建版本记录，支持版本描述、更新日志等元数据
// @Tags 管理员/版本管理
// @Accept mpfd
// @Produce json  
// @Param id path int true "软件包ID" minimum(1) example(1) 
// @Param file formData file true "软件包文件，支持zip、exe、bin等格式"
// @Param version formData string true "版本号，遵循语义化版本规范(SemVer)" format:"SemVer" example(2.1.0)
// @Param description formData string false "版本描述，解释本次更新的内容"
// @Param changelog formData string false "更新日志，在多个换行中详述改进和修复"
// @Param min_version formData string false "最低兼容版本，低于此版本需重新安装" format:"SemVer" example(2.0.0)
// @Param is_force_upgrade formData bool false "是否强制升级，强制覆盖当前版本(1:是, 0:否)" 
// @Param is_recommend formData bool false "是否推荐更新，作为默认更新选项(1:是, 0:否)"
// @Param is_stable formData bool false "是否稳定版，标记稳定性(1:稳定版, 0:测试版)"
// @Success 200 {object} pkg/response.Response{data=handler.UploadVersionResponse} "上传成功"
// @Failure 400 {object} pkg/response.Response "参数错误：版本格式不规范或文件类型不支持"
// @Failure 401 {object} pkg/response.Response "未认证或令牌失效" 
// @Failure 403 {object} pkg/response.Response "权限不足，非管理员或无对应包管理权限"
// @Failure 413 {object} pkg/response.Response "文件过大，超出限制"
// @Failure 500 {object} pkg/response.Response "服务器错误：存储失败或文件校验异常"
// @Security BearerAuth []
// @Router /admin/packages/{id}/versions [post]
func (h *PackageHandler) UploadVersion(c *gin.Context) {
    packageId, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil || packageId == 0 {
        response.BadRequest(c, "无效软件包ID")
        return
    }
    
    // 解析 multipart 携带的文件
    file, err := c.FormFile("file")
    if err != nil {
        response.Error(c, 400, "上传文件失败: "+err.Error())
        return
    }
    
    // 验证文件是否已存在
    if exists, _ := h.packageService.CheckVersionExists(uint(packageId), file.Filename); exists {
        response.Error(c, 409, "同名文件已在同一软件包中存在")
        return
    }
    
    // 获取各种表单参数
    version := c.PostForm("version")
    description := c.PostForm("description")
    changelog := c.PostForm("changelog")
    
    // 转换布尔型参数
    isRecommend := int8(0)
    if c.PostForm("is_recommend") == "1" {
        isRecommend = 1
    }
    
    isForceUpgrade := int8(0)
    if c.PostForm("is_force_upgrade") == "1" {
        isForceUpgrade = 1
    }
    
    isStable := int8(1) // 默认为稳定版
    if c.PostForm("is_stable") == "0" {
        isStable = 0
    }
    
    // 创建业务请求结构体并保存
    createReq := &service.CreateVersionRequest{
        PackageID:     uint(packageId),
        Version:       version,
        Description:   description,
        Changelog:     changelog,
        MinVersion:    c.PostForm("min_version"),
        IsForceUpgrade: isForceUpgrade,
        IsRecommend:   isRecommend,
        IsStable:      isStable,
    }
    
    // 使用c.SaveUploadedFile保存上传的文件到临时路径
    tempPath := path.Join(os.TempDir(), file.Filename)
    if err := c.SaveUploadedFile(file, tempPath); err != nil {
        response.InternalError(c, "保存上传文件失败: "+err.Error())
        return
    }
    
    defer os.Remove(tempPath) // 处理结束时删除临时文件
    
    // 调用业务层处理
    versionRecord, err := h.packageService.CreateVersion(
        c.Request.Context(),
        createReq,
        tempPath,
        file.Filename,
    )
    
    if err != nil {
        response.Error(c, 500, "创建版本失败: "+err.Error())
        return
    }
    
    // 构建返回结果
    resp := UploadVersionResponse{
        ID:        versionRecord.ID,
        PackageID: versionRecord.PackageID,
        Version:   versionRecord.Version,
        FileName:  versionRecord.FileName,
        FileSize:  versionRecord.FileSize,
        FileSHA256: versionRecord.FileSHA256,
        Description: versionRecord.Description,
        Changelog: versionRecord.Changelog,
        IsRecommend: versionRecord.IsRecommend,
        MinVersion: versionRecord.MinVersion,
        IsStable:   versionRecord.IsStable,
        IsForceUpgrade: versionRecord.IsForceUpgrade,
        ReleaseDate: versionRecord.ReleaseDate.Format(time.RFC3339),
        CreatedAt:   versionRecord.CreatedAt.Format(time.RFC3339),
    }
    
    response.Success(c, resp)
}
```

## 测试

当前项目没有测试文件。如需添加测试：

```bash
# 运行所有测试
go test ./...

# 运行单个包的测试
go test ./internal/service/...

# 运行单个测试
go test -run TestFunctionName ./path/to/package

# 带覆盖率
go test -cover ./...
```

## 代码风格指南

### Go 代码规范

#### 导入顺序

导入按以下顺序分组，组间用空行分隔：

```go
import (
    // 1. 标准库
    "context"
    "errors"
    
    // 2. 外部包
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // 3. 项目内部包
    "github.com/yourorg/vpublish/internal/model"
    "github.com/yourorg/vpublish/pkg/response"
)
```

#### 命名规范

- **包名**: 小写单词，不使用下划线
- **结构体/接口**: PascalCase（导出），camelCase（私有）
- **常量**: PascalCase 或全大写下划线分隔
- **错误变量**: `Err` 前缀，如 `ErrUserNotFound`
- **接口**: 动词+er 后缀，如 `Handler`, `Repository`

#### 错误处理

```go
// 在 service 层定义业务错误
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
)

// 返回错误，不使用 panic
func (s *UserService) GetByID(ctx context.Context, id uint) (*model.User, error) {
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, ErrUserNotFound
    }
    return user, nil
}

// handler 层统一响应处理
func (h *UserHandler) Get(c *gin.Context) {
    user, err := h.userService.GetByID(ctx, id)
    if err != nil {
        response.NotFound(c, "user not found")
        return
    }
    response.Success(c, user)
}
```

#### Context 使用

所有跨层方法必须传入 `context.Context` 作为第一个参数：

```go
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    return &user, err
}
```

#### 结构体定义

```go
// Handler 结构体
type CategoryHandler struct {
    categoryService *service.CategoryService
}

// 构造函数
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
    return &CategoryHandler{categoryService: categoryService}
}

// Request 结构体放在 service 层
type CreateCategoryRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    SortOrder   int    `json:"sort_order"`
}

// Model 使用 GORM 标签
type Category struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Name        string         `gorm:"size:100;uniqueIndex;not null" json:"name"`
    Description string         `gorm:"size:500" json:"description"`
    CreatedAt   time.Time      `json:"created_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

#### 方法命名规范

Handler 层:
- `List` - 列表
- `Get` - 获取单个
- `Create` - 创建
- `Update` - 更新
- `Delete` - 删除

Repository 层:
- `Create`, `Update`, `Delete`
- `GetByID`, `GetByName`, `GetByCode`
- `List`, `ListActive`
- `ExistsByCode`

Service 层:
- `Create`, `Update`, `Delete`
- `GetByID`, `GetByCode`, `List`
- 业务方法如 `ChangePassword`, `ResetPassword`

### 前端代码规范 (Vue 3 + TypeScript)

#### API 定义

```typescript
// src/api/example.ts
import { get, post, put, del, type ApiResponse, type PageResponse } from './request'

export interface User {
  id: number
  username: string
  nickname: string
}

export const userApi = {
  list(params: { page?: number }): Promise<ApiResponse<PageResponse<User>>> {
    return get('/admin/users', { params })
  },
  
  get(id: number): Promise<ApiResponse<User>> {
    return get(`/admin/users/${id}`)
  },
  
  create(data: CreateUserRequest): Promise<ApiResponse<User>> {
    return post('/admin/users', data)
  },
}
```

#### 类型定义

- 使用 TypeScript 接口定义所有数据结构
- 接口命名: PascalCase
- 使用 `type` 定义联合类型、工具类型

## 项目结构说明

```
vpublish/
├── cmd/
│   ├── server/main.go      # HTTP 服务入口
│   └── mcp/main.go         # MCP 服务入口
├── internal/
│   ├── config/             # 配置加载
│   ├── database/           # 数据库连接和迁移
│   ├── handler/            # HTTP 处理器（控制器）
│   ├── middleware/         # 中间件（JWT、签名、CORS）
│   ├── model/              # GORM 数据模型
│   ├── repository/         # 数据访问层
│   ├── service/            # 业务逻辑层
│   └── cron/               # 定时任务
├── pkg/                    # 公共工具包
│   ├── jwt/                # JWT 工具
│   ├── pinyin/             # 拼音转换（生成类别代码）
│   ├── response/           # HTTP 响应封装
│   ├── signature/          # HMAC-SHA256 签名
│   └── storage/            # 文件存储
├── web/                    # Vue 3 前端
│   └── src/
│       ├── api/            # API 接口定义
│       ├── views/          # 页面组件
│       ├── stores/         # Pinia 状态管理
│       ├── router/         # 路由配置
│       └── utils/          # 工具函数
├── configs/config.yaml     # 配置文件
└── migrations/             # 数据库迁移
```

## 分层架构

遵循三层架构，依赖方向: `handler → service → repository → model`

1. **handler**: 处理 HTTP 请求，参数验证，调用 service，返回响应
2. **service**: 业务逻辑，事务处理，定义业务错误
3. **repository**: 数据库操作，纯 CRUD
4. **model**: 数据模型定义

## 认证机制

- **管理端 API**: JWT Token 认证，请求头 `Authorization: Bearer <token>`
- **APP端 API**: AppKey + HMAC-SHA256 签名认证
  - 请求头: `X-App-Key`, `X-Timestamp`, `X-Signature`
- **下载链接**: 带签名的临时 URL

## 注意事项

1. 密码等敏感字段使用 `json:"-"` 排除序列化
2. 所有数据库操作使用 `WithContext(ctx)`
3. 使用 `response` 包统一返回格式: `{code, message, data}`
4. 分页使用 `response.Page(c, list, total, page, pageSize)`
5. 中文注释和中文错误信息用于用户友好提示