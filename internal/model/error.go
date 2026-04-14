package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrorRecord 报错记录模型
// swagger:model ErrorRecord
type ErrorRecord struct {
	// 主键ID
	ID uint `gorm:"primaryKey" json:"id" example:"1"`
	// 接口请求唯一标识
	RequestID string `gorm:"size:128;not null" json:"request_id" example:"req-20260330-001"`
	// 报错发生时间戳（毫秒）
	Timestamp int64 `gorm:"not null;index" json:"timestamp" example:"1711737600000"`
	// 报错模块（自动识别）
	Module string `gorm:"size:64;index" json:"module" example:"user"`
	// 应用类型：app（来自 /api/v1/app/* 接口）/platform（来自 /api/v1/admin/* 接口）
	// 系统根据报错的原始接口路径自动识别
	AppType string `gorm:"size:32;not null" json:"app_type" example:"app"`
	// 接口路径
	Path string `gorm:"size:255;not null" json:"path" example:"/api/v1/app/packages"`
	// 接口返回的错误码
	Code string `gorm:"size:64;not null;index" json:"code" example:"500"`
	// 接口返回的消息信息
	ErrorMessage string `gorm:"size:500;not null" json:"error_message" example:"Internal Server Error"`
	// 接口入参
	RequestParams JSONMap `gorm:"type:json" json:"request_params"`
	// 设备信息
	DeviceInfo JSONMap `gorm:"type:json" json:"device_info"`
	// 报错类型：system_error/business_error
	ErrorType string `gorm:"size:32;not null" json:"error_type" example:"system_error"`
	// 备注
	Remark string `gorm:"size:500" json:"remark" example:"已处理"`
	// 创建时间
	CreatedAt time.Time `gorm:"default:0" json:"created_at" example:"2026-03-30T10:00:00Z"`
	// 更新时间
	UpdatedAt time.Time `gorm:"default:0" json:"updated_at" example:"2026-03-30T15:30:00Z"`
	// 软删除时间
	DeletedAt gorm.DeletedAt `gorm:"default:0" json:"-"`
}

// TableName 表名
func (ErrorRecord) TableName() string {
	return "error_records"
}

// JSONMap 用于存储JSON对象的map类型
type JSONMap map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}
