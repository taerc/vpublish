package model

import (
	"time"

	"gorm.io/gorm"
)

// User 管理员用户
// swagger:model User
type User struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 用户名
	Username string `gorm:"size:50;uniqueIndex;not null" json:"username" example:"admin"`
	// 密码哈希 (不返回)
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	// 昵称
	Nickname string `gorm:"size:50" json:"nickname" example:"系统管理员"`
	// 邮箱
	Email string `gorm:"size:100" json:"email" example:"admin@example.com"`
	// 角色
	Role string `gorm:"size:20;default:user" json:"role" example:"admin"`
	// 是否启用
	IsActive bool `gorm:"default:true" json:"is_active" example:"true"`
	// 最后登录时间
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2024-03-12T15:30:00Z"`
	// 最后登录IP
	LastLoginIP string `gorm:"size:45" json:"last_login_ip,omitempty" example:"192.168.1.100"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T08:00:00Z"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-12T15:30:00Z"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// AppKey APP认证密钥
// swagger:model AppKey
type AppKey struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 应用名称
	AppName string `gorm:"size:100;not null" json:"app_name" example:"移动客户端"`
	// AppKey
	AppKey string `gorm:"size:64;uniqueIndex;not null" json:"app_key" example:"abc123def456"`
	// AppSecret (不返回)
	AppSecret string `gorm:"size:64;not null" json:"-"`
	// 描述
	Description string `gorm:"size:200" json:"description" example:"用于移动APP的API访问密钥"`
	// 是否启用
	IsActive bool `gorm:"default:true" json:"is_active" example:"true"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-03-12T10:00:00Z"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-12T15:30:00Z"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AppKey) TableName() string {
	return "app_keys"
}
