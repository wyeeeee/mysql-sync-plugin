package repository

import (
	"database/sql"
	"fmt"
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/models"
	"strings"
	"sync"
	"time"
)

// SQLiteRepository SQLite数据访问实现
type SQLiteRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewSQLiteRepository 创建SQLite Repository实例
func NewSQLiteRepository(db *sql.DB) Repository {
	return &SQLiteRepository{
		db: db,
	}
}

// ==================== 用户管理 ====================

// CreateUser 创建用户
func (r *SQLiteRepository) CreateUser(user *auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.Exec(
		"INSERT INTO users (username, password, role, display_name, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user.Username, user.Password, user.Role, user.DisplayName, user.Status, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	id, _ := result.LastInsertId()
	user.ID = id
	return nil
}

// GetUserByID 根据ID获取用户
func (r *SQLiteRepository) GetUserByID(id int64) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var user auth.User
	err := r.db.QueryRow(
		"SELECT id, username, password, role, display_name, status, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *SQLiteRepository) GetUserByUsername(username string) (*auth.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var user auth.User
	err := r.db.QueryRow(
		"SELECT id, username, password, role, display_name, status, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	return &user, nil
}

// ListUsers 获取用户列表(分页)
func (r *SQLiteRepository) ListUsers(query *models.UserQuery) ([]*auth.User, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if query.Role != "" {
		conditions = append(conditions, "role = ?")
		args = append(args, query.Role)
	}
	if query.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, query.Status)
	}
	if query.Keyword != "" {
		conditions = append(conditions, "(username LIKE ? OR display_name LIKE ?)")
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT id, username, password, role, display_name, status, created_at, updated_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, query.PageSize, offset)
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		var user auth.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.DisplayName, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描用户数据失败: %w", err)
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
func (r *SQLiteRepository) UpdateUser(user *auth.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		"UPDATE users SET role = ?, display_name = ?, status = ?, updated_at = ? WHERE id = ?",
		user.Role, user.DisplayName, user.Status, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (r *SQLiteRepository) DeleteUser(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// UpdateUserStatus 更新用户状态
func (r *SQLiteRepository) UpdateUserStatus(id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"UPDATE users SET status = ?, updated_at = ? WHERE id = ?",
		status, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	return nil
}

// UpdateUserPassword 更新用户密码
func (r *SQLiteRepository) UpdateUserPassword(id int64, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"UPDATE users SET password = ?, updated_at = ? WHERE id = ?",
		hashedPassword, time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("更新用户密码失败: %w", err)
	}

	return nil
}

// ==================== 数据源管理 ====================

// CreateDatasource 创建数据源
func (r *SQLiteRepository) CreateDatasource(ds *models.Datasource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.Exec(
		"INSERT INTO datasources (name, description, host, port, database_name, username, password, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ds.Name, ds.Description, ds.Host, ds.Port, ds.DatabaseName, ds.Username, ds.Password, ds.CreatedBy, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("创建数据源失败: %w", err)
	}

	id, _ := result.LastInsertId()
	ds.ID = id
	return nil
}

// GetDatasourceByID 根据ID获取数据源
func (r *SQLiteRepository) GetDatasourceByID(id int64) (*models.Datasource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var ds models.Datasource
	err := r.db.QueryRow(
		"SELECT id, name, description, host, port, database_name, username, password, created_by, created_at, updated_at FROM datasources WHERE id = ?",
		id,
	).Scan(&ds.ID, &ds.Name, &ds.Description, &ds.Host, &ds.Port, &ds.DatabaseName, &ds.Username, &ds.Password, &ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询数据源失败: %w", err)
	}

	return &ds, nil
}

// ListDatasources 获取数据源列表(分页)
func (r *SQLiteRepository) ListDatasources(query *models.DatasourceQuery) ([]*models.Datasource, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 构建查询条件
	var conditions []string
	var args []interface{}

	if query.Keyword != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM datasources %s", whereClause)
	var total int64
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询数据源总数失败: %w", err)
	}

	// 分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	offset := (query.Page - 1) * query.PageSize

	// 查询数据
	dataQuery := fmt.Sprintf(`
		SELECT id, name, description, host, port, database_name, username, password, created_by, created_at, updated_at
		FROM datasources %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, query.PageSize, offset)
	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询数据源列表失败: %w", err)
	}
	defer rows.Close()

	var datasources []*models.Datasource
	for rows.Next() {
		var ds models.Datasource
		if err := rows.Scan(&ds.ID, &ds.Name, &ds.Description, &ds.Host, &ds.Port, &ds.DatabaseName, &ds.Username, &ds.Password, &ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描数据源数据失败: %w", err)
		}
		datasources = append(datasources, &ds)
	}

	return datasources, total, nil
}

// UpdateDatasource 更新数据源
func (r *SQLiteRepository) UpdateDatasource(ds *models.Datasource) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ds.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		"UPDATE datasources SET name = ?, description = ?, host = ?, port = ?, database_name = ?, username = ?, password = ?, updated_at = ? WHERE id = ?",
		ds.Name, ds.Description, ds.Host, ds.Port, ds.DatabaseName, ds.Username, ds.Password, ds.UpdatedAt, ds.ID,
	)
	if err != nil {
		return fmt.Errorf("更新数据源失败: %w", err)
	}

	return nil
}

// DeleteDatasource 删除数据源
func (r *SQLiteRepository) DeleteDatasource(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec("DELETE FROM datasources WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除数据源失败: %w", err)
	}

	return nil
}

// ==================== 数据源表管理 ====================

// CreateDatasourceTable 创建数据源表配置
func (r *SQLiteRepository) CreateDatasourceTable(table *models.DatasourceTable) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.Exec(
		"INSERT INTO datasource_tables (datasource_id, table_name, table_alias, query_mode, custom_sql, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		table.DatasourceID, table.TableName, table.TableAlias, table.QueryMode, table.CustomSQL, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("创建数据源表配置失败: %w", err)
	}

	id, _ := result.LastInsertId()
	table.ID = id
	return nil
}

// GetDatasourceTableByID 根据ID获取数据源表配置
func (r *SQLiteRepository) GetDatasourceTableByID(id int64) (*models.DatasourceTable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var table models.DatasourceTable
	err := r.db.QueryRow(
		"SELECT id, datasource_id, table_name, table_alias, query_mode, custom_sql, created_at, updated_at FROM datasource_tables WHERE id = ?",
		id,
	).Scan(&table.ID, &table.DatasourceID, &table.TableName, &table.TableAlias, &table.QueryMode, &table.CustomSQL, &table.CreatedAt, &table.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询数据源表配置失败: %w", err)
	}

	return &table, nil
}

// ListDatasourceTables 获取数据源的所有表配置
func (r *SQLiteRepository) ListDatasourceTables(datasourceID int64) ([]*models.DatasourceTable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query(
		"SELECT id, datasource_id, table_name, table_alias, query_mode, custom_sql, created_at, updated_at FROM datasource_tables WHERE datasource_id = ? ORDER BY created_at DESC",
		datasourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询数据源表配置列表失败: %w", err)
	}
	defer rows.Close()

	var tables []*models.DatasourceTable
	for rows.Next() {
		var table models.DatasourceTable
		if err := rows.Scan(&table.ID, &table.DatasourceID, &table.TableName, &table.TableAlias, &table.QueryMode, &table.CustomSQL, &table.CreatedAt, &table.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描数据源表配置数据失败: %w", err)
		}
		tables = append(tables, &table)
	}

	return tables, nil
}

// UpdateDatasourceTable 更新数据源表配置
func (r *SQLiteRepository) UpdateDatasourceTable(table *models.DatasourceTable) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	table.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		"UPDATE datasource_tables SET table_alias = ?, query_mode = ?, custom_sql = ?, updated_at = ? WHERE id = ?",
		table.TableAlias, table.QueryMode, table.CustomSQL, table.UpdatedAt, table.ID,
	)
	if err != nil {
		return fmt.Errorf("更新数据源表配置失败: %w", err)
	}

	return nil
}

// DeleteDatasourceTable 删除数据源表配置
func (r *SQLiteRepository) DeleteDatasourceTable(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec("DELETE FROM datasource_tables WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("删除数据源表配置失败: %w", err)
	}

	return nil
}

// ==================== 字段映射管理 ====================

// BatchCreateFieldMappings 批量创建字段映射
func (r *SQLiteRepository) BatchCreateFieldMappings(tableID int64, mappings []*models.DatasourceFieldMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 先删除旧的映射
	if _, err := r.db.Exec("DELETE FROM field_mappings WHERE datasource_table_id = ?", tableID); err != nil {
		return fmt.Errorf("删除旧字段映射失败: %w", err)
	}

	// 批量插入新映射
	for _, mapping := range mappings {
		enabled := 1
		if !mapping.Enabled {
			enabled = 0
		}
		_, err := r.db.Exec(
			"INSERT INTO field_mappings (datasource_table_id, field_name, field_alias, enabled, created_at) VALUES (?, ?, ?, ?, ?)",
			tableID, mapping.FieldName, mapping.FieldAlias, enabled, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("创建字段映射失败: %w", err)
		}
	}

	return nil
}

// ListFieldMappings 获取表的所有字段映射
func (r *SQLiteRepository) ListFieldMappings(tableID int64) ([]*models.DatasourceFieldMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query(
		"SELECT id, datasource_table_id, field_name, field_alias, enabled, created_at FROM field_mappings WHERE datasource_table_id = ?",
		tableID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询字段映射列表失败: %w", err)
	}
	defer rows.Close()

	var mappings []*models.DatasourceFieldMapping
	for rows.Next() {
		var mapping models.DatasourceFieldMapping
		var enabled int
		if err := rows.Scan(&mapping.ID, &mapping.DatasourceTableID, &mapping.FieldName, &mapping.FieldAlias, &enabled, &mapping.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描字段映射数据失败: %w", err)
		}
		mapping.Enabled = enabled == 1
		mappings = append(mappings, &mapping)
	}

	return mappings, nil
}

// DeleteFieldMappingsByTableID 删除表的所有字段映射
func (r *SQLiteRepository) DeleteFieldMappingsByTableID(tableID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec("DELETE FROM field_mappings WHERE datasource_table_id = ?", tableID)
	if err != nil {
		return fmt.Errorf("删除字段映射失败: %w", err)
	}

	return nil
}

// ==================== 权限管理 ====================

// GrantDatasourcePermission 授予用户数据源权限
func (r *SQLiteRepository) GrantDatasourcePermission(userID, datasourceID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"INSERT IGNORE INTO user_datasource_permissions (user_id, datasource_id, created_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE created_at=created_at",
		userID, datasourceID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("授予数据源权限失败: %w", err)
	}

	return nil
}

// RevokeDatasourcePermission 撤销用户数据源权限
func (r *SQLiteRepository) RevokeDatasourcePermission(userID, datasourceID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"DELETE FROM user_datasource_permissions WHERE user_id = ? AND datasource_id = ?",
		userID, datasourceID,
	)
	if err != nil {
		return fmt.Errorf("撤销数据源权限失败: %w", err)
	}

	return nil
}

// ListUserDatasources 获取用户可访问的数据源列表
func (r *SQLiteRepository) ListUserDatasources(userID int64) ([]*models.Datasource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query(`
		SELECT d.id, d.name, d.description, d.host, d.port, d.database_name, d.username, d.password, d.created_by, d.created_at, d.updated_at
		FROM datasources d
		INNER JOIN user_datasource_permissions p ON d.id = p.datasource_id
		WHERE p.user_id = ?
		ORDER BY d.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户数据源列表失败: %w", err)
	}
	defer rows.Close()

	var datasources []*models.Datasource
	for rows.Next() {
		var ds models.Datasource
		if err := rows.Scan(&ds.ID, &ds.Name, &ds.Description, &ds.Host, &ds.Port, &ds.DatabaseName, &ds.Username, &ds.Password, &ds.CreatedBy, &ds.CreatedAt, &ds.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描数据源数据失败: %w", err)
		}
		datasources = append(datasources, &ds)
	}

	return datasources, nil
}

// CheckDatasourcePermission 检查用户是否有数据源权限
func (r *SQLiteRepository) CheckDatasourcePermission(userID, datasourceID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM user_datasource_permissions WHERE user_id = ? AND datasource_id = ?",
		userID, datasourceID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查数据源权限失败: %w", err)
	}

	return count > 0, nil
}

