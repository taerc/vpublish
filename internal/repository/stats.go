package repository

import (
	"context"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// CreateDownloadLog 创建下载日志
func (r *StatsRepository) CreateDownloadLog(ctx context.Context, log *model.DownloadLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetDownloadStats 获取下载统计
func (r *StatsRepository) GetDownloadStats(ctx context.Context, categoryID uint, startDate, endDate time.Time) ([]model.DownloadStat, error) {
	var stats []model.DownloadStat
	query := r.db.WithContext(ctx).Model(&model.DownloadStat{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate)

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Order("stat_date ASC").Find(&stats).Error
	return stats, err
}

// GetDailyTrend 获取每日下载趋势（按日期分组汇总）
func (r *StatsRepository) GetDailyTrend(ctx context.Context, categoryID uint, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	query := r.db.WithContext(ctx).
		Table("download_stats").
		Select("stat_date as date, SUM(download_count) as count").
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Group("stat_date").
		Order("stat_date ASC")

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Find(&results).Error
	return results, err
}

// GetCategoryStats 按类别统计
func (r *StatsRepository) GetCategoryStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.WithContext(ctx).
		Table("download_stats").
		Select("categories.name as category_name, categories.code as category_code, SUM(download_stats.download_count) as total_count").
		Joins("JOIN categories ON categories.id = download_stats.category_id").
		Where("download_stats.stat_date >= ? AND download_stats.stat_date <= ?", startDate, endDate).
		Group("download_stats.category_id").
		Order("total_count DESC").
		Find(&results).Error

	return results, err
}

// UpsertDownloadStat 创建或更新下载统计
func (r *StatsRepository) UpsertDownloadStat(ctx context.Context, stat *model.DownloadStat) error {
	return r.db.WithContext(ctx).
		Assign(map[string]interface{}{
			"download_count": gorm.Expr("download_count + ?", stat.DownloadCount),
		}).
		FirstOrCreate(stat, model.DownloadStat{
			VersionID: stat.VersionID,
			StatDate:  stat.StatDate,
		}).Error
}

// GetDailyStats 获取每日统计
func (r *StatsRepository) GetDailyStats(ctx context.Context, categoryID uint, date time.Time) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.DownloadStat{}).
		Where("stat_date = ?", date)

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Select("COALESCE(SUM(download_count), 0)").Scan(&count).Error
	return count, err
}

// GetMonthlyStats 获取月度统计
func (r *StatsRepository) GetMonthlyStats(ctx context.Context, categoryID uint, year int, month int) (int64, error) {
	var count int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, -1)

	query := r.db.WithContext(ctx).Model(&model.DownloadStat{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate)

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Select("COALESCE(SUM(download_count), 0)").Scan(&count).Error
	return count, err
}

// GetYearlyStats 获取年度统计
func (r *StatsRepository) GetYearlyStats(ctx context.Context, categoryID uint, year int) (int64, error) {
	var count int64
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)

	query := r.db.WithContext(ctx).Model(&model.DownloadStat{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate)

	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Select("COALESCE(SUM(download_count), 0)").Scan(&count).Error
	return count, err
}
