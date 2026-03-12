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
//
// @Summary 获取AppKey列表
// @Description 获取AppKey分页列表，用于管理APP端认证密钥
// @Tags 管理员/AppKey管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.AppKey}} "获取成功，返回AppKey分页列表"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/appkeys [get]
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
//
// @Summary 获取AppKey详情
// @Description 根据ID获取单个AppKey的详细信息
// @Tags 管理员/AppKey管理
// @Accept json
// @Produce json
// @Param id path int true "AppKey ID" minimum(1)
// @Success 200 {object} response.Response{data=model.AppKey} "获取成功，返回AppKey详情"
// @Failure 400 {object} response.Response "请求参数错误，无效的ID"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "AppKey不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/appkeys/{id} [get]
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
//
// @Summary 创建AppKey
// @Description 创建新的APP认证密钥，系统自动生成AppKey和AppSecret。注意：AppSecret仅在创建时返回一次，请妥善保存。
// @Tags 管理员/AppKey管理
// @Accept json
// @Produce json
// @Param request body service.CreateAppKeyRequest true "创建AppKey请求参数"
// @Success 200 {object} response.Response{data=service.AppKeyResponse} "创建成功，返回AppKey信息（包含AppSecret）"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/appkeys [post]
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
//
// @Summary 更新AppKey
// @Description 更新AppKey信息，可修改应用名称、描述和启用状态
// @Tags 管理员/AppKey管理
// @Accept json
// @Produce json
// @Param id path int true "AppKey ID" minimum(1)
// @Param request body service.UpdateAppKeyRequest true "更新AppKey请求参数"
// @Success 200 {object} response.Response{data=model.AppKey} "更新成功，返回更新后的AppKey信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "AppKey不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/appkeys/{id} [put]
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
//
// @Summary 删除AppKey
// @Description 删除指定的AppKey，删除后该密钥将无法使用
// @Tags 管理员/AppKey管理
// @Accept json
// @Produce json
// @Param id path int true "AppKey ID" minimum(1)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误，无效的ID"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "AppKey不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/appkeys/{id} [delete]
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
//
// @Summary 重新生成AppSecret
// @Description 为指定AppKey重新生成AppSecret，旧的Secret将立即失效。注意：新的AppSecret仅在此次返回，请妥善保存。
// @Tags 管理员/AppKey管理
// @Accept json
// @Produce json
// @Param id path int true "AppKey ID" minimum(1)
// @Success 200 {object} response.Response{data=service.AppKeyResponse} "重新生成成功，返回新的AppSecret"
// @Failure 400 {object} response.Response "请求参数错误，无效的ID"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "AppKey不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/appkeys/{id}/regenerate [post]
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
