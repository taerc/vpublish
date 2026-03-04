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