package repository

import (
	"context"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

type VersionRepository struct {
	db *gorm.DB
}

func NewVersionRepository(db *gorm.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

func (r *VersionRepository) Create(ctx context.Context, version *model.Version) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *VersionRepository) Update(ctx context.Context, version *model.Version) error {
	return r.db.WithContext(ctx).Save(version).Error
}

func (r *VersionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Version{}, id).Error
}

func (r *VersionRepository) GetByID(ctx context.Context, id uint) (*model.Version, error) {
	var version model.Version
	err := r.db.WithContext(ctx).Preload("Package").Preload("Package.Category").First(&version, id).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *VersionRepository) GetByPackageAndVersion(ctx context.Context, packageID uint, version string) (*model.Version, error) {
	var v model.Version
	err := r.db.WithContext(ctx).Where("package_id = ? AND version = ?", packageID, version).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VersionRepository) ListByPackage(ctx context.Context, packageID uint, page, pageSize int) ([]model.Version, int64, error) {
	var versions []model.Version
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Version{}).Where("package_id = ?", packageID)
	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("version_code DESC").Find(&versions).Error
	return versions, total, err
}

func (r *VersionRepository) GetLatestByPackage(ctx context.Context, packageID uint) (*model.Version, error) {
	var version model.Version
	err := r.db.WithContext(ctx).
		Where("package_id = ? AND is_latest = ?", packageID, true).
		First(&version).Error
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *VersionRepository) GetLatestByCategoryCode(ctx context.Context, categoryCode string) (*model.Version, error) {
	var version model.Version
	err := r.db.WithContext(ctx).
		Joins("JOIN packages ON packages.id = versions.package_id").
		Joins("JOIN categories ON categories.id = packages.category_id").
		Where("categories.code = ? AND versions.is_latest = ? AND packages.is_active = ?", categoryCode, true, true).
		First(&version).Error
	if err != nil {
		return nil, err
	}

	// 手动加载关联
	var pkg model.Package
	if err := r.db.WithContext(ctx).First(&pkg, version.PackageID).Error; err == nil {
		version.Package = &pkg
		var category model.Category
		if err := r.db.WithContext(ctx).First(&category, pkg.CategoryID).Error; err == nil {
			pkg.Category = &category
		}
	}

	return &version, nil
}


func (r *VersionRepository) ClearLatestFlag(ctx context.Context, packageID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Version{}).
		Where("package_id = ?", packageID).
		Update("is_latest", false).Error
}

func (r *VersionRepository) SetLatestFlag(ctx context.Context, versionID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Version{}).
		Where("id = ?", versionID).
		Update("is_latest", true).Error
}

func (r *VersionRepository) IncrementDownloadCount(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Version{}).
		Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error
}

func (r *VersionRepository) ExistsByPackageAndVersion(ctx context.Context, packageID uint, version string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Version{}).
		Where("package_id = ? AND version = ?", packageID, version).
		Count(&count).Error
	return count > 0, err
}

func (r *VersionRepository) GetMaxVersionCode(ctx context.Context, packageID uint) (int, error) {
	var maxCode int
	err := r.db.WithContext(ctx).
		Model(&model.Version{}).
		Where("package_id = ?", packageID).
		Select("COALESCE(MAX(version_code), 0)").
		Scan(&maxCode).Error
	return maxCode, err
}
