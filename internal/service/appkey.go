package service

import (
	"context"
	"errors"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/utils"
)

var (
	ErrAppKeyNotFound      = errors.New("app key not found")
	ErrAppKeyAlreadyExists = errors.New("app key already exists")
)

type AppKeyService struct {
	appKeyRepo *repository.AppKeyRepository
}

func NewAppKeyService(appKeyRepo *repository.AppKeyRepository) *AppKeyService {
	return &AppKeyService{appKeyRepo: appKeyRepo}
}

// CreateAppKeyRequest 创建 AppKey 请求
type CreateAppKeyRequest struct {
	AppName     string `json:"app_name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=200"`
}

// UpdateAppKeyRequest 更新 AppKey 请求
type UpdateAppKeyRequest struct {
	AppName     string `json:"app_name" binding:"omitempty,min=1,max=100"`
	Description string `json:"description" binding:"omitempty,max=200"`
	IsActive    *bool  `json:"is_active"`
}

// AppKeyResponse AppKey 响应（包含 Secret，仅在创建/重新生成时返回）
type AppKeyResponse struct {
	ID          uint   `json:"id"`
	AppName     string `json:"app_name"`
	AppKey      string `json:"app_key"`
	AppSecret   string `json:"app_secret,omitempty"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// Create 创建新的 AppKey
func (s *AppKeyService) Create(ctx context.Context, req *CreateAppKeyRequest) (*AppKeyResponse, error) {
	// 生成 AppKey 和 AppSecret
	appKey, err := utils.GenerateAppKey()
	if err != nil {
		return nil, err
	}

	appSecret, err := utils.GenerateAppSecret()
	if err != nil {
		return nil, err
	}

	record := &model.AppKey{
		AppName:     req.AppName,
		AppKey:      appKey,
		AppSecret:   appSecret,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.appKeyRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return &AppKeyResponse{
		ID:          record.ID,
		AppName:     record.AppName,
		AppKey:      record.AppKey,
		AppSecret:   appSecret, // 创建时返回 Secret
		Description: record.Description,
		IsActive:    record.IsActive,
		CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// Update 更新 AppKey
func (s *AppKeyService) Update(ctx context.Context, id uint, req *UpdateAppKeyRequest) (*model.AppKey, error) {
	record, err := s.appKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAppKeyNotFound
	}

	if req.AppName != "" {
		record.AppName = req.AppName
	}
	if req.Description != "" {
		record.Description = req.Description
	}
	if req.IsActive != nil {
		record.IsActive = *req.IsActive
	}

	if err := s.appKeyRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// Delete 删除 AppKey
func (s *AppKeyService) Delete(ctx context.Context, id uint) error {
	return s.appKeyRepo.Delete(ctx, id)
}

// GetByID 根据 ID 获取 AppKey
func (s *AppKeyService) GetByID(ctx context.Context, id uint) (*model.AppKey, error) {
	return s.appKeyRepo.GetByID(ctx, id)
}

// List 获取 AppKey 列表
func (s *AppKeyService) List(ctx context.Context, page, pageSize int) ([]model.AppKey, int64, error) {
	return s.appKeyRepo.List(ctx, page, pageSize)
}

// RegenerateSecret 重新生成 AppSecret
func (s *AppKeyService) RegenerateSecret(ctx context.Context, id uint) (*AppKeyResponse, error) {
	record, err := s.appKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrAppKeyNotFound
	}

	// 生成新的 AppSecret
	newSecret, err := utils.GenerateAppSecret()
	if err != nil {
		return nil, err
	}

	record.AppSecret = newSecret
	if err := s.appKeyRepo.Update(ctx, record); err != nil {
		return nil, err
	}

	return &AppKeyResponse{
		ID:          record.ID,
		AppName:     record.AppName,
		AppKey:      record.AppKey,
		AppSecret:   newSecret, // 重新生成时返回新 Secret
		Description: record.Description,
		IsActive:    record.IsActive,
		CreatedAt:   record.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   record.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
