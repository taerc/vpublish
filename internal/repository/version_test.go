package repository

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

func setupVersionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.Category{}, &model.Package{}, &model.Version{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func createTestCategoryAndPackage(t *testing.T, db *gorm.DB) (*model.Category, *model.Package) {
	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "Test Package",
		Description: "Test Description",
		IsActive:    true,
		CreatedBy:   1,
	}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("failed to create package: %v", err)
	}

	return category, pkg
}

func TestVersionRepository_Create(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	version := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file.apk",
		FileName:     "file.apk",
		FileSize:     1024,
		FileHash:     "abc123",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}

	err := repo.Create(ctx, version)
	if err != nil {
		t.Fatalf("failed to create version: %v", err)
	}

	if version.ID == 0 {
		t.Error("version ID should not be zero after creation")
	}
}

func TestVersionRepository_GetByID(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	version := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file.apk",
		FileName:     "file.apk",
		FileSize:     1024,
		FileHash:     "abc123",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	db.Create(version)

	found, err := repo.GetByID(ctx, version.ID)
	if err != nil {
		t.Fatalf("failed to get version: %v", err)
	}

	if found.Version != version.Version {
		t.Errorf("expected version %s, got %s", version.Version, found.Version)
	}
}

func TestVersionRepository_ExistsByPackageAndVersion(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	version := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file.apk",
		FileName:     "file.apk",
		FileSize:     1024,
		FileHash:     "abc123",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	db.Create(version)

	tests := []struct {
		name      string
		packageID uint
		version   string
		wantExist bool
	}{
		{
			name:      "exists",
			packageID: pkg.ID,
			version:   "1.0.0",
			wantExist: true,
		},
		{
			name:      "not exists - different version",
			packageID: pkg.ID,
			version:   "2.0.0",
			wantExist: false,
		},
		{
			name:      "not exists - different package",
			packageID: 999,
			version:   "1.0.0",
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := repo.ExistsByPackageAndVersion(ctx, tt.packageID, tt.version)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.wantExist {
				t.Errorf("expected exists %v, got %v", tt.wantExist, exists)
			}
		})
	}
}

func TestVersionRepository_GetMaxVersionCode(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	versions := []*model.Version{
		{
			PackageID:    pkg.ID,
			Version:      "1.0.0",
			VersionCode:  1000000,
			FilePath:     "/test/path/file1.apk",
			FileName:     "file1.apk",
			FileSize:     1024,
			FileHash:     "hash1",
			ForceUpgrade: false,
			IsLatest:     false,
			IsStable:     true,
			CreatedBy:    1,
			PublishedAt:  &now,
		},
		{
			PackageID:    pkg.ID,
			Version:      "2.0.0",
			VersionCode:  2000000,
			FilePath:     "/test/path/file2.apk",
			FileName:     "file2.apk",
			FileSize:     2048,
			FileHash:     "hash2",
			ForceUpgrade: false,
			IsLatest:     true,
			IsStable:     true,
			CreatedBy:    1,
			PublishedAt:  &now,
		},
		{
			PackageID:    pkg.ID,
			Version:      "1.5.0",
			VersionCode:  1005000,
			FilePath:     "/test/path/file3.apk",
			FileName:     "file3.apk",
			FileSize:     1536,
			FileHash:     "hash3",
			ForceUpgrade: false,
			IsLatest:     false,
			IsStable:     true,
			CreatedBy:    1,
			PublishedAt:  &now,
		},
	}

	for _, v := range versions {
		db.Create(v)
	}

	maxCode, err := repo.GetMaxVersionCode(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("failed to get max version code: %v", err)
	}

	if maxCode != 2000000 {
		t.Errorf("expected max version code 2000000, got %d", maxCode)
	}

	t.Run("empty package returns 0", func(t *testing.T) {
		maxCode, err := repo.GetMaxVersionCode(ctx, 999)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if maxCode != 0 {
			t.Errorf("expected max version code 0 for empty package, got %d", maxCode)
		}
	})
}

