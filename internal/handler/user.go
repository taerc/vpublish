package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/password"
	"github.com/taerc/vpublish/pkg/response"
)

// UserListRequest 用户列表请求参数
type UserListRequest struct {
	// 页码
	Page int `form:"page" json:"page" example:"1"`
	// 每页数量
	PageSize int `form:"page_size" json:"page_size" example:"20"`
}

// UserCreateRequest 创建用户请求参数
type UserCreateRequest struct {
	// 用户名 (必填，长度3-50)
	Username string `json:"username" binding:"required,min=3,max=50" example:"zhangsan"`
	// 密码 (必填，长度8-72，需符合复杂度要求)
	Password string `json:"password" binding:"required,min=8,max=72" example:"Password@123" swaggertype:"string" format:"password"`
	// 昵称
	Nickname string `json:"nickname" example:"张三"`
	// 邮箱
	Email string `json:"email" example:"zhangsan@example.com"`
	// 角色 (admin/user)
	Role string `json:"role" example:"user"`
}

// UserUpdateRequest 更新用户请求参数
type UserUpdateRequest struct {
	// 昵称
	Nickname string `json:"nickname" example:"李四"`
	// 邮箱
	Email string `json:"email" example:"lisi@example.com"`
	// 角色 (admin/user)
	Role string `json:"role" example:"admin"`
	// 是否激活
	IsActive *bool `json:"is_active" example:"true"`
}

// UserResetPasswordRequest 重置密码请求参数
type UserResetPasswordRequest struct {
	// 新密码 (必填，长度8-72，需符合复杂度要求)
	Password string `json:"password" binding:"required,min=8,max=72" example:"NewPassword@456" swaggertype:"string" format:"password"`
}

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// List 用户列表
//
// @Summary 获取用户分页列表
// @Description 获取管理员用户分页列表，支持分页参数
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.User}} "获取成功，返回用户分页列表"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/users [get]
func (h *UserHandler) List(c *gin.Context) {
	page := middleware.ParseIntQuery(c, "page", 1)
	pageSize := middleware.ParseIntQuery(c, "page_size", 20)

	users, total, err := h.userService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get users")
		return
	}

	// 清除敏感信息
	for i := range users {
		users[i].PasswordHash = ""
	}

	response.Page(c, users, total, page, pageSize)
}

// Get 获取单个用户
//
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID" minimum(1)
// @Success 200 {object} response.Response{data=model.User} "获取成功，返回用户详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	user.PasswordHash = ""
	response.Success(c, user)
}

// Create 创建用户
//
// @Summary 创建用户
// @Description 创建新的管理员用户，用户名必须唯一，密码需符合复杂度要求（至少8位，包含大小写字母、数字和特殊字符）
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param request body handler.UserCreateRequest true "创建用户请求参数"
// @Success 200 {object} response.Response{data=model.User} "创建成功，返回新用户信息"
// @Failure 400 {object} response.Response "请求参数错误或用户名已存在"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	user.PasswordHash = ""
	response.Success(c, user)
}

// Update 更新用户
//
// @Summary 更新用户信息
// @Description 更新指定用户的信息，包括昵称、邮箱、角色和激活状态
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID" minimum(1)
// @Param request body handler.UserUpdateRequest true "更新用户请求参数"
// @Success 200 {object} response.Response{data=model.User} "更新成功，返回更新后的用户信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	user, err := h.userService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	user.PasswordHash = ""
	response.Success(c, user)
}

// Delete 删除用户
//
// @Summary 删除用户
// @Description 删除指定的管理员用户，不允许删除自己
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID" minimum(1)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误或尝试删除自己"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	// 不能删除自己
	currentUserID := middleware.GetUserID(c)
	if uint(id) == currentUserID {
		response.BadRequest(c, "cannot delete yourself")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetPassword 重置密码
//
// @Summary 重置用户密码
// @Description 重置指定用户的密码，新密码需符合复杂度要求（至少8位，包含大小写字母、数字和特殊字符）
// @Tags 管理员/用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID" minimum(1)
// @Param request body handler.UserResetPasswordRequest true "重置密码请求参数"
// @Success 200 {object} response.Response "重置成功"
// @Failure 400 {object} response.Response "请求参数错误或密码不符合复杂度要求"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/users/{id}/password [put]
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=8,max=72"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	// 验证密码复杂度
	if err := password.Validate(req.Password); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.ResetPassword(c.Request.Context(), uint(id), req.Password); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}
