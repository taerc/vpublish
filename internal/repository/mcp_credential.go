package repository

import (
	"context"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

type MCPCredentialRepository struct {
	db *gorm.DB
}

func NewMCPCredentialRepository(db *gorm.DB) *MCPCredentialRepository {
	return &MCPCredentialRepository{db: db}
}

func (r *MCPCredentialRepository) Create(ctx context.Context, cred *model.MCPCredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

func (r *MCPCredentialRepository) Update(ctx context.Context, cred *model.MCPCredential) error {
	return r.db.WithContext(ctx).Save(cred).Error
}

func (r *MCPCredentialRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.MCPCredential{}, id).Error
}

func (r *MCPCredentialRepository) GetByID(ctx context.Context, id uint) (*model.MCPCredential, error) {
	var cred model.MCPCredential
	err := r.db.WithContext(ctx).First(&cred, id).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *MCPCredentialRepository) GetByKey(ctx context.Context, appKey string) (*model.MCPCredential, error) {
	var cred model.MCPCredential
	err := r.db.WithContext(ctx).Where("app_key = ?", appKey).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *MCPCredentialRepository) List(ctx context.Context, page, pageSize int) ([]model.MCPCredential, int64, error) {
	var creds []model.MCPCredential
	var total int64

	db := r.db.WithContext(ctx).Model(&model.MCPCredential{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&creds).Error
	return creds, total, err
}

// ValidateCredential validates app key and secret, returns the credential if valid
func (r *MCPCredentialRepository) ValidateCredential(ctx context.Context, appKey, appSecret string) (*model.MCPCredential, error) {
	var cred model.MCPCredential
	err := r.db.WithContext(ctx).Where("app_key = ? AND app_secret = ?", appKey, appSecret).First(&cred).Error
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

// UpdateLastUsedAt updates the last used timestamp
func (r *MCPCredentialRepository) UpdateLastUsedAt(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.MCPCredential{}).Where("id = ?", id).Update("last_used_at", now).Error
}
