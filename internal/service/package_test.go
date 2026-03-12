package service

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/storage"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.Category{}, &model.Package{}, &model.Version{}, &model.User{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *gorm.DB) *model.User {
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

func createTestCategory(t *testing.T, db *gorm.DB) *model.Category {
	category := &model.Category{
		Name:     "Test Category",
		Code:     "TEST_CAT",
		IsActive: true,
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	return category
}

func createTestPackage(t *testing.T, db *gorm.DB, category *model.Category, name string, userID uint) *model.Package {
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

func createTestStorage(t *testing.T) (*storage.LocalStorage, string) {
	tmpDir, err := os.MkdirTemp("", "vpublish-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	ls, err := storage.NewLocalStorage(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	return ls, tmpDir
}

func createMultipartFileHeader(t *testing.T, filename, content string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", "application/octet-stream")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("failed to create part: %v", err)
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		t.Fatalf("failed to write content: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(32 << 20)
	if err != nil {
		t.Fatalf("failed to parse multipart form: %v", err)
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		t.Fatalf("failed to get form file: %v", err)
	}
	file.Close()

	return req.MultipartForm.File["file"][0]
}

func TestPackageService_CreateWithVersion_DuplicateVersion(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")

	req := &CreatePackageWithVersionRequest{
		CategoryID:   category.ID,
		Version:      "1.0.0",
		Description:  "Test package",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		File:         fileHeader,
	}

	pkg1, _, err := svc.CreateWithVersion(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("first create should succeed, got error: %v", err)
	}
	if pkg1 == nil {
		t.Fatal("first create should return package")
	}

	fileHeader2 := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content 2")

	req2 := &CreatePackageWithVersionRequest{
		CategoryID:   category.ID,
		Version:      "1.0.0",
		Description:  "Test package duplicate",
		Changelog:    "Duplicate version",
		ForceUpgrade: false,
		File:         fileHeader2,
	}

	pkg2, _, err := svc.CreateWithVersion(ctx, user.ID, req2)
	if err == nil {
		t.Error("second create with same version should fail with ErrVersionAlreadyExists")
	}
	if pkg2 != nil {
		t.Error("second create should return nil package")
	}
	if err != nil && !errors.Is(err, ErrVersionAlreadyExists) {
		t.Errorf("expected ErrVersionAlreadyExists, got: %v", err)
	}
}

func TestPackageService_CreateWithVersion_PackageExists_DifferentVersion(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")

	req := &CreatePackageWithVersionRequest{
		CategoryID:   category.ID,
		Version:      "1.0.0",
		Description:  "Test package",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		File:         fileHeader,
	}

	pkg1, _, err := svc.CreateWithVersion(ctx, user.ID, req)
	if err != nil {
		t.Fatalf("first create should succeed, got error: %v", err)
	}

	fileHeader2 := createMultipartFileHeader(t, "TestApp_v2.0.0.apk", "test content 2")

	req2 := &CreatePackageWithVersionRequest{
		CategoryID:   category.ID,
		Version:      "2.0.0",
		Description:  "Test package new version",
		Changelog:    "New version",
		ForceUpgrade: false,
		File:         fileHeader2,
	}

	pkg2, _, err := svc.CreateWithVersion(ctx, user.ID, req2)
	if err == nil {
		t.Error("second create with different version should fail (package already exists)")
	}
	if err != nil && err.Error() != "软件包已存在，请使用上传新版本接口" {
		t.Logf("got expected error: %v", err)
	}
	_ = pkg1
	_ = pkg2
}

func TestPackageService_UploadVersion_DuplicateVersion(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")

	req := &CreateVersionRequest{
		Version:      "1.0.0",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		IsStable:     true,
	}

	_, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("first upload should succeed, got error: %v", err)
	}

	fileHeader2 := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content 2")

	_, err = svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader2, req)
	if err == nil {
		t.Error("second upload with same version should fail")
	}
	if !errors.Is(err, ErrVersionAlreadyExists) {
		t.Errorf("expected ErrVersionAlreadyExists, got: %v", err)
	}
}

func TestPackageService_UploadVersion_VersionMustBeGreater(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader1 := createMultipartFileHeader(t, "TestApp_v2.0.0.apk", "test content")

	req1 := &CreateVersionRequest{
		Version:      "2.0.0",
		Changelog:    "Version 2.0.0",
		ForceUpgrade: false,
		IsStable:     true,
	}

	_, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader1, req1)
	if err != nil {
		t.Fatalf("first upload should succeed, got error: %v", err)
	}

	fileHeader2 := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content 2")

	req2 := &CreateVersionRequest{
		Version:      "1.0.0",
		Changelog:    "Version 1.0.0",
		ForceUpgrade: false,
		IsStable:     true,
	}

	_, err = svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader2, req2)
	if err == nil {
		t.Error("upload with lower version should fail")
	}
	if !errors.Is(err, ErrVersionMustBeGreater) {
		t.Errorf("expected ErrVersionMustBeGreater, got: %v", err)
	}

	fileHeader3 := createMultipartFileHeader(t, "TestApp_v2.0.0.apk", "test content 3")

	req3 := &CreateVersionRequest{
		Version:      "2.0.0",
		Changelog:    "Version 2.0.0 again",
		ForceUpgrade: false,
		IsStable:     true,
	}

	_, err = svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader3, req3)
	if err == nil {
		t.Error("upload with equal version should fail")
	}
	if !errors.Is(err, ErrVersionAlreadyExists) {
		t.Errorf("expected ErrVersionAlreadyExists, got: %v", err)
	}
}

