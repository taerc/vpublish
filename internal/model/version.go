package model

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)
// Version 软件版本
type Version struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	PackageID   uint   `gorm:"not null;index" json:"package_id"`
	Version     string `gorm:"size:50;not null" json:"version"`    // 版本号字符串 1.0.0
	VersionCode int    `gorm:"not null;index" json:"version_code"` // 版本号数值 10000

	// 文件信息
	FilePath string `gorm:"size:500;not null" json:"-"`         // 存储路径（内部）
	FileName string `gorm:"size:255;not null" json:"file_name"` // 原始文件名
	FileSize int64  `gorm:"not null" json:"file_size"`          // 文件大小
	FileHash string `gorm:"size:64;not null" json:"file_hash"`  // SHA256

	// 版本信息
	Changelog    string `gorm:"type:text" json:"changelog"`     // 更新日志
	ReleaseNotes string `gorm:"type:text" json:"release_notes"` // 发布说明
	MinVersion   string `gorm:"size:50" json:"min_version"`     // 最低兼容版本

	// 升级控制
	ForceUpgrade bool `gorm:"default:false" json:"force_upgrade"` // 强制升级
	IsLatest     bool `gorm:"default:false" json:"is_latest"`     // 最新版本
	IsStable     bool `gorm:"default:true" json:"is_stable"`      // 稳定版

	// 统计
	DownloadCount int `gorm:"default:0" json:"download_count"`

	// 审计
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
	PublishedAt *time.Time     `json:"published_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Package *Package `gorm:"foreignKey:PackageID" json:"package,omitempty"`
	Creator *User    `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Version) TableName() string {
	return "versions"
}

// GetDownloadURL 生成下载URL
func (v *Version) GetDownloadURL(baseURL string) string {
	return baseURL + "/api/v1/app/download/" + strconv.FormatUint(uint64(v.ID), 10)
}
