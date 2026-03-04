package repository

import (
	"context"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

type AppKeyRepository struct {
	db *gorm.DB
}

func NewAppKeyRepository(db *gorm.DB) *AppKeyRepository {
	return &AppKeyRepository{db: db}
}

func (r *AppKeyRepository) Create(ctx context.Context, appKey *model.AppKey) error {
	return r.db.WithContext(ctx).Create(appKey).Error
}

func (r *AppKeyRepository) Update(ctx context.Context, appKey *model.AppKey) error {
	return r.db.WithContext(ctx).Save(appKey).Error
}

func (r *AppKeyRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AppKey{}, id).Error
}

func (r *AppKeyRepository) GetByID(ctx context.Context, id uint) (*model.AppKey, error) {
	var appKey model.AppKey
	err := r.db.WithContext(ctx).First(&appKey, id).Error
	if err != nil {
		return nil, err
	}
	return &appKey, nil
}

func (r *AppKeyRepository) GetByKey(ctx context.Context, key string) (*model.AppKey, error) {
	var appKey model.AppKey
	err := r.db.WithContext(ctx).Where("app_key = ?", key).First(&appKey).Error
	if err != nil {
		return nil, err
	}
	return &appKey, nil
}

func (r *AppKeyRepository) List(ctx context.Context, page, pageSize int) ([]model.AppKey, int64, error) {
	var keys []model.AppKey
	var total int64

	db := r.db.WithContext(ctx).Model(&model.AppKey{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&keys).Error
	return keys, total, err
}
