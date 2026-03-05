package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
	"github.com/taerc/vpublish/pkg/signature"
)

type PackageHandler struct {
	packageService *service.PackageService
	statsRepo      *repository.StatsRepository
	appKeyRepo     *repository.AppKeyRepository
}

func NewPackageHandler(
	packageService *service.PackageService,
	statsRepo *repository.StatsRepository,
	appKeyRepo *repository.AppKeyRepository,
) *PackageHandler {
	return &PackageHandler{
		packageService: packageService,
		statsRepo:      statsRepo,
		appKeyRepo:     appKeyRepo,
	}
}

// List 软件包列表
func (h *PackageHandler) List(c *gin.Context) {
	categoryID := middleware.ParseIntQuery(c, "category_id", 0)
	page := middleware.ParseIntQuery(c, "page", 1)
	pageSize := middleware.ParseIntQuery(c, "page_size", 20)

	packages, total, err := h.packageService.List(c.Request.Context(), uint(categoryID), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get packages")
		return
	}

	response.Page(c, packages, total, page, pageSize)
}

// Get 获取单个软件包
func (h *PackageHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	pkg, err := h.packageService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "package not found")
		return
	}

	response.Success(c, pkg)
}

// Create 创建软件包（含第一个版本上传）
func (h *PackageHandler) Create(c *gin.Context) {
	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	// 解析表单数据
	var req service.CreatePackageWithVersionRequest
	categoryIDStr := c.PostForm("category_id")
	if categoryIDStr == "" {
		response.BadRequest(c, "category_id is required")
		return
	}
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid category_id")
		return
	}

	req.CategoryID = uint(categoryID)
	req.Version = c.PostForm("version")
	req.Description = c.PostForm("description")
	req.Changelog = c.PostForm("changelog")
	req.ForceUpgrade = c.PostForm("force_upgrade") == "true"
	req.File = file

	if req.Version == "" {
		response.BadRequest(c, "version is required")
		return
	}

	userID := middleware.GetUserID(c)
	pkg, version, err := h.packageService.CreateWithVersion(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, gin.H{
		"package": pkg,
		"version": version,
	})
}

// Update 更新软件包
func (h *PackageHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req service.UpdatePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request")
		return
	}

	pkg, err := h.packageService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, pkg)
}

// Delete 删除软件包
func (h *PackageHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.packageService.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListVersions 获取软件包版本列表
func (h *PackageHandler) ListVersions(c *gin.Context) {
	packageID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid package id")
		return
	}

	page := middleware.ParseIntQuery(c, "page", 1)
	pageSize := middleware.ParseIntQuery(c, "page_size", 20)

	versions, total, err := h.packageService.ListVersions(c.Request.Context(), uint(packageID), page, pageSize)
	if err != nil {
		response.InternalError(c, "failed to get versions")
		return
	}

	response.Page(c, versions, total, page, pageSize)
}

// UploadVersion 上传新版本
func (h *PackageHandler) UploadVersion(c *gin.Context) {
	packageID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid package id")
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	// 解析表单数据
	var req service.CreateVersionRequest
	req.Version = c.PostForm("version")
	req.Changelog = c.PostForm("changelog")
	req.ReleaseNotes = c.PostForm("release_notes")
	req.MinVersion = c.PostForm("min_version")
	req.ForceUpgrade = c.PostForm("force_upgrade") == "true"
	req.IsStable = c.PostForm("is_stable") != "false" // 默认为稳定版

	if req.Version == "" {
		response.BadRequest(c, "version is required")
		return
	}

	userID := middleware.GetUserID(c)
	version, err := h.packageService.UploadVersion(c.Request.Context(), uint(packageID), userID, file, &req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, version)
}

// DeleteVersion 删除版本
func (h *PackageHandler) DeleteVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid version id")
		return
	}

	if err := h.packageService.DeleteVersion(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetLatestByCategory APP端：根据类别代码获取最新版本
func (h *PackageHandler) GetLatestByCategory(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "category code is required")
		return
	}

	version, err := h.packageService.GetLatestByCategoryCode(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, "no version found for this category")
		return
	}

	// 获取 AppKey 信息用于生成签名链接
	appKey := middleware.GetAppKey(c)
	downloadURL, err := h.packageService.GenerateDownloadURL(c.Request.Context(), version.ID, appKey.AppSecret)
	if err != nil {
		response.InternalError(c, "failed to generate download url")
		return
	}

	// 构建 package 信息（防止空指针）
	packageInfo := gin.H{
		"id":       version.PackageID,
		"name":     "",
		"category": nil,
	}
	if version.Package != nil {
		packageInfo["name"] = version.Package.Name
		if version.Package.Category != nil {
			packageInfo["category"] = gin.H{
				"id":   version.Package.Category.ID,
				"name": version.Package.Category.Name,
				"code": version.Package.Category.Code,
			}
		}
	}

	// 返回版本信息
	resp := gin.H{
		"id":            version.ID,
		"version":       version.Version,
		"version_code":  version.VersionCode,
		"file_name":     version.FileName,
		"file_size":     version.FileSize,
		"file_hash":     version.FileHash,
		"changelog":     version.Changelog,
		"release_notes": version.ReleaseNotes,
		"min_version":   version.MinVersion,
		"force_upgrade": version.ForceUpgrade,
		"is_stable":     version.IsStable,
		"download_url":  downloadURL,
		"published_at":  version.PublishedAt,
		"package":       packageInfo,
	}

	response.Success(c, resp)
}

