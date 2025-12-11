package handler

import (
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService *service.UserService
	log         *logger.Logger
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         logger.New("user-handler"),
	}
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		h.log.Errorf("创建用户", "创建用户失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "创建用户失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("创建用户", "成功创建用户: %s", user.Username)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: user,
	})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		h.log.Errorf("获取用户", "获取用户失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取用户失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: user,
	})
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	var query models.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	users, total, err := h.userService.ListUsers(&query)
	if err != nil {
		h.log.Errorf("获取用户列表", "获取用户列表失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取用户列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: gin.H{
			"list":     users,
			"total":    total,
			"page":     query.Page,
			"pageSize": query.PageSize,
		},
	})
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(id, &req)
	if err != nil {
		h.log.Errorf("更新用户", "更新用户失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "更新用户失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("更新用户", "成功更新用户: %s", user.Username)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: user,
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	if err := h.userService.DeleteUser(id); err != nil {
		h.log.Errorf("删除用户", "删除用户失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "删除用户失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("删除用户", "成功删除用户ID: %d", id)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "删除成功",
	})
}

// UpdateUserStatus 更新用户状态
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.userService.UpdateUserStatus(id, req.Status); err != nil {
		h.log.Errorf("更新用户状态", "更新用户状态失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "更新用户状态失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("更新用户状态", "成功更新用户ID %d 状态为: %s", id, req.Status)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "更新成功",
	})
}

// ResetPassword 重置用户密码
func (h *UserHandler) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的用户ID",
		})
		return
	}

	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.userService.ResetPassword(id, req.NewPassword); err != nil {
		h.log.Errorf("重置密码", "重置密码失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "重置密码失败: " + err.Error(),
		})
		return
	}

	h.log.Infof("重置密码", "成功重置用户ID %d 的密码", id)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "重置成功",
	})
}
