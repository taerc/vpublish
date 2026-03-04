package model

import (
	"time"
)

// DownloadLog 下载日志
type DownloadLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	VersionID    uint      `gorm:"not null;index" json:"version_id"`
	AppKey       string    `gorm:"size:64" json:"app_key"`
	ClientIP     string    `gorm:"size:45" json:"client_ip"`
	UserAgent    string    `gorm:"size:500" json:"user_agent"`
	DownloadedAt time.Time `gorm:"index" json:"downloaded_at"`
}

func (DownloadLog) TableName() string {
	return "download_logs"
}

// DownloadStat 下载统计（按天聚合）
type DownloadStat struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	VersionID     uint      `gorm:"not null;uniqueIndex:uk_version_date" json:"version_id"`
	CategoryID    uint      `gorm:"not null;index" json:"category_id"`
	StatDate      time.Time `gorm:"type:date;not null;uniqueIndex:uk_version_date" json:"stat_date"`
	DownloadCount int       `gorm:"default:0" json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (DownloadStat) TableName() string {
	return "download_stats"
}

// OperationLog 操作日志
type OperationLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"user_id"`
	Action       string    `gorm:"size:50;not null;index" json:"action"`
	ResourceType string    `gorm:"size:50" json:"resource_type"`
	ResourceID   uint      `json:"resource_id"`
	Detail       string    `gorm:"type:text" json:"detail"`
	IP           string    `gorm:"size:45" json:"ip"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
