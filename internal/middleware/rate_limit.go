package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 创建基于滑动窗口的限流中间件
// rps: 每秒允许的请求数
// burst: 突发情况下允许的最大请求数
func RateLimiter(rps int, burst int) gin.HandlerFunc {
	limiter := NewSimpleLimiter(rps, burst)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"code":    429,
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SimpleLimiter 简单的限流器实现
type SimpleLimiter struct {
	mu      sync.Mutex
	rps     int        // 每秒请求数
	burst   int        // 突发请求数
	tokens  int        // 当前可用令牌数
	lastRef time.Time  // 上次刷新时间
}

// NewSimpleLimiter 创建简单限流器
func NewSimpleLimiter(rps, burst int) *SimpleLimiter {
	return &SimpleLimiter{
		rps:     rps,
		burst:   burst,
		tokens:  burst,
		lastRef: time.Now(),
	}
}

// Allow 检查是否允许通过
func (l *SimpleLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastRef).Seconds()

	// 补充令牌
	if elapsed > 0 {
		newTokens := int(elapsed * float64(l.rps))
		l.tokens += newTokens
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
		l.lastRef = now
	}

	// 检查是否有可用令牌
	if l.tokens > 0 {
		l.tokens--
		return true
	}

	return false
}

// IPRateLimiter 基于IP的限流器（可选）
// 为每个IP地址创建独立的限流器
type IPRateLimiter struct {
	ips map[string]*SimpleLimiter
	mu  sync.RWMutex
	rps int
	b   int
}

func NewIPRateLimiter(rps int, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*SimpleLimiter),
		rps: rps,
		b:   burst,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *SimpleLimiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = NewSimpleLimiter(i.rps, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// GlobalIPRateLimiter 全局IP限流器实例
var GlobalIPRateLimiter *IPRateLimiter

func InitIPRateLimiter(rps int, burst int) {
	GlobalIPRateLimiter = NewIPRateLimiter(rps, burst)
}

// IPMiddleware 基于IP的限流中间件
func IPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := GlobalIPRateLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"code":    429,
				"message": "Too many requests from this IP, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}