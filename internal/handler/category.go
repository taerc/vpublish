package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
)

// CreateCategoryRequest 创建类别请求参数
type CreateCategoryRequest struct {
	// 类别名称（必填，长度3-100字符）
	Name string `json:"name" binding:"required,min=3,max=100" example:"无人机应用"`
	// 描述信息（最大500字符）
	Description string `json:"description,omitempty" binding:"max=500" example:"包含各类无人机飞行、遥控、管理相关应用软件"`
	// 排序权重（数值越大越靠后）
	SortOrder int `json:"sort_order,omitempty" example:"100"`
}

// UpdateCategoryRequest 更新类别请求参数
type UpdateCategoryRequest struct {
	// 类别名称（长度3-100字符）
	Name string `json:"name,omitempty" binding:"omitempty,min=3,max=100" example:"无人机应用"`
	// 描述信息（最大500字符）
	Description string `json:"description,omitempty" binding:"max=500" example:"包含各类无人机飞行、遥控、管理相关应用软件"`
	// 排序权重（数值越大越靠后）
	SortOrder *int `json:"sort_order,omitempty" example:"100"`
	// 是否启用
	IsActive *bool `json:"is_active,omitempty" example:"true"`
}

// CategoryHandler 类别处理器
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler 创建类别处理器
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List 类别列表
//
// @Summary 获取类别分页列表
// @Description 获取软件类别的分页列表，支持按页码和每页数量进行分页查询
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(50)
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.Category}} "获取成功，返回类别分页列表"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	page := middleware.ParseIntQuery(c, "page", 1)
	pageSize := middleware.ParseIntQuery(c, "page_size", 50)

	categories, total, err := h.categoryService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get categories")
		return
	}

	response.Page(c, categories, total, page, pageSize)
}

// ListActive 获取启用的类别列表
//
// @Summary 获取所有启用的类别列表
// @Description APP端获取所有启用状态的软件类别，用于展示可用的软件种类
// @Tags APP端
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]model.Category} "返回启用的类别列表"
// @Failure 401 {object} response.Response "认证失败，AppKey或签名验证不正确"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security SignatureAuth []
// @Security TimestampAuth []
// @Security SignatureValueAuth []
// @Router /app/categories [get]
func (h *CategoryHandler) ListActive(c *gin.Context) {
	categories, err := h.categoryService.ListActive(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get categories")
		return
	}

	response.Success(c, categories)
}

// Get 获取单个类别
//
// @Summary 根据ID获取类别详情
// @Description 通过ID获取指定软件类别的详细信息
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param id path int true "类别ID" minimum(1) example(1)
// @Success 200 {object} response.Response{data=model.Category} "返回类别详细信息"
// @Failure 400 {object} response.Response "请求参数错误，无效的类别ID"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	category, err := h.categoryService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "category not found")
		return
	}

	response.Success(c, category)
}

// GetByCode 根据代码获取类别
//
// @Summary 根据代码获取类别详情
// @Description 通过类别代码获取指定软件类别的详细信息
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param code path string true "类别代码" example(TYPE_WU_REN_JI_YING_YONG)
// @Success 200 {object} response.Response{data=model.Category} "返回类别详细信息"
// @Failure 400 {object} response.Response "请求参数错误，代码不能为空"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/categories/code/{code} [get]
func (h *CategoryHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "code is required")
		return
	}

	category, err := h.categoryService.GetByCode(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, "category not found")
		return
	}

	response.Success(c, category)
}

// Create 创建类别
//
// @Summary 创建软件类别
// @Description 创建一个新的软件类别，系统会自动根据中文名称生成拼音代码（如：无人机 -> TYPE_WU_REN_JI）
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true "创建类别的请求参数"
// @Success 200 {object} response.Response{data=model.Category} "创建成功，返回新创建的类别信息"
// @Failure 400 {object} response.Response "请求参数错误，例如参数未满足约束条件或类别名称已存在"
// @Failure 401 {object} response.Response "未认证，需要有效的管理员Token"
// @Failure 403 {object} response.Response "权限不足，当前用户角色不允许此操作"
// @Failure 409 {object} response.Response "冲突，相同名称的类别已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 转换为service层请求结构
	svcReq := &service.CreateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	category, err := h.categoryService.Create(c.Request.Context(), svcReq)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, category)
}

// Update 更新类别
//
// @Summary 更新软件类别
// @Description 更新指定ID的软件类别信息，如果更新名称则系统会自动重新生成拼音代码
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param id path int true "类别ID" minimum(1) example(1)
// @Param request body UpdateCategoryRequest true "更新类别的请求参数"
// @Success 200 {object} response.Response{data=model.Category} "更新成功，返回更新后的类别信息"
// @Failure 400 {object} response.Response "请求参数错误，无效的类别ID或参数不满足约束条件"
// @Failure 401 {object} response.Response "未认证，需要有效的管理员Token"
// @Failure 403 {object} response.Response "权限不足，当前用户角色不允许此操作"
// @Failure 404 {object} response.Response "类别不存在"
// @Failure 409 {object} response.Response "冲突，相同名称的类别已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	// 转换为service层请求结构
	svcReq := &service.UpdateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}

	category, err := h.categoryService.Update(c.Request.Context(), uint(id), svcReq)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, category)
}

// Delete 删除类别
//
// @Summary 删除软件类别
// @Description 删除指定ID的软件类别，软删除操作
// @Tags 管理员/类别管理
// @Accept json
// @Produce json
// @Param id path int true "类别ID" minimum(1) example(1)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误，无效的类别ID"
// @Failure 401 {object} response.Response "未认证，需要有效的管理员Token"
// @Failure 403 {object} response.Response "权限不足，当前用户角色不允许此操作"
// @Failure 404 {object} response.Response "类别不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.categoryService.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}
