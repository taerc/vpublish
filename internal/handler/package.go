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
// @Summary 获取软件包分页列表
// @Description 获取软件包列表，支持按类别筛选和分页
// @Tags 管理员/软件包管理
// @Accept json
// @Produce json
// @Param category_id query int false "类别ID" minimum(0)
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.Package}} "软件包列表"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/packages [get]
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
// @Summary 获取软件包详情
// @Description 根据ID获取单个软件包的详细信息
// @Tags 管理员/软件包管理
// @Accept json
// @Produce json
// @Param id path int true "软件包ID" minimum(1)
// @Success 200 {object} response.Response{data=model.Package} "软件包详情"
// @Failure 400 {object} response.Response "无效的ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "软件包不存在"
// @Security BearerAuth []
// @Router /admin/packages/{id} [get]
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
// @Summary 创建软件包
// @Description 创建新的软件包并上传第一个版本文件
// @Tags 管理员/软件包管理
// @Accept mpfd
// @Produce json
// @Param file formData file true "软件包文件"
// @Param category_id formData string true "类别ID"
// @Param version formData string true "版本号" example("1.0.0")
// @Param description formData string false "软件包描述"
// @Param changelog formData string false "更新日志"
// @Param force_upgrade formData bool false "是否强制升级" default(false)
// @Success 200 {object} response.Response{data=map[string]interface{}} "创建成功，返回软件包和版本信息"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/packages [post]
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
// @Summary 更新软件包
// @Description 更新软件包信息
// @Tags 管理员/软件包管理
// @Accept json
// @Produce json
// @Param id path int true "软件包ID" minimum(1)
// @Param request body service.UpdatePackageRequest true "更新请求参数"
// @Success 200 {object} response.Response{data=model.Package} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "软件包不存在"
// @Security BearerAuth []
// @Router /admin/packages/{id} [put]
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
// @Summary 删除软件包
// @Description 删除指定的软件包及其所有版本
// @Tags 管理员/软件包管理
// @Accept json
// @Produce json
// @Param id path int true "软件包ID" minimum(1)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "无效的ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "软件包不存在"
// @Security BearerAuth []
// @Router /admin/packages/{id} [delete]
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
// @Summary 获取版本列表
// @Description 获取指定软件包的所有版本列表，支持分页
// @Tags 管理员/版本管理
// @Accept json
// @Produce json
// @Param id path int true "软件包ID" minimum(1)
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20)
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.Version}} "版本列表"
// @Failure 400 {object} response.Response "无效的软件包ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/packages/{id}/versions [get]
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
// @Summary 上传新版本
// @Description 为指定软件包上传新版本文件
// @Tags 管理员/版本管理
// @Accept mpfd
// @Produce json
// @Param id path int true "软件包ID" minimum(1)
// @Param file formData file true "版本文件"
// @Param version formData string true "版本号" example("1.0.1")
// @Param changelog formData string false "更新日志"
// @Param release_notes formData string false "发布说明"
// @Param min_version formData string false "最低兼容版本" example("1.0.0")
// @Param force_upgrade formData bool false "是否强制升级" default(false)
// @Param is_stable formData bool false "是否稳定版" default(true)
// @Success 200 {object} response.Response{data=model.Version} "上传成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "软件包不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/packages/{id}/versions [post]
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
// @Summary 删除版本
// @Description 删除指定的软件包版本
// @Tags 管理员/版本管理
// @Accept json
// @Produce json
// @Param id path int true "版本ID" minimum(1)
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "无效的版本ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "版本不存在"
// @Security BearerAuth []
// @Router /admin/versions/{id} [delete]
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
// @Summary 获取类别最新版本
// @Description APP端根据类别代码获取该类别下软件包的最新版本信息，包含带签名的下载链接
// @Tags APP端
// @Accept json
// @Produce json
// @Param code path string true "类别代码" example("TYPE_WU_REN_JI")
// @Success 200 {object} response.Response{data=map[string]interface{}} "最新版本信息"
// @Failure 400 {object} response.Response "类别代码不能为空"
// @Failure 401 {object} response.Response "认证失败"
// @Failure 404 {object} response.Response "该类别下没有找到版本"
// @Security SignatureAuth []
// @Security TimestampAuth []
// @Security SignatureValueAuth []
// @Router /app/categories/{code}/latest [get]
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

// GetVersionsByCategory APP端：根据类别代码获取版本列表
// @Summary 获取类别版本列表
// @Description APP端根据类别代码获取该类别下软件包的最新5个版本信息列表，包含带签名的下载链接
// @Tags APP端
// @Accept json
// @Produce json
// @Param code path string true "类别代码" example("TYPE_WU_REN_JI")
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "版本列表信息"
// @Failure 400 {object} response.Response "类别代码不能为空"
// @Failure 401 {object} response.Response "认证失败"
// @Failure 404 {object} response.Response "该类别下没有找到版本"
// @Security SignatureAuth []
// @Security TimestampAuth []
// @Security SignatureValueAuth []
// @Router /app/categories/{code}/versions [get]
func (h *PackageHandler) GetVersionsByCategory(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "category code is required")
		return
	}

	// 获取 AppKey 信息用于生成签名链接
	appKey := middleware.GetAppKey(c)

	versions, err := h.packageService.GetVersionsByCategoryCode(c.Request.Context(), code, appKey.AppSecret)
	if err != nil {
		response.NotFound(c, "no versions found for this category")
		return
	}

	if len(versions) == 0 {
		response.NotFound(c, "no versions found for this category")
		return
	}

	response.Success(c, versions)
}

// Download 下载软件包
// @Summary 下载软件包
// @Description APP端下载软件包版本文件，需要提供有效的下载令牌
// @Tags APP端
// @Accept json
// @Produce octet-stream
// @Param id path int true "版本ID" minimum(1)
// @Param token query string true "下载令牌"
// @Param expires query string true "过期时间戳"
// @Success 200 {file} file "文件流"
// @Failure 400 {object} response.Response "无效的版本ID或参数"
// @Failure 401 {object} response.Response "下载令牌无效或已过期"
// @Failure 404 {object} response.Response "版本不存在"
// @Security SignatureAuth []
// @Security TimestampAuth []
// @Security SignatureValueAuth []
// @Router /app/download/{id} [get]
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
// @Summary 管理端下载版本
// @Description 管理员下载软件包版本文件，无需签名验证
// @Tags 管理员/版本管理
// @Accept json
// @Produce octet-stream
// @Param id path int true "版本ID" minimum(1)
// @Success 200 {file} file "文件流"
// @Failure 400 {object} response.Response "无效的版本ID"
// @Failure 401 {object} response.Response "未认证"
// @Failure 404 {object} response.Response "版本不存在"
// @Security BearerAuth []
// @Router /admin/versions/{id}/download [get]
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
