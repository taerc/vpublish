package service

import (
	"context"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
)

type StatsService struct {
	statsRepo *repository.StatsRepository
}

func NewStatsService(statsRepo *repository.StatsRepository) *StatsService {
	return &StatsService{statsRepo: statsRepo}
}

// RecordDownload 记录下载
func (s *StatsService) RecordDownload(ctx context.Context, versionID, categoryID uint, appKey, clientIP, userAgent string) error {
	// 记录原始日志
	log := &model.DownloadLog{
		VersionID:    versionID,
		AppKey:       appKey,
		ClientIP:     clientIP,
		UserAgent:    userAgent,
		DownloadedAt: time.Now(),
	}
	if err := s.statsRepo.CreateDownloadLog(ctx, log); err != nil {
		return err
	}

	// 更新日统计
	today := time.Now().Truncate(24 * time.Hour)
	stat := &model.DownloadStat{
		VersionID:     versionID,
		CategoryID:    categoryID,
		StatDate:      today,
		DownloadCount: 1,
	}
	return s.statsRepo.UpsertDownloadStat(ctx, stat)
}

// GetDailyStats 获取每日统计
func (s *StatsService) GetDailyStats(ctx context.Context, categoryID uint, date time.Time) (int64, error) {
	return s.statsRepo.GetDailyStats(ctx, categoryID, date)
}

// GetMonthlyStats 获取月度统计
func (s *StatsService) GetMonthlyStats(ctx context.Context, categoryID uint, year, month int) (int64, error) {
	return s.statsRepo.GetMonthlyStats(ctx, categoryID, year, month)
}

// GetYearlyStats 获取年度统计
func (s *StatsService) GetYearlyStats(ctx context.Context, categoryID uint, year int) (int64, error) {
	return s.statsRepo.GetYearlyStats(ctx, categoryID, year)
}

// GetCategoryStats 按类别统计
func (s *StatsService) GetCategoryStats(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return s.statsRepo.GetCategoryStats(ctx, startDate, endDate)
}

// GetDailyTrend 获取每日下载趋势
func (s *StatsService) GetDailyTrend(ctx context.Context, categoryID uint, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	return s.statsRepo.GetDailyTrend(ctx, categoryID, startDate, endDate)
}
