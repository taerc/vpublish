package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
)

type AppKeyHandler struct {
	appKeyService *service.AppKeyService
}

func NewAppKeyHandler(appKeyService *service.AppKeyService) *AppKeyHandler {
	return &AppKeyHandler{appKeyService: appKeyService}
}

// List 获取 AppKey 列表
func (h *AppKeyHandler) List(c *gin.Context) {
	page := middleware.ParseIntQuery(c, "page", 1)
	pageSize := middleware.ParseIntQuery(c, "page_size", 20)

	keys, total, err := h.appKeyService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get app keys")
		return
	}

	response.Page(c, keys, total, page, pageSize)
}

// Get 获取单个 AppKey
func (h *AppKeyHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	key, err := h.appKeyService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "app key not found")
		return
	}

	response.Success(c, key)
}

// Create 创建 AppKey
func (h *AppKeyHandler) Create(c *gin.Context) {
	var req service.CreateAppKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	result, err := h.appKeyService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result)
}

// Update 更新 AppKey
func (h *AppKeyHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req service.UpdateAppKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	key, err := h.appKeyService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, key)
}

// Delete 删除 AppKey
func (h *AppKeyHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.appKeyService.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// RegenerateSecret 重新生成 AppSecret
func (h *AppKeyHandler) RegenerateSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	result, err := h.appKeyService.RegenerateSecret(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result)
}
