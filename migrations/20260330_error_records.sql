-- ============================================
-- 报错记录表
-- 用于存储APP端和平台前端的接口报错信息
-- 优化时间：2026-03-31
-- 更新时间：2026-04-14
-- ============================================

-- 报错记录表
CREATE TABLE IF NOT EXISTS error_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    request_id VARCHAR(128) NOT NULL COMMENT '接口请求唯一标识',
    timestamp BIGINT NOT NULL COMMENT '报错发生时间戳（毫秒）',
    module VARCHAR(64) COMMENT '报错模块（自动识别）',
    app_type VARCHAR(32) NOT NULL COMMENT '应用类型：app/platform',
    code VARCHAR(64) NOT NULL COMMENT '接口返回的错误码',
    error_message TEXT NOT NULL COMMENT '接口返回的消息信息',
    request_params JSON COMMENT '接口入参（JSON格式）',
    device_info JSON COMMENT '设备信息（JSON格式，包含应用版本号、设备型号、操作系统版本等）',
    error_type VARCHAR(32) NOT NULL COMMENT '报错类型：system_error/business_error',
    remark TEXT COMMENT '备注',
    created_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at datetime(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    deleted_at datetime(3) NULL COMMENT '软删除时间',
    INDEX idx_request_id (request_id),
    INDEX idx_timestamp (timestamp),
    INDEX idx_module (module),
    INDEX idx_app_type (app_type),
    INDEX idx_code (code),
    INDEX idx_error_type (error_type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='报错记录表';
