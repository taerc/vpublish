# vpublish 部署文档

## 目录

- [系统概述](#系统概述)
- [环境要求](#环境要求)
- [系统架构](#系统架构)
- [数据库配置](#数据库配置)
- [配置文件说明](#配置文件说明)
- [部署方式](#部署方式)
- [启动与停止](#启动与停止)
- [安全配置](#安全配置)
- [监控与日志](#监控与日志)
- [故障排查](#故障排查)

---

## 系统概述

vpublish 是一个低空智能平台软件包管理系统，提供以下核心功能：

- **软件版本管理**: 支持多类别软件包管理、语义化版本控制、SHA256文件校验
- **下载统计**: 日/月/年下载量统计、类别分布统计、图表可视化
- **APP端API**: 提供软件查询、下载等接口
- **MCP服务**: 支持 AI 助手直接调用软件包管理功能

### 技术栈

| 组件 | 技术选型 |
|------|----------|
| 后端框架 | Go 1.21+ / Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0+ |
| 前端框架 | Vue 3 / Vite / TypeScript |
| UI组件库 | Element Plus |
| 状态管理 | Pinia |
| 图表库 | ECharts |
| 认证方式 | JWT + HMAC-SHA256 签名 |

---

## 环境要求

### 硬件要求

| 资源 | 最低配置 | 推荐配置 |
|------|----------|----------|
| CPU | 2核 | 4核+ |
| 内存 | 4GB | 8GB+ |
| 磁盘 | 50GB | 200GB+ (根据软件包存储需求) |

### 软件依赖

| 软件 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.21+ | 后端运行环境 |
| Node.js | 18+ | 前端构建环境 |
| MySQL | 8.0+ | 数据库 |
| Nginx | 1.18+ | 反向代理（可选） |

---

## 系统架构

### 目录结构

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
│       └── router/         # 路由配置
├── configs/config.yaml     # 配置文件
└── uploads/                # 上传文件存储
```

### 分层架构

```
┌─────────────────────────────────────────────┐
│                  Handler 层                  │  HTTP 请求处理、参数验证
├─────────────────────────────────────────────┤
│                  Service 层                  │  业务逻辑、事务处理
├─────────────────────────────────────────────┤
│                Repository 层                 │  数据库操作、纯 CRUD
├─────────────────────────────────────────────┤
│                  Model 层                    │  数据模型定义
└─────────────────────────────────────────────┘
```

### 数据模型

| 表名 | 说明 |
|------|------|
| `users` | 管理员用户 |
| `app_keys` | APP 认证密钥 |
| `categories` | 软件类别 |
| `packages` | 软件包 |
| `versions` | 软件版本 |
| `download_logs` | 下载日志 |
| `download_stats` | 下载统计（按天聚合） |
| `operation_logs` | 操作日志 |

### API 路由

#### 管理端 API (需要 JWT 认证)

```
POST   /api/v1/admin/auth/login           # 登录
POST   /api/v1/admin/auth/refresh         # 刷新 Token
POST   /api/v1/admin/auth/logout          # 登出
GET    /api/v1/admin/auth/profile         # 获取用户信息
PUT    /api/v1/admin/auth/password        # 修改密码

# 用户管理
GET    /api/v1/admin/users                # 用户列表
POST   /api/v1/admin/users                # 创建用户
PUT    /api/v1/admin/users/:id            # 更新用户
DELETE /api/v1/admin/users/:id            # 删除用户

# 类别管理
GET    /api/v1/admin/categories           # 类别列表
POST   /api/v1/admin/categories           # 创建类别
PUT    /api/v1/admin/categories/:id       # 更新类别
DELETE /api/v1/admin/categories/:id       # 删除类别

# 软件包管理
GET    /api/v1/admin/packages             # 软件包列表
POST   /api/v1/admin/packages             # 创建软件包
PUT    /api/v1/admin/packages/:id         # 更新软件包
DELETE /api/v1/admin/packages/:id         # 删除软件包

# 版本管理
GET    /api/v1/admin/packages/:id/versions  # 版本列表
POST   /api/v1/admin/packages/:id/versions  # 上传版本
DELETE /api/v1/admin/versions/:id           # 删除版本

# 统计
GET    /api/v1/admin/stats/overview       # 总览统计
GET    /api/v1/admin/stats/daily          # 日统计
GET    /api/v1/admin/stats/monthly        # 月统计
GET    /api/v1/admin/stats/yearly         # 年统计
GET    /api/v1/admin/stats/category       # 类别分布
```

#### APP 端 API (需要签名认证)

```
GET    /api/v1/app/categories             # 获取软件类别列表
GET    /api/v1/app/categories/:code/latest # 获取某类别最新版本
GET    /api/v1/app/download/:id           # 下载软件包
```

---

## 数据库配置

### 创建数据库

```bash
mysql -u root -p -e "CREATE DATABASE vpublish CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 数据库用户授权（推荐）

```sql
-- 创建专用用户
CREATE USER 'vpublish'@'%' IDENTIFIED BY 'your_strong_password';

-- 授权
GRANT ALL PRIVILEGES ON vpublish.* TO 'vpublish'@'%';
FLUSH PRIVILEGES;
```

### 数据库迁移

程序启动时会自动执行数据库迁移（AutoMigrate），无需手动执行迁移脚本。

---

## 配置文件说明

### 配置文件路径

```
configs/config.yaml
```

### 完整配置示例

```yaml
# 服务配置
server:
  port: 8080                    # 服务端口
  mode: release                 # 运行模式: debug, release
  read_timeout: 30s             # 读取超时
  write_timeout: 30s            # 写入超时

# 数据库配置
database:
  host: 172.16.10.56            # 数据库地址
  port: 3306                    # 数据库端口
  user: root                    # 数据库用户
  password: your_password       # 数据库密码
  dbname: vpublish              # 数据库名称
  max_open_conns: 100           # 最大连接数
  max_idle_conns: 10            # 最大空闲连接数

# JWT 配置
jwt:
  secret: your_jwt_secret_key   # JWT 密钥（请修改为复杂字符串）
  expire: 24h                   # Token 过期时间
  refresh_expire: 168h          # 刷新 Token 过期时间

# 存储配置
storage:
  type: local                   # 存储类型: local
  path: ./uploads               # 存储路径
  max_file_size: 104857600      # 最大文件大小 (100MB)

# 日志配置
log:
  level: info                   # 日志级别: debug, info, warn, error

# CORS 配置
cors:
  enabled: true
  allow_origins:
    - "http://localhost:3000"
    - "http://127.0.0.1:3000"
    - "https://your-domain.com"
  allow_methods:
    - GET
    - POST
    - PUT
    - DELETE
    - OPTIONS
  allow_headers:
    - Origin
    - Content-Type
    - Authorization
    - X-App-Key
    - X-Timestamp
    - X-Signature
```

### 配置项详解

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.port` | HTTP 服务端口 | 8080 |
| `server.mode` | 运行模式 (debug/release) | debug |
| `database.host` | MySQL 主机地址 | - |
| `database.port` | MySQL 端口 | 3306 |
| `jwt.secret` | JWT 签名密钥 | - |
| `jwt.expire` | Token 有效期 | 24h |
| `storage.path` | 文件存储路径 | ./uploads |
| `storage.max_file_size` | 最大上传文件大小 | 100MB |

---

## 部署方式

### 方式一：二进制部署

#### 1. 构建后端

```bash
# 安装依赖
go mod tidy

# 构建 Linux 版本
GOOS=linux GOARCH=amd64 go build -o vpublish-server ./cmd/server

# 构建 MCP 服务（可选）
GOOS=linux GOARCH=amd64 go build -o vpublish-mcp ./cmd/mcp
```

#### 2. 构建前端

```bash
cd web

# 安装依赖
npm install

# 构建生产版本
npm run build
```

构建产物位于 `web/dist/` 目录。

#### 3. 部署目录结构

```
/opt/vpublish/
├── vpublish-server      # 后端二进制
├── configs/
│   └── config.yaml      # 配置文件
├── uploads/             # 上传文件存储
└── web/
    └── dist/            # 前端静态文件
```

#### 4. 配置 Systemd 服务

创建服务文件 `/etc/systemd/system/vpublish.service`:

```ini
[Unit]
Description=vpublish Server
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/vpublish
ExecStart=/opt/vpublish/vpublish-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
systemctl daemon-reload
systemctl enable vpublish
systemctl start vpublish
```

### 方式二：Nginx 反向代理部署

#### Nginx 配置示例

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /opt/vpublish/web/dist;
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 文件上传大小限制
        client_max_body_size 100M;
    }

    # 健康检查
    location /health {
        proxy_pass http://127.0.0.1:8080/health;
    }
}
```

### 方式三：开发环境部署

```bash
# 后端
go run cmd/server/main.go

# 前端（另一个终端）
cd web && npm run dev
```

---

## 启动与停止

### 启动服务

```bash
# Systemd 方式
systemctl start vpublish

# 直接运行
./vpublish-server

# 指定配置文件
./vpublish-server -config /path/to/config.yaml
```

### 停止服务

```bash
# Systemd 方式
systemctl stop vpublish

# 发送信号优雅关闭
kill -SIGTERM <pid>
```

### 健康检查

```bash
curl http://localhost:8080/health
# 响应: {"status":"ok"}
```

---

## 安全配置

### 1. JWT 密钥

```yaml
jwt:
  secret: "请使用至少32位的随机字符串"
```

生成随机密钥：

```bash
openssl rand -base64 32
```

### 2. APP 端签名认证

APP 端 API 使用 HMAC-SHA256 签名认证：

#### 请求头

| 请求头 | 说明 |
|--------|------|
| `X-App-Key` | 应用 Key |
| `X-Timestamp` | 时间戳 (RFC3339 格式) |
| `X-Signature` | HMAC-SHA256 签名 |

#### 签名生成流程

```
1. 将请求参数按 key 排序
2. 拼接为 key1=value1&key2=value2 格式
3. 追加 &timestamp=<timestamp>
4. 使用 AppSecret 进行 HMAC-SHA256 计算
5. 转换为十六进制字符串
```

#### 签名有效期

- 签名有效期：300 秒（5 分钟）
- 时间戳必须在有效期内

### 3. CORS 配置

生产环境请配置正确的允许域名：

```yaml
cors:
  enabled: true
  allow_origins:
    - "https://your-domain.com"
```

### 4. 文件上传安全

- 文件大小限制：100MB（可配置）
- 自动计算 SHA256 文件哈希
- 文件名安全处理

---

## 监控与日志

### 日志级别

| 级别 | 说明 |
|------|------|
| debug | 调试信息 |
| info | 常规信息 |
| warn | 警告信息 |
| error | 错误信息 |

### 日志输出

日志输出到标准输出，可通过 Systemd 或日志收集工具管理：

```bash
# 查看 Systemd 日志
journalctl -u vpublish -f

# 日志轮转配置
/var/log/vpublish/*.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
```

### 定时任务

系统包含以下定时任务：

| 任务 | 执行时间 | 说明 |
|------|----------|------|
| 下载统计聚合 | 每天凌晨 1 点 | 将下载日志聚合到统计表 |

---

## 故障排查

### 常见问题

#### 1. 数据库连接失败

```
Error: connect database: Error 1045: Access denied for user
```

**解决方案**：
- 检查数据库用户名和密码
- 确认数据库用户有正确的权限
- 检查数据库是否允许远程连接

#### 2. JWT Token 无效

```
Error: token is expired
```

**解决方案**：
- 检查系统时间是否正确
- 使用 refresh token 刷新
- 重新登录获取新 token

#### 3. 签名验证失败

```
Error: invalid signature
```

**解决方案**：
- 确认 AppKey 和 AppSecret 正确
- 检查时间戳格式（RFC3339）
- 确认签名算法实现正确
- 检查签名是否在有效期内（5分钟）

#### 4. 文件上传失败

```
Error: file too large
```

**解决方案**：
- 检查 `storage.max_file_size` 配置
- 如使用 Nginx，调整 `client_max_body_size`

#### 5. CORS 错误

```
Error: CORS policy blocked
```

**解决方案**：
- 检查 `cors.allow_origins` 配置
- 确认请求域名在允许列表中

### 性能优化建议

1. **数据库连接池**：根据并发量调整 `max_open_conns` 和 `max_idle_conns`
2. **文件存储**：考虑使用对象存储（如 MinIO、OSS）替代本地存储
3. **缓存**：可引入 Redis 缓存热点数据
4. **负载均衡**：使用 Nginx 或云负载均衡器

---

## MCP 服务部署

MCP 服务支持 AI 助手直接调用软件包管理功能。

### 构建

```bash
go build -o vpublish-mcp ./cmd/mcp
```

### MCP 工具列表

| 工具名 | 说明 |
|--------|------|
| `list_categories` | 获取所有软件类别 |
| `list_packages` | 获取软件包列表 |
| `list_versions` | 获取软件包版本列表 |
| `get_latest_version` | 获取指定类别的最新版本 |
| `create_category` | 创建新的软件类别 |
| `create_package` | 创建新的软件包 |
| `get_download_stats` | 获取下载统计 |
| `delete_version` | 删除指定版本 |

---

## 附录

### 默认管理员账户

首次部署需要通过数据库创建管理员账户：

```sql
-- 密码为 bcrypt 加密后的 "admin123"
INSERT INTO users (username, password_hash, nickname, role, is_active, created_at, updated_at)
VALUES ('admin', '$2a$10$...', '管理员', 'admin', true, NOW(), NOW());
```

### 端口清单

| 端口 | 服务 | 说明 |
|------|------|------|
| 8080 | HTTP Server | 后端 API 服务 |
| 3000 | Vite Dev Server | 前端开发服务器 |

### 相关文档

- [AGENTS.md](./AGENTS.md) - 项目开发指南
- [README.md](./README.md) - 项目简介

---

*文档版本: 1.0.0*
*最后更新: 2026-03-04*