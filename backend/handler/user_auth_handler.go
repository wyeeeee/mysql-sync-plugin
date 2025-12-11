package handler

import (
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"mysql-sync-plugin/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserAuthHandler 前端用户认证处理器
type UserAuthHandler struct {
	userService       *service.UserService
	permissionService *service.PermissionService
	authStore         *auth.Store
	log               *logger.Logger
}

// NewUserAuthHandler 创建前端用户认证处理器
func NewUserAuthHandler(userService *service.UserService, permissionService *service.PermissionService, authStore *auth.Store) *UserAuthHandler {
	return &UserAuthHandler{
		userService:       userService,
		permissionService: permissionService,
		authStore:         authStore,
		log:               logger.New("user-auth-handler"),
	}
}

// Login 用户登录
func (h *UserAuthHandler) Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	// 验证用户登录
	userInfo, err := h.userService.ValidateLogin(req.Username, req.Password)
	if err != nil {
		h.log.Errorf("用户登录", "登录失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  err.Error(),
		})
		return
	}

	// 创建会话(普通用户token有效期7天)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := h.authStore.CreateSession(userInfo.ID)
	if err != nil {
		h.log.Errorf("用户登录", "创建会话失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "创建会话失败",
		})
		return
	}

	h.log.Infof("用户登录", "用户 %s 登录成功", userInfo.Username)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: models.UserLoginResponse{
			Token:     session.Token,
			ExpiresAt: expiresAt.Unix(),
			User:      *userInfo,
		},
	})
}

// Logout 用户登出
func (h *UserAuthHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeSuccess,
			Msg:  "登出成功",
		})
		return
	}

	// 移除 "Bearer " 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if err := h.authStore.DeleteSession(token); err != nil {
		h.log.Errorf("用户登出", "删除会话失败: %v", err)
	}

	h.log.Info("用户登出", "用户登出成功")
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Msg:  "登出成功",
	})
}

// GetCurrentUser 获取当前用户信息
func (h *UserAuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "未找到用户信息",
		})
		return
	}

	userInfo, err := h.userService.GetUserByID(userID.(int64))
	if err != nil {
		h.log.Errorf("获取当前用户", "获取用户信息失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取用户信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: userInfo,
	})
}

// GetUserDatasources 获取当前用户可访问的数据源列表
func (h *UserAuthHandler) GetUserDatasources(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "未找到用户信息",
		})
		return
	}

	datasources, err := h.permissionService.ListUserDatasources(userID.(int64))
	if err != nil {
		h.log.Errorf("获取用户数据源", "获取用户数据源失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取数据源列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: datasources,
	})
}

// GetUserTables 获取当前用户在指定数据源下可访问的表列表
func (h *UserAuthHandler) GetUserTables(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "未找到用户信息",
		})
		return
	}

	datasourceIDStr := c.Param("id")
	datasourceID, err := strconv.ParseInt(datasourceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "无效的数据源ID",
		})
		return
	}

	// 检查用户是否有数据源权限
	hasPermission, err := h.permissionService.CheckDatasourcePermission(userID.(int64), datasourceID)
	if err != nil {
		h.log.Errorf("检查数据源权限", "检查权限失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "检查权限失败",
		})
		return
	}

	if !hasPermission {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeInsufficientAuth,
			Msg:  "无权访问该数据源",
		})
		return
	}

	tables, err := h.permissionService.ListUserTables(userID.(int64), datasourceID)
	if err != nil {
		h.log.Errorf("获取用户表", "获取用户表失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "获取表列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: tables,
	})
}
