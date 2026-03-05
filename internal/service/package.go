package service

import (
	"context"
"errors"
	"regexp"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/taerc/vpublish/pkg/signature"
	"github.com/taerc/vpublish/pkg/storage"
)

var (
	ErrPackageNotFound      = errors.New("package not found")
	ErrVersionNotFound      = errors.New("version not found")
	ErrVersionAlreadyExists = errors.New("version already exists")
	ErrInvalidVersionFormat = errors.New("版本号格式必须为 v1.0.0 或 1.0.0")
	ErrVersionMustBeGreater = errors.New("新版本号必须大于已有版本号")
)

// versionRegex 版本号格式正则表达式
// 允许格式: v1.0.0, V1.0.0, 1.0.0
var versionRegex = regexp.MustCompile(`^(v|V)?\d+\.\d+\.\d+$`)

// validateVersionFormat 校验版本号格式
// 允许格式: v1.0.0, V1.0.0, 1.0.0
// 必须为三段数字，可选 v/V 前缀
func validateVersionFormat(version string) error {
	version = strings.TrimSpace(version)
	if !versionRegex.MatchString(version) {
		return ErrInvalidVersionFormat
	}
	return nil
}

type PackageService struct {
	packageRepo  *repository.PackageRepository
	versionRepo  *repository.VersionRepository
	categoryRepo *repository.CategoryRepository
	storage      *storage.LocalStorage
	baseURL      string
}

func NewPackageService(
	packageRepo *repository.PackageRepository,
	versionRepo *repository.VersionRepository,
	categoryRepo *repository.CategoryRepository,
	storage *storage.LocalStorage,
	baseURL string,
) *PackageService {
	return &PackageService{
		packageRepo:  packageRepo,
		versionRepo:  versionRepo,
		categoryRepo: categoryRepo,
		storage:      storage,
		baseURL:      baseURL,
	}
}

// CreatePackageWithVersion 创建软件包并上传第一个版本
type CreatePackageWithVersionRequest struct {
	CategoryID   uint                  `json:"category_id" binding:"required"`
	Version      string                `json:"version" binding:"required"`
	Description  string                `json:"description"`
	Changelog    string                `json:"changelog"`
	ForceUpgrade bool                  `json:"force_upgrade"` // 是否强制升级
	File         *multipart.FileHeader `json:"-"`
}

type UpdatePackageRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type CreateVersionRequest struct {
	Version      string `json:"version" binding:"required"`
	Changelog    string `json:"changelog"`
	ReleaseNotes string `json:"release_notes"`
	MinVersion   string `json:"min_version"`
	ForceUpgrade bool   `json:"force_upgrade"`
	IsStable     bool   `json:"is_stable"`
}

// CreateWithVersion 创建软件包并上传第一个版本
func (s *PackageService) CreateWithVersion(
	ctx context.Context,
	userID uint,
	req *CreatePackageWithVersionRequest,
) (*model.Package, *model.Version, error) {
	// 验证类别存在
	category, err := s.categoryRepo.GetByID(ctx, req.CategoryID)
	if err != nil {
		return nil, nil, errors.New("category not found")
	}

	// 从文件名提取软件包名称
	packageName := extractPackageName(req.File.Filename)

	// 检查同一类别下是否已存在同名软件包
	existingPkg, err := s.packageRepo.GetByCategoryAndName(ctx, req.CategoryID, packageName)
	if err == nil {
		if exists, _ := s.versionRepo.ExistsByPackageAndVersion(ctx, existingPkg.ID, req.Version); exists {
			return nil, nil, ErrVersionAlreadyExists
		}
		return nil, nil, errors.New("软件包已存在，请使用上传新版本接口")
	}

	// 校验版本号格式
	if err := validateVersionFormat(req.Version); err != nil {
		return nil, nil, err
	}

	// 解析版本号
	versionCode, err := parseVersionCode(req.Version)
	if err != nil {
		return nil, nil, ErrInvalidVersionFormat
	}

	// 保存文件
	filePath, fileSize, fileHash, err := s.storage.Save(req.File, category.Code)
	if err != nil {
		return nil, nil, fmt.Errorf("save file: %w", err)
	}

	// 创建软件包
	pkg := &model.Package{
		CategoryID:  category.ID,
		Name:        packageName,
		Description: req.Description,
		IsActive:    true,
		CreatedBy:   userID,
	}

	if err := s.packageRepo.Create(ctx, pkg); err != nil {
		s.storage.Delete(filePath)
		return nil, nil, err
	}

	// 创建版本
	now := time.Now()
	version := &model.Version{
		PackageID:    pkg.ID,
		Version:      req.Version,
		VersionCode:  versionCode,
		FilePath:     filePath,
		FileName:     req.File.Filename,
		FileSize:     fileSize,
		FileHash:     fileHash,
		Changelog:    req.Changelog,
		ForceUpgrade: req.ForceUpgrade,
		IsLatest:     true,
		IsStable:     true,
		CreatedBy:    userID,
		PublishedAt:  &now,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		s.packageRepo.Delete(ctx, pkg.ID)
		s.storage.Delete(filePath)
		return nil, nil, err
	}

	return pkg, version, nil
}

