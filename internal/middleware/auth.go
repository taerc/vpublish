package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/pkg/jwt"
	"github.com/taerc/vpublish/pkg/response"
)

func JWTAuth(jwtService *jwt.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取 token
		token := c.GetHeader("Authorization")
		if token == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// 去掉 Bearer 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 解析 token
		claims, err := jwtService.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// 存储用户信息到 context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminOnly 仅管理员可访问
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Forbidden(c, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID 从 context 获取用户ID
func GetUserID(c *gin.Context) uint {
	userID, _ := c.Get("user_id")
	return userID.(uint)
}

// GetUsername 从 context 获取用户名
func GetUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	return username.(string)
}

// GetRole 从 context 获取角色
func GetRole(c *gin.Context) string {
	role, _ := c.Get("role")
	return role.(string)
}

// GetCurrentTime 获取当前时间
func GetCurrentTime() time.Time {
	return time.Now()
}

// GetClientIP 获取客户端IP
func GetClientIP(c *gin.Context) string {
	// 先检查 X-Forwarded-For
	ip := c.GetHeader("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	// 再检查 X-Real-IP
	ip = c.GetHeader("X-Real-IP")
	if ip != "" {
		return ip
	}
	// 最后使用 RemoteAddr
	return c.ClientIP()
}

// ParseIntParam 解析整数参数
func ParseIntParam(c *gin.Context, key string) (int, error) {
	val := c.Param(key)
	return strconv.Atoi(val)
}

// ParseIntQuery 解析整数查询参数
func ParseIntQuery(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return result
}
