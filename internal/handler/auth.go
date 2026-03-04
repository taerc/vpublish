package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/jwt"
	"github.com/taerc/vpublish/pkg/response"
)

type AuthHandler struct {
	userService *service.UserService
	jwtService  *jwt.JWT
}

func NewAuthHandler(userService *service.UserService, jwtService *jwt.JWT) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Role     string `json:"role"`
	} `json:"user"`
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	// 生成 token
	token, err := h.jwtService.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.InternalError(c, "generate token failed")
		return
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID, user.Username, user.Role)
	if err != nil {
		response.InternalError(c, "generate refresh token failed")
		return
	}

	// 更新最后登录信息
	go func() {
		h.userService.UpdateLastLogin(c.Request.Context(), user.ID, middleware.GetClientIP(c))
	}()

	resp := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(24 * time.Hour.Seconds()),
	}
	resp.User.ID = user.ID
	resp.User.Username = user.Username
	resp.User.Nickname = user.Nickname
	resp.User.Role = user.Role

	response.Success(c, resp)
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		response.Unauthorized(c, "missing refresh token")
		return
	}

	claims, err := h.jwtService.ParseToken(refreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	token, err := h.jwtService.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		response.InternalError(c, "generate token failed")
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"expires_in": int64(24 * time.Hour.Seconds()),
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 无状态，服务端不需要特殊处理
	// 如果需要，可以将 token 加入黑名单
	response.Success(c, nil)
}

// GetProfile 获取当前用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"nickname":  user.Nickname,
		"email":     user.Email,
		"role":      user.Role,
		"is_active": user.IsActive,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}
