-- ============================================
-- 报错记录表
-- 用于存储APP端和平台前端的接口报错信息
-- 优化时间：2026-03-31
-- 更新时间：2026-04-14
-- ============================================

-- 报错记录表
CREATE TABLE IF NOT EXISTS `error_records` (
    `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `request_id` VARCHAR(128) NOT NULL COMMENT '接口请求唯一标识，用于关联单次请求',
    `timestamp` BIGINT NOT NULL COMMENT '报错发生时间戳（毫秒）',
    `module` VARCHAR(64) DEFAULT NULL COMMENT '报错模块（自动识别），如: login, package, version等',
    `app_type` VARCHAR(32) NOT NULL COMMENT '应用类型：app（APP端）/ platform（平台前端）',
    `path` VARCHAR(255) NOT NULL COMMENT '接口路径',
    `code` VARCHAR(64) NOT NULL COMMENT '接口返回的错误码，如: 401, 403, 500等',
    `error_message` VARCHAR(500) NOT NULL COMMENT '接口返回的错误消息信息',
    `request_params` JSON DEFAULT NULL COMMENT '接口入参（JSON格式），便于问题复现和定位',
    `device_info` JSON DEFAULT NULL COMMENT '设备信息（JSON格式），包含应用版本号、设备型号、操作系统版本等',
    `error_type` VARCHAR(32) NOT NULL COMMENT '报错类型：system_error（系统错误）/ business_error（业务错误）',
    `remark` VARCHAR(500) DEFAULT NULL COMMENT '备注信息，可记录额外的上下文信息',
    `created_at` DATETIME(3) NOT NULL DEFAULT 0 COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT 0 COMMENT '更新时间',
    `deleted_at` DATETIME(3) NOT NULL DEFAULT 0 COMMENT '软删除时间，不为空表示已删除',
    PRIMARY KEY (`id`),
    KEY `idx_timestamp` (`timestamp`),
    KEY `idx_module` (`module`),
    KEY `idx_code` (`code`),
    KEY `idx_created_at` (`created_at`),
    KEY `idx_deleted_at` (`deleted_at`),
    KEY `idx_code_timestamp` (`code`, `timestamp`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报错记录表';
