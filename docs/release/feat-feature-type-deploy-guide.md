# 部署更新操作说明：Feature Type 功能

## 一、更新前准备

### 1.1 备份数据库

```bash
mysqldump -h 172.16.10.56 -u root -p vpublish-test > vpublish-backup-$(date +%Y%m%d).sql
```

### 1.2 备份当前服务

```bash
# 备份当前二进制
cp /path/to/current/vpublish-server /path/to/backup/vpublish-server-backup-$(date +%Y%m%d)
```

### 1.3 确认分支代码

```bash
cd /wkspace/git/vpublish
git fetch origin feat-add-build-type-fm
git checkout feat-add-build-type-fm
git log --oneline -3
# 预期看到两个提交：
# e5d0b00 fix: resolve TSC type errors in feature_type implementation
# 881e827 feat: add feature_type (debug/release/demo) to versions
```

---

## 二、后端部署

### 2.1 构建后端

```bash
cd /wkspace/git/vpublish
go build -o vpublish-server ./cmd/server
```

### 2.2 验证构建

```bash
go vet ./...
# middleware 有预存 UTF-8 报错可忽略，确认其他模块无异常
```

### 2.3 停止服务

```bash
# 停止当前运行的服务
systemctl stop vpublish  # 或 kill <pid>
```

### 2.4 替换二进制

```bash
cp vpublish-server /path/to/deploy/vpublish-server
chmod +x /path/to/deploy/vpublish-server
```

### 2.5 启动服务（自动触发数据库迁移）

```bash
systemctl start vpublish  # 或 /path/to/deploy/vpublish-server
```

### 2.6 验证数据库迁移

```bash
# 检查 feature_type 列是否已创建
mysql -h 172.16.10.56 -u root -p -e "
  SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT 
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA='vpublish-test' 
    AND TABLE_NAME='versions' 
    AND COLUMN_NAME='feature_type';
"

# 预期输出:
# COLUMN_NAME   | DATA_TYPE | IS_NULLABLE | COLUMN_DEFAULT
# feature_type  | varchar   | NO          | release
```

```bash
# 检查已有记录的 feature_type 是否已迁移为 release
mysql -h 172.16.10.56 -u root -p -e "
  SELECT COUNT(*) as null_count 
  FROM vpublish-test.versions 
  WHERE feature_type IS NULL OR feature_type = '';
"

# 预期输出: null_count = 0
```

### 2.7 验证服务健康

```bash
curl http://localhost:8080/health
# 预期返回: {"status":"ok","version":"..."}
```

---

## 三、前端部署

### 3.1 构建前端

```bash
cd /wkspace/git/vpublish/web
npm install
npx vue-tsc --noEmit   # 类型检查
npx vite build          # 构建生产产物
```

### 3.2 部署前端产物

```bash
# 将 web/dist/ 目录部署到 Web 服务器
rsync -avz dist/ /path/to/web-server/html/
```

---

## 四、功能验证

### 4.1 管理端上传测试

```bash
# 1. 登录获取 JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 2. 上传 release 版本（默认）
curl -s -X POST http://localhost:8080/api/v1/admin/packages/1/versions \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test-app.apk" \
  -F "version=1.0.0" \
  | jq '.data.feature_type'
# 预期输出: "release"

# 3. 上传 debug 版本
curl -s -X POST http://localhost:8080/api/v1/admin/packages/1/versions \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test-app-debug.apk" \
  -F "version=1.0.1" \
  -F "feature_type=debug" \
  | jq '.data.feature_type'
# 预期输出: "debug"
```

### 4.2 APP 端查询测试

```bash
# 获取类别为 TYPE_WU_REN_JI 的最新 release 版本
curl -s "http://localhost:8080/api/v1/app/categories/TYPE_WU_REN_JI/latest?type=release" \
  | jq '.data.feature_type'
# 预期输出: "release"

# 获取同一类别的最新 debug 版本
curl -s "http://localhost:8080/api/v1/app/categories/TYPE_WU_REN_JI/latest?type=debug" \
  | jq '.data.feature_type'
# 预期输出: "debug"

# 不传 type 参数，应该默认返回 release
curl -s "http://localhost:8080/api/v1/app/categories/TYPE_WU_REN_JI/latest" \
  | jq '.data.feature_type'
# 预期输出: "release"
```

### 4.3 管理端版本列表筛选

```bash
# 筛选 release 版本
curl -s "http://localhost:8080/api/v1/admin/packages/1/versions?feature_type=release" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data.list[].feature_type'
# 预期全部输出: "release"
```

---

## 五、回滚方案

如果部署后出现问题，按以下步骤回滚：

### 5.1 停止服务

```bash
systemctl stop vpublish
```

### 5.2 恢复二进制

```bash
cp /path/to/backup/vpublish-server-backup-* /path/to/deploy/vpublish-server
```

### 5.3 恢复前端

```bash
# 从备份恢复前端
rsync -avz /path/to/backup/dist/ /path/to/web-server/html/
```

### 5.4 恢复数据库

```bash
mysql -h 172.16.10.56 -u root -p vpublish-test < /path/to/backup/vpublish-backup-*.sql
```

### 5.5 重启服务

```bash
systemctl start vpublish
```

---

## 六、影响范围评估

| 组件 | 影响 |
|------|------|
| **数据库** | versions 表增加1列，历史数据自动迁移，无数据丢失风险 |
| **APP 客户端** | 不传 type 参数行为不变（默认 release），无需客户端更新 |
| **管理端 Web** | 前端页面增加类型选择，更新后可使用新功能 |
| **MCP 服务** | 可选，如有独立部署需一并更新 |

---

## 七、检查清单

- [x] 数据库备份完成
- [x] 后端构建成功
- [x] `feature_type` 列存在且默认值为 `release`
- [x] 历史数据 feature_type 全部非空
- [x] 前端构建成功
- [x] 管理端上传 release 版本成功
- [x] 管理端上传 debug 版本成功
- [x] APP latest 接口返回 release 类型
- [x] APP latest 接口返回 debug 类型（传 type=debug）
- [x] 不传 type 参数默认返回 release
- [x] 服务重启无异常
