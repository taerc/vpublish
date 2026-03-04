package repository

import (
	"context"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

type PackageRepository struct {
	db *gorm.DB
}

func NewPackageRepository(db *gorm.DB) *PackageRepository {
	return &PackageRepository{db: db}
}

func (r *PackageRepository) Create(ctx context.Context, pkg *model.Package) error {
	return r.db.WithContext(ctx).Create(pkg).Error
}

func (r *PackageRepository) Update(ctx context.Context, pkg *model.Package) error {
	return r.db.WithContext(ctx).Save(pkg).Error
}

func (r *PackageRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Package{}, id).Error
}

func (r *PackageRepository) GetByID(ctx context.Context, id uint) (*model.Package, error) {
	var pkg model.Package
	err := r.db.WithContext(ctx).Preload("Category").Preload("Creator").First(&pkg, id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *PackageRepository) List(ctx context.Context, categoryID uint, page, pageSize int) ([]model.Package, int64, error) {
	var packages []model.Package
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Package{}).Preload("Category").Preload("Creator")
	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&packages).Error
	if err != nil {
		return nil, 0, err
	}

	// 加载每个软件包的最新版本信息
	for i := range packages {
		var latestVersion model.Version
		err := r.db.WithContext(ctx).
			Where("package_id = ? AND is_latest = ?", packages[i].ID, true).
			First(&latestVersion).Error
		if err == nil {
			packages[i].LatestVersion = &latestVersion
		}
	}

	return packages, total, nil
}

func (r *PackageRepository) ListByCategory(ctx context.Context, categoryID uint) ([]model.Package, error) {
	var packages []model.Package
	err := r.db.WithContext(ctx).Where("category_id = ? AND is_active = ?", categoryID, true).Find(&packages).Error
	return packages, err
}

func (r *PackageRepository) ListActive(ctx context.Context) ([]model.Package, error) {
	var packages []model.Package
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Preload("Category").Find(&packages).Error
	return packages, err
}

func (r *PackageRepository) GetByCategoryAndName(ctx context.Context, categoryID uint, name string) (*model.Package, error) {
	var pkg model.Package
	err := r.db.WithContext(ctx).Where("category_id = ? AND name = ?", categoryID, name).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}
