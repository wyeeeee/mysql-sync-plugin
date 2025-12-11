package service

import (
	"fmt"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/repository"
)

// PermissionService 权限管理服务
type PermissionService struct {
	repo repository.Repository
}

// NewPermissionService 创建权限管理服务实例
func NewPermissionService(repo repository.Repository) *PermissionService {
	return &PermissionService{
		repo: repo,
	}
}

// GrantDatasourcePermissions 授予用户数据源权限(批量)
func (s *PermissionService) GrantDatasourcePermissions(userID int64, datasourceIDs []int64) error {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查所有数据源是否存在
	for _, dsID := range datasourceIDs {
		ds, err := s.repo.GetDatasourceByID(dsID)
		if err != nil {
			return fmt.Errorf("获取数据源失败: %w", err)
		}
		if ds == nil {
			return fmt.Errorf("数据源 %d 不存在", dsID)
		}
	}

	// 批量授予权限
	if err := s.repo.BatchGrantDatasourcePermissions(userID, datasourceIDs); err != nil {
		return fmt.Errorf("授予数据源权限失败: %w", err)
	}

	return nil
}

// RevokeDatasourcePermissions 撤销用户数据源权限(批量)
func (s *PermissionService) RevokeDatasourcePermissions(userID int64, datasourceIDs []int64) error {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 批量撤销权限
	if err := s.repo.BatchRevokeDatasourcePermissions(userID, datasourceIDs); err != nil {
		return fmt.Errorf("撤销数据源权限失败: %w", err)
	}

	return nil
}

// ListUserDatasources 获取用户可访问的数据源列表
func (s *PermissionService) ListUserDatasources(userID int64) ([]*models.Datasource, error) {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	datasources, err := s.repo.ListUserDatasources(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户数据源列表失败: %w", err)
	}

	// 返回时不包含密码
	for _, ds := range datasources {
		ds.Password = ""
	}

	return datasources, nil
}

// CheckDatasourcePermission 检查用户是否有数据源权限
func (s *PermissionService) CheckDatasourcePermission(userID, datasourceID int64) (bool, error) {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return false, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("用户不存在")
	}

	// 管理员拥有所有权限
	if user.Role == "admin" {
		return true, nil
	}

	hasPermission, err := s.repo.CheckDatasourcePermission(userID, datasourceID)
	if err != nil {
		return false, fmt.Errorf("检查数据源权限失败: %w", err)
	}

	return hasPermission, nil
}

// GrantTablePermissions 授予用户表权限(批量)
func (s *PermissionService) GrantTablePermissions(userID int64, tableIDs []int64) error {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 检查所有表是否存在
	for _, tableID := range tableIDs {
		table, err := s.repo.GetDatasourceTableByID(tableID)
		if err != nil {
			return fmt.Errorf("获取数据源表配置失败: %w", err)
		}
		if table == nil {
			return fmt.Errorf("数据源表配置 %d 不存在", tableID)
		}
	}

	// 批量授予权限
	if err := s.repo.BatchGrantTablePermissions(userID, tableIDs); err != nil {
		return fmt.Errorf("授予表权限失败: %w", err)
	}

	return nil
}

// RevokeTablePermissions 撤销用户表权限(批量)
func (s *PermissionService) RevokeTablePermissions(userID int64, tableIDs []int64) error {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 批量撤销权限
	if err := s.repo.BatchRevokeTablePermissions(userID, tableIDs); err != nil {
		return fmt.Errorf("撤销表权限失败: %w", err)
	}

	return nil
}

// ListUserTables 获取用户在指定数据源下可访问的表列表
func (s *PermissionService) ListUserTables(userID, datasourceID int64) ([]*models.DatasourceTable, error) {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查数据源是否存在
	ds, err := s.repo.GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("数据源不存在")
	}

	tables, err := s.repo.ListUserTables(userID, datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取用户表列表失败: %w", err)
	}

	return tables, nil
}

// CheckTablePermission 检查用户是否有表权限
func (s *PermissionService) CheckTablePermission(userID, tableID int64) (bool, error) {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return false, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return false, fmt.Errorf("用户不存在")
	}

	// 管理员拥有所有权限
	if user.Role == "admin" {
		return true, nil
	}

	hasPermission, err := s.repo.CheckTablePermission(userID, tableID)
	if err != nil {
		return false, fmt.Errorf("检查表权限失败: %w", err)
	}

	return hasPermission, nil
}

// ListAllDatasourcesWithPermission 获取所有数据源及用户权限状态
func (s *PermissionService) ListAllDatasourcesWithPermission(userID int64) ([]*models.UserDatasourceWithPermission, error) {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 获取所有数据源
	query := &models.DatasourceQuery{Page: 1, PageSize: 1000}
	allDatasources, _, err := s.repo.ListDatasources(query)
	if err != nil {
		return nil, fmt.Errorf("获取数据源列表失败: %w", err)
	}

	// 获取用户已有权限的数据源
	userDatasources, err := s.repo.ListUserDatasources(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户数据源列表失败: %w", err)
	}

	// 构建权限映射
	permissionMap := make(map[int64]bool)
	for _, ds := range userDatasources {
		permissionMap[ds.ID] = true
	}

	// 构建结果
	result := make([]*models.UserDatasourceWithPermission, len(allDatasources))
	for i, ds := range allDatasources {
		ds.Password = "" // 不返回密码
		result[i] = &models.UserDatasourceWithPermission{
			Datasource:    *ds,
			HasPermission: permissionMap[ds.ID],
		}
	}

	return result, nil
}

// ListAllTablesWithPermission 获取指定数据源的所有表及用户权限状态
func (s *PermissionService) ListAllTablesWithPermission(userID, datasourceID int64) ([]*models.UserTableWithPermission, error) {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查数据源是否存在
	ds, err := s.repo.GetDatasourceByID(datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取数据源失败: %w", err)
	}
	if ds == nil {
		return nil, fmt.Errorf("数据源不存在")
	}

	// 获取数据源的所有表
	allTables, err := s.repo.ListDatasourceTables(datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取数据源表列表失败: %w", err)
	}

	// 获取用户已有权限的表
	userTables, err := s.repo.ListUserTables(userID, datasourceID)
	if err != nil {
		return nil, fmt.Errorf("获取用户表列表失败: %w", err)
	}

	// 构建权限映射
	permissionMap := make(map[int64]bool)
	for _, table := range userTables {
		permissionMap[table.ID] = true
	}

	// 构建结果
	result := make([]*models.UserTableWithPermission, len(allTables))
	for i, table := range allTables {
		result[i] = &models.UserTableWithPermission{
			DatasourceTable: *table,
			HasPermission:   permissionMap[table.ID],
		}
	}

	return result, nil
}
