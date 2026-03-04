package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/middleware"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
)

type StatsHandler struct {
	statsService *service.StatsService
}

func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// DailyStats 获取每日下载统计
func (h *StatsHandler) DailyStats(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.BadRequest(c, "invalid date format")
		return
	}

	categoryID := middleware.ParseIntQuery(c, "category_id", 0)
	count, err := h.statsService.GetDailyStats(c.Request.Context(), uint(categoryID), date)
	if err != nil {
		response.InternalError(c, "failed to get stats")
		return
	}

	response.Success(c, gin.H{
		"date":  dateStr,
		"count": count,
	})
}

// DailyTrend 获取每日下载趋势
func (h *StatsHandler) DailyTrend(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	categoryID := middleware.ParseIntQuery(c, "category_id", 0)

	if startDateStr == "" || endDateStr == "" {
		// 默认最近30天
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		startDateStr = startDate.Format("2006-01-02")
		endDateStr = endDate.Format("2006-01-02")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.BadRequest(c, "invalid start_date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.BadRequest(c, "invalid end_date format")
		return
	}

	trend, err := h.statsService.GetDailyTrend(c.Request.Context(), uint(categoryID), startDate, endDate)
	if err != nil {
		response.InternalError(c, "failed to get trend data")
		return
	}

	response.Success(c, gin.H{
		"start_date": startDateStr,
		"end_date":   endDateStr,
		"trend":      trend,
	})
}

// MonthlyStats 获取月度下载统计
func (h *StatsHandler) MonthlyStats(c *gin.Context) {
	year := middleware.ParseIntQuery(c, "year", time.Now().Year())
	month := middleware.ParseIntQuery(c, "month", int(time.Now().Month()))
	categoryID := middleware.ParseIntQuery(c, "category_id", 0)

	count, err := h.statsService.GetMonthlyStats(c.Request.Context(), uint(categoryID), year, month)
	if err != nil {
		response.InternalError(c, "failed to get stats")
		return
	}

	response.Success(c, gin.H{
		"year":        year,
		"month":       month,
		"count":       count,
		"category_id": categoryID,
	})
}

// YearlyStats 获取年度下载统计
func (h *StatsHandler) YearlyStats(c *gin.Context) {
	year := middleware.ParseIntQuery(c, "year", time.Now().Year())
	categoryID := middleware.ParseIntQuery(c, "category_id", 0)

	count, err := h.statsService.GetYearlyStats(c.Request.Context(), uint(categoryID), year)
	if err != nil {
		response.InternalError(c, "failed to get stats")
		return
	}

	response.Success(c, gin.H{
		"year":        year,
		"count":       count,
		"category_id": categoryID,
	})
}

// CategoryStats 按类别统计
func (h *StatsHandler) CategoryStats(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		// 默认最近30天
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		startDateStr = startDate.Format("2006-01-02")
		endDateStr = endDate.Format("2006-01-02")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		response.BadRequest(c, "invalid start_date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		response.BadRequest(c, "invalid end_date format")
		return
	}

	stats, err := h.statsService.GetCategoryStats(c.Request.Context(), startDate, endDate)
	if err != nil {
		response.InternalError(c, "failed to get stats")
		return
	}

	response.Success(c, gin.H{
		"start_date": startDateStr,
		"end_date":   endDateStr,
		"stats":      stats,
	})
}

// Overview 统计概览
func (h *StatsHandler) Overview(c *gin.Context) {
	now := time.Now()

	// 今日统计
	todayCount, _ := h.statsService.GetDailyStats(c.Request.Context(), 0, now)

	// 本月统计
	monthlyCount, _ := h.statsService.GetMonthlyStats(c.Request.Context(), 0, now.Year(), int(now.Month()))

	// 本年统计
	yearlyCount, _ := h.statsService.GetYearlyStats(c.Request.Context(), 0, now.Year())

	// 类别分布
	categoryStats, _ := h.statsService.GetCategoryStats(c.Request.Context(),
		now.AddDate(0, 0, -30), now)

	response.Success(c, gin.H{
		"today":       todayCount,
		"monthly":     monthlyCount,
		"yearly":      yearlyCount,
		"by_category": categoryStats,
	})
}
