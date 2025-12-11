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
