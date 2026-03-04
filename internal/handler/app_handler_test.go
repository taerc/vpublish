package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/signature"
	"github.com/taerc/vpublish/pkg/storage"
	"gorm.io/gorm"
)

func setupHandlerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.User{}, &model.Category{}, &model.Package{}, &model.Version{}, &model.AppKey{}, &model.DownloadLog{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func setupHandlerTestStorage(t *testing.T) (*storage.LocalStorage, string) {
	tmpDir, err := os.MkdirTemp("", "vpublish-handler-test-*")
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

func createHandlerTestUser(t *testing.T, db *gorm.DB) *model.User {
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

func createHandlerTestCategory(t *testing.T, db *gorm.DB, name, code string) *model.Category {
	category := &model.Category{
		Name:     name,
		Code:     code,
		IsActive: true,
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	return category
}

func createHandlerTestPackage(t *testing.T, db *gorm.DB, categoryID, userID uint, name string) *model.Package {
	pkg := &model.Package{
		CategoryID:  categoryID,
		Name:        name,
		Description: "Test package description",
		IsActive:    true,
		CreatedBy:   userID,
	}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("failed to create package: %v", err)
	}
	return pkg
}

func createHandlerTestVersion(t *testing.T, db *gorm.DB, packageID, userID uint, version string) *model.Version {
	now := time.Now()
	v := &model.Version{
		PackageID:    packageID,
		Version:      version,
		VersionCode:  parseTestVersionCode(version),
		FilePath:     "test/path/file.apk",
		FileName:     "test_" + version + ".apk",
		FileSize:     1024,
		FileHash:     "test_hash_" + version,
		Changelog:    "Version " + version,
		ForceUpgrade: false,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    userID,
		PublishedAt:  &now,
	}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("failed to create version: %v", err)
	}
	return v
}

func createHandlerTestAppKey(t *testing.T, db *gorm.DB) *model.AppKey {
	appKey := &model.AppKey{
		AppName:     "Test App",
		AppKey:      "test_app_key_handler",
		AppSecret:   "test_app_secret_handler",
		Description: "Test app key for handler tests",
		IsActive:    true,
	}
	if err := db.Create(appKey).Error; err != nil {
		t.Fatalf("failed to create app key: %v", err)
	}
	return appKey
}

func parseTestVersionCode(version string) int {
	parts := splitVersion(version)
	code := 0
	for i, part := range parts {
		num := 0
		for _, c := range part {
			if c >= '0' && c <= '9' {
				num = num*10 + int(c-'0')
			}
		}
		multiplier := 1
		for j := 0; j < 2-i; j++ {
			multiplier *= 1000
		}
		code += num * multiplier
	}
	return code
}

func splitVersion(version string) []string {
	result := []string{}
	current := ""
	for _, c := range version {
		if c == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	for len(result) < 3 {
		result = append(result, "0")
	}
	return result[:3]
}

// TestGetLatestByCategory 测试获取类别最新版�?func TestGetLatestByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")
	version := createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	appKey := createHandlerTestAppKey(t, db)

	// 设置仓储和服�?	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	// 创建路由
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("app_key", appKey)
		c.Next()
	})
	router.GET("/app/categories/:code/latest", packageHandler.GetLatestByCategory)

	// 创建请求
	req := httptest.NewRequest(http.MethodGet, "/app/categories/TYPE_WU_REN_JI/latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data in response")
	}

	if data["version"] != version.Version {
		t.Errorf("expected version %s, got %v", version.Version, data["version"])
	}

	if data["id"].(float64) != float64(version.ID) {
		t.Errorf("expected version id %d, got %v", version.ID, data["id"])
	}

	// 检查是否包含下载链�?	if _, ok := data["download_url"]; !ok {
		t.Error("expected download_url in response")
	}
}

// TestGetLatestByCategory_NotFound 测试类别不存在的情况
func TestGetLatestByCategory_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	_ = createHandlerTestUser(t, db)
	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("app_key", appKey)
		c.Next()
	})
	router.GET("/app/categories/:code/latest", packageHandler.GetLatestByCategory)

	req := httptest.NewRequest(http.MethodGet, "/app/categories/NONEXISTENT/latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for non-existent category, got %d", http.StatusNotFound, w.Code)
	}
}

// TestGetLatestByCategory_EmptyCode 测试空类别代�?func TestGetLatestByCategory_EmptyCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("app_key", appKey)
		c.Next()
	})
	router.GET("/app/categories/:code/latest", packageHandler.GetLatestByCategory)

	req := httptest.NewRequest(http.MethodGet, "/app/categories//latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for empty category code, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetLatestByCategory_InactivePackage 测试非活跃软件包
func TestGetLatestByCategory_InactivePackage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")
	// 设置为非活跃
	pkg.IsActive = false
	db.Save(pkg)
	_ = createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("app_key", appKey)
		c.Next()
	})
	router.GET("/app/categories/:code/latest", packageHandler.GetLatestByCategory)

	req := httptest.NewRequest(http.MethodGet, "/app/categories/TYPE_WU_REN_JI/latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 非活跃软件包的版本不应该被返�?	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for inactive package, got %d", http.StatusNotFound, w.Code)
	}
}

