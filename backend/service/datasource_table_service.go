package service

import (
	"fmt"
	"mysql-sync-plugin/models"
	"time"
)

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

	// 自动从数据库加载字段映射
	if err := s.autoLoadFieldMappings(table, ds); err != nil {
		// 字段加载失败不影响表创建,只记录错误
		fmt.Printf("自动加载字段映射失败: %v\n", err)
	}

	// 如果提供了字段映射,批量创建(会覆盖自动加载的字段)
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

// autoLoadFieldMappings 自动从数据库加载字段映射
func (s *DatasourceService) autoLoadFieldMappings(table *models.DatasourceTable, ds *models.Datasource) error {
	var fields []FieldInfo
	var err error

	// 根据查询模式获取字段列表
	if table.QueryMode == "sql" {
		if table.CustomSQL == "" {
			return fmt.Errorf("自定义SQL为空")
		}
		fields, err = s.GetFieldListFromSQL(ds.ID, ds.DatabaseName, table.CustomSQL)
	} else {
		fields, err = s.GetFieldList(ds.ID, ds.DatabaseName, table.TableName)
	}

	if err != nil {
		return fmt.Errorf("获取字段列表失败: %w", err)
	}

	// 创建字段映射
	if len(fields) > 0 {
		mappings := make([]*models.DatasourceFieldMapping, len(fields))
		for i, field := range fields {
			// 使用字段名作为别名
			mappings[i] = &models.DatasourceFieldMapping{
				DatasourceTableID: table.ID,
				FieldName:         field.Name,
				FieldAlias:        field.Name,
				CreatedAt:         time.Now(),
			}
		}
		if err := s.repo.BatchCreateFieldMappings(table.ID, mappings); err != nil {
			return fmt.Errorf("批量创建字段映射失败: %w", err)
		}
	}

	return nil
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

// BatchCreateDatasourceTables 批量创建数据源表配置
func (s *DatasourceService) BatchCreateDatasourceTables(datasourceID int64, req *models.BatchCreateDatasourceTablesRequest) ([]*models.DatasourceTable, error) {
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

	// 批量创建表配置
	tables := make([]*models.DatasourceTable, 0, len(req.TableNames))
	for _, tableName := range req.TableNames {
		tableReq := &models.CreateDatasourceTableRequest{
			TableName: tableName,
			QueryMode: req.QueryMode,
		}
		table, err := s.CreateDatasourceTable(datasourceID, tableReq)
		if err != nil {
			// 记录错误但继续处理其他表
			fmt.Printf("创建表配置失败 [%s]: %v\n", tableName, err)
			continue
		}
		tables = append(tables, table)
	}

	return tables, nil
}
