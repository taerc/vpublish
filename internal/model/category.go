package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 软件类别
type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:100;uniqueIndex;not null" json:"name"` // 中文名称
	Code        string         `gorm:"size:100;uniqueIndex;not null" json:"code"` // 代码枚举 TYPE_WU_REN_JI
	Description string         `gorm:"size:500" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Packages []Package `gorm:"foreignKey:CategoryID" json:"packages,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// Package 软件包
type Package struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CategoryID  uint           `gorm:"not null;index" json:"category_id"`
	Name        string         `gorm:"size:200;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedBy   uint           `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Category      *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Versions      []Version `gorm:"foreignKey:PackageID" json:"versions,omitempty"`
	Creator       *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	LatestVersion *Version  `gorm:"-" json:"latest_version,omitempty"` // 最新版本（非数据库字段）
}

func (Package) TableName() string {
	return "packages"
}
