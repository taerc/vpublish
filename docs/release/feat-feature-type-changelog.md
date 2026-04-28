# 版本更新说明：Feature Type 功能

## 版本信息

- **分支**: `feat-add-build-type-fm`
- **提交**: `881e827` + `e5d0b00`
- **日期**: 2026-04-27

---

## 功能概述

新增 `feature_type`（功能类型）字段，支持同一软件包按 **调试版本（debug）**、**正式版本（release）**、**演示版本（demo）** 分别维护最新版本。获取最新版本时，查询维度从单一 `code` 扩展为 `code + feature_type` 组合。

---

## 变更清单

### 数据库

| 变更项 | 说明 |
|--------|------|
| `versions.feature_type` | 新增字段，`VARCHAR(20) NOT NULL DEFAULT 'release'` |
| 历史数据迁移 | 已有记录自动设为 `release`，保持现有行为一致 |

### API 变更

| 接口 | 变更 |
|------|------|
| `POST /api/v1/admin/packages` | 新增 `feature_type` 表单参数（可选，默认 `release`） |
| `POST /api/v1/admin/packages/:id/versions` | 新增 `feature_type` 表单参数（可选，默认 `release`） |
| `GET /api/v1/admin/packages/:id/versions` | 新增 `feature_type` 查询参数（可选，不过滤） |
| `GET /api/v1/app/categories/:code/latest` | 新增 `type` 查询参数（可选，默认 `release`） |
| `GET /api/v1/app/categories/:code/versions` | 新增 `type` 查询参数（可选，默认 `release`） |
| 所有接口返回数据 | 新增 `feature_type` 字段（追加，不删除已有字段） |

### 前端变更

| 页面 | 变更 |
|------|------|
| 软件包管理列表 | 创建软件包对话框增加功能类型下拉选择 |
| 软件包版本管理 | 上传版本对话框增加功能类型选择；版本列表增加功能类型列和筛选 |

### 业务逻辑变更

| 变更项 | 原逻辑 | 新逻辑 |
|--------|--------|--------|
| `is_latest` 维护 | 按 `package_id` 唯一 | 按 `(package_id, feature_type)` 组合维护 |
| 版本号比较 | 同一 package 下所有版本比较 | 仅同 `package_id + feature_type` 内比较 |
| 最新版本查询 | 按 `code` 取最高版本 | 按 `code + feature_type` 取最高版本 |
| APP 端默认行为 | 取全类最新 | 默认取 `release` 类型最新 |

### MCP 变更

| 工具 | 变更 |
|------|------|
| `get_latest_version` | 新增 `feature_type` 参数（非必填，默认 `release`） |

---

## 兼容性说明

### 向后兼容
- APP 端不传 `type` 参数时，默认查询 `release` 类型，行为与升级前一致
- 管理端上传不传 `feature_type` 时，默认为 `release`
- API 返回新增字段 `feature_type`，不删除或修改已有字段
- 历史数据 `feature_type` 默认 `release`

### 不兼容项
- **无**。此次更新完全向后兼容。

### 注意事项
- 如果客户端之前依赖 `GET /app/categories/:code/latest` 返回的是"绝对最新版本"（不限类型），升级后将返回 `release` 类型最新版本。如需获取其他类型，需显式传参 `?type=debug`。

---

## 测试情况

| 测试项 | 结果 | 说明 |
|--------|------|------|
| Go 编译 | PASS | `go build ./...` |
| Go vet | PASS | （middleware 有预存 UTF-8 问题，与此改动无关） |
| Model 测试 | PASS | 14 个用例 |
| Repository 测试 | PASS | feature_type 过滤验证 |
| Service 测试 | PASS | 版本比较/上传逻辑验证 |
| Handler 测试 | PASS | 参数解析验证 |
| TypeScript 检查 | PASS | `vue-tsc --noEmit` |
| 前端构建 | PASS | `vite build` |
| Swagger 文档 | PASS | `swag init` 生成成功 |
