package model

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Version 软件版本
// swagger:model Version
type Version struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 软件包ID
	PackageID uint `gorm:"not null;index" json:"package_id" example:"1"`
	// 版本号
	Version string `gorm:"size:50;not null" json:"version" example:"2.1.0"`
	// 版本代码 (数值)
	VersionCode int `gorm:"not null;index" json:"version_code" example:"20100"`

	// 文件信息
	// 文件存储路径 (内部使用)
	FilePath string `gorm:"size:500;not null" json:"-"`
	// 原始文件名
	FileName string `gorm:"size:255;not null" json:"file_name" example:"drone-control-2.1.0.zip"`
	// 文件大小 (字节)
	FileSize int64 `gorm:"not null" json:"file_size" example:"15728640"`
	// 文件SHA256哈希
	FileHash string `gorm:"size:64;not null" json:"file_hash" example:"a1b2c3d4e5f6..."`

	// 版本信息
	// 更新日志
	Changelog string `gorm:"type:text" json:"changelog" example:"修复关键安全漏洞\n优化性能"`
	// 发布说明
	ReleaseNotes string `gorm:"type:text" json:"release_notes" example:"本次更新修复了多个已知问题"`
	// 最低兼容版本
	MinVersion string `gorm:"size:50" json:"min_version" example:"2.0.0"`

	// 升级控制
	// 是否强制升级
	ForceUpgrade bool `gorm:"default:false" json:"force_upgrade" example:"false"`
	// 是否最新版本
	IsLatest bool `gorm:"default:false" json:"is_latest" example:"true"`
	// 是否稳定版
	IsStable bool `gorm:"default:true" json:"is_stable" example:"true"`

	// 统计
	// 下载次数
	DownloadCount int `gorm:"default:0" json:"download_count" example:"1523"`

	// 审计
	// 创建者ID
	CreatedBy uint `gorm:"not null" json:"created_by" example:"1"`
	// 发布时间
	PublishedAt *time.Time `json:"published_at,omitempty" example:"2024-03-12T16:00:00Z"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-03-12T15:45:00Z"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-12T16:00:00Z"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

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
