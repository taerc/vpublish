package model

import (
	"time"

	"gorm.io/gorm"
)

// User 管理员用户
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:50" json:"nickname"`
	Email        string         `gorm:"size:100" json:"email"`
	Role         string         `gorm:"size:20;default:user" json:"role"` // admin, user
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"size:45" json:"last_login_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// AppKey APP认证密钥
type AppKey struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	AppName     string         `gorm:"size:100;not null" json:"app_name"`
	AppKey      string         `gorm:"size:64;uniqueIndex;not null" json:"app_key"`
	AppSecret   string         `gorm:"size:64;not null" json:"-"`
	Description string         `gorm:"size:200" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AppKey) TableName() string {
	return "app_keys"
}