// GrantTablePermission 授予用户表权限
func (r *SQLiteRepository) GrantTablePermission(userID, tableID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"INSERT IGNORE INTO user_table_permissions (user_id, datasource_table_id, created_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE created_at=created_at",
		userID, tableID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("授予表权限失败: %w", err)
	}

	return nil
}

// RevokeTablePermission 撤销用户表权限
func (r *SQLiteRepository) RevokeTablePermission(userID, tableID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, err := r.db.Exec(
		"DELETE FROM user_table_permissions WHERE user_id = ? AND datasource_table_id = ?",
		userID, tableID,
	)
	if err != nil {
		return fmt.Errorf("撤销表权限失败: %w", err)
	}

	return nil
}

// ListUserTables 获取用户在指定数据源下可访问的表列表
func (r *SQLiteRepository) ListUserTables(userID, datasourceID int64) ([]*models.DatasourceTable, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rows, err := r.db.Query(`
		SELECT t.id, t.datasource_id, t.table_name, t.table_alias, t.query_mode, t.custom_sql, t.created_at, t.updated_at
		FROM datasource_tables t
		INNER JOIN user_table_permissions p ON t.id = p.datasource_table_id
		WHERE p.user_id = ? AND t.datasource_id = ?
		ORDER BY t.created_at DESC
	`, userID, datasourceID)
	if err != nil {
		return nil, fmt.Errorf("查询用户表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []*models.DatasourceTable
	for rows.Next() {
		var table models.DatasourceTable
		if err := rows.Scan(&table.ID, &table.DatasourceID, &table.TableName, &table.TableAlias, &table.QueryMode, &table.CustomSQL, &table.CreatedAt, &table.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描表数据失败: %w", err)
		}
		tables = append(tables, &table)
	}

	return tables, nil
}

// CheckTablePermission 检查用户是否有表权限
func (r *SQLiteRepository) CheckTablePermission(userID, tableID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM user_table_permissions WHERE user_id = ? AND datasource_table_id = ?",
		userID, tableID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查表权限失败: %w", err)
	}

	return count > 0, nil
}

// BatchGrantDatasourcePermissions 批量授予数据源权限
func (r *SQLiteRepository) BatchGrantDatasourcePermissions(userID int64, datasourceIDs []int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, dsID := range datasourceIDs {
		_, err := r.db.Exec(
			"INSERT IGNORE INTO user_datasource_permissions (user_id, datasource_id, created_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE created_at=created_at",
			userID, dsID, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("批量授予数据源权限失败: %w", err)
		}
	}

	return nil
}

// BatchRevokeDatasourcePermissions 批量撤销数据源权限
func (r *SQLiteRepository) BatchRevokeDatasourcePermissions(userID int64, datasourceIDs []int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, dsID := range datasourceIDs {
		_, err := r.db.Exec(
			"DELETE FROM user_datasource_permissions WHERE user_id = ? AND datasource_id = ?",
			userID, dsID,
		)
		if err != nil {
			return fmt.Errorf("批量撤销数据源权限失败: %w", err)
		}
	}

	return nil
}

// BatchGrantTablePermissions 批量授予表权限
func (r *SQLiteRepository) BatchGrantTablePermissions(userID int64, tableIDs []int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, tableID := range tableIDs {
		_, err := r.db.Exec(
			"INSERT IGNORE INTO user_table_permissions (user_id, datasource_table_id, created_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE created_at=created_at",
			userID, tableID, time.Now(),
		)
		if err != nil {
			return fmt.Errorf("批量授予表权限失败: %w", err)
		}
	}

	return nil
}

// BatchRevokeTablePermissions 批量撤销表权限
func (r *SQLiteRepository) BatchRevokeTablePermissions(userID int64, tableIDs []int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, tableID := range tableIDs {
		_, err := r.db.Exec(
			"DELETE FROM user_table_permissions WHERE user_id = ? AND datasource_table_id = ?",
			userID, tableID,
		)
		if err != nil {
			return fmt.Errorf("批量撤销表权限失败: %w", err)
		}
	}

	return nil
}
