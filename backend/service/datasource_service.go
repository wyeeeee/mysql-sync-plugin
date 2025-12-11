package service

import (
	"database/sql"
	"fmt"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/repository"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DatasourceService 数据源管理服务
type DatasourceService struct {
	repo   repository.Repository
	crypto *CryptoService
}

// NewDatasourceService 创建数据源管理服务实例
func NewDatasourceService(repo repository.Repository, crypto *CryptoService) *DatasourceService {
	return &DatasourceService{
		repo:   repo,
		crypto: crypto,
	}
}

// CreateDatasource 创建数据源
func (s *DatasourceService) CreateDatasource(req *models.CreateDatasourceRequest, createdBy int64) (*models.Datasource, error) {
	// 加密密码
	encryptedPassword, err := s.crypto.Encrypt(req.Password)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败: %w", err)
	}

	// 创建数据源
	ds := &models.Datasource{
		Name:         req.Name,
		Description:  req.Description,
		Host:         req.Host,
		Port:         req.Port,
		DatabaseName: req.DatabaseName,
		Username:     req.Username,
		Password:     encryptedPassword,
		CreatedBy:    createdBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateDatasource(ds); err != nil {
		return nil, fmt.Errorf("创建数据源失败: %w", err)
	}

	// 返回时不包含密码
	ds.Password = ""
	return ds, nil
}

// GetDatasourceByID 根据ID获取数据源
func (s *DatasourceService) GetDatasourceByID(id int64) (*models.Datasource, error) {
	ds, err := s.repo.GetDatasourceByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("数据源不存在")
	}

	// 返回时不包含密码
	ds.Password = ""
	return ds, nil
}

// GetDatasourceByIDWithPassword 根据ID获取数据源(包含解密后的密码)
func (s *DatasourceService) GetDatasourceByIDWithPassword(id int64) (*models.Datasource, error) {
	ds, err := s.repo.GetDatasourceByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("数据源不存在")
	}

	// 解密密码
	decryptedPassword, err := s.crypto.Decrypt(ds.Password)
	if err != nil {
		return nil, fmt.Errorf("解密密码失败: %w", err)
	}
	ds.Password = decryptedPassword

	return ds, nil
}

// ListDatasources 获取数据源列表
func (s *DatasourceService) ListDatasources(query *models.DatasourceQuery) ([]*models.Datasource, int64, error) {
	datasources, total, err := s.repo.ListDatasources(query)
	if err != nil {
		return nil, 0, fmt.Errorf("获取数据源列表失败: %w", err)
	}

	// 返回时不包含密码
	for _, ds := range datasources {
		ds.Password = ""
	}

	return datasources, total, nil
}

// UpdateDatasource 更新数据源
func (s *DatasourceService) UpdateDatasource(id int64, req *models.UpdateDatasourceRequest) (*models.Datasource, error) {
	// 获取数据源
	ds, err := s.repo.GetDatasourceByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("数据源不存在")
	}

	// 更新字段
	if req.Name != "" {
		ds.Name = req.Name
	}
	if req.Description != "" {
		ds.Description = req.Description
	}
	if req.Host != "" {
		ds.Host = req.Host
	}
	if req.Port > 0 {
		ds.Port = req.Port
	}
	if req.DatabaseName != "" {
		ds.DatabaseName = req.DatabaseName
	}
	if req.Username != "" {
		ds.Username = req.Username
	}
	// 如果提供了新密码,则加密并更新
	if req.Password != "" {
		encryptedPassword, err := s.crypto.Encrypt(req.Password)
		if err != nil {
			return nil, fmt.Errorf("加密密码失败: %w", err)
		}
		ds.Password = encryptedPassword
	}

	ds.UpdatedAt = time.Now()

	if err := s.repo.UpdateDatasource(ds); err != nil {
		return nil, fmt.Errorf("更新数据源失败: %w", err)
	}

	// 返回时不包含密码
	ds.Password = ""
	return ds, nil
}

// DeleteDatasource 删除数据源
func (s *DatasourceService) DeleteDatasource(id int64) error {
	// 检查数据源是否存在
	ds, err := s.repo.GetDatasourceByID(id)
	if err != nil {
		return fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return fmt.Errorf("数据源不存在")
	}

	if err := s.repo.DeleteDatasource(id); err != nil {
		return fmt.Errorf("删除数据源失败: %w", err)
	}

	return nil
}

// TestConnection 测试数据源连接
func (s *DatasourceService) TestConnection(id int64) error {
	// 获取数据源(包含密码)
	ds, err := s.GetDatasourceByIDWithPassword(id)
	if err != nil {
		return err
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		ds.Username,
		ds.Password,
		ds.Host,
		ds.Port,
		ds.DatabaseName,
	)

	// 尝试连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接测试失败: %w", err)
	}

	return nil
}

