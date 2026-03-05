package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/config"
	"github.com/taerc/vpublish/internal/database"
	"github.com/taerc/vpublish/internal/handler"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/jwt"
	"github.com/taerc/vpublish/pkg/storage"
)

func main() {
	// 加载配置
	cfg, err := config.Load("./configs/config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 初始化数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	// 自动迁移
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	// 初始化存储
	store, err := storage.NewLocalStorage(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	// 初始化 JWT
	jwtService := jwt.New(cfg.JWT.Secret, cfg.JWT.Expire, cfg.JWT.RefreshExpire)

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	appKeyRepo := repository.NewAppKeyRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	packageRepo := repository.NewPackageRepository(db)
	versionRepo := repository.NewVersionRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	// 初始化 Service
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	packageService := service.NewPackageService(packageRepo, versionRepo, categoryRepo, store, "")
	statsService := service.NewStatsService(statsRepo)
	appKeyService := service.NewAppKeyService(appKeyRepo)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(userService, jwtService)
	userHandler := handler.NewUserHandler(userService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	packageHandler := handler.NewPackageHandler(packageService, statsRepo, appKeyRepo)
	statsHandler := handler.NewStatsHandler(statsService)
	appKeyHandler := handler.NewAppKeyHandler(appKeyService)

	// 创建路由
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS(&cfg.CORS))

	setupRoutes(r, authHandler, userHandler, categoryHandler, packageHandler, statsHandler,
		appKeyHandler, jwtService, appKeyRepo)

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server started on port %d", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRoutes(
	r *gin.Engine,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	categoryHandler *handler.CategoryHandler,
	packageHandler *handler.PackageHandler,
	statsHandler *handler.StatsHandler,
	appKeyHandler *handler.AppKeyHandler,
	jwtService *jwt.JWT,
	appKeyRepo *repository.AppKeyRepository,
) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// ============ 管理端 API ============
		admin := v1.Group("/admin")
		{
			// 认证
			admin.POST("/auth/login", authHandler.Login)
			admin.POST("/auth/refresh", authHandler.RefreshToken)
			admin.POST("/auth/logout", authHandler.Logout)

			// 需要登录的路由
			auth := admin.Group("")
			auth.Use(middleware.JWTAuth(jwtService))
			{
				// 用户信息
				auth.GET("/auth/profile", authHandler.GetProfile)
				auth.PUT("/auth/password", authHandler.ChangePassword)

				// 用户管理
				auth.GET("/users", userHandler.List)
				auth.GET("/users/:id", userHandler.Get)
				auth.POST("/users", userHandler.Create)
				auth.PUT("/users/:id", userHandler.Update)
				auth.DELETE("/users/:id", userHandler.Delete)
				auth.PUT("/users/:id/password", userHandler.ResetPassword)

				// 类别管理
				auth.GET("/categories", categoryHandler.List)
				auth.GET("/categories/active", categoryHandler.ListActive) // 获取启用的类别列表
				auth.GET("/categories/:id", categoryHandler.Get)
				auth.POST("/categories", categoryHandler.Create)
				auth.PUT("/categories/:id", categoryHandler.Update)
				auth.DELETE("/categories/:id", categoryHandler.Delete)

				// 软件包管理
				auth.GET("/packages", packageHandler.List)
				auth.GET("/packages/:id", packageHandler.Get)
				auth.POST("/packages", packageHandler.Create)
				auth.PUT("/packages/:id", packageHandler.Update)
				auth.DELETE("/packages/:id", packageHandler.Delete)

				// 版本管理
				auth.GET("/packages/:id/versions", packageHandler.ListVersions)
				auth.POST("/packages/:id/versions", packageHandler.UploadVersion)
				auth.DELETE("/versions/:id", packageHandler.DeleteVersion)
				auth.GET("/versions/:id/download", packageHandler.DownloadVersion) // 管理端下载

				// 统计
				auth.GET("/stats/daily", statsHandler.DailyStats)
				auth.GET("/stats/trend", statsHandler.DailyTrend)
				auth.GET("/stats/monthly", statsHandler.MonthlyStats)
				auth.GET("/stats/yearly", statsHandler.YearlyStats)
				auth.GET("/stats/category", statsHandler.CategoryStats)
				auth.GET("/stats/overview", statsHandler.Overview)

				// AppKey 管理
				auth.GET("/appkeys", appKeyHandler.List)
				auth.GET("/appkeys/:id", appKeyHandler.Get)
				auth.POST("/appkeys", appKeyHandler.Create)
				auth.PUT("/appkeys/:id", appKeyHandler.Update)
				auth.DELETE("/appkeys/:id", appKeyHandler.Delete)
				auth.POST("/appkeys/:id/regenerate", appKeyHandler.RegenerateSecret)
			}
		}

		// ============ APP端 API ============
		app := v1.Group("/app")
		app.Use(middleware.SignatureAuth(appKeyRepo))
		{
			// 类别列表
			app.GET("/categories", categoryHandler.ListActive)

			// 获取某类别的最新版本
			app.GET("/categories/:code/latest", packageHandler.GetLatestByCategory)

			// 下载
			app.GET("/download/:id", packageHandler.Download)
		}
	}
}
