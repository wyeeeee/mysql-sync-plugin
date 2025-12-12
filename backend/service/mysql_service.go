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

	var fields []models.Field
	var sheetName string

	// 根据取数模式获取字段
	if config.QueryMode == "sql" && config.CustomSQL != "" {
		fields, err = s.getSQLSchema(db, config.CustomSQL)
		sheetName = "自定义查询"
	} else {
		fields, err = s.getTableSchema(db, config.Database, config.Table)
		sheetName = config.Table
	}
	if err != nil {
		return nil, err
	}

	// 应用字段映射
	fields = s.applyFieldMappings(fields, config.FieldMappings)

	return &models.SheetMetaResponse{
		SheetName: sheetName,
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

	var total int
	var fields []models.Field
	var records []models.Record

	// 根据取数模式获取数据
	if config.QueryMode == "sql" && config.CustomSQL != "" {
		total, err = s.getSQLRecordCount(db, config.CustomSQL)
		if err != nil {
			return nil, err
		}
		fields, err = s.getSQLSchema(db, config.CustomSQL)
		if err != nil {
			return nil, err
		}
		records, err = s.getSQLRecords(db, config.CustomSQL, fields, offset, maxResults)
	} else {
		total, err = s.getRecordCount(db, config.Table)
		if err != nil {
			return nil, err
		}
		fields, err = s.getTableSchema(db, config.Database, config.Table)
		if err != nil {
			return nil, err
		}
		records, err = s.getTableRecords(db, config.Table, fields, offset, maxResults)
	}
	if err != nil {
		return nil, err
	}

	// 应用字段映射到记录
	records = s.applyRecordFieldMappings(records, config.FieldMappings)

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

	// 构建字段类型映射
	fieldTypeMap := make(map[string]string)
	for _, f := range fields {
		columnName := strings.TrimPrefix(f.ID, "fid_")
		fieldTypeMap[columnName] = f.Type
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
			fieldType := fieldTypeMap[col]

			// 根据字段类型正确转换数据
			record.Fields[fieldID] = s.convertValue(value, fieldType)

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
		strings.Contains(mysqlType, "double") ||
		strings.Contains(mysqlType, "numeric") ||
		strings.Contains(mysqlType, "real") ||
		strings.Contains(mysqlType, "money") ||
		mysqlType == "bit" {
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

// convertValue 根据字段类型转换值
func (s *MySQLService) convertValue(value interface{}, fieldType string) interface{} {
	if value == nil {
		return nil
	}

	// 数字类型处理
	if fieldType == "number" {
		return s.toNumber(value)
	}

	// 非数字类型，转换为字符串
	switch v := value.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	case int64:
		return v
	case float64:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toNumber 将值转换为数字类型
func (s *MySQLService) toNumber(value interface{}) interface{} {
	if value == nil {
		// 修复：飞书数字字段不接受 nil，返回 0
		return 0
	}

	switch v := value.(type) {
	case int64:
		return v
	case int32:
		return int64(v)
	case int:
		return int64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	case []byte:
		strVal := string(v)
		if strVal == "" {
			// 修复：空字符串返回 0
			return 0
		}
		if f, err := strconv.ParseFloat(strVal, 64); err == nil {
			return f
		}
		// 修复：解析失败返回 0
		return 0
	case string:
		if v == "" {
			// 修复：空字符串返回 0
			return 0
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		// 修复：解析失败返回 0
		return 0
	default:
		// 尝试转换任何其他类型
		strVal := fmt.Sprintf("%v", v)
		if strVal == "" {
			// 修复：空字符串返回 0
			return 0
		}
		if f, err := strconv.ParseFloat(strVal, 64); err == nil {
			return f
		}
		// 修复：解析失败返回 0
		return 0
	}
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

// applyFieldMappings 应用字段映射到字段列表
func (s *MySQLService) applyFieldMappings(fields []models.Field, mappings []models.FieldMapping) []models.Field {
	if len(mappings) == 0 {
		return fields
	}

	// 构建映射表
	aliasMap := make(map[string]string)
	for _, m := range mappings {
		if m.AliasField != "" && m.AliasField != m.MysqlField {
			aliasMap[m.MysqlField] = m.AliasField
		}
	}

	// 应用映射：同时修改 Name 和 ID（钉钉需要这样来显示别名）
	for i := range fields {
		if alias, ok := aliasMap[fields[i].Name]; ok {
			fields[i].Name = alias
			fields[i].ID = fmt.Sprintf("fid_%s", alias)
		}
	}

	return fields
}

// applyRecordFieldMappings 应用字段映射到记录数据
func (s *MySQLService) applyRecordFieldMappings(records []models.Record, mappings []models.FieldMapping) []models.Record {
	if len(mappings) == 0 {
		return records
	}

	// 构建映射表: fid_原字段名 -> fid_别名
	aliasMap := make(map[string]string)
	for _, m := range mappings {
		if m.AliasField != "" && m.AliasField != m.MysqlField {
			oldKey := fmt.Sprintf("fid_%s", m.MysqlField)
			newKey := fmt.Sprintf("fid_%s", m.AliasField)
			aliasMap[oldKey] = newKey
		}
	}

	// 应用映射到每条记录
	for i := range records {
		newFields := make(map[string]interface{})
		for key, value := range records[i].Fields {
			if newKey, ok := aliasMap[key]; ok {
				newFields[newKey] = value
			} else {
				newFields[key] = value
			}
		}
		records[i].Fields = newFields
	}

	return records
}

// getSQLSchema 通过执行SQL获取结果集的字段结构
func (s *MySQLService) getSQLSchema(db *sql.DB, customSQL string) ([]models.Field, error) {
	// 添加 LIMIT 1 来只获取一行用于分析结构
	query := fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT 1", strings.TrimSuffix(strings.TrimSpace(customSQL), ";"))

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("执行SQL失败: %w", err)
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("获取列信息失败: %w", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("获取列类型失败: %w", err)
	}

	// 获取当前数据库所有字段的备注映射
	commentMap := s.getAllColumnComments(db)

	var fields []models.Field
	for i, col := range columns {
		dbType := ""
		if i < len(columnTypes) {
			dbType = columnTypes[i].DatabaseTypeName()
		}

		field := models.Field{
			ID:          fmt.Sprintf("fid_%s", col),
			Name:        col,
			Type:        s.mapMySQLTypeToAITable(dbType),
			IsPrimary:   i == 0,
			Description: commentMap[col],
		}

		if field.Type == "number" {
			field.Property = s.getNumberProperty(dbType)
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// getAllColumnComments 获取当前数据库所有表的字段备注
func (s *MySQLService) getAllColumnComments(db *sql.DB) map[string]string {
	commentMap := make(map[string]string)

	query := `
		SELECT COLUMN_NAME, COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_COMMENT != ''
	`

	rows, err := db.Query(query)
	if err != nil {
		return commentMap
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, columnComment string
		if err := rows.Scan(&columnName, &columnComment); err != nil {
			continue
		}
		// 如果同名字段有多个备注，保留第一个
		if _, exists := commentMap[columnName]; !exists {
			commentMap[columnName] = columnComment
		}
	}

	return commentMap
}

// getSQLRecordCount 获取自定义SQL的记录总数
func (s *MySQLService) getSQLRecordCount(db *sql.DB, customSQL string) (int, error) {
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS t", strings.TrimSuffix(strings.TrimSpace(customSQL), ";"))

	var count int
	err := db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("获取记录数失败: %w", err)
	}

	return count, nil
}

// getSQLRecords 获取自定义SQL的记录数据(分页)
func (s *MySQLService) getSQLRecords(db *sql.DB, customSQL string, fields []models.Field, offset, limit int) ([]models.Record, error) {
	query := fmt.Sprintf("SELECT * FROM (%s) AS t LIMIT %d OFFSET %d",
		strings.TrimSuffix(strings.TrimSpace(customSQL), ";"),
		limit,
		offset,
	)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("执行SQL失败: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 构建字段类型映射
	fieldTypeMap := make(map[string]string)
	for _, f := range fields {
		columnName := strings.TrimPrefix(f.ID, "fid_")
		fieldTypeMap[columnName] = f.Type
	}

	var records []models.Record
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		record := models.Record{
			Fields: make(map[string]interface{}),
		}

		for i, col := range columns {
			fieldID := fmt.Sprintf("fid_%s", col)
			value := values[i]
			fieldType := fieldTypeMap[col]

			// 根据字段类型正确转换数据
			record.Fields[fieldID] = s.convertValue(value, fieldType)

			if i == 0 && value != nil {
				record.ID = fmt.Sprintf("%v", value)
			}
		}

		records = append(records, record)
	}

	return records, rows.Err()
}

// PreviewSQL 预览SQL执行结果（获取字段列表）
func (s *MySQLService) PreviewSQL(config *models.MySQLConfig) ([]models.Field, error) {
	db, err := s.connectDB(config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return s.getSQLSchema(db, config.CustomSQL)
}
