package repository

import (
	"context"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) GetByCode(ctx context.Context, code string) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) GetByName(ctx context.Context, name string) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) List(ctx context.Context, page, pageSize int) ([]model.Category, int64, error) {
	var categories []model.Category
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Category{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, total, err
}

func (r *CategoryRepository) ListActive(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Category{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

func (r *CategoryRepository) CountPackagesByCategory(ctx context.Context, categoryID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Package{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}
