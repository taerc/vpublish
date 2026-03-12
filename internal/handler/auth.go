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

// LoginRequest 登录请求数据结构
type LoginRequest struct {
	// 用户名
	Username string `json:"username" binding:"required" example:"admin"`
	// 密码
	Password string `json:"password" binding:"required" example:"123456" swaggertype:"string" format:"password"`
}

// LoginResponse 登录响应数据结构
type LoginResponse struct {
	// JWT Token
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// 刷新Token
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// Token过期时间(秒)
	ExpiresIn int64 `json:"expires_in" example:"86400"`
	// 用户信息
	User struct {
		// 用户ID
		ID uint `json:"id" example:"1"`
		// 用户名
		Username string `json:"username" example:"admin"`
		// 昵称
		Nickname string `json:"nickname" example:"管理员"`
		// 角色
		Role string `json:"role" example:"admin"`
	} `json:"user"`
}

// ProfileResponse 用户信息响应数据结构
type ProfileResponse struct {
	// 用户ID
	ID uint `json:"id" example:"1"`
	// 用户名
	Username string `json:"username" example:"admin"`
	// 昵称
	Nickname string `json:"nickname" example:"管理员"`
	// 邮箱
	Email string `json:"email" example:"admin@example.com"`
	// 角色
	Role string `json:"role" example:"admin"`
	// 是否启用 (1:启用, 0:禁用)
	IsActive int8 `json:"is_active" example:"1"`
}

// Login 登录
//
// @Summary 用户登录
// @Description 用户通过用户名和密码完成身份认证，成功后返回JWT token和用户信息
// @Tags 管理员/认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=LoginResponse} "登录成功，返回token和用户信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "认证失败，用户名或密码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/auth/login [post]
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
//
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌，刷新令牌通过请求头X-Refresh-Token传递
// @Tags 管理员/认证
// @Accept json
// @Produce json
// @Param X-Refresh-Token header string true "刷新令牌"
// @Success 200 {object} response.Response{data=map[string]interface{}} "刷新成功，返回新的访问令牌"
// @Failure 401 {object} response.Response "刷新令牌无效或已过期"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/auth/refresh [post]
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
//
// @Summary 用户登出
// @Description 用户登出系统，由于JWT是无状态认证，服务端不需要特殊处理。客户端应删除本地存储的token。
// @Tags 管理员/认证
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "登出成功"
// @Router /admin/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT 无状态，服务端不需要特殊处理
	// 如果需要，可以将 token 加入黑名单
	response.Success(c, nil)
}

// GetProfile 获取当前用户信息
//
// @Summary 获取当前登录用户信息
// @Description 获取当前已登录用户的详细信息，包括用户名、昵称、邮箱、角色等
// @Tags 管理员/认证
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=ProfileResponse} "获取成功，返回用户信息"
// @Failure 401 {object} response.Response "未认证，需要有效的JWT token"
// @Failure 404 {object} response.Response "用户不存在"
// @Security BearerAuth []
// @Router /admin/auth/profile [get]
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
//
// @Summary 修改当前用户密码
// @Description 用户修改自己的登录密码，需要提供旧密码和新密码。新密码长度要求8-72个字符。
// @Tags 管理员/认证
// @Accept json
// @Produce json
// @Param request body service.ChangePasswordRequest true "修改密码请求参数"
// @Success 200 {object} response.Response "密码修改成功"
// @Failure 400 {object} response.Response "请求参数错误或旧密码不正确"
// @Failure 401 {object} response.Response "未认证，需要有效的JWT token"
// @Security BearerAuth []
// @Router /admin/auth/password [put]
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
