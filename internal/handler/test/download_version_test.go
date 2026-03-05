package handler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/handler"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/jwt"
	"github.com/taerc/vpublish/pkg/storage"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	err = db.AutoMigrate(&model.User{}, &model.Category{}, &model.Package{}, &model.Version{}, &model.AppKey{}, &model.DownloadLog{}, &model.DownloadStat{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func setupTestStorage(t *testing.T) (*storage.LocalStorage, string) {
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

func createTestUser(t *testing.T, db *gorm.DB) *model.User {
	user := &model.User{Username: "testuser", PasswordHash: "hash", Nickname: "Test", Role: "admin", IsActive: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func createTestCategory(t *testing.T, db *gorm.DB) *model.Category {
	cat := &model.Category{Name: "TestCat", Code: "TEST_CAT", IsActive: true}
	if err := db.Create(cat).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	return cat
}

func createTestPackage(t *testing.T, db *gorm.DB, catID, userID uint) *model.Package {
	pkg := &model.Package{CategoryID: catID, Name: "TestPkg", Description: "test", IsActive: true, CreatedBy: userID}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	return pkg
}

func createTestVersion(t *testing.T, db *gorm.DB, pkgID uint) *model.Version {
	now := time.Now()
	v := &model.Version{
		PackageID: pkgID, Version: "1.0.0", VersionCode: 1000000,
		FilePath: "test/file.apk", FileName: "test.apk", FileSize: 1024, FileHash: "abc123",
		IsStable: true, IsLatest: true, DownloadCount: 0, PublishedAt: &now,
	}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("failed to create version: %v", err)
	}
	return v
}

func TestDownloadVersion_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	ls, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createTestUser(t, db)
	cat := createTestCategory(t, db)
	pkg := createTestPackage(t, db, cat.ID, user.ID)
	version := createTestVersion(t, db, pkg.ID)

	testContent := []byte("test download content")
	testFilePath := ls.GetFilePath(version.FilePath)
	os.MkdirAll(tmpDir+"/test", 0755)
	os.WriteFile(testFilePath, testContent, 0644)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	h := handler.NewPackageHandler(packageService, statsRepo, appKeyRepo)

	jwtService := jwt.New("test-secret", time.Hour, time.Hour*24)
	token, _ := jwtService.GenerateToken(user.ID, user.Username, user.Role)

	router := gin.New()
	router.GET("/admin/versions/:id/download", middleware.JWTAuth(jwtService), h.DownloadVersion)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/versions/%d/download", version.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Body.String() != string(testContent) {
		t.Errorf("expected body '%s', got '%s'", string(testContent), w.Body.String())
	}

	time.Sleep(100 * time.Millisecond)

	var updated model.Version
	db.First(&updated, version.ID)
	if updated.DownloadCount != 1 {
		t.Errorf("expected download count 1, got %d", updated.DownloadCount)
	}

	var logs []model.DownloadLog
	db.Where("version_id = ?", version.ID).Find(&logs)
	if len(logs) != 1 {
		t.Errorf("expected 1 download log, got %d", len(logs))
	}
	if logs[0].AppKey != fmt.Sprintf("admin_%d", user.ID) {
		t.Errorf("expected app key 'admin_%d', got '%s'", user.ID, logs[0].AppKey)
	}

	// Verify download stats (for charts)
	var stats []model.DownloadStat
	db.Where("version_id = ?", version.ID).Find(&stats)
	if len(stats) != 1 {
		t.Errorf("expected 1 download stat, got %d", len(stats))
	}
	if len(stats) > 0 && stats[0].DownloadCount != 1 {
		t.Errorf("expected download stat count 1, got %d", stats[0].DownloadCount)
	}
}

func TestDownloadVersion_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	ls, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	h := handler.NewPackageHandler(packageService, statsRepo, appKeyRepo)

	jwtService := jwt.New("test-secret", time.Hour, time.Hour*24)

	router := gin.New()
	router.GET("/admin/versions/:id/download", middleware.JWTAuth(jwtService), h.DownloadVersion)

	req := httptest.NewRequest(http.MethodGet, "/admin/versions/1/download", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestDownloadVersion_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	ls, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createTestUser(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	h := handler.NewPackageHandler(packageService, statsRepo, appKeyRepo)

	jwtService := jwt.New("test-secret", time.Hour, time.Hour*24)
	token, _ := jwtService.GenerateToken(user.ID, user.Username, user.Role)

	router := gin.New()
	router.GET("/admin/versions/:id/download", middleware.JWTAuth(jwtService), h.DownloadVersion)

	req := httptest.NewRequest(http.MethodGet, "/admin/versions/invalid/download", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestDownloadVersion_VersionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	ls, tmpDir := setupTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createTestUser(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	h := handler.NewPackageHandler(packageService, statsRepo, appKeyRepo)

	jwtService := jwt.New("test-secret", time.Hour, time.Hour*24)
	token, _ := jwtService.GenerateToken(user.ID, user.Username, user.Role)

	router := gin.New()
	router.GET("/admin/versions/:id/download", middleware.JWTAuth(jwtService), h.DownloadVersion)

	req := httptest.NewRequest(http.MethodGet, "/admin/versions/999/download", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
