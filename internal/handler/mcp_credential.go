package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
)

type MCPCredentialHandler struct {
	mcpCredService *service.MCPCredentialService
}

func NewMCPCredentialHandler(mcpCredService *service.MCPCredentialService) *MCPCredentialHandler {
	return &MCPCredentialHandler{mcpCredService: mcpCredService}
}

// CreateMCPCredentialRequest 创建 MCP 凭证请求
type CreateMCPCredentialRequest struct {
	// 凭证名称 (必填，长度1-100)
	Name string `json:"name" binding:"required,min=1,max=100" example:"AI助手凭证"`
	// 权限级别 (必填，可选值: read_only, read_write)
	PermissionLevel string `json:"permission_level" binding:"required,oneof=read_only read_write" example:"read_only"`
	// 描述信息 (最大200字符)
	Description string `json:"description" binding:"max=200" example:"用于AI助手访问MCP服务"`
	// 过期时间 (可选，不填表示永不过期)
	ExpiresAt string `json:"expires_at" example:"2026-12-31T23:59:59Z"`
}

// UpdateMCPCredentialRequest 更新 MCP 凭证请求
type UpdateMCPCredentialRequest struct {
	// 凭证名称 (长度1-100)
	Name string `json:"name" binding:"omitempty,min=1,max=100" example:"AI助手凭证"`
	// 权限级别 (可选值: read_only, read_write)
	PermissionLevel string `json:"permission_level" binding:"omitempty,oneof=read_only read_write" example:"read_write"`
	// 描述信息 (最大200字符)
	Description string `json:"description" binding:"omitempty,max=200" example:"更新后的描述"`
	// 是否启用
	IsActive *bool `json:"is_active" example:"true"`
	// 过期时间
	ExpiresAt string `json:"expires_at" example:"2026-12-31T23:59:59Z"`
}

// MCPCredentialResponse MCP 凭证响应
type MCPCredentialResponse struct {
	// 凭证ID
	ID uint `json:"id" example:"1"`
	// 凭证名称
	Name string `json:"name" example:"AI助手凭证"`
	// AppKey (应用标识)
	AppKey string `json:"app_key" example:"ak_abc123def456"`
	// AppSecret (应用密钥，仅在创建/重新生成时返回)
	AppSecret string `json:"app_secret,omitempty" example:"sk_xyz789uvw012"`
	// 权限级别 (read_only, read_write)
	PermissionLevel string `json:"permission_level" example:"read_only"`
	// 描述信息
	Description string `json:"description" example:"用于AI助手访问MCP服务"`
	// 是否启用
	IsActive bool `json:"is_active" example:"true"`
	// 最后使用时间
	LastUsedAt string `json:"last_used_at,omitempty" example:"2026-03-12 10:30:00"`
	// 过期时间
	ExpiresAt string `json:"expires_at,omitempty" example:"2026-12-31 23:59:59"`
	// 创建人ID
	CreatedBy uint `json:"created_by" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2026-03-12 10:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2026-03-12 15:30:00"`
}

// List 获取 MCP 凭证列表
// @Summary 获取MCP凭证分页列表
// @Description 获取所有MCP服务认证凭证的分页列表，包含AppKey、权限级别等信息
// @Tags 管理员/MCP凭证管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.MCPCredential}} "获取成功，返回凭证分页列表"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/mcp-credentials [get]
func (h *MCPCredentialHandler) List(c *gin.Context) {
	page := middleware.ParseIntQuery(c, "page", 1)
	pageSize := middleware.ParseIntQuery(c, "page_size", 20)

	creds, total, err := h.mcpCredService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get MCP credentials")
		return
	}

	response.Page(c, creds, total, page, pageSize)
}

// Get 获取单个 MCP 凭证
// @Summary 获取单个MCP凭证详情
// @Description 根据ID获取指定MCP凭证的详细信息
// @Tags 管理员/MCP凭证管理
// @Accept json
// @Produce json
// @Param id path int true "凭证ID" minimum(1)
// @Success 200 {object} response.Response{data=model.MCPCredential} "获取成功，返回凭证详情"
// @Failure 400 {object} response.Response "无效的ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "凭证不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/mcp-credentials/{id} [get]
func (h *MCPCredentialHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	cred, err := h.mcpCredService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "MCP credential not found")
		return
	}

	response.Success(c, cred)
}

// Create 创建 MCP 凭证
// @Summary 创建MCP凭证
// @Description 创建新的MCP服务认证凭证，系统自动生成AppKey和AppSecret。注意：AppSecret仅在创建时返回一次，请妥善保存。
// @Tags 管理员/MCP凭证管理
// @Accept json
// @Produce json
// @Param request body handler.CreateMCPCredentialRequest true "创建凭证请求参数"
// @Success 200 {object} response.Response{data=handler.MCPCredentialResponse} "创建成功，返回凭证信息（含AppSecret）"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/mcp-credentials [post]
func (h *MCPCredentialHandler) Create(c *gin.Context) {
	var req service.CreateMCPCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 从上下文获取当前用户 ID
	if userID, exists := c.Get("user_id"); exists {
		req.CreatedBy = userID.(uint)
	}

	result, err := h.mcpCredService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result)
}

// Update 更新 MCP 凭证
// @Summary 更新MCP凭证
// @Description 更新指定MCP凭证的信息，如名称、权限级别、描述、启用状态等
// @Tags 管理员/MCP凭证管理
// @Accept json
// @Produce json
// @Param id path int true "凭证ID" minimum(1)
// @Param request body handler.UpdateMCPCredentialRequest true "更新凭证请求参数"
// @Success 200 {object} response.Response{data=model.MCPCredential} "更新成功，返回更新后的凭证信息"
// @Failure 400 {object} response.Response "请求参数错误或ID无效"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "凭证不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/mcp-credentials/{id} [put]
func (h *MCPCredentialHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req service.UpdateMCPCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	cred, err := h.mcpCredService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, cred)
}

// Delete 删除 MCP 凭证
// @Summary 删除MCP凭证
// @Description 删除指定的MCP服务认证凭证，删除后该凭证将无法继续使用
// @Tags 管理员/MCP凭证管理
// @Accept json
// @Produce json
// @Param id path int true "凭证ID" minimum(1)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "无效的ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "凭证不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/mcp-credentials/{id} [delete]
func (h *MCPCredentialHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.mcpCredService.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// RegenerateSecret 重新生成 AppSecret
// @Summary 重新生成AppSecret
// @Description 为指定MCP凭证重新生成AppSecret，旧的Secret将立即失效。注意：新的AppSecret仅在本次返回，请妥善保存。
// @Tags 管理员/MCP凭证管理
// @Accept json
// @Produce json
// @Param id path int true "凭证ID" minimum(1)
// @Success 200 {object} response.Response{data=handler.MCPCredentialResponse} "重新生成成功，返回新凭证信息（含新AppSecret）"
// @Failure 400 {object} response.Response "无效的ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "凭证不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/mcp-credentials/{id}/regenerate [post]
func (h *MCPCredentialHandler) RegenerateSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	result, err := h.mcpCredService.RegenerateSecret(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, result)
}
