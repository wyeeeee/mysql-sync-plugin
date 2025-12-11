package service

import (
	"database/sql"
	"fmt"
	"strings"
)

// FieldInfo 字段信息
type FieldInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Comment      string `json:"comment"`
	IsPrimary    bool   `json:"isPrimary"`
	IsNullable   bool   `json:"isNullable"`
	DefaultValue string `json:"defaultValue"`
}

// GetDatabaseList 获取数据源的数据库列表
func (s *DatasourceService) GetDatabaseList(datasourceID int64) ([]string, error) {
	// 获取数据源(包含密码)
	ds, err := s.GetDatasourceByIDWithPassword(datasourceID)
	if err != nil {
		return nil, err
	}

	// 构建DSN(不指定数据库)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		ds.Username,
		ds.Password,
		ds.Host,
		ds.Port,
	)

	// 连接MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()

	// 查询数据库列表
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("查询数据库列表失败: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, fmt.Errorf("读取数据库名失败: %w", err)
		}
		// 过滤系统数据库
		if dbName != "information_schema" && dbName != "mysql" && dbName != "performance_schema" && dbName != "sys" {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// GetTableList 获取数据源指定数据库的表列表
func (s *DatasourceService) GetTableList(datasourceID int64, databaseName string) ([]string, error) {
	// 获取数据源(包含密码)
	ds, err := s.GetDatasourceByIDWithPassword(datasourceID)
	if err != nil {
		return nil, err
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.Username,
		ds.Password,
		ds.Host,
		ds.Port,
		databaseName,
	)

	// 连接MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()

	// 查询表列表
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("读取表名失败: %w", err)
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// GetFieldList 获取数据源指定表的字段列表
func (s *DatasourceService) GetFieldList(datasourceID int64, databaseName, tableName string) ([]FieldInfo, error) {
	// 获取数据源(包含密码)
	ds, err := s.GetDatasourceByIDWithPassword(datasourceID)
	if err != nil {
		return nil, err
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.Username,
		ds.Password,
		ds.Host,
		ds.Port,
		databaseName,
	)

	// 连接MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()

	// 查询字段信息
	query := `
		SELECT
			COLUMN_NAME,
			COLUMN_TYPE,
			COLUMN_COMMENT,
			COLUMN_KEY,
			IS_NULLABLE,
			COLUMN_DEFAULT
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`
	rows, err := db.Query(query, databaseName, tableName)
	if err != nil {
		return nil, fmt.Errorf("查询字段列表失败: %w", err)
	}
	defer rows.Close()

	var fields []FieldInfo
	for rows.Next() {
		var field FieldInfo
		var columnKey, isNullable string
		var defaultValue sql.NullString

		if err := rows.Scan(&field.Name, &field.Type, &field.Comment, &columnKey, &isNullable, &defaultValue); err != nil {
			return nil, fmt.Errorf("读取字段信息失败: %w", err)
		}

		field.IsPrimary = columnKey == "PRI"
		field.IsNullable = isNullable == "YES"
		if defaultValue.Valid {
			field.DefaultValue = defaultValue.String
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// GetFieldListFromSQL 从自定义SQL获取字段列表
func (s *DatasourceService) GetFieldListFromSQL(datasourceID int64, databaseName, customSQL string) ([]FieldInfo, error) {
	// 获取数据源(包含密码)
	ds, err := s.GetDatasourceByIDWithPassword(datasourceID)
	if err != nil {
		return nil, err
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.Username,
		ds.Password,
		ds.Host,
		ds.Port,
		databaseName,
	)

	// 连接MySQL
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()

	// 清理SQL:移除末尾的分号和空白字符
	cleanSQL := strings.TrimSpace(customSQL)
	cleanSQL = strings.TrimSuffix(cleanSQL, ";")

	// 执行SQL查询(LIMIT 0只获取结构,不获取数据)
	query := fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT 0", cleanSQL)
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("执行SQL失败: %w", err)
	}
	defer rows.Close()

	// 获取列信息
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %w", err)
	}

	// 构建字段列表
	var fields []FieldInfo
	for _, ct := range columnTypes {
		field := FieldInfo{
			Name:    ct.Name(),
			Type:    ct.DatabaseTypeName(),
			Comment: "", // SQL查询无法获取备注
		}

		// 判断是否可为空
		nullable, ok := ct.Nullable()
		if ok {
			field.IsNullable = nullable
		}

		fields = append(fields, field)
	}

	return fields, nil
}
