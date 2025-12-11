package models

import "time"

// UserInfo 用户信息(用于API响应,不包含密码)
type UserInfo struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	Role        string `json:"role" binding:"required"` // "admin" 或 "user"
	DisplayName string `json:"displayName"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"` // "active" 或 "disabled"
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UserQuery 用户查询参数
type UserQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Role     string `form:"role"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

// UserLoginRequest 用户登录请求(前端用户)
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserLoginResponse 用户登录响应
type UserLoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expiresAt"`
	User      UserInfo  `json:"user"`
}
