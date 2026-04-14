-- ============================================
-- 更新报错记录表结构
-- 更新时间：2026-04-14
-- 变更内容：
-- 1. 取消 request_id 的唯一索引，支持重复
-- 2. 新增 path 字段
-- 3. 将 error_message 和 remark 从 TEXT 改为 VARCHAR(500)
-- 4. 修改时间字段的默认值为 NOT NULL DEFAULT 0
-- 5. 移除 app_type 和 error_type 的索引
-- ============================================

-- 1. 新增 path 字段（如果不存在）
SET @column_exists = (SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE()
    AND table_name = 'error_records'
    AND column_name = 'path');
SET @sql = IF(@column_exists = 0,
    'ALTER TABLE `error_records` ADD COLUMN `path` VARCHAR(255) NOT NULL COMMENT ''接口路径'' AFTER `app_type`',
    'SELECT ''Column path already exists''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 2. 修改 error_message 字段类型
ALTER TABLE `error_records` MODIFY COLUMN `error_message` VARCHAR(500) NOT NULL COMMENT '接口返回的错误消息信息';

-- 3. 修改 remark 字段类型
ALTER TABLE `error_records` MODIFY COLUMN `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注信息，可记录额外的上下文信息';

-- 4. 修改时间字段的默认值
ALTER TABLE `error_records` MODIFY COLUMN `created_at` DATETIME(3) NOT NULL DEFAULT 0 COMMENT '创建时间';
ALTER TABLE `error_records` MODIFY COLUMN `updated_at` DATETIME(3) NOT NULL DEFAULT 0 COMMENT '更新时间';
ALTER TABLE `error_records` MODIFY COLUMN `deleted_at` DATETIME(3) NOT NULL DEFAULT 0 COMMENT '软删除时间，不为空表示已删除';

-- 5. 安全删除索引（如果存在）
-- 删除 idx_request_id
SET @count = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'error_records'
    AND index_name = 'idx_request_id');
SET @sql = IF(@count > 0,
    'ALTER TABLE `error_records` DROP INDEX `idx_request_id`',
    'SELECT ''Index idx_request_id does not exist''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 删除 idx_app_type
SET @count = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'error_records'
    AND index_name = 'idx_app_type');
SET @sql = IF(@count > 0,
    'ALTER TABLE `error_records` DROP INDEX `idx_app_type`',
    'SELECT ''Index idx_app_type does not exist''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 删除 idx_app_type_created_at
SET @count = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'error_records'
    AND index_name = 'idx_app_type_created_at');
SET @sql = IF(@count > 0,
    'ALTER TABLE `error_records` DROP INDEX `idx_app_type_created_at`',
    'SELECT ''Index idx_app_type_created_at does not exist''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 删除 idx_error_type
SET @count = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'error_records'
    AND index_name = 'idx_error_type');
SET @sql = IF(@count > 0,
    'ALTER TABLE `error_records` DROP INDEX `idx_error_type`',
    'SELECT ''Index idx_error_type does not exist''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 删除 idx_error_type_created_at
SET @count = (SELECT COUNT(*) FROM information_schema.statistics
    WHERE table_schema = DATABASE()
    AND table_name = 'error_records'
    AND index_name = 'idx_error_type_created_at');
SET @sql = IF(@count > 0,
    'ALTER TABLE `error_records` DROP INDEX `idx_error_type_created_at`',
    'SELECT ''Index idx_error_type_created_at does not exist''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;