-- MySQL 诊断脚本
-- 用于排查"数据源表配置不存在"错误

USE mysql_sync_plugin;

-- 1. 检查数据库连接
SELECT 'Database connection OK' AS status;

-- 2. 查看所有数据源
SELECT
    id,
    name,
    host,
    port,
    database_name,
    username,
    created_at
FROM datasources
ORDER BY id;

-- 3. 查看所有数据源表配置
SELECT
    dt.id AS table_id,
    dt.datasource_id,
    ds.name AS datasource_name,
    dt.table_name,
    dt.table_alias,
    dt.query_mode,
    dt.created_at
FROM datasource_tables dt
LEFT JOIN datasources ds ON dt.datasource_id = ds.id
ORDER BY dt.id;

-- 4. 查看字段映射数量
SELECT
    dt.id AS table_id,
    dt.table_name,
    COUNT(fm.id) AS field_count
FROM datasource_tables dt
LEFT JOIN field_mappings fm ON dt.id = fm.datasource_table_id
GROUP BY dt.id, dt.table_name
ORDER BY dt.id;

-- 5. 查看所有用户（不包含密码）
SELECT
    id,
    username,
    role,
    display_name,
    status,
    created_at
FROM users
ORDER BY id;

-- 6. 查看用户数据源权限
SELECT
    u.id AS user_id,
    u.username,
    ds.id AS datasource_id,
    ds.name AS datasource_name,
    udp.created_at AS granted_at
FROM user_datasource_permissions udp
JOIN users u ON udp.user_id = u.id
JOIN datasources ds ON udp.datasource_id = ds.id
ORDER BY u.id, ds.id;

-- 7. 查看用户表权限
SELECT
    u.id AS user_id,
    u.username,
    dt.id AS table_id,
    ds.name AS datasource_name,
    dt.table_name,
    utp.created_at AS granted_at
FROM user_table_permissions utp
JOIN users u ON utp.user_id = u.id
JOIN datasource_tables dt ON utp.datasource_table_id = dt.id
JOIN datasources ds ON dt.datasource_id = ds.id
ORDER BY u.id, dt.id;

-- 8. 检查是否有孤立的权限记录（表配置已删除但权限还在）
SELECT
    'Orphaned table permissions' AS issue,
    utp.id,
    utp.user_id,
    utp.datasource_table_id
FROM user_table_permissions utp
LEFT JOIN datasource_tables dt ON utp.datasource_table_id = dt.id
WHERE dt.id IS NULL;

-- 9. 检查是否有孤立的字段映射（表配置已删除但字段映射还在）
SELECT
    'Orphaned field mappings' AS issue,
    fm.id,
    fm.datasource_table_id,
    fm.field_name
FROM field_mappings fm
LEFT JOIN datasource_tables dt ON fm.datasource_table_id = dt.id
WHERE dt.id IS NULL;

-- 10. 统计信息
SELECT
    'Total datasources' AS metric,
    COUNT(*) AS count
FROM datasources
UNION ALL
SELECT
    'Total datasource tables' AS metric,
    COUNT(*) AS count
FROM datasource_tables
UNION ALL
SELECT
    'Total field mappings' AS metric,
    COUNT(*) AS count
FROM field_mappings
UNION ALL
SELECT
    'Total users' AS metric,
    COUNT(*) AS count
FROM users
UNION ALL
SELECT
    'Total datasource permissions' AS metric,
    COUNT(*) AS count
FROM user_datasource_permissions
UNION ALL
SELECT
    'Total table permissions' AS metric,
    COUNT(*) AS count
FROM user_table_permissions;