func TestVersionRepository_GetByPackageAndVersion(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	version := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file.apk",
		FileName:     "file.apk",
		FileSize:     1024,
		FileHash:     "abc123",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	db.Create(version)

	found, err := repo.GetByPackageAndVersion(ctx, pkg.ID, "1.0.0")
	if err != nil {
		t.Fatalf("failed to get version: %v", err)
	}

	if found.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", found.Version)
	}

	_, err = repo.GetByPackageAndVersion(ctx, pkg.ID, "2.0.0")
	if err == nil {
		t.Error("expected error when getting non-existent version")
	}
}

func TestVersionRepository_ListByPackage(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	for i := 1; i <= 5; i++ {
		version := &model.Version{
			PackageID:    pkg.ID,
			Version:      "1.0." + string(rune('0'+i)),
			VersionCode:  1000000 + i,
			FilePath:     "/test/path/file.apk",
			FileName:     "file.apk",
			FileSize:     1024,
			FileHash:     "abc123",
			ForceUpgrade: false,
			IsLatest:     i == 5,
			IsStable:     true,
			CreatedBy:    1,
			PublishedAt:  &now,
		}
		db.Create(version)
	}

	versions, total, err := repo.ListByPackage(ctx, pkg.ID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list versions: %v", err)
	}

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	if len(versions) != 5 {
		t.Errorf("expected 5 versions, got %d", len(versions))
	}
}

func TestVersionRepository_ClearLatestFlag(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	v1 := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file1.apk",
		FileName:     "file1.apk",
		FileSize:     1024,
		FileHash:     "hash1",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	v2 := &model.Version{
		PackageID:    pkg.ID,
		Version:      "2.0.0",
		VersionCode:  2000000,
		FilePath:     "/test/path/file2.apk",
		FileName:     "file2.apk",
		FileSize:     2048,
		FileHash:     "hash2",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	db.Create(v1)
	db.Create(v2)

	err := repo.ClearLatestFlag(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("failed to clear latest flag: %v", err)
	}

	var count int64
	db.Model(&model.Version{}).Where("package_id = ? AND is_latest = ?", pkg.ID, true).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 versions with is_latest=true, got %d", count)
	}
}

func TestVersionRepository_GetLatestByPackage(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	v1 := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file1.apk",
		FileName:     "file1.apk",
		FileSize:     1024,
		FileHash:     "hash1",
		ForceUpgrade: false,
		IsLatest:     false,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	v2 := &model.Version{
		PackageID:    pkg.ID,
		Version:      "2.0.0",
		VersionCode:  2000000,
		FilePath:     "/test/path/file2.apk",
		FileName:     "file2.apk",
		FileSize:     2048,
		FileHash:     "hash2",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	db.Create(v1)
	db.Create(v2)

	latest, err := repo.GetLatestByPackage(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("failed to get latest version: %v", err)
	}

	if latest.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", latest.Version)
	}
}

func TestVersionRepository_Delete(t *testing.T) {
	db := setupVersionTestDB(t)
	repo := NewVersionRepository(db)
	ctx := context.Background()

	_, pkg := createTestCategoryAndPackage(t, db)

	now := time.Now()
	version := &model.Version{
		PackageID:    pkg.ID,
		Version:      "1.0.0",
		VersionCode:  1000000,
		FilePath:     "/test/path/file.apk",
		FileName:     "file.apk",
		FileSize:     1024,
		FileHash:     "abc123",
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    1,
		PublishedAt:  &now,
	}
	db.Create(version)

	err := repo.Delete(ctx, version.ID)
	if err != nil {
		t.Fatalf("failed to delete version: %v", err)
	}

	_, err = repo.GetByID(ctx, version.ID)
	if err == nil {
		t.Error("expected error when getting deleted version")
	}
}
