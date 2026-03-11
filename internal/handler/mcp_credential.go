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

// List 获取 MCP 凭证列表
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
