# vpublish

低空智能平台软件包管理系统

## 项目结构

```
vpublish/
├── cmd/
│   ├── server/          # 后端服务入口
│   └── mcp/             # MCP服务入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── handler/         # HTTP处理器
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   ├── repository/      # 数据访问层
│   ├── service/         # 业务逻辑层
│   └── cron/            # 定时任务
├── pkg/                 # 公共工具包
│   ├── jwt/             # JWT工具
│   ├── pinyin/          # 拼音转换
│   ├── response/        # 响应封装
│   ├── signature/       # 签名工具
│   └── storage/         # 文件存储
├── configs/             # 配置文件
├── uploads/             # 上传文件存储
├── web/                 # 前端项目
│   ├── src/
│   │   ├── api/         # API接口
│   │   ├── views/       # 页面组件
│   │   ├── stores/      # 状态管理
│   │   ├── router/      # 路由配置
│   │   ├── utils/       # 工具函数
│   │   └── assets/      # 静态资源
│   └── package.json
├── go.mod
└── README.md
```

## 技术栈

### 后端
- Go 1.21+
- Gin (Web框架)
- GORM (ORM)
- MySQL 8.0
- JWT认证
- HMAC-SHA256签名

### 前端
- Vue 3
- Vite
- TypeScript
- Element Plus
- Pinia
- ECharts

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 18+
- MySQL 8.0+

### 后端启动

```bash
# 安装依赖
go mod tidy

# 创建数据库
mysql -u root -p -e "CREATE DATABASE vpublish CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 修改配置
cp configs/config.yaml.example configs/config.yaml
# 编辑 config.yaml 填入数据库密码等

# 运行
go run cmd/server/main.go
```

### 前端启动

```bash
cd web

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建
npm run build
```

## API 文档

### APP端 API

所有APP端API需要签名认证，请求头包含：
- `X-App-Key`: 应用Key
- `X-Timestamp`: 时间戳
- `X-Signature`: 签名

```
GET  /api/v1/app/categories              # 获取软件类别列表
GET  /api/v1/app/categories/:code/latest # 获取某类别最新版本
GET  /api/v1/app/download/:id            # 下载软件包
```

### 管理端 API

管理端API需要JWT认证，请求头包含：
- `Authorization: Bearer <token>`

```
# 认证
POST /api/v1/admin/auth/login    # 登录
POST /api/v1/admin/auth/logout   # 登出

# 用户管理
GET/POST/PUT/DELETE /api/v1/admin/users

# 类别管理
GET/POST/PUT/DELETE /api/v1/admin/categories

# 软件包管理
GET/POST/PUT/DELETE /api/v1/admin/packages
POST /api/v1/admin/packages/:id/versions  # 上传版本

# 统计
GET /api/v1/admin/stats/overview
GET /api/v1/admin/stats/daily
GET /api/v1/admin/stats/monthly
GET /api/v1/admin/stats/yearly
```

## MCP 服务

MCP服务提供了完整的软件包管理功能，支持AI助手直接调用。

### 构建MCP服务

```bash
go build -o vpublish-mcp ./cmd/mcp
```

### MCP工具列表

- `list_categories` - 获取所有软件类别
- `list_packages` - 获取软件包列表
- `list_versions` - 获取软件包版本列表
- `get_latest_version` - 获取指定类别的最新版本
- `create_category` - 创建新的软件类别
- `create_package` - 创建新的软件包
- `get_download_stats` - 获取下载统计
- `delete_version` - 删除指定版本

## 功能特性

### 软件类别管理
- 中文类别名称自动生成拼音代码（如："无人机" → "TYPE_WU_REN_JI"）
- 支持排序、启用/禁用

### 软件包管理
- 多类别管理
- 版本号解析（语义化版本）
- 文件上传与存储
- SHA256文件校验

### 版本发布
- 支持强制升级标记
- 稳定版/测试版区分
- 更新日志和发布说明
- 最低兼容版本设置

### 安全认证
- APP端：AppKey + HMAC-SHA256签名
- 管理端：JWT Token认证
- 下载链接签名验证

### 统计分析
- 日/月/年下载量统计
- 类别分布统计
- 图表可视化展示

## 部署

### 二进制部署

```bash
# 构建后端
go build -o vpublish-server ./cmd/server

# 构建前端
cd web && npm run build

# 运行
./vpublish-server
```

## License

MIT