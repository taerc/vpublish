-- ============================================
-- vpublish 数据库部署脚本
-- 低空智能平台软件包管理系统
-- ============================================

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS vpublish 
    CHARACTER SET utf8mb4 
    COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE vpublish;

-- ============================================
-- 用户与认证
-- ============================================

-- 管理员用户表
DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    nickname VARCHAR(50) COMMENT '昵称',
    email VARCHAR(100) COMMENT '邮箱',
    role ENUM('admin', 'user') DEFAULT 'user' COMMENT '角色',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    last_login_ip VARCHAR(45) COMMENT '最后登录IP',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_username (username),
    INDEX idx_active (is_active),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员用户表';

-- APP密钥表
DROP TABLE IF EXISTS app_keys;
CREATE TABLE app_keys (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '密钥ID',
    app_name VARCHAR(100) NOT NULL COMMENT '应用名称',
    app_key VARCHAR(64) UNIQUE NOT NULL COMMENT '应用Key',
    app_secret VARCHAR(64) NOT NULL COMMENT '应用Secret',
    description VARCHAR(200) COMMENT '描述',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_app_key (app_key),
    INDEX idx_active (is_active),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='APP认证密钥表';

-- ============================================
-- 软件管理
-- ============================================

-- 软件类别表
DROP TABLE IF EXISTS categories;
CREATE TABLE categories (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '类别ID',
    name VARCHAR(100) UNIQUE NOT NULL COMMENT '类别中文名称',
    code VARCHAR(100) UNIQUE NOT NULL COMMENT '类别代码枚举，如 TYPE_WU_REN_JI',
    description VARCHAR(500) COMMENT '类别描述',
    sort_order INT DEFAULT 0 COMMENT '排序序号',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_code (code),
    INDEX idx_active (is_active),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='软件类别表';

-- 软件包表
DROP TABLE IF EXISTS packages;
CREATE TABLE packages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '软件包ID',
    category_id BIGINT NOT NULL COMMENT '类别ID',
    name VARCHAR(200) NOT NULL COMMENT '软件包名称',
    description TEXT COMMENT '软件包描述',
    icon VARCHAR(255) COMMENT '图标URL',
    developer VARCHAR(100) COMMENT '开发者',
    website VARCHAR(255) COMMENT '官网地址',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_by BIGINT NOT NULL COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_category (category_id),
    INDEX idx_name (name),
    INDEX idx_active (is_active),
    INDEX idx_deleted_at (deleted_at),
    CONSTRAINT fk_package_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT,
    CONSTRAINT fk_package_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='软件包表';

-- 版本表
DROP TABLE IF EXISTS versions;
CREATE TABLE versions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '版本ID',
    package_id BIGINT NOT NULL COMMENT '软件包ID',
    version VARCHAR(50) NOT NULL COMMENT '版本号，如 1.0.0',
    version_code INT NOT NULL COMMENT '版本号数值，用于比较大小',
    
    -- 文件信息
    file_path VARCHAR(500) NOT NULL COMMENT '文件存储路径',
    file_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    file_size BIGINT NOT NULL COMMENT '文件大小(字节)',
    file_hash VARCHAR(64) NOT NULL COMMENT '文件SHA256哈希',
    
    -- 版本信息
    changelog TEXT COMMENT '更新日志',
    release_notes TEXT COMMENT '发布说明',
    min_version VARCHAR(50) COMMENT '最低兼容版本',
    
    -- 升级控制
    force_upgrade BOOLEAN DEFAULT FALSE COMMENT '是否强制升级',
    is_latest BOOLEAN DEFAULT FALSE COMMENT '是否最新版本',
    is_stable BOOLEAN DEFAULT TRUE COMMENT '是否稳定版',
    
    -- 统计
    download_count INT DEFAULT 0 COMMENT '下载次数',
    
    -- 审计
    created_by BIGINT NOT NULL COMMENT '发布人ID',
    published_at TIMESTAMP NULL COMMENT '发布时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    
    UNIQUE KEY uk_package_version (package_id, version),
    INDEX idx_package_latest (package_id, is_latest),
    INDEX idx_version_code (version_code),
    INDEX idx_deleted_at (deleted_at),
    CONSTRAINT fk_version_package FOREIGN KEY (package_id) REFERENCES packages(id) ON DELETE CASCADE,
    CONSTRAINT fk_version_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='软件版本表';

-- ============================================
-- 统计分析
-- ============================================

-- 下载记录表（原始日志）
DROP TABLE IF EXISTS download_logs;
CREATE TABLE download_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    version_id BIGINT NOT NULL COMMENT '版本ID',
    app_key VARCHAR(64) COMMENT 'APP Key',
    client_ip VARCHAR(45) COMMENT '客户端IP',
    user_agent VARCHAR(500) COMMENT 'User-Agent',
    downloaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '下载时间',
    INDEX idx_version_time (version_id, downloaded_at),
    INDEX idx_time (downloaded_at),
    CONSTRAINT fk_log_version FOREIGN KEY (version_id) REFERENCES versions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='下载记录日志表';

-- 下载统计表（按天聚合）
DROP TABLE IF EXISTS download_stats;
CREATE TABLE download_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '统计ID',
    version_id BIGINT NOT NULL COMMENT '版本ID',
    category_id BIGINT NOT NULL COMMENT '类别ID（冗余，方便统计）',
    stat_date DATE NOT NULL COMMENT '统计日期',
    download_count INT DEFAULT 0 COMMENT '下载次数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_version_date (version_id, stat_date),
    INDEX idx_category_date (category_id, stat_date),
    INDEX idx_date (stat_date),
    CONSTRAINT fk_stat_version FOREIGN KEY (version_id) REFERENCES versions(id) ON DELETE CASCADE,
    CONSTRAINT fk_stat_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='下载统计汇总表';

-- ============================================
-- 操作日志
-- ============================================

-- 操作日志表
DROP TABLE IF EXISTS operation_logs;
CREATE TABLE operation_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    user_id BIGINT COMMENT '操作用户ID',
    action VARCHAR(50) NOT NULL COMMENT '操作类型',
    resource_type VARCHAR(50) COMMENT '资源类型',
    resource_id BIGINT COMMENT '资源ID',
    detail TEXT COMMENT '操作详情JSON',
    ip VARCHAR(45) COMMENT '操作IP',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user (user_id),
    INDEX idx_action (action),
    INDEX idx_time (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- ============================================
-- 初始数据
-- ============================================

-- 插入默认管理员账户
-- 密码: admin@123 (bcrypt哈希)
INSERT INTO users (username, password_hash, nickname, role, is_active) VALUES
('admin', '$2b$10$kyvUfABj.s2MWDOjvnCJs.F//EKmMgxKTswymQhK.988N2LsX0d/a', '系统管理员', 'admin', TRUE);

-- 插入示例APP密钥
-- AppKey: test_app_key
-- AppSecret: test_app_secret
INSERT INTO app_keys (app_name, app_key, app_secret, description, is_active) VALUES
('测试应用', 'test_app_key_12345678', 'test_app_secret_abcdefgh', '测试用的APP密钥', TRUE);

-- 插入示例软件类别
INSERT INTO categories (name, code, description, sort_order, is_active) VALUES
('无人机', 'TYPE_WU_REN_JI', '无人机相关软件', 1, TRUE),
('地面站', 'TYPE_DI_MIAN_ZHAN', '地面站控制软件', 2, TRUE),
('飞控系统', 'TYPE_FEI_KONG_XI_TONG', '飞行控制系统', 3, TRUE),
('数据链路', 'TYPE_SHU_JU_LIAN_LU', '数据链路通信软件', 4, TRUE);

-- ============================================
-- 完成提示
-- ============================================
SELECT '数据库部署完成!' AS message;
SELECT '默认管理员账户: admin / admin@123' AS admin_account;
SELECT '测试APP密钥: test_app_key_12345678' AS app_key;