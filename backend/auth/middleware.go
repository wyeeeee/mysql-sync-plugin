package auth

import (
	"mysql-sync-plugin/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理后台认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "未提供认证信息",
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "认证格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		store := GetStore()
		session, err := store.GetSessionByToken(token)
		if err != nil || session == nil {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "认证已过期或无效",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := store.GetUserByID(session.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if user.Status != "active" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "用户已被禁用",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("session", session)

		c.Next()
	}
}

// RequireAdminRole 要求管理员角色中间件
func RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "未找到用户信息",
			})
			c.Abort()
			return
		}

		u := user.(*User)
		if u.Role != "admin" {
			c.JSON(http.StatusForbidden, models.Response{
				Code: models.CodeInsufficientAuth,
				Msg:  "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserAuthMiddleware 普通用户认证中间件(用于钉钉/飞书前端)
func UserAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "未提供认证信息",
			})
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "认证格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证token
		store := GetStore()
		session, err := store.GetSessionByToken(token)
		if err != nil || session == nil {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "认证已过期或无效",
			})
			c.Abort()
			return
		}

		// 获取用户信息
		user, err := store.GetUserByID(session.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "用户不存在",
			})
			c.Abort()
			return
		}

		// 检查用户状态
		if user.Status != "active" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "用户已被禁用",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("session", session)

		c.Next()
	}
}
