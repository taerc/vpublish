package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/signature"
	"gorm.io/gorm"
)

func setupSignatureTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&model.AppKey{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func createTestAppKey(t *testing.T, db *gorm.DB) *model.AppKey {
	appKey := &model.AppKey{
		AppName:     "Test App",
		AppKey:      "test_app_key_123456",
		AppSecret:   "test_app_secret_abcdef",
		Description: "Test application key",
		IsActive:    true,
	}
	if err := db.Create(appKey).Error; err != nil {
		t.Fatalf("failed to create app key: %v", err)
	}
	return appKey
}

func createDisabledAppKey(t *testing.T, db *gorm.DB) *model.AppKey {
	appKey := &model.AppKey{
		AppName:     "Disabled App",
		AppKey:      "disabled_app_key",
		AppSecret:   "disabled_app_secret",
		Description: "Disabled application key",
		IsActive:    false,
	}
	if err := db.Create(appKey).Error; err != nil {
		t.Fatalf("failed to create disabled app key: %v", err)
	}
	return appKey
}

func TestSignatureAuth_MissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)

	tests := []struct {
		name       string
		headers    map[string]string
		wantStatus int
	}{
		{
			name:       "missing all headers",
			headers:    map[string]string{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing app key",
			headers: map[string]string{
				signature.HeaderTimestamp: time.Now().Format(time.RFC3339),
				signature.HeaderSignature: "somesignature",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing timestamp",
			headers: map[string]string{
				signature.HeaderAppKey:    "test_key",
				signature.HeaderSignature: "somesignature",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing signature",
			headers: map[string]string{
				signature.HeaderAppKey:    "test_key",
				signature.HeaderTimestamp: time.Now().Format(time.RFC3339),
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			for k, v := range tt.headers {
				c.Request.Header.Set(k, v)
			}

			handler := SignatureAuth(appKeyRepo)
			handler(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestSignatureAuth_InvalidTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)

	tests := []struct {
		name       string
		timestamp  string
		wantStatus int
	}{
		{
			name:       "invalid timestamp format",
			timestamp:  "invalid-timestamp",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty timestamp",
			timestamp:  "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			c.Request.Header.Set(signature.HeaderAppKey, "test_key")
			c.Request.Header.Set(signature.HeaderTimestamp, tt.timestamp)
			c.Request.Header.Set(signature.HeaderSignature, "somesignature")

			handler := SignatureAuth(appKeyRepo)
			handler(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestSignatureAuth_ExpiredSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	// 使用过期的时间戳�?00秒前，超过了300秒的有效期）
	expiredTime := time.Now().Add(-600 * time.Second)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, expiredTime.Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, "somesignature")

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for expired signature, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSignatureAuth_FutureTimestamp(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	// 使用未来的时间戳�?00秒后，超过了300秒的允许误差�?	futureTime := time.Now().Add(600 * time.Second)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, futureTime.Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, "somesignature")

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for future timestamp, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSignatureAuth_InvalidAppKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(signature.HeaderAppKey, "nonexistent_key")
	c.Request.Header.Set(signature.HeaderTimestamp, time.Now().Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, "somesignature")

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for invalid app key, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSignatureAuth_DisabledAppKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createDisabledAppKey(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, time.Now().Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, "somesignature")

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for disabled app key, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSignatureAuth_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	timestamp := time.Now()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, timestamp.Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, "invalid_signature")

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for invalid signature, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSignatureAuth_ValidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	timestamp := time.Now()
	params := map[string]string{"param": "value"}
	sig := signature.GenerateSignature(params, appKey.AppSecret, timestamp.Unix())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, timestamp.Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, sig)

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	// 检查是否通过了中间件（没有调�?c.Abort()�?	if c.IsAborted() {
		t.Error("expected request to pass signature validation")
	}

	// 检�?context 中是否存储了 app_key
	storedAppKey, exists := c.Get("app_key")
	if !exists {
		t.Error("expected app_key to be stored in context")
	}
	if storedAppKey.(*model.AppKey).AppKey != appKey.AppKey {
		t.Errorf("expected app_key %s, got %s", appKey.AppKey, storedAppKey.(*model.AppKey).AppKey)
	}
}

func TestSignatureAuth_ValidSignature_NoParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	timestamp := time.Now()
	params := map[string]string{}
	sig := signature.GenerateSignature(params, appKey.AppSecret, timestamp.Unix())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, timestamp.Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, sig)

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if c.IsAborted() {
		t.Error("expected request to pass signature validation without params")
	}
}

func TestSignatureAuth_ValidSignature_MultipleParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	timestamp := time.Now()
	params := map[string]string{
		"param1": "value1",
		"param2": "value2",
		"param3": "value3",
	}
	sig := signature.GenerateSignature(params, appKey.AppSecret, timestamp.Unix())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test?param1=value1&param2=value2&param3=value3", nil)
	c.Request.Header.Set(signature.HeaderAppKey, appKey.AppKey)
	c.Request.Header.Set(signature.HeaderTimestamp, timestamp.Format(time.RFC3339))
	c.Request.Header.Set(signature.HeaderSignature, sig)

	handler := SignatureAuth(appKeyRepo)
	handler(c)

	if c.IsAborted() {
		t.Error("expected request to pass signature validation with multiple params")
	}
}

func TestGetAppKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appKey := &model.AppKey{
		AppName:   "Test App",
		AppKey:    "test_key",
		AppSecret: "test_secret",
		IsActive:  true,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("app_key", appKey)

	retrievedAppKey := GetAppKey(c)
	if retrievedAppKey.AppKey != appKey.AppKey {
		t.Errorf("expected app_key %s, got %s", appKey.AppKey, retrievedAppKey.AppKey)
	}
}

// 集成测试：测试完整的签名验证流程
func TestSignatureAuth_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSignatureTestDB(t)
	appKeyRepo := repository.NewAppKeyRepository(db)
	appKey := createTestAppKey(t, db)

	// 创建一个测试路�?	router := gin.New()
	router.Use(SignatureAuth(appKeyRepo))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name         string
		setupRequest func(req *http.Request)
		wantStatus   int
		wantBody     string
	}{
		{
			name: "valid signature",
			setupRequest: func(req *http.Request) {
				timestamp := time.Now()
				params := map[string]string{"q": "test"}
				sig := signature.GenerateSignature(params, appKey.AppSecret, timestamp.Unix())
				req.Header.Set(signature.HeaderAppKey, appKey.AppKey)
				req.Header.Set(signature.HeaderTimestamp, timestamp.Format(time.RFC3339))
				req.Header.Set(signature.HeaderSignature, sig)
			},
			wantStatus: http.StatusOK,
			wantBody:   "success",
		},
		{
			name: "invalid signature",
			setupRequest: func(req *http.Request) {
				req.Header.Set(signature.HeaderAppKey, appKey.AppKey)
				req.Header.Set(signature.HeaderTimestamp, time.Now().Format(time.RFC3339))
				req.Header.Set(signature.HeaderSignature, "wrong_signature")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing headers",
			setupRequest: func(req *http.Request) {
				// 不设置任�?header
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected?q=test", nil)
			tt.setupRequest(req)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.wantBody != "" && !contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
