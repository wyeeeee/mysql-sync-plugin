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

	"github.com/gin-gonic/gin"
)

// SignatureAuth 签名验证中间件
func SignatureAuth(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求头
		timestamp := c.GetHeader("Ding-Docs-Timestamp")
		signature := c.GetHeader("Ding-Docs-Signature")

		if timestamp == "" || signature == "" {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "缺少签名信息",
			})
			c.Abort()
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.Response{
				Code: models.CodeParamError,
				Msg:  "读取请求体失败",
			})
			c.Abort()
			return
		}

		// 恢复请求体供后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 验证签名
		if !verifySignature(secretKey, string(bodyBytes), timestamp, signature) {
			c.JSON(http.StatusUnauthorized, models.Response{
				Code: models.CodeAuthFailed,
				Msg:  "签名验证失败",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// verifySignature 验证HMAC-SHA256签名
func verifySignature(secretKey, body, timestamp, signature string) bool {
	// 将时间戳转换为int64验证有效性
	_, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	// 计算签名: HMAC-SHA256(secretKey, body + timestamp)
	content := body + timestamp
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(content))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
