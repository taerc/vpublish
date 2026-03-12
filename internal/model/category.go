package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 软件类别
// swagger:model Category
type Category struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 类别名称
	Name string `gorm:"size:100;uniqueIndex;not null" json:"name" example:"无人机应用"`
	// 类别代码 (自动生成的拼音代码)
	Code string `gorm:"size:100;uniqueIndex;not null" json:"code" example:"TYPE_WU_REN_JI_YING_YONG"`
	// 描述
	Description string `gorm:"size:500" json:"description" example:"各类无人机相关应用程序"`
	// 排序值
	SortOrder int `gorm:"default:0" json:"sort_order" example:"10"`
	// 是否启用
	IsActive bool `gorm:"default:true" json:"is_active" example:"true"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-03-12T10:00:00Z"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-12T15:30:00Z"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Packages []Package `gorm:"foreignKey:CategoryID" json:"packages,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// Package 软件包
// swagger:model Package
type Package struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 类别ID
	CategoryID uint `gorm:"not null;index" json:"category_id" example:"1"`
	// 软件包名称
	Name string `gorm:"size:200;not null" json:"name" example:"无人机控制系统"`
	// 描述
	Description string `gorm:"type:text" json:"description" example:"专业级无人机飞行控制系统"`
	// 是否启用
	IsActive bool `gorm:"default:true" json:"is_active" example:"true"`
	// 创建者ID
	CreatedBy uint `gorm:"not null" json:"created_by" example:"1"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-03-12T10:00:00Z"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-12T15:30:00Z"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Category      *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Versions      []Version `gorm:"foreignKey:PackageID" json:"versions,omitempty"`
	Creator       *User     `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	LatestVersion *Version  `gorm:"-" json:"latest_version,omitempty"`
}
