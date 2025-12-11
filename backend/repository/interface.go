package repository

import (
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/models"
)

// Repository 数据访问接口
type Repository interface {
	// ==================== 用户管理 ====================

	// CreateUser 创建用户
	CreateUser(user *auth.User) error

	// GetUserByID 根据ID获取用户
	GetUserByID(id int64) (*auth.User, error)

	// GetUserByUsername 根据用户名获取用户
	GetUserByUsername(username string) (*auth.User, error)

	// ListUsers 获取用户列表(分页)
	ListUsers(query *models.UserQuery) ([]*auth.User, int64, error)

	// UpdateUser 更新用户信息
	UpdateUser(user *auth.User) error

	// DeleteUser 删除用户
	DeleteUser(id int64) error

	// UpdateUserStatus 更新用户状态
	UpdateUserStatus(id int64, status string) error

	// UpdateUserPassword 更新用户密码
	UpdateUserPassword(id int64, hashedPassword string) error

	// ==================== 数据源管理 ====================

	// CreateDatasource 创建数据源
	CreateDatasource(ds *models.Datasource) error

	// GetDatasourceByID 根据ID获取数据源
	GetDatasourceByID(id int64) (*models.Datasource, error)

	// ListDatasources 获取数据源列表(分页)
	ListDatasources(query *models.DatasourceQuery) ([]*models.Datasource, int64, error)

	// UpdateDatasource 更新数据源
	UpdateDatasource(ds *models.Datasource) error

	// DeleteDatasource 删除数据源
	DeleteDatasource(id int64) error

	// ==================== 数据源表管理 ====================

	// CreateDatasourceTable 创建数据源表配置
	CreateDatasourceTable(table *models.DatasourceTable) error

	// GetDatasourceTableByID 根据ID获取数据源表配置
	GetDatasourceTableByID(id int64) (*models.DatasourceTable, error)

	// ListDatasourceTables 获取数据源的所有表配置
	ListDatasourceTables(datasourceID int64) ([]*models.DatasourceTable, error)

	// UpdateDatasourceTable 更新数据源表配置
	UpdateDatasourceTable(table *models.DatasourceTable) error

	// DeleteDatasourceTable 删除数据源表配置
	DeleteDatasourceTable(id int64) error

	// ==================== 字段映射管理 ====================

	// BatchCreateFieldMappings 批量创建字段映射
	BatchCreateFieldMappings(tableID int64, mappings []*models.DatasourceFieldMapping) error

	// ListFieldMappings 获取表的所有字段映射
	ListFieldMappings(tableID int64) ([]*models.DatasourceFieldMapping, error)

	// DeleteFieldMappingsByTableID 删除表的所有字段映射
	DeleteFieldMappingsByTableID(tableID int64) error

	// ==================== 权限管理 ====================

	// GrantDatasourcePermission 授予用户数据源权限
	GrantDatasourcePermission(userID, datasourceID int64) error

	// RevokeDatasourcePermission 撤销用户数据源权限
	RevokeDatasourcePermission(userID, datasourceID int64) error

	// ListUserDatasources 获取用户可访问的数据源列表
	ListUserDatasources(userID int64) ([]*models.Datasource, error)

	// CheckDatasourcePermission 检查用户是否有数据源权限
	CheckDatasourcePermission(userID, datasourceID int64) (bool, error)

	// GrantTablePermission 授予用户表权限
	GrantTablePermission(userID, tableID int64) error

	// RevokeTablePermission 撤销用户表权限
	RevokeTablePermission(userID, tableID int64) error

	// ListUserTables 获取用户在指定数据源下可访问的表列表
	ListUserTables(userID, datasourceID int64) ([]*models.DatasourceTable, error)

	// CheckTablePermission 检查用户是否有表权限
	CheckTablePermission(userID, tableID int64) (bool, error)

	// BatchGrantDatasourcePermissions 批量授予数据源权限
	BatchGrantDatasourcePermissions(userID int64, datasourceIDs []int64) error

	// BatchRevokeDatasourcePermissions 批量撤销数据源权限
	BatchRevokeDatasourcePermissions(userID int64, datasourceIDs []int64) error

	// BatchGrantTablePermissions 批量授予表权限
	BatchGrantTablePermissions(userID int64, tableIDs []int64) error

	// BatchRevokeTablePermissions 批量撤销表权限
	BatchRevokeTablePermissions(userID int64, tableIDs []int64) error
}
