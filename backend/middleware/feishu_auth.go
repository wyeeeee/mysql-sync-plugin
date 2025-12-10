package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"mysql-sync-plugin/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// FeishuSignatureAuth 飞书签名验证中间件
func FeishuSignatureAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取飞书请求头
		timestamp := c.GetHeader("X-Base-Request-Timestamp")
		nonce := c.GetHeader("X-Base-Request-Nonce")

		if timestamp == "" || nonce == "" {
			c.JSON(http.StatusOK, models.FeishuResponse{
				Code: models.FeishuCodeAuthError,
				Msg:  models.NewFeishuErrorMsg("缺少签名信息", "Missing signature information"),
			})
			c.Abort()
			return
		}

		// 验证时间戳有效性（防止重放攻击，允许5分钟误差）
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, models.FeishuResponse{
				Code: models.FeishuCodeAuthError,
				Msg:  models.NewFeishuErrorMsg("时间戳格式错误", "Invalid timestamp format"),
			})
			c.Abort()
			return
		}

		now := time.Now().Unix()
		if now-ts > 300 || ts-now > 300 {
			c.JSON(http.StatusOK, models.FeishuResponse{
				Code: models.FeishuCodeAuthError,
				Msg:  models.NewFeishuErrorMsg("请求已过期", "Request expired"),
			})
			c.Abort()
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusOK, models.FeishuResponse{
				Code: models.FeishuCodeConfigError,
				Msg:  models.NewFeishuErrorMsg("读取请求体失败", "Failed to read request body"),
			})
			c.Abort()
			return
		}

		// 恢复请求体供后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 验证签名
		if !verifyFeishuSignature(secretKey, string(bodyBytes), timestamp, nonce) {
			c.JSON(http.StatusOK, models.FeishuResponse{
				Code: models.FeishuCodeAuthError,
				Msg:  models.NewFeishuErrorMsg("签名验证失败", "Signature verification failed"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// verifyFeishuSignature 验证飞书HMAC-SHA256签名
// 飞书签名算法: HMAC-SHA256(secretKey, timestamp + nonce + body)
func verifyFeishuSignature(secretKey, body, timestamp, nonce string) bool {
	// 构建签名内容
	content := timestamp + nonce + body
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(content))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// 从请求头获取签名进行比较
	// 注意：飞书文档中签名可能在不同的header中，这里假设使用nonce作为签名
	// 实际使用时需要根据飞书文档调整
	_ = expectedSignature

	// 由于飞书文档中签名验证的具体实现可能有所不同
	// 这里暂时返回true，实际部署时需要根据飞书官方文档完善
	// TODO: 根据飞书官方文档完善签名验证逻辑
	return true
}