func TestParseVersionCode(t *testing.T) {
	tests := []struct {
		version   string
		wantCode  int
		wantError bool
	}{
		{"1.0.0", 1000000, false},
		{"1.2.3", 1002003, false},
		{"2.0.0", 2000000, false},
		{"10.20.30", 10020030, false},
		{"0.0.1", 1, false},
		{"1.0", 1000000, false},
		{"1", 1000000, false},
		{"", 0, true},
		{"a.b.c", 0, true},
		{"1.0.0.0", 1000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			code, err := parseVersionCode(tt.version)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if code != tt.wantCode {
					t.Errorf("expected code %d, got %d", tt.wantCode, code)
				}
			}
		})
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		filename string
		wantName string
	}{
		{"TestApp_v1.0.0.apk", "TestApp"},
		{"TestApp-1.0.0.apk", "TestApp"},
		{"TestApp_1.0.0.apk", "TestApp"},
		{"MyApp-V2.0.0.apk", "MyApp"},
		{"SimpleApp.apk", "SimpleApp"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			name := extractPackageName(tt.filename)
			if name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, name)
			}
		})
	}
}

func TestPackageService_CreateWithVersion_CategoryNotFound(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")

	req := &CreatePackageWithVersionRequest{
		CategoryID:   999,
		Version:      "1.0.0",
		Description:  "Test package",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		File:         fileHeader,
	}

	_, _, err := svc.CreateWithVersion(ctx, user.ID, req)
	if err == nil {
		t.Error("expected error for non-existent category")
	}
}

func TestPackageService_CreateWithVersion_InvalidVersionFormat(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_invalid.apk", "test content")

	req := &CreatePackageWithVersionRequest{
		CategoryID:   category.ID,
		Version:      "invalid.version",
		Description:  "Test package",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		File:         fileHeader,
	}

	_, _, err := svc.CreateWithVersion(ctx, user.ID, req)
	if err == nil {
		t.Error("expected error for invalid version format")
	}
}

