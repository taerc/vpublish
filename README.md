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

### 传输模式

MCP服务支持两种传输模式：

| 模式 | 命令参数 | 说明 |
|------|---------|------|
| **stdio** | `-transport=stdio` (默认) | 标准输入/输出，适合本地客户端 |
| **HTTP** | `-transport=http` | HTTP + SSE，适合远程访问 |

### 构建MCP服务

```bash
go build -o vpublish-mcp ./cmd/mcp
```

### stdio 模式配置

**Trae / Claude Desktop 配置：**

```json
{
  "mcpServers": [
    {
      "name": "vpublish-mcp",
      "command": ["D:/wkspace/git/vpublish/vpublish-mcp.exe"],
      "env": {
        "MCP_APP_KEY": "your_app_key",
        "MCP_APP_SECRET": "your_app_secret"
      }
    }
  ]
}
```

### HTTP 模式配置

**1. 启动服务：**

```bash
./vpublish-mcp -transport=http
```

**2. Trae 配置：**

```json
{
  "mcpServers": {
    "vpublish-mcp": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "X-MCP-App-Key": "your_app_key",
        "X-MCP-App-Secret": "your_app_secret"
      }
    }
  }
}
```

**3. 配置文件 (`configs/config.yaml`)：**

```yaml
mcp:
  http:
    enabled: true          # 是否启用 HTTP 传输
    host: localhost        # 监听地址
    port: 8080             # 与主服务共用端口
    endpoint_path: /mcp    # MCP 端点路径
```

### 认证方式

| 方式 | Header 格式 | 说明 |
|------|------------|------|
| 自定义 Header | `X-MCP-App-Key` + `X-MCP-App-Secret` | 推荐方式 |
| Bearer Token | `Authorization: Bearer <token>` | 标准方式 |

### MCP工具列表

| 工具 | 权限 | 说明 |
|------|------|------|
| `list_categories` | 只读 | 获取所有软件类别 |
| `list_packages` | 只读 | 获取软件包列表 |
| `list_versions` | 只读 | 获取软件包版本列表 |
| `get_latest_version` | 只读 | 获取指定类别的最新版本 |
| `get_download_stats` | 只读 | 获取下载统计 |
| `create_category` | 读写 | 创建新的软件类别 |
| `create_package` | 读写 | 创建新的软件包 |
| `delete_version` | 读写 | 删除指定版本 |
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