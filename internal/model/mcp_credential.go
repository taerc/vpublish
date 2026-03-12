package model

import (
	"time"

	"gorm.io/gorm"
)

// MCPCredential MCP服务认证凭证
// swagger:model MCPCredential
type MCPCredential struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 凭证名称
	Name string `gorm:"size:100;not null" json:"name" example:"AI助手集成"`
	// AppKey
	AppKey string `gorm:"size:64;uniqueIndex;not null" json:"app_key" example:"mcp_abc123def456"`
	// AppSecret (不返回)
	AppSecret string `gorm:"size:128;not null" json:"-"`
	// 权限级别
	PermissionLevel string `gorm:"size:20;not null;default:'read_only'" json:"permission_level" example:"read_only"`
	// 描述
	Description string `gorm:"size:200" json:"description" example:"用于AI助手集成的MCP凭证"`
	// 是否启用
	IsActive bool `gorm:"default:true" json:"is_active" example:"true"`
	// 最后使用时间
	LastUsedAt *time.Time `json:"last_used_at,omitempty" example:"2024-03-12T15:30:00Z"`
	// 过期时间
	ExpiresAt *time.Time `json:"expires_at,omitempty" example:"2025-03-12T00:00:00Z"`
	// 创建者ID
	CreatedBy uint `json:"created_by" example:"1"`
	// 创建时间
	CreatedAt time.Time `json:"created_at" example:"2024-03-12T10:00:00Z"`
	// 更新时间
	UpdatedAt time.Time `json:"updated_at" example:"2024-03-12T15:30:00Z"`
	// 删除时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MCPCredential) TableName() string {
	return "mcp_credentials"
}

// PermissionLevel constants
const (
	PermissionReadOnly  = "read_only"
	PermissionReadWrite = "read_write"
)

// CanWrite checks if the credential has write permission
func (c *MCPCredential) CanWrite() bool {
	return c.PermissionLevel == PermissionReadWrite
}

// IsExpired checks if the credential is expired
func (c *MCPCredential) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsValid checks if the credential is valid for use
func (c *MCPCredential) IsValid() bool {
	return c.IsActive && !c.IsExpired()
}
