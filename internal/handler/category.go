package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
)

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List 类别列表
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
func (h *CategoryHandler) ListActive(c *gin.Context) {
	categories, err := h.categoryService.ListActive(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get categories")
		return
	}

	response.Success(c, categories)
}

// Get 获取单个类别
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
func (h *CategoryHandler) Create(c *gin.Context) {
	var req service.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	category, err := h.categoryService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, category)
}

// Update 更新类别
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req service.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	category, err := h.categoryService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, category)
}

// Delete 删除类别
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
