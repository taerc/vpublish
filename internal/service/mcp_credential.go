package service

import (
	"context"
	"errors"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/utils"
)

var (
	ErrMCPCredentialNotFound      = errors.New("MCP credential not found")
	ErrMCPCredentialAlreadyExists = errors.New("MCP credential already exists")
	ErrMCPCredentialInvalid       = errors.New("MCP credential is invalid or expired")
	ErrMCPCredentialNoPermission  = errors.New("MCP credential does not have write permission")
)

type MCPCredentialService struct {
	credRepo *repository.MCPCredentialRepository
}

func NewMCPCredentialService(credRepo *repository.MCPCredentialRepository) *MCPCredentialService {
	return &MCPCredentialService{credRepo: credRepo}
}

// CreateMCPCredentialRequest 创建 MCP 凭证请求
type CreateMCPCredentialRequest struct {
	Name            string     `json:"name" binding:"required,min=1,max=100"`
	PermissionLevel string     `json:"permission_level" binding:"required,oneof=read_only read_write"`
	Description     string     `json:"description" binding:"max=200"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedBy       uint       `json:"created_by"`
}

// UpdateMCPCredentialRequest 更新 MCP 凭证请求
type UpdateMCPCredentialRequest struct {
	Name            string     `json:"name" binding:"omitempty,min=1,max=100"`
	PermissionLevel string     `json:"permission_level" binding:"omitempty,oneof=read_only read_write"`
	Description     string     `json:"description" binding:"omitempty,max=200"`
	IsActive        *bool      `json:"is_active"`
	ExpiresAt       *time.Time `json:"expires_at"`
}

// MCPCredentialResponse MCP 凭证响应（包含 Secret，仅在创建/重新生成时返回）
type MCPCredentialResponse struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	AppKey          string `json:"app_key"`
	AppSecret       string `json:"app_secret,omitempty"`
	PermissionLevel string `json:"permission_level"`
	Description     string `json:"description"`
	IsActive        bool   `json:"is_active"`
	LastUsedAt      string `json:"last_used_at,omitempty"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	CreatedBy       uint   `json:"created_by"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// Create 创建新的 MCP 凭证
func (s *MCPCredentialService) Create(ctx context.Context, req *CreateMCPCredentialRequest) (*MCPCredentialResponse, error) {
	// 生成 AppKey 和 AppSecret
	appKey, err := utils.GenerateAppKey()
	if err != nil {
		return nil, err
	}

	appSecret, err := utils.GenerateAppSecret()
	if err != nil {
		return nil, err
	}

	// 验证权限级别
	if req.PermissionLevel != model.PermissionReadOnly && req.PermissionLevel != model.PermissionReadWrite {
		req.PermissionLevel = model.PermissionReadOnly
	}

	record := &model.MCPCredential{
		Name:            req.Name,
		AppKey:          appKey,
		AppSecret:       appSecret,
		PermissionLevel: req.PermissionLevel,
		Description:     req.Description,
		IsActive:        true,
		ExpiresAt:       req.ExpiresAt,
		CreatedBy:       req.CreatedBy,
	}

	if err := s.credRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return &MCPCredentialResponse{
		ID:              record.ID,
		Name:            record.Name,
		AppKey:          record.AppKey,
		AppSecret:       appSecret, // 创建时返回 Secret
		PermissionLevel: record.PermissionLevel,
		Description:     record.Description,
		IsActive:        record.IsActive,
		ExpiresAt:       formatTime(record.ExpiresAt),
		CreatedBy:       record.CreatedBy,
		CreatedAt:       record.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       record.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Update 更新 MCP 凭证
func (s *MCPCredentialService) Update(ctx context.Context, id uint, req *UpdateMCPCredentialRequest) (*model.MCPCredential, error) {
	record, err := s.credRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrMCPCredentialNotFound
	}

	if req.Name != "" {
		record.Name = req.Name
	}
	if req.PermissionLevel != "" {
		record.PermissionLevel = req.PermissionLevel
	}
	if req.Description != "" {
		record.Description = req.Description
	}
	if req.IsActive != nil {
		record.IsActive = *req.IsActive
	}
	if req.ExpiresAt != nil {
		record.ExpiresAt = req.ExpiresAt
	}

	if err := s.credRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// Delete 删除 MCP 凭证
func (s *MCPCredentialService) Delete(ctx context.Context, id uint) error {
	return s.credRepo.Delete(ctx, id)
}

// GetByID 根据 ID 获取 MCP 凭证
func (s *MCPCredentialService) GetByID(ctx context.Context, id uint) (*model.MCPCredential, error) {
	return s.credRepo.GetByID(ctx, id)
}

// List 获取 MCP 凭证列表
func (s *MCPCredentialService) List(ctx context.Context, page, pageSize int) ([]model.MCPCredential, int64, error) {
	return s.credRepo.List(ctx, page, pageSize)
}

// RegenerateSecret 重新生成 AppSecret
func (s *MCPCredentialService) RegenerateSecret(ctx context.Context, id uint) (*MCPCredentialResponse, error) {
	record, err := s.credRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrMCPCredentialNotFound
	}

	// 生成新的 AppSecret
	newSecret, err := utils.GenerateAppSecret()
	if err != nil {
		return nil, err
	}

	record.AppSecret = newSecret
	if err := s.credRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	return &MCPCredentialResponse{
		ID:              record.ID,
		Name:            record.Name,
		AppKey:          record.AppKey,
		AppSecret:       newSecret, // 重新生成时返回新 Secret
		PermissionLevel: record.PermissionLevel,
		Description:     record.Description,
		IsActive:        record.IsActive,
		LastUsedAt:      formatTime(record.LastUsedAt),
		ExpiresAt:       formatTime(record.ExpiresAt),
		CreatedBy:       record.CreatedBy,
		CreatedAt:       record.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       record.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Validate 验证 MCP 凭证，返回凭证信息和是否有效
func (s *MCPCredentialService) Validate(ctx context.Context, appKey, appSecret string) (*model.MCPCredential, error) {
	record, err := s.credRepo.ValidateCredential(ctx, appKey, appSecret)
	if err != nil {
		return nil, ErrMCPCredentialInvalid
	}

	if !record.IsValid() {
		return nil, ErrMCPCredentialInvalid
	}

	// 更新最后使用时间
	go func() {
		// 异步更新，不阻塞主流程
		_ = s.credRepo.UpdateLastUsedAt(context.Background(), record.ID)
	}()

	return record, nil
}

// ValidateAndCheckPermission 验证凭证并检查权限
func (s *MCPCredentialService) ValidateAndCheckPermission(ctx context.Context, appKey, appSecret, requiredPermission string) (*model.MCPCredential, error) {
	record, err := s.Validate(ctx, appKey, appSecret)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if requiredPermission == model.PermissionReadWrite && !record.CanWrite() {
		return nil, ErrMCPCredentialNoPermission
	}

	return record, nil
}

// formatTime 格式化时间指针
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