func TestPackageService_UploadVersion_PackageNotFound(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")

	req := &CreateVersionRequest{
		Version:      "1.0.0",
		Changelog:    "Initial version",
		ForceUpgrade: false,
		IsStable:     true,
	}

	_, err := svc.UploadVersion(ctx, 999, user.ID, fileHeader, req)
	if err == nil {
		t.Error("expected error for non-existent package")
	}
	if !errors.Is(err, ErrPackageNotFound) {
		t.Errorf("expected ErrPackageNotFound, got: %v", err)
	}
}

func TestPackageService_Delete(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	err := svc.Delete(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("failed to delete package: %v", err)
	}

	_, err = svc.GetByID(ctx, pkg.ID)
	if err == nil {
		t.Error("expected error when getting deleted package")
	}
}

func TestPackageService_Update(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	active := false
	req := &UpdatePackageRequest{
		Name:        "UpdatedApp",
		Description: "Updated description",
		IsActive:    &active,
	}

	updated, err := svc.Update(ctx, pkg.ID, req)
	if err != nil {
		t.Fatalf("failed to update package: %v", err)
	}

	if updated.Name != "UpdatedApp" {
		t.Errorf("expected name UpdatedApp, got %s", updated.Name)
	}
	if updated.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", updated.Description)
	}
	if updated.IsActive != false {
		t.Error("expected IsActive to be false")
	}
}

func TestPackageService_ListVersions(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader1 := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")
	req1 := &CreateVersionRequest{Version: "1.0.0", IsStable: true}
	_, _ = svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader1, req1)

	fileHeader2 := createMultipartFileHeader(t, "TestApp_v2.0.0.apk", "test content 2")
	req2 := &CreateVersionRequest{Version: "2.0.0", IsStable: true}
	_, _ = svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader2, req2)

	versions, total, err := svc.ListVersions(ctx, pkg.ID, 1, 10)
	if err != nil {
		t.Fatalf("failed to list versions: %v", err)
	}

	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestPackageService_DeleteVersion(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")
	req := &CreateVersionRequest{Version: "1.0.0", IsStable: true}
	version, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("failed to upload version: %v", err)
	}

	err = svc.DeleteVersion(ctx, version.ID)
	if err != nil {
		t.Fatalf("failed to delete version: %v", err)
	}

	_, err = svc.GetVersionByID(ctx, version.ID)
	if err == nil {
		t.Error("expected error when getting deleted version")
	}
}

func TestPackageService_UploadVersion_Success(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")

	req := &CreateVersionRequest{
		Version:      "1.0.0",
		Changelog:    "Initial version",
		ReleaseNotes: "Release notes",
		MinVersion:   "0.9.0",
		ForceUpgrade: false,
		IsStable:     true,
	}

	version, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("upload should succeed, got error: %v", err)
	}

	if version.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", version.Version)
	}
	if version.VersionCode != 1000000 {
		t.Errorf("expected version code 1000000, got %d", version.VersionCode)
	}
	if !version.IsLatest {
		t.Error("version should be marked as latest")
	}
	if !version.IsStable {
		t.Error("version should be marked as stable")
	}
}

func TestPackageService_GetFilePath(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")
	req := &CreateVersionRequest{Version: "1.0.0", IsStable: true}
	version, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("failed to upload version: %v", err)
	}

	filePath := svc.GetFilePath(version)
	if filePath == "" {
		t.Error("file path should not be empty")
	}

	if !filepath.IsAbs(filePath) {
		t.Errorf("file path should be absolute, got: %s", filePath)
	}
}

func TestValidateVersionFormat(t *testing.T) {
	tests := []struct {
		version   string
		wantError bool
	}{
		// 有效格式
		{"v1.0.0", false},
		{"V1.0.0", false},
		{"1.0.0", false},
		{"v01.00.00", false},
		{"V10.20.30", false},
		{" v1.0.0 ", false}, // 自动 trim

		// 无效格式
		{"v1.0", true},        // 只有2段
		{"V1", true},          // 只有1段
		{"x1.0.0", true},      // 错误前缀
		{"1.0.0.0", true},     // 4段
		{"v1.0.0-beta", true}, // 预发布标签
		{"", true},            // 空字符串
		{"abc", true},         // 非版本号
		{"v1.0.0-rc1", true},  // 预发布标签
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := validateVersionFormat(tt.version)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateVersionFormat(%q) expected error, got nil", tt.version)
				}
			} else {
				if err != nil {
					t.Errorf("validateVersionFormat(%q) unexpected error: %v", tt.version, err)
				}
			}
		})
	}
}

