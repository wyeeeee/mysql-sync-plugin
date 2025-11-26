package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mysql-sync-plugin/models"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLService MySQL数据源服务
type MySQLService struct{}

// NewMySQLService 创建MySQL服务实例
func NewMySQLService() *MySQLService {
	return &MySQLService{}
}

// GetDatabases 获取数据库列表
func (s *MySQLService) GetDatabases(config *models.MySQLConfig) ([]string, error) {
	// 连接MySQL(不指定数据库)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接测试失败: %w", err)
	}

	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, err
		}
		// 过滤系统数据库
		if dbName != "information_schema" && dbName != "mysql" && dbName != "performance_schema" && dbName != "sys" {
			databases = append(databases, dbName)
		}
	}

	return databases, nil
}

// GetTables 获取数据表列表
func (s *MySQLService) GetTables(config *models.MySQLConfig) ([]string, error) {
	db, err := s.connectDB(config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// GetFields 获取表字段信息
func (s *MySQLService) GetFields(config *models.MySQLConfig) ([]models.Field, error) {
	db, err := s.connectDB(config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return s.getTableSchema(db, config.Database, config.Table)
}

// GetSheetMeta 获取表结构
func (s *MySQLService) GetSheetMeta(req *models.SheetMetaRequest) (*models.SheetMetaResponse, error) {
	// 解析MySQL配置
	var config models.MySQLConfig
	if err := json.Unmarshal([]byte(req.Params), &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 连接数据库
	db, err := s.connectDB(&config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 获取表结构
	fields, err := s.getTableSchema(db, config.Database, config.Table)
	if err != nil {
		return nil, err
	}

	return &models.SheetMetaResponse{
		SheetName: config.Table,
		Fields:    fields,
	}, nil
}

// GetRecords 获取表记录(分页)
func (s *MySQLService) GetRecords(req *models.RecordsRequest) (*models.RecordsResponse, error) {
	// 解析MySQL配置
	var config models.MySQLConfig
	if err := json.Unmarshal([]byte(req.Params), &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 连接数据库
	db, err := s.connectDB(&config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// 解析分页参数
	offset := 0
	if req.NextToken != "" {
		// nextToken格式: "offset:300"
		parts := strings.Split(req.NextToken, ":")
		if len(parts) == 2 {
			offset, _ = strconv.Atoi(parts[1])
		}
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 300
	}

	// 获取总记录数
	total, err := s.getRecordCount(db, config.Table)
	if err != nil {
		return nil, err
	}

	// 获取表结构(用于字段映射)
	fields, err := s.getTableSchema(db, config.Database, config.Table)
	if err != nil {
		return nil, err
	}

	// 获取记录数据
	records, err := s.getTableRecords(db, config.Table, fields, offset, maxResults)
	if err != nil {
		return nil, err
	}

	// 计算下一页token
	nextOffset := offset + maxResults
	hasMore := nextOffset < total
	nextToken := ""
	if hasMore {
		nextToken = fmt.Sprintf("offset:%d", nextOffset)
	}

	return &models.RecordsResponse{
		NextToken: nextToken,
		HasMore:   hasMore,
		Records:   records,
		Total:     total,
	}, nil
}

// connectDB 连接MySQL数据库
func (s *MySQLService) connectDB(config *models.MySQLConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return db, nil
}

// getTableSchema 获取表结构
func (s *MySQLService) getTableSchema(db *sql.DB, database, table string) ([]models.Field, error) {
	query := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			COLUMN_KEY,
			COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query, database, table)
	if err != nil {
		return nil, fmt.Errorf("查询表结构失败: %w", err)
	}
	defer rows.Close()

	var fields []models.Field
	for rows.Next() {
		var columnName, dataType, columnKey, columnComment string
		if err := rows.Scan(&columnName, &dataType, &columnKey, &columnComment); err != nil {
			return nil, err
		}

		field := models.Field{
			ID:          fmt.Sprintf("fid_%s", columnName),
			Name:        columnName,
			Type:        s.mapMySQLTypeToAITable(dataType),
			IsPrimary:   columnKey == "PRI",
			Description: columnComment,
		}

		// 设置字段属性
		if field.Type == "number" {
			field.Property = s.getNumberProperty(dataType)
		}

		fields = append(fields, field)
	}

	return fields, rows.Err()
}

// getRecordCount 获取记录总数
func (s *MySQLService) getRecordCount(db *sql.DB, table string) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", table)
	err := db.QueryRow(query).Scan(&count)
	return count, err
}

// getTableRecords 获取表记录
func (s *MySQLService) getTableRecords(db *sql.DB, table string, fields []models.Field, offset, limit int) ([]models.Record, error) {
	// 构建字段列表
	var columnNames []string
	for _, field := range fields {
		// 从字段ID中提取列名 (fid_xxx -> xxx)
		columnName := strings.TrimPrefix(field.ID, "fid_")
		columnNames = append(columnNames, fmt.Sprintf("`%s`", columnName))
	}

	query := fmt.Sprintf("SELECT %s FROM `%s` LIMIT ? OFFSET ?",
		strings.Join(columnNames, ", "),
		table,
	)

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("查询记录失败: %w", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var records []models.Record
	for rows.Next() {
		// 创建扫描目标
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// 构建记录
		record := models.Record{
			Fields: make(map[string]interface{}),
		}

		for i, col := range columns {
			fieldID := fmt.Sprintf("fid_%s", col)
			value := values[i]

			// 处理NULL值和类型转换
			if value != nil {
				switch v := value.(type) {
				case []byte:
					record.Fields[fieldID] = string(v)
				default:
					record.Fields[fieldID] = v
				}
			} else {
				record.Fields[fieldID] = nil
			}

			// 第一个字段作为记录ID(通常是主键)
			if i == 0 && value != nil {
				record.ID = fmt.Sprintf("%v", value)
			}
		}

		records = append(records, record)
	}

	return records, rows.Err()
}

// mapMySQLTypeToAITable 映射MySQL类型到AI表格类型
func (s *MySQLService) mapMySQLTypeToAITable(mysqlType string) string {
	mysqlType = strings.ToLower(mysqlType)

	// 数字类型
	if strings.Contains(mysqlType, "int") ||
		strings.Contains(mysqlType, "decimal") ||
		strings.Contains(mysqlType, "float") ||
		strings.Contains(mysqlType, "double") {
		return "number"
	}

	// 日期类型
	if strings.Contains(mysqlType, "date") ||
		strings.Contains(mysqlType, "time") {
		return "date"
	}

	// 默认为文本
	return "text"
}

// getNumberProperty 获取数字类型属性
func (s *MySQLService) getNumberProperty(mysqlType string) map[string]interface{} {
	mysqlType = strings.ToLower(mysqlType)

	if strings.Contains(mysqlType, "int") {
		return map[string]interface{}{
			"formatter": "INT",
		}
	}

	if strings.Contains(mysqlType, "decimal") || strings.Contains(mysqlType, "float") {
		return map[string]interface{}{
			"formatter": "FLOAT_2",
		}
	}

	return nil
}