// Download 下载软件包
func (h *PackageHandler) Download(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid version id")
		return
	}

	// 获取签名参数
	token := c.Query("token")
	expiresStr := c.Query("expires")
	if token == "" || expiresStr == "" {
		response.Unauthorized(c, "missing download token")
		return
	}

	// 验证签名
	version, err := h.packageService.GetVersionByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "version not found")
		return
	}

	// 获取 AppKey（需要从请求中识别）
	appKeyHeader := c.GetHeader(signature.HeaderAppKey)
	if appKeyHeader == "" {
		response.Unauthorized(c, "missing app key")
		return
	}

	appKey, err := h.appKeyRepo.GetByKey(c.Request.Context(), appKeyHeader)
	if err != nil {
		response.Unauthorized(c, "invalid app key")
		return
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid expires")
		return
	}

	if !signature.VerifyDownloadToken(uint(id), appKey.AppSecret, expires, token) {
		response.Unauthorized(c, "invalid download token")
		return
	}

	// 提前保存需要的数据，避免在 goroutine 中使用请求上下文
	versionID := uint(id)
	appKeyVal := appKeyHeader
	clientIP := middleware.GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	downloadedAt := middleware.GetCurrentTime()

	// 记录下载日志和统计（使用独立的 context，避免请求结束后 context 被取消）
	go func() {
		ctx := context.Background()

		// 1. 记录下载日志
		log := &model.DownloadLog{
			VersionID:    versionID,
			AppKey:       appKeyVal,
			ClientIP:     clientIP,
			UserAgent:    userAgent,
			DownloadedAt: downloadedAt,
		}
		if err := h.statsRepo.CreateDownloadLog(ctx, log); err != nil {
			fmt.Printf("failed to create download log: %v\n", err)
		}

		// 2. 增加版本下载计数
		if err := h.packageService.IncrementDownloadCount(ctx, versionID); err != nil {
			fmt.Printf("failed to increment download count: %v\n", err)
		}

		// 3. 更新统计表
		categoryID := version.Package.CategoryID
		today := downloadedAt.Truncate(24 * time.Hour)
		stat := &model.DownloadStat{
			VersionID:     versionID,
			CategoryID:    categoryID,
			StatDate:      today,
			DownloadCount: 1,
		}
		if err := h.statsRepo.UpsertDownloadStat(ctx, stat); err != nil {
			fmt.Printf("failed to update download stats: %v\n", err)
		}
	}()
	// 发送文件
	filePath := h.packageService.GetFilePath(version)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+version.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.FileAttachment(filePath, version.FileName)
}
// DownloadVersion 管理端下载版本（无需签名验证）
func (h *PackageHandler) DownloadVersion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid version id")
		return
	}

	// 获取版本信息
	version, err := h.packageService.GetVersionByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "version not found")
		return
	}

	// 获取当前用户信息
	userKey := "admin"
	userID := middleware.GetUserID(c)
	if userID > 0 {
		userKey = fmt.Sprintf("admin_%d", userID)
	}

	// 提前保存需要的数据，避免在 goroutine 中使用请求上下文
	versionID := uint(id)
	categoryID := version.Package.CategoryID
	clientIP := middleware.GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	downloadedAt := middleware.GetCurrentTime()

	// 记录下载日志和统计（使用独立的 context，避免请求结束后 context 被取消）
	go func() {
		ctx := context.Background()

		// 1. 记录下载日志
		log := &model.DownloadLog{
			VersionID:    versionID,
			AppKey:       userKey,
			ClientIP:     clientIP,
			UserAgent:    userAgent,
			DownloadedAt: downloadedAt,
		}
		if err := h.statsRepo.CreateDownloadLog(ctx, log); err != nil {
			fmt.Printf("failed to create download log: %v\n", err)
		}

		// 2. 增加版本下载计数
		if err := h.packageService.IncrementDownloadCount(ctx, versionID); err != nil {
			fmt.Printf("failed to increment download count: %v\n", err)
		}

		// 3. 更新统计表（用于图表展示）
		today := downloadedAt.Truncate(24 * time.Hour)
		stat := &model.DownloadStat{
			VersionID:     versionID,
			CategoryID:    categoryID,
			StatDate:      today,
			DownloadCount: 1,
		}
		if err := h.statsRepo.UpsertDownloadStat(ctx, stat); err != nil {
			fmt.Printf("failed to update download stats: %v\n", err)
		}
	}()

	// 发送文件
	filePath := h.packageService.GetFilePath(version)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+version.FileName)
	c.Header("Content-Type", "application/octet-stream")
	c.FileAttachment(filePath, version.FileName)
}