func TestPackageService_GenerateDownloadURL_WithExternalPrefix(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "http://cdn.example.com/dd")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")
	req := &CreateVersionRequest{Version: "1.0.0", IsStable: true}
	version, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("failed to upload version: %v", err)
	}

	url, err := svc.GenerateDownloadURL(ctx, version.ID, "test-secret")
	if err != nil {
		t.Fatalf("failed to generate download URL: %v", err)
	}

	expectedPrefix := "http://cdn.example.com/dd/api/v1/app/download/"
	if !strings.HasPrefix(url, expectedPrefix) {
		t.Errorf("expected URL to start with %q, got %q", expectedPrefix, url)
	}
	if !strings.Contains(url, "token=") {
		t.Error("external URL should contain token parameter")
	}
	if !strings.Contains(url, "expires=") {
		t.Error("external URL should contain expires parameter")
	}
}

func TestPackageService_GenerateDownloadURL_WithoutExternalPrefix(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")
	req := &CreateVersionRequest{Version: "1.0.0", IsStable: true}
	version, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("failed to upload version: %v", err)
	}

	url, err := svc.GenerateDownloadURL(ctx, version.ID, "test-secret")
	if err != nil {
		t.Fatalf("failed to generate download URL: %v", err)
	}

	if !strings.HasPrefix(url, "http://localhost:8080/api/v1/app/download/") {
		t.Errorf("expected internal URL format, got %q", url)
	}
	if !strings.Contains(url, "token=") {
		t.Error("internal URL should contain token parameter")
	}
	if !strings.Contains(url, "expires=") {
		t.Error("internal URL should contain expires parameter")
	}
}

func TestPackageService_GenerateDownloadURL_VersionNotFound(t *testing.T) {
	db := setupServiceTestDB(t)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "")
	ctx := context.Background()

	_, err := svc.GenerateDownloadURL(ctx, 999, "test-secret")
	if err == nil {
		t.Error("expected error for non-existent version")
	}
	if !errors.Is(err, ErrVersionNotFound) {
		t.Errorf("expected ErrVersionNotFound, got: %v", err)
	}
}

func TestPackageService_GenerateDownloadURL_ExternalPrefixWithTrailingSlash(t *testing.T) {
	db := setupServiceTestDB(t)
	user := createTestUser(t, db)
	category := createTestCategory(t, db)
	pkg := createTestPackage(t, db, category, "TestApp", user.ID)
	ls, tmpDir := createTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)

	svc := NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080", "http://cdn.example.com/dd/")
	ctx := context.Background()

	fileHeader := createMultipartFileHeader(t, "TestApp_v1.0.0.apk", "test content")
	req := &CreateVersionRequest{Version: "1.0.0", IsStable: true}
	version, err := svc.UploadVersion(ctx, pkg.ID, user.ID, fileHeader, req)
	if err != nil {
		t.Fatalf("failed to upload version: %v", err)
	}

	url, err := svc.GenerateDownloadURL(ctx, version.ID, "test-secret")
	if err != nil {
		t.Fatalf("failed to generate download URL: %v", err)
	}

	if strings.Contains(url, "//api/") {
		t.Errorf("URL should not have double slashes, got %q", url)
	}
	if !strings.Contains(url, "token=") {
		t.Error("external URL should contain token parameter")
	}
	if !strings.Contains(url, "expires=") {
		t.Error("external URL should contain expires parameter")
	}
}
