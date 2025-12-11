package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/repository"
	"time"
)

// UserService 用户管理服务
type UserService struct {
	repo repository.Repository
}

// NewUserService 创建用户管理服务实例
func NewUserService(repo repository.Repository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.UserInfo, error) {
	// 验证角色
	if req.Role != "admin" && req.Role != "user" {
		return nil, fmt.Errorf("无效的角色: %s", req.Role)
	}

	// 检查用户名是否已存在
	existingUser, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 创建用户
	user := &auth.User{
		Username:    req.Username,
		Password:    hashPassword(req.Password),
		Role:        req.Role,
		DisplayName: req.DisplayName,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return s.toUserInfo(user), nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int64) (*models.UserInfo, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	return s.toUserInfo(user), nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(query *models.UserQuery) ([]*models.UserInfo, int64, error) {
	users, total, err := s.repo.ListUsers(query)
	if err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}

	userInfos := make([]*models.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = s.toUserInfo(user)
	}

	return userInfos, total, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id int64, req *models.UpdateUserRequest) (*models.UserInfo, error) {
	// 获取用户
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 更新字段
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Role != "" {
		if req.Role != "admin" && req.Role != "user" {
			return nil, fmt.Errorf("无效的角色: %s", req.Role)
		}
		user.Role = req.Role
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return s.toUserInfo(user), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int64) error {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 不允许删除管理员账户(至少保留一个)
	if user.Role == "admin" {
		// 检查是否还有其他管理员
		query := &models.UserQuery{Role: "admin", Page: 1, PageSize: 10}
		admins, _, err := s.repo.ListUsers(query)
		if err != nil {
			return fmt.Errorf("检查管理员数量失败: %w", err)
		}
		if len(admins) <= 1 {
			return fmt.Errorf("不能删除最后一个管理员账户")
		}
	}

	if err := s.repo.DeleteUser(id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// UpdateUserStatus 更新用户状态
func (s *UserService) UpdateUserStatus(id int64, status string) error {
	// 验证状态
	if status != "active" && status != "disabled" {
		return fmt.Errorf("无效的状态: %s", status)
	}

	// 检查用户是否存在
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 不允许禁用管理员账户(至少保留一个活跃的)
	if user.Role == "admin" && status == "disabled" {
		query := &models.UserQuery{Role: "admin", Status: "active", Page: 1, PageSize: 10}
		activeAdmins, _, err := s.repo.ListUsers(query)
		if err != nil {
			return fmt.Errorf("检查活跃管理员数量失败: %w", err)
		}
		if len(activeAdmins) <= 1 {
			return fmt.Errorf("不能禁用最后一个活跃的管理员账户")
		}
	}

	if err := s.repo.UpdateUserStatus(id, status); err != nil {
		return fmt.Errorf("更新用户状态失败: %w", err)
	}

	return nil
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(id int64, newPassword string) error {
	// 检查用户是否存在
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证密码长度
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度不能少于6位")
	}

	hashedPassword := hashPassword(newPassword)
	if err := s.repo.UpdateUserPassword(id, hashedPassword); err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	return nil
}

// ValidateLogin 验证用户登录
func (s *UserService) ValidateLogin(username, password string) (*models.UserInfo, error) {
	// 获取用户
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, fmt.Errorf("用户已被禁用")
	}

	// 验证密码
	if !verifyPassword(password, user.Password) {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	return s.toUserInfo(user), nil
}

// toUserInfo 转换为UserInfo
func (s *UserService) toUserInfo(user *auth.User) *models.UserInfo {
	return &models.UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Role:        user.Role,
		DisplayName: user.DisplayName,
		Status:      user.Status,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// hashPassword 哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// verifyPassword 验证密码
func verifyPassword(inputPassword, storedPassword string) bool {
	return hashPassword(inputPassword) == storedPassword
}
