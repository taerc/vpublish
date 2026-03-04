package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/response"
	"github.com/taerc/vpublish/pkg/signature"
)

// SignatureAuth APP端签名认证中间件
func SignatureAuth(appKeyRepo *repository.AppKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		appKey := c.GetHeader(signature.HeaderAppKey)
		timestampStr := c.GetHeader(signature.HeaderTimestamp)
		clientSignature := c.GetHeader(signature.HeaderSignature)

		if appKey == "" || timestampStr == "" || clientSignature == "" {
			response.Unauthorized(c, "missing signature headers")
			c.Abort()
			return
		}

		// 解析时间戳
		timestamp, err := time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			response.Unauthorized(c, "invalid timestamp format")
			c.Abort()
			return
		}

		// 检查时间戳是否在有效期内（允许前后 SignatureExpire 秒的误差）
		now := time.Now()
		diff := now.Sub(timestamp)
		if diff < 0 {
			diff = -diff // 取绝对值
		}
		if diff > signature.SignatureExpire*time.Second {
			response.Unauthorized(c, "signature expired")
			c.Abort()
			return
		}

		// 查询 AppKey
		appKeyRecord, err := appKeyRepo.GetByKey(c.Request.Context(), appKey)
		if err != nil {
			response.Unauthorized(c, "invalid app key")
			c.Abort()
			return
		}

		if !appKeyRecord.IsActive {
			response.Unauthorized(c, "app key is disabled")
			c.Abort()
			return
		}

		// 构建参数map
		params := make(map[string]string)
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}

		// 验证签名
		if !signature.VerifySignature(params, appKeyRecord.AppSecret, timestamp.Unix(), clientSignature) {
			response.Unauthorized(c, "invalid signature")
			c.Abort()
			return
		}

		// 存储应用信息到 context
		c.Set("app_key", appKeyRecord)
		c.Set("app_name", appKeyRecord.AppName)

		c.Next()
	}
}

// GetAppKey 从 context 获取 AppKey
func GetAppKey(c *gin.Context) *model.AppKey {
	appKey, _ := c.Get("app_key")
	if appKey == nil {
		return nil
	}
	return appKey.(*model.AppKey)
}
