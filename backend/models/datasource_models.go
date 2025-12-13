package models

import "time"

// Datasource 数据源模型
type Datasource struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	DatabaseName string    `json:"databaseName"`
	Username     string    `json:"username"`
	Password     string    `json:"-"` // 不输出到JSON
	CreatedBy    int64     `json:"createdBy"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// DatasourceTable 数据源表配置模型
type DatasourceTable struct {
	ID           int64     `json:"id"`
	DatasourceID int64     `json:"datasourceId"`
	TableName    string    `json:"tableName"`
	TableAlias   string    `json:"tableAlias"`
	QueryMode    string    `json:"queryMode"` // "table" 或 "sql"
	CustomSQL    string    `json:"customSql"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// DatasourceFieldMapping 字段映射模型
type DatasourceFieldMapping struct {
	ID                int64     `json:"id"`
	DatasourceTableID int64     `json:"datasourceTableId"`
	FieldName         string    `json:"fieldName"`
	FieldAlias        string    `json:"fieldAlias"`
	Enabled           bool      `json:"enabled"`
	CreatedAt         time.Time `json:"createdAt"`
}

// CreateDatasourceRequest 创建数据源请求
type CreateDatasourceRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Host         string `json:"host" binding:"required"`
	Port         int    `json:"port" binding:"required"`
	DatabaseName string `json:"databaseName" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

// UpdateDatasourceRequest 更新数据源请求
type UpdateDatasourceRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"databaseName"`
	Username     string `json:"username"`
	Password     string `json:"password"` // 如果为空则不更新密码
}

// CreateDatasourceTableRequest 创建数据源表配置请求
type CreateDatasourceTableRequest struct {
	TableName  string                   `json:"tableName" binding:"required"`
	TableAlias string                   `json:"tableAlias"`
	QueryMode  string                   `json:"queryMode" binding:"required"` // "table" 或 "sql"
	CustomSQL  string                   `json:"customSql"`
	FieldMappings []FieldMapping        `json:"fieldMappings"` // 字段映射列表
}

// UpdateDatasourceTableRequest 更新数据源表配置请求
type UpdateDatasourceTableRequest struct {
	TableAlias string          `json:"tableAlias"`
	QueryMode  string          `json:"queryMode"`
	CustomSQL  string          `json:"customSql"`
	FieldMappings []FieldMapping `json:"fieldMappings"`
}

// DatasourceQuery 数据源查询参数
type DatasourceQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
}

// DatasourceWithTables 数据源及其表配置
type DatasourceWithTables struct {
	Datasource
	Tables []DatasourceTable `json:"tables"`
}

// BatchCreateDatasourceTablesRequest 批量创建数据源表配置请求
type BatchCreateDatasourceTablesRequest struct {
	TableNames []string `json:"tableNames" binding:"required"` // 表名列表
	QueryMode  string   `json:"queryMode" binding:"required"`  // 查询模式: "table" 或 "sql"
}
