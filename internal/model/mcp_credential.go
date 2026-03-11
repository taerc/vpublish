package model

import (
	"time"

	"gorm.io/gorm"
)

// MCPCredential MCP服务认证凭证
type MCPCredential struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:100;not null" json:"name"`
	AppKey          string         `gorm:"size:64;uniqueIndex;not null" json:"app_key"`
	AppSecret       string         `gorm:"size:128;not null" json:"-"`
	PermissionLevel string         `gorm:"size:20;not null;default:'read_only'" json:"permission_level"` // read_only, read_write
	Description     string         `gorm:"size:200" json:"description"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	LastUsedAt      *time.Time     `json:"last_used_at"`
	ExpiresAt       *time.Time     `json:"expires_at"`
	CreatedBy       uint           `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
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
