-- MySQL 数据库 Schema
-- 用于 mysql-sync-plugin 项目

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS mysql_sync_plugin DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE mysql_sync_plugin;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' COMMENT '用户角色: admin 或 user',
    display_name VARCHAR(255) COMMENT '显示名称',
    status VARCHAR(50) NOT NULL DEFAULT 'active' COMMENT '用户状态: active 或 disabled',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_role (role),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 会话表
CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token (token),
    INDEX idx_expires_at (expires_at),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- 数据源表
CREATE TABLE IF NOT EXISTS datasources (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '数据源名称',
    description TEXT COMMENT '数据源描述',
    host VARCHAR(255) NOT NULL COMMENT 'MySQL主机',
    port INT NOT NULL COMMENT 'MySQL端口',
    database_name VARCHAR(255) NOT NULL COMMENT '数据库名',
    username VARCHAR(255) NOT NULL COMMENT 'MySQL用户名',
    password TEXT NOT NULL COMMENT 'MySQL密码(加密存储)',
    created_by BIGINT NOT NULL COMMENT '创建者ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    INDEX idx_created_by (created_by),
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据源配置表';

-- 数据源表配置
CREATE TABLE IF NOT EXISTS datasource_tables (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    datasource_id BIGINT NOT NULL COMMENT '数据源ID',
    table_name VARCHAR(255) NOT NULL COMMENT '表名',
    table_alias VARCHAR(255) COMMENT '表别名',
    query_mode VARCHAR(50) NOT NULL DEFAULT 'table' COMMENT '查询模式: table 或 sql',
    custom_sql TEXT COMMENT '自定义SQL语句',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE,
    UNIQUE KEY uk_datasource_table (datasource_id, table_name),
    INDEX idx_datasource_id (datasource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据源表配置';

-- 字段映射表
CREATE TABLE IF NOT EXISTS field_mappings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    datasource_table_id BIGINT NOT NULL COMMENT '数据源表ID',
    field_name VARCHAR(255) NOT NULL COMMENT '原始字段名',
    field_alias VARCHAR(255) NOT NULL COMMENT '字段别名',
    enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用: 1启用 0禁用',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (datasource_table_id) REFERENCES datasource_tables(id) ON DELETE CASCADE,
    UNIQUE KEY uk_table_field (datasource_table_id, field_name),
    INDEX idx_datasource_table_id (datasource_table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字段映射表';

-- 数据迁移: 为已有记录添加 enabled 字段默认值
-- ALTER TABLE field_mappings ADD COLUMN enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用: 1启用 0禁用' AFTER field_alias;

-- 用户数据源权限表
CREATE TABLE IF NOT EXISTS user_datasource_permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    datasource_id BIGINT NOT NULL COMMENT '数据源ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (datasource_id) REFERENCES datasources(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_datasource (user_id, datasource_id),
    INDEX idx_user_id (user_id),
    INDEX idx_datasource_id (datasource_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户数据源权限表';

-- 用户表权限表
CREATE TABLE IF NOT EXISTS user_table_permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    datasource_table_id BIGINT NOT NULL COMMENT '数据源表ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (datasource_table_id) REFERENCES datasource_tables(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_table (user_id, datasource_table_id),
    INDEX idx_user_id (user_id),
    INDEX idx_datasource_table_id (datasource_table_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表权限表';

-- 日志表
CREATE TABLE IF NOT EXISTS logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    level VARCHAR(50) NOT NULL COMMENT '日志级别',
    module VARCHAR(255) NOT NULL COMMENT '模块名称',
    action VARCHAR(255) NOT NULL COMMENT '操作名称',
    message TEXT COMMENT '日志消息',
    detail TEXT COMMENT '详细信息',
    ip VARCHAR(100) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    duration BIGINT COMMENT '执行时长(毫秒)',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_level (level),
    INDEX idx_module (module),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

-- 插入默认管理员账户
-- 密码: admin123 (SHA256哈希)
INSERT INTO users (username, password, role, display_name, status)
VALUES ('admin', '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9', 'admin', '系统管理员', 'active')
ON DUPLICATE KEY UPDATE username=username;