// CreateDatasourceTable 创建数据源表配置
func (s *DatasourceService) CreateDatasourceTable(datasourceID int64, req *models.CreateDatasourceTableRequest) (*models.DatasourceTable, error) {
	// 检查数据源是否存在
	ds, err := s.repo.GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("数据源不存在")
	}

	// 验证查询模式
	if req.QueryMode != "table" && req.QueryMode != "sql" {
		return nil, fmt.Errorf("无效的查询模式: %s", req.QueryMode)
	}

	// 如果是SQL模式,必须提供自定义SQL
	if req.QueryMode == "sql" && req.CustomSQL == "" {
		return nil, fmt.Errorf("SQL模式必须提供自定义SQL")
	}

	// 创建表配置
	table := &models.DatasourceTable{
		DatasourceID: datasourceID,
		TableName:    req.TableName,
		TableAlias:   req.TableAlias,
		QueryMode:    req.QueryMode,
		CustomSQL:    req.CustomSQL,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateDatasourceTable(table); err != nil {
		return nil, fmt.Errorf("创建数据源表配置失败: %w", err)
	}

	// 如果提供了字段映射,批量创建
	if len(req.FieldMappings) > 0 {
		mappings := make([]*models.DatasourceFieldMapping, len(req.FieldMappings))
		for i, fm := range req.FieldMappings {
			mappings[i] = &models.DatasourceFieldMapping{
				DatasourceTableID: table.ID,
				FieldName:         fm.MysqlField,
				FieldAlias:        fm.AliasField,
				CreatedAt:         time.Now(),
			}
		}
		if err := s.repo.BatchCreateFieldMappings(table.ID, mappings); err != nil {
			return nil, fmt.Errorf("创建字段映射失败: %w", err)
		}
	}

	return table, nil
}

// GetDatasourceTableByID 根据ID获取数据源表配置
func (s *DatasourceService) GetDatasourceTableByID(id int64) (*models.DatasourceTable, error) {
	table, err := s.repo.GetDatasourceTableByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取数据源表配置失败: %w", err)
	}
	if table == nil {
		return nil, fmt.Errorf("数据源表配置不存在")
	}

	return table, nil
}

// ListDatasourceTables 获取数据源的所有表配置
func (s *DatasourceService) ListDatasourceTables(datasourceID int64) ([]*models.DatasourceTable, error) {
	tables, err := s.repo.ListDatasourceTables(datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取数据源表配置列表失败: %w", err)
	}

	return tables, nil
}

// UpdateDatasourceTable 更新数据源表配置
func (s *DatasourceService) UpdateDatasourceTable(id int64, req *models.UpdateDatasourceTableRequest) (*models.DatasourceTable, error) {
	// 获取表配置
	table, err := s.repo.GetDatasourceTableByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取数据源表配置失败: %w", err)
	}
	if table == nil {
		return nil, fmt.Errorf("数据源表配置不存在")
	}

	// 更新字段
	if req.TableAlias != "" {
		table.TableAlias = req.TableAlias
	}
	if req.QueryMode != "" {
		if req.QueryMode != "table" && req.QueryMode != "sql" {
			return nil, fmt.Errorf("无效的查询模式: %s", req.QueryMode)
		}
		table.QueryMode = req.QueryMode
	}
	if req.CustomSQL != "" {
		table.CustomSQL = req.CustomSQL
	}

	table.UpdatedAt = time.Now()

	if err := s.repo.UpdateDatasourceTable(table); err != nil {
		return nil, fmt.Errorf("更新数据源表配置失败: %w", err)
	}

	// 如果提供了字段映射,批量更新
	if len(req.FieldMappings) > 0 {
		mappings := make([]*models.DatasourceFieldMapping, len(req.FieldMappings))
		for i, fm := range req.FieldMappings {
			mappings[i] = &models.DatasourceFieldMapping{
				DatasourceTableID: table.ID,
				FieldName:         fm.MysqlField,
				FieldAlias:        fm.AliasField,
				CreatedAt:         time.Now(),
			}
		}
		if err := s.repo.BatchCreateFieldMappings(table.ID, mappings); err != nil {
			return nil, fmt.Errorf("更新字段映射失败: %w", err)
		}
	}

	return table, nil
}

// DeleteDatasourceTable 删除数据源表配置
func (s *DatasourceService) DeleteDatasourceTable(id int64) error {
	// 检查表配置是否存在
	table, err := s.repo.GetDatasourceTableByID(id)
	if err != nil {
		return fmt.Errorf("获取数据源表配置失败: %w", err)
	}
	if table == nil {
		return fmt.Errorf("数据源表配置不存在")
	}

	if err := s.repo.DeleteDatasourceTable(id); err != nil {
		return fmt.Errorf("删除数据源表配置失败: %w", err)
	}

	return nil
}

// GetFieldMappings 获取表的字段映射
func (s *DatasourceService) GetFieldMappings(tableID int64) ([]*models.DatasourceFieldMapping, error) {
	mappings, err := s.repo.ListFieldMappings(tableID)
	if err != nil {
		return nil, fmt.Errorf("获取字段映射失败: %w", err)
	}

	return mappings, nil
}

// UpdateFieldMappings 更新表的字段映射
func (s *DatasourceService) UpdateFieldMappings(tableID int64, fieldMappings []models.FieldMapping) error {
	// 检查表配置是否存在
	table, err := s.repo.GetDatasourceTableByID(tableID)
	if err != nil {
		return fmt.Errorf("获取数据源表配置失败: %w", err)
	}
	if table == nil {
		return fmt.Errorf("数据源表配置不存在")
	}

	// 转换为DatasourceFieldMapping
	mappings := make([]*models.DatasourceFieldMapping, len(fieldMappings))
	for i, fm := range fieldMappings {
		mappings[i] = &models.DatasourceFieldMapping{
			DatasourceTableID: tableID,
			FieldName:         fm.MysqlField,
			FieldAlias:        fm.AliasField,
			CreatedAt:         time.Now(),
		}
	}

	if err := s.repo.BatchCreateFieldMappings(tableID, mappings); err != nil {
		return fmt.Errorf("更新字段映射失败: %w", err)
	}

	return nil
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

// FieldInfo 字段信息
type FieldInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Comment     string `json:"comment"`
	IsPrimary   bool   `json:"isPrimary"`
	IsNullable  bool   `json:"isNullable"`
	DefaultValue string `json:"defaultValue"`
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
