package database

import (
	"fmt"
	"log"
	"time"

	"github.com/taerc/vpublish/internal/config"
	"github.com/taerc/vpublish/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.AppKey{},
		&model.Category{},
		&model.Package{},
		&model.Version{},
		&model.DownloadLog{},
		&model.DownloadStat{},
		&model.MCPCredential{},
		&model.ErrorRecord{},
	); err != nil {
		return err
	}

	// feature_type 迁移：设置已有记录的默认值
	// GORM AutoMigrate 只对新行设置默认值，需手动更新已有数据
	if err := db.Exec("UPDATE versions SET feature_type = 'release' WHERE feature_type IS NULL OR feature_type = ''").Error; err != nil {
		log.Printf("warning: migrate feature_type default value: %v", err)
		// 不阻断迁移，feature_type 可能已经存在
	}

	// 设置列默认值和 NOT NULL 约束
	if err := db.Exec("ALTER TABLE versions MODIFY COLUMN feature_type VARCHAR(20) NOT NULL DEFAULT 'release'").Error; err != nil {
		log.Printf("warning: modify feature_type column definition: %v", err)
	}

	return nil
}
