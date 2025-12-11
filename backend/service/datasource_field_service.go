package service

import (
	"fmt"
	"mysql-sync-plugin/models"
	"time"
)

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
