package models

import "time"

// UserDatasourcePermission 用户数据源权限模型
type UserDatasourcePermission struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	DatasourceID int64     `json:"datasourceId"`
	CreatedAt    time.Time `json:"createdAt"`
}

// UserTablePermission 用户表权限模型
type UserTablePermission struct {
	ID                int64     `json:"id"`
	UserID            int64     `json:"userId"`
	DatasourceTableID int64     `json:"datasourceTableId"`
	CreatedAt         time.Time `json:"createdAt"`
}

// GrantDatasourcePermissionRequest 授予数据源权限请求
type GrantDatasourcePermissionRequest struct {
	DatasourceIDs []int64 `json:"datasourceIds" binding:"required"`
}

// GrantTablePermissionRequest 授予表权限请求
type GrantTablePermissionRequest struct {
	TableIDs []int64 `json:"tableIds" binding:"required"`
}

// UserDatasourceWithPermission 用户数据源及权限信息
type UserDatasourceWithPermission struct {
	Datasource
	HasPermission bool `json:"hasPermission"`
}

// UserTableWithPermission 用户表及权限信息
type UserTableWithPermission struct {
	DatasourceTable
	HasPermission bool `json:"hasPermission"`
}
