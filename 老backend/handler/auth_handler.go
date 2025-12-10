package handler

import (
	"mysql-sync-plugin/auth"
	"mysql-sync-plugin/logger"
	"mysql-sync-plugin/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	log *logger.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		log: logger.New("auth"),
	}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	store := auth.GetStore()

	// 查找用户
	user, err := store.GetUserByUsername(req.Username)
	if err != nil {
		h.log.Errorf("登录", "查询用户失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "系统错误",
		})
		return
	}

	if user == nil {
		h.log.Warnf("登录", "用户不存在: %s", req.Username)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "用户名或密码错误",
		})
		return
	}

	// 验证密码
	if !auth.VerifyPassword(req.Password, user.Password) {
		h.log.Warnf("登录", "密码错误: %s", req.Username)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "用户名或密码错误",
		})
		return
	}

	// 创建会话
	session, err := store.CreateSession(user.ID)
	if err != nil {
		h.log.Errorf("登录", "创建会话失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "登录失败",
		})
		return
	}

	h.log.Infof("登录", "用户登录成功: %s", req.Username)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: auth.LoginResponse{
			Token:     session.Token,
			ExpiresAt: session.ExpiresAt.Unix(),
		},
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	session, exists := c.Get("session")
	if !exists {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeSuccess,
		})
		return
	}

	s := session.(*auth.Session)
	store := auth.GetStore()
	store.DeleteSession(s.Token)

	h.log.Infof("登出", "用户登出: userID=%d", s.UserID)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "未登录",
		})
		return
	}

	u := user.(*auth.User)
	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
		Data: gin.H{
			"id":       u.ID,
			"username": u.Username,
		},
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req auth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeParamError,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	user, _ := c.Get("user")
	u := user.(*auth.User)

	// 验证旧密码
	if !auth.VerifyPassword(req.OldPassword, u.Password) {
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeAuthFailed,
			Msg:  "原密码错误",
		})
		return
	}

	// 更新密码
	store := auth.GetStore()
	if err := store.UpdatePassword(u.ID, req.NewPassword); err != nil {
		h.log.Errorf("修改密码", "更新密码失败: %v", err)
		c.JSON(http.StatusOK, models.Response{
			Code: models.CodeThirdPartyError,
			Msg:  "修改密码失败",
		})
		return
	}

	h.log.Infof("修改密码", "用户修改密码成功: %s", u.Username)

	c.JSON(http.StatusOK, models.Response{
		Code: models.CodeSuccess,
	})
}
