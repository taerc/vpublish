package repository

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

func setupPackageTestDB(t *testing.T) *gorm.DB {
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

func TestPackageRepository_Create(t *testing.T) {
	db := setupPackageTestDB(t)
	repo := NewPackageRepository(db)
	ctx := context.Background()

	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	db.Create(category)

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "Test Package",
		Description: "Test Description",
		IsActive:    true,
		CreatedBy:   1,
	}

	err := repo.Create(ctx, pkg)
	if err != nil {
		t.Fatalf("failed to create package: %v", err)
	}

	if pkg.ID == 0 {
		t.Error("package ID should not be zero after creation")
	}
}

func TestPackageRepository_GetByID(t *testing.T) {
	db := setupPackageTestDB(t)
	repo := NewPackageRepository(db)
	ctx := context.Background()

	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	db.Create(category)

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "Test Package",
		Description: "Test Description",
		IsActive:    true,
		CreatedBy:   1,
	}
	db.Create(pkg)

	found, err := repo.GetByID(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("failed to get package: %v", err)
	}

	if found.Name != pkg.Name {
		t.Errorf("expected name %s, got %s", pkg.Name, found.Name)
	}
}

func TestPackageRepository_GetByCategoryAndName(t *testing.T) {
	db := setupPackageTestDB(t)
	repo := NewPackageRepository(db)
	ctx := context.Background()

	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	db.Create(category)

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "Test Package",
		Description: "Test Description",
		IsActive:    true,
		CreatedBy:   1,
	}
	db.Create(pkg)

	tests := []struct {
		name       string
		categoryID uint
		pkgName    string
		wantFound  bool
	}{
		{
			name:       "found",
			categoryID: category.ID,
			pkgName:    "Test Package",
			wantFound:  true,
		},
		{
			name:       "not found - different name",
			categoryID: category.ID,
			pkgName:    "Other Package",
			wantFound:  false,
		},
		{
			name:       "not found - different category",
			categoryID: 999,
			pkgName:    "Test Package",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.GetByCategoryAndName(ctx, tt.categoryID, tt.pkgName)
			if tt.wantFound {
				if err != nil {
					t.Fatalf("expected to find package, got error: %v", err)
				}
				if found.Name != tt.pkgName {
					t.Errorf("expected name %s, got %s", tt.pkgName, found.Name)
				}
			} else {
				if err == nil {
					t.Errorf("expected not to find package, but found one")
				}
			}
		})
	}
}

func TestPackageRepository_List(t *testing.T) {
	db := setupPackageTestDB(t)
	repo := NewPackageRepository(db)
	ctx := context.Background()

	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	db.Create(category)

	for i := 1; i <= 5; i++ {
		pkg := &model.Package{
			CategoryID:  category.ID,
			Name:        "Test Package",
			Description: "Test Description",
			IsActive:    true,
			CreatedBy:   1,
		}
		db.Create(pkg)
	}

	packages, total, err := repo.List(ctx, category.ID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list packages: %v", err)
	}

	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}

	if len(packages) != 5 {
		t.Errorf("expected 5 packages, got %d", len(packages))
	}
}

func TestPackageRepository_Update(t *testing.T) {
	db := setupPackageTestDB(t)
	repo := NewPackageRepository(db)
	ctx := context.Background()

	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	db.Create(category)

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "Test Package",
		Description: "Test Description",
		IsActive:    true,
		CreatedBy:   1,
	}
	db.Create(pkg)

	pkg.Name = "Updated Package"
	pkg.Description = "Updated Description"

	err := repo.Update(ctx, pkg)
	if err != nil {
		t.Fatalf("failed to update package: %v", err)
	}

	found, _ := repo.GetByID(ctx, pkg.ID)
	if found.Name != "Updated Package" {
		t.Errorf("expected name Updated Package, got %s", found.Name)
	}
}

func TestPackageRepository_Delete(t *testing.T) {
	db := setupPackageTestDB(t)
	repo := NewPackageRepository(db)
	ctx := context.Background()

	category := &model.Category{
		Name: "Test Category",
		Code: "TEST_CAT",
	}
	db.Create(category)

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "Test Package",
		Description: "Test Description",
		IsActive:    true,
		CreatedBy:   1,
	}
	db.Create(pkg)

	err := repo.Delete(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("failed to delete package: %v", err)
	}

	_, err = repo.GetByID(ctx, pkg.ID)
	if err == nil {
		t.Error("expected error when getting deleted package")
	}
}

// TestPackageRepository_PreloadSoftDeletedCategory tests that Preload with Unscoped
// returns the category object even if it was soft-deleted
func TestPackageRepository_PreloadSoftDeletedCategory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.Category{}, &model.Package{}, &model.User{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Create test data
	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hash",
		Role:         "admin",
		IsActive:     true,
	}
	db.Create(user)

	category := &model.Category{
		Name:     "TestCategory",
		Code:     "TEST_CAT",
		IsActive: true,
	}
	db.Create(category)

	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        "TestPackage",
		Description: "Test",
		IsActive:    true,
		CreatedBy:   user.ID,
	}
	db.Create(pkg)

	// Soft-delete the category
	db.Delete(&model.Category{}, category.ID)

	// Verify category is soft-deleted
	var catCount int64
	db.Model(&model.Category{}).Where("id = ?", category.ID).Count(&catCount)
	if catCount != 0 {
		t.Error("category should be soft-deleted and not count in normal query")
	}

	// Now test that Preload with Unscoped returns the category
	repo := NewPackageRepository(db)
	ctx := context.Background()

	// Test GetByID
	found, err := repo.GetByID(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Category == nil {
		t.Error("Category should be loaded even though it's soft-deleted")
	} else {
		if found.Category.Name != "TestCategory" {
			t.Errorf("expected category name 'TestCategory', got %s", found.Category.Name)
		}
		t.Logf("SUCCESS: Category loaded: %+v", found.Category)
	}

	// Test List
	packages, total, err := repo.List(ctx, 0, 1, 10)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}

	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}

	if packages[0].Category == nil {
		t.Error("Category should be loaded in List even though it's soft-deleted")
	} else {
		if packages[0].Category.Name != "TestCategory" {
			t.Errorf("expected category name 'TestCategory', got %s", packages[0].Category.Name)
		}
		t.Logf("SUCCESS: Category loaded in List: %+v", packages[0].Category)
	}
}
