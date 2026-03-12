package service

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"gorm.io/gorm"
)

func setupCategoryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.Category{}, &model.Package{}, &model.User{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func createCategoryTestUser(t *testing.T, db *gorm.DB) *model.User {
	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Nickname:     "Test User",
		Role:         "admin",
		IsActive:     true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func createTestCategoryForDelete(t *testing.T, db *gorm.DB, name string) *model.Category {
	category := &model.Category{
		Name:     name,
		Code:     "TEST_" + name,
		IsActive: true,
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	return category
}

func createTestPackageForCategory(t *testing.T, db *gorm.DB, category *model.Category, name string, userID uint) *model.Package {
	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        name,
		Description: "Test package",
		IsActive:    true,
		CreatedBy:   userID,
	}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	return pkg
}

// TestCategoryService_Delete_WithPackages tests that deleting a category with packages fails
func TestCategoryService_Delete_WithPackages(t *testing.T) {
	db := setupCategoryTestDB(t)
	user := createCategoryTestUser(t, db)
	category := createTestCategoryForDelete(t, db, "TestCategory")
	_ = createTestPackageForCategory(t, db, category, "TestPackage", user.ID)

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	// Attempt to delete category with packages - should fail
	err := svc.Delete(ctx, category.ID)
	if err == nil {
		t.Error("expected error when deleting category with packages")
	}
	if !errors.Is(err, ErrCategoryHasPackages) {
		t.Errorf("expected ErrCategoryHasPackages, got: %v", err)
	}

	// Verify category still exists
	_, err = svc.GetByID(ctx, category.ID)
	if err != nil {
		t.Error("category should still exist after failed delete")
	}
}

// TestCategoryService_Delete_WithoutPackages tests that deleting a category without packages succeeds
func TestCategoryService_Delete_WithoutPackages(t *testing.T) {
	db := setupCategoryTestDB(t)
	_ = createCategoryTestUser(t, db)
	category := createTestCategoryForDelete(t, db, "EmptyCategory")

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	// Delete category without packages - should succeed
	err := svc.Delete(ctx, category.ID)
	if err != nil {
		t.Fatalf("expected no error when deleting category without packages, got: %v", err)
	}

	// Verify category is soft-deleted
	_, err = svc.GetByID(ctx, category.ID)
	if err == nil {
		t.Error("category should not be found after delete")
	}
}

// TestCategoryService_Create tests successful category creation
func TestCategoryService_Create(t *testing.T) {
	db := setupCategoryTestDB(t)

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	req := &CreateCategoryRequest{
		Name:        "New Category",
		Description: "Test description",
		SortOrder:   10,
	}

	category, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if category.Name != "New Category" {
		t.Errorf("expected name 'New Category', got %s", category.Name)
	}
	if category.Code == "" {
		t.Error("expected code to be generated")
	}
	if !category.IsActive {
		t.Error("expected IsActive to be true")
	}
}

// TestCategoryService_Create_DuplicateName tests that creating a category with duplicate name fails
func TestCategoryService_Create_DuplicateName(t *testing.T) {
	db := setupCategoryTestDB(t)
	_ = createTestCategoryForDelete(t, db, "ExistingCategory")

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	req := &CreateCategoryRequest{
		Name:        "ExistingCategory",
		Description: "Duplicate category",
	}

	_, err := svc.Create(ctx, req)
	if err == nil {
		t.Error("expected error when creating category with duplicate name")
	}
	if !errors.Is(err, ErrCategoryAlreadyExists) {
		t.Errorf("expected ErrCategoryAlreadyExists, got: %v", err)
	}
}

// TestCategoryService_Update tests successful category update
func TestCategoryService_Update(t *testing.T) {
	db := setupCategoryTestDB(t)
	category := createTestCategoryForDelete(t, db, "CategoryToUpdate")

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	newSortOrder := 20
	req := &UpdateCategoryRequest{
		Name:        "UpdatedCategory",
		Description: "Updated description",
		SortOrder:   &newSortOrder,
	}

	updated, err := svc.Update(ctx, category.ID, req)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Name != "UpdatedCategory" {
		t.Errorf("expected name 'UpdatedCategory', got %s", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", updated.Description)
	}
	if updated.SortOrder != 20 {
		t.Errorf("expected sort order 20, got %d", updated.SortOrder)
	}
}

// TestCategoryService_GetByID tests getting a category by ID
func TestCategoryService_GetByID(t *testing.T) {
	db := setupCategoryTestDB(t)
	category := createTestCategoryForDelete(t, db, "TestCategory")

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	found, err := svc.GetByID(ctx, category.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if found.Name != "TestCategory" {
		t.Errorf("expected name 'TestCategory', got %s", found.Name)
	}
}

// TestCategoryService_GetByID_NotFound tests getting a non-existent category
func TestCategoryService_GetByID_NotFound(t *testing.T) {
	db := setupCategoryTestDB(t)

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, 999)
	if err == nil {
		t.Error("expected error when getting non-existent category")
	}
}

// TestCategoryService_List tests listing categories
func TestCategoryService_List(t *testing.T) {
	db := setupCategoryTestDB(t)
	_ = createTestCategoryForDelete(t, db, "Category1")
	_ = createTestCategoryForDelete(t, db, "Category2")

	categoryRepo := repository.NewCategoryRepository(db)
	svc := NewCategoryService(categoryRepo)
	ctx := context.Background()

	categories, total, err := svc.List(ctx, 1, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}
