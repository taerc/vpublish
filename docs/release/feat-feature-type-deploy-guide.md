# 生产环境升级指南：v2.1.0 (Feature Type)

> 分支 `feat-add-build-type-fm` 已合并至 `master`，以下为生产环境升级标准流程。

## 变更概要

- **数据库**: `versions` 表新增 `feature_type` 列 (`VARCHAR(20) NOT NULL DEFAULT 'release'`)，已有记录自动迁移
- **API**: 版本上传和查询接口新增 `feature_type` 参数（可选，默认 `release`）
- **前端**: 版本管理页面增加功能类型选择和筛选
- **兼容性**: 完全向后兼容，无破坏性变更

详见 [CHANGELOG](./feat-feature-type-changelog.md)

---

## 方案一：一键升级（推荐）

使用 `deploy/upgrade.sh` 脚本自动完成备份→升级→验证全流程。

### 1. 在开发机器上打包

```bash
cd /path/to/vpublish
./scripts/build-release.sh v2.1.0
# 产物: dist/vpublish-v2.1.0-linux-amd64.tar.gz
```

### 2. 上传到生产服务器

```bash
scp dist/vpublish-v2.1.0-linux-amd64.tar.gz user@prod-server:/tmp/
```

### 3. 在生产服务器上执行升级

```bash
ssh user@prod-server
cd /tmp
tar -xzf vpublish-v2.1.0-linux-amd64.tar.gz
cd vpublish-v2.1.0-linux-amd64

# 预览操作步骤（不实际执行）
sudo ./deploy/upgrade.sh --dry-run

# 执行升级
sudo ./deploy/upgrade.sh
```

升级脚本自动完成：
1. 备份当前二进制、前端、配置、数据库
2. 停止服务
3. 替换文件
4. 启动服务（GORM AutoMigrate 自动处理数据库迁移）
5. 验证健康状态和数据库列

### 回滚

```bash
sudo ./deploy/upgrade.sh --rollback
```

---

## 方案二：手动升级

如果不想使用自动脚本，按以下步骤手动操作。

### 1. 备份

```bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
mkdir -p /opt/vpublish/backups/backup_${TIMESTAMP}

# 备份二进制
cp /opt/vpublish/vpublish-server /opt/vpublish/backups/backup_${TIMESTAMP}/
[ -f /opt/vpublish/vpublish-mcp ] && cp /opt/vpublish/vpublish-mcp /opt/vpublish/backups/backup_${TIMESTAMP}/

# 备份前端
cp -r /opt/vpublish/web/dist /opt/vpublish/backups/backup_${TIMESTAMP}/

# 备份数据库
mysqldump -h 172.16.10.56 -u root -p vpublish > /opt/vpublish/backups/backup_${TIMESTAMP}/vpublish_${TIMESTAMP}.sql
```

### 2. 替换文件

```bash
# 停止服务
systemctl stop vpublish

# 替换二进制
cp vpublish-server /opt/vpublish/vpublish-server
chmod +x /opt/vpublish/vpublish-server
[ -f vpublish-mcp ] && cp vpublish-mcp /opt/vpublish/vpublish-mcp && chmod +x /opt/vpublish/vpublish-mcp

# 替换前端
rm -rf /opt/vpublish/web/dist
cp -r web/dist /opt/vpublish/web/dist

# 启动服务（自动执行数据库迁移）
systemctl start vpublish
sleep 3
```

### 3. 验证

```bash
# 健康检查
curl http://127.0.0.1:8080/health

# 检查 feature_type 列
mysql -h 172.16.10.56 -u root -p -e "
  SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
  FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA='vpublish' AND TABLE_NAME='versions' AND COLUMN_NAME='feature_type';
"

# 检查空值记录（应为 0）
mysql -h 172.16.10.56 -u root -p -e "
  SELECT COUNT(*) as null_count FROM vpublish.versions
  WHERE feature_type IS NULL OR feature_type = '';
"
```

### 手动回滚

```bash
LATEST_BACKUP=$(ls -dt /opt/vpublish/backups/backup_* | head -1)
systemctl stop vpublish
cp "$LATEST_BACKUP/vpublish-server" /opt/vpublish/vpublish-server
rm -rf /opt/vpublish/web/dist
cp -r "$LATEST_BACKUP/dist" /opt/vpublish/web/dist
systemctl start vpublish
```

---

## 功能验证

升级完成后，建议验证新功能是否正常。

```bash
# 1. 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 2. 上传 debug 版本（验证 feature_type 字段）
curl -s -X POST http://localhost:8080/api/v1/admin/packages/1/versions \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test-app-debug.apk" \
  -F "version=1.0.1" \
  -F "feature_type=debug" \
  | jq '.data.feature_type'
# 预期: "debug"

# 3. APP 端获取 release 版本（默认行为不变）
curl -s "http://localhost:8080/api/v1/app/categories/TYPE_WU_REN_JI/latest" \
  | jq '.data.feature_type'
# 预期: "release"
```

---

## 影响范围

| 组件 | 影响 | 是否需要客户端更新 |
|------|------|---|
| **数据库** | versions 表新增 1 列，自动迁移 | 否 |
| **APP 客户端** | 不传 type 参数行为不变（默认 release） | 否 |
| **管理端 Web** | 页面新增功能类型选择和筛选 | 是（需重新构建前端） |
| **MCP 服务** | 可选，独立部署时建议一并更新 | 否（参数可选） |

---

## 升级检查清单

- [ ] 开发机器打包成功 (`./scripts/build-release.sh v2.1.0`)
- [ ] 上传到生产服务器
- [ ] 数据库备份完成
- [ ] 二进制替换完成
- [ ] 前端文件替换完成
- [ ] 服务启动成功
- [ ] 健康检查通过
- [ ] `feature_type` 列存在且默认值为 `release`
- [ ] 历史数据 feature_type 全部非空
- [ ] 回滚方案已确认
