# MCP Configuration for vpublish

This MCP server provides tools for managing software packages.

## Available Tools

### list_categories
获取所有软件类别列表

### list_packages
获取软件包列表
- `category_id` (可选): 类别ID

### list_versions
获取软件包的所有版本
- `package_id` (必需): 软件包ID

### get_latest_version
获取指定类别的最新版本
- `category_code` (必需): 类别代码，如 TYPE_WU_REN_JI

### create_category
创建新的软件类别
- `name` (必需): 类别中文名称
- `description` (可选): 类别描述

### create_package
创建新的软件包
- `category_id` (必需): 类别ID
- `name` (必需): 软件包名称
- `description` (可选): 软件包描述
- `created_by` (可选): 创建者用户ID，默认为0

### get_download_stats
获取下载统计数据
- `type` (必需): 统计类型 daily/monthly/yearly
- `category_id` (可选): 类别ID
- `year` (可选): 年份
- `month` (可选): 月份
- `date` (可选): 日期 (YYYY-MM-DD)

### delete_version
删除指定版本
- `version_id` (必需): 版本ID

## Configuration

Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "vpublish": {
      "command": "vpublish-mcp",
      "args": []
    }
  }
}
```

## Building

```bash
go build -o vpublish-mcp ./cmd/mcp
```