// TestGetLatestByCategory_MultipleVersions 测试多个版本只返回最�?func TestGetLatestByCategory_MultipleVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")

	// 创建多个版本
	v1 := createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	v1.IsLatest = false
	db.Save(v1)

	v2 := createHandlerTestVersion(t, db, pkg.ID, user.ID, "2.0.0")
	v2.IsLatest = true
	db.Save(v2)

	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("app_key", appKey)
		c.Next()
	})
	router.GET("/app/categories/:code/latest", packageHandler.GetLatestByCategory)

	req := httptest.NewRequest(http.MethodGet, "/app/categories/TYPE_WU_REN_JI/latest", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if data["version"] != "2.0.0" {
		t.Errorf("expected latest version 2.0.0, got %v", data["version"])
	}
}

// TestDownload_MissingToken 测试下载缺少token
func TestDownload_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")
	_ = createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.GET("/app/download/:id", packageHandler.Download)

	// 没有token参数
	req := httptest.NewRequest(http.MethodGet, "/app/download/1", nil)
	req.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for missing token, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestDownload_InvalidToken 测试下载无效token
func TestDownload_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")
	_ = createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.GET("/app/download/:id", packageHandler.Download)

	// 使用无效token
	req := httptest.NewRequest(http.MethodGet, "/app/download/1?token=invalid_token&expires=1234567890", nil)
	req.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for invalid token, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestDownload_VersionNotFound 测试下载不存在的版本
func TestDownload_VersionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.GET("/app/download/:id", packageHandler.Download)

	// 不存在的版本ID
	token := signature.GenerateDownloadToken(999, appKey.AppSecret, time.Now().Add(time.Hour).Unix())
	expires := time.Now().Add(time.Hour).Unix()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/app/download/999?token=%s&expires=%d", token, expires), nil)
	req.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for non-existent version, got %d", http.StatusNotFound, w.Code)
	}
}

// TestDownload_ValidToken 测试有效token下载
func TestDownload_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")
	version := createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	appKey := createHandlerTestAppKey(t, db)

	// 创建测试文件
	testContent := []byte("test file content")
	testFilePath := ls.GetFilePath(version.FilePath)
	if err := os.MkdirAll(tmpDir+"/test/path", 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")
	packageHandler := NewPackageHandler(packageService, statsRepo, appKeyRepo)

	router := gin.New()
	router.GET("/app/download/:id", packageHandler.Download)

	// 生成有效token
	expires := time.Now().Add(time.Hour).Unix()
	token := signature.GenerateDownloadToken(version.ID, appKey.AppSecret, expires)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/app/download/%d?token=%s&expires=%d", version.ID, token, expires), nil)
	req.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestGenerateDownloadURL 测试下载URL生成
func TestGenerateDownloadURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupHandlerTestDB(t)
	ls, tmpDir := setupHandlerTestStorage(t)
	defer os.RemoveAll(tmpDir)

	user := createHandlerTestUser(t, db)
	category := createHandlerTestCategory(t, db, "无人�?, "TYPE_WU_REN_JI")
	pkg := createHandlerTestPackage(t, db, category.ID, user.ID, "DroneApp")
	version := createHandlerTestVersion(t, db, pkg.ID, user.ID, "1.0.0")
	appKey := createHandlerTestAppKey(t, db)

	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, ls, "http://localhost:8080")

	ctx := context.Background()
	downloadURL, err := packageService.GenerateDownloadURL(ctx, version.ID, appKey.AppSecret)
	if err != nil {
		t.Fatalf("failed to generate download URL: %v", err)
	}

	if downloadURL == "" {
		t.Error("expected non-empty download URL")
	}

	// 检查URL格式
	if !bytes.Contains([]byte(downloadURL), []byte("/api/v1/app/download/")) {
		t.Error("download URL should contain /api/v1/app/download/")
	}

	if !bytes.Contains([]byte(downloadURL), []byte("token=")) {
		t.Error("download URL should contain token parameter")
	}

	if !bytes.Contains([]byte(downloadURL), []byte("expires=")) {
		t.Error("download URL should contain expires parameter")
	}
}

// TestVersionModel_GetDownloadURL 测试Version模型的GetDownloadURL方法
func TestVersionModel_GetDownloadURL(t *testing.T) {
	version := &model.Version{
		ID: 100,
	}

	baseURL := "http://localhost:8080"
	url := version.GetDownloadURL(baseURL)

	expectedURL := "http://localhost:8080/api/v1/app/download/100"
	if url != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, url)
	}
}

// TestVersionModel_GetDownloadURL_LargeID 测试大ID�?func TestVersionModel_GetDownloadURL_LargeID(t *testing.T) {
	version := &model.Version{
		ID: 999999,
	}

	baseURL := "http://localhost:8080"
	url := version.GetDownloadURL(baseURL)

	expectedURL := "http://localhost:8080/api/v1/app/download/999999"
	if url != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, url)
	}
}