// extractPackageName 从文件名提取软件包名称
func extractPackageName(filename string) string {
	// 去除扩展名
	name := filename
	ext := filepath.Ext(filename)
	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}

	// 去除版本号等后缀（如 _v1.0.0, -1.0.0, _20231201）
	// 常见格式: app_v1.0.0.apk -> app
	for _, sep := range []string{"_v", "-v", "_V", "-V", "_", "-"} {
		if idx := strings.Index(name, sep); idx > 0 {
			// 检查分隔符后面是否是版本号格式
			suffix := name[idx+1:]
			if isVersionLike(suffix) {
				name = name[:idx]
				break
			}
		}
	}

	return name
}

// isVersionLike 检查字符串是否像版本号
func isVersionLike(s string) bool {
	// 移除开头的 v 或 V
	s = strings.TrimLeft(s, "vV")

	// 检查是否符合版本号格式 (如 1.0.0, 1.0, 1.0.0.0)
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return false
	}

	digitCount := 0
	for _, p := range parts {
		if _, err := strconv.Atoi(p); err == nil {
			digitCount++
		}
	}

	return digitCount >= len(parts)/2
}

func (s *PackageService) Update(ctx context.Context, id uint, req *UpdatePackageRequest) (*model.Package, error) {
	pkg, err := s.packageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPackageNotFound
	}

	if req.Name != "" {
		pkg.Name = req.Name
	}
	if req.Description != "" {
		pkg.Description = req.Description
	}
	if req.IsActive != nil {
		pkg.IsActive = *req.IsActive
	}

	if err := s.packageRepo.Update(ctx, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

func (s *PackageService) Delete(ctx context.Context, id uint) error {
	return s.packageRepo.Delete(ctx, id)
}

func (s *PackageService) GetByID(ctx context.Context, id uint) (*model.Package, error) {
	return s.packageRepo.GetByID(ctx, id)
}

func (s *PackageService) List(ctx context.Context, categoryID uint, page, pageSize int) ([]model.Package, int64, error) {
	return s.packageRepo.List(ctx, categoryID, page, pageSize)
}

// UploadVersion 上传新版本
func (s *PackageService) UploadVersion(
	ctx context.Context,
	packageID uint,
	userID uint,
	file *multipart.FileHeader,
	req *CreateVersionRequest,
) (*model.Version, error) {
	// 验证软件包存在
	pkg, err := s.packageRepo.GetByID(ctx, packageID)
	if err != nil {
		return nil, ErrPackageNotFound
	}

	// 检查版本是否已存在
	if exists, _ := s.versionRepo.ExistsByPackageAndVersion(ctx, packageID, req.Version); exists {
		return nil, ErrVersionAlreadyExists
	}

	// 校验版本号格式
	if err := validateVersionFormat(req.Version); err != nil {
		return nil, err
	}

	// 解析版本号
	versionCode, err := parseVersionCode(req.Version)
	if err != nil {
		return nil, ErrInvalidVersionFormat
	}

	// 检查新版本号必须大于已有最大版本号
	maxVersionCode, err := s.versionRepo.GetMaxVersionCode(ctx, packageID)
	if err != nil {
		return nil, fmt.Errorf("get max version: %w", err)
	}
	if versionCode <= maxVersionCode {
		return nil, ErrVersionMustBeGreater
	}

	// 保存文件
	categoryCode := pkg.Category.Code
	filePath, fileSize, fileHash, err := s.storage.Save(file, categoryCode)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	// 清除之前的最新版本标记
	s.versionRepo.ClearLatestFlag(ctx, packageID)

	now := time.Now()
	version := &model.Version{
		PackageID:    packageID,
		Version:      req.Version,
		VersionCode:  versionCode,
		FilePath:     filePath,
		FileName:     file.Filename,
		FileSize:     fileSize,
		FileHash:     fileHash,
		Changelog:    req.Changelog,
		ReleaseNotes: req.ReleaseNotes,
		MinVersion:   req.MinVersion,
		ForceUpgrade: req.ForceUpgrade,
		IsLatest:     true,
		IsStable:     req.IsStable,
		CreatedBy:    userID,
		PublishedAt:  &now,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		// 删除已上传的文件
		s.storage.Delete(filePath)
		return nil, err
	}

	return version, nil
}

// GetLatestByCategoryCode 根据类别代码获取最新版本
func (s *PackageService) GetLatestByCategoryCode(ctx context.Context, categoryCode string) (*model.Version, error) {
	return s.versionRepo.GetLatestByCategoryCode(ctx, categoryCode)
}

// GetVersionByID 获取版本详情
func (s *PackageService) GetVersionByID(ctx context.Context, id uint) (*model.Version, error) {
	return s.versionRepo.GetByID(ctx, id)
}

// ListVersions 列出软件包的所有版本
func (s *PackageService) ListVersions(ctx context.Context, packageID uint, page, pageSize int) ([]model.Version, int64, error) {
	return s.versionRepo.ListByPackage(ctx, packageID, page, pageSize)
}

// DeleteVersion 删除版本
func (s *PackageService) DeleteVersion(ctx context.Context, id uint) error {
	version, err := s.versionRepo.GetByID(ctx, id)
	if err != nil {
		return ErrVersionNotFound
	}

	// 删除文件
	if err := s.storage.Delete(version.FilePath); err != nil {
		// 记录错误但不阻止删除
	}

	return s.versionRepo.Delete(ctx, id)
}

// GenerateDownloadURL 生成带签名的下载URL
func (s *PackageService) GenerateDownloadURL(ctx context.Context, versionID uint, appSecret string) (string, error) {
	_, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return "", ErrVersionNotFound
	}

	// 生成过期时间（1小时后）
	expires := time.Now().Add(time.Hour).Unix()

	// 生成签名令牌
	token := signature.GenerateDownloadToken(versionID, appSecret, expires)

	// 构建下载URL
	return fmt.Sprintf("%s/api/v1/app/download/%d?token=%s&expires=%d",
		s.baseURL, versionID, token, expires), nil
}

// GetFilePath 获取文件路径
func (s *PackageService) GetFilePath(version *model.Version) string {
	return s.storage.GetFilePath(version.FilePath)
}

// IncrementDownloadCount 增加版本下载计数
func (s *PackageService) IncrementDownloadCount(ctx context.Context, versionID uint) error {
	return s.versionRepo.IncrementDownloadCount(ctx, versionID)
}

// parseVersionCode 将版本字符串转换为数字
// 例如: "1.2.3" -> 100020003
func parseVersionCode(version string) (int, error) {
	parts := strings.Split(version, ".")
	if len(parts) > 3 {
		parts = parts[:3]
	}

	code := 0
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return 0, err
		}
		multiplier := 1
		for j := 0; j < 2-i; j++ {
			multiplier *= 1000
		}
		code += num * multiplier
	}

	return code, nil
}
