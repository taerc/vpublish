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
//
// @Summary 获取每日下载统计
// @Description 获取指定日期的下载次数统计，可按类别筛选
// @Tags 管理员/统计
// @Accept json
// @Produce json
// @Param date query string false "日期 (格式: 2006-01-02)" example("2024-03-12")
// @Param category_id query int false "类别ID，为0表示全部类别" example(1)
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回每日下载统计，包含date和count字段"
// @Failure 400 {object} response.Response "请求参数错误，日期格式不正确"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/stats/daily [get]
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
//
// @Summary 获取每日下载趋势
// @Description 获取指定日期范围内的每日下载趋势数据，可按类别筛选
// @Tags 管理员/统计
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期 (格式: 2006-01-02)" example("2024-03-01")
// @Param end_date query string false "结束日期 (格式: 2006-01-02)" example("2024-03-12")
// @Param category_id query int false "类别ID，为0表示全部类别" example(1)
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回下载趋势数据，包含start_date、end_date和trend字段"
// @Failure 400 {object} response.Response "请求参数错误，日期格式不正确"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/stats/trend [get]
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
//
// @Summary 获取月度下载统计
// @Description 获取指定年月的下载次数统计，可按类别筛选
// @Tags 管理员/统计
// @Accept json
// @Produce json
// @Param year query int false "年份" minimum(2000) example(2024)
// @Param month query int false "月份" minimum(1) maximum(12) example(3)
// @Param category_id query int false "类别ID，为0表示全部类别" example(1)
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回月度下载统计，包含year、month、count和category_id字段"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/stats/monthly [get]
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
//
// @Summary 获取年度下载统计
// @Description 获取指定年份的下载次数统计，可按类别筛选
// @Tags 管理员/统计
// @Accept json
// @Produce json
// @Param year query int false "年份" minimum(2000) example(2024)
// @Param category_id query int false "类别ID，为0表示全部类别" example(1)
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回年度下载统计，包含year、count和category_id字段"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/stats/yearly [get]
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
//
// @Summary 获取按类别统计的下载量
// @Description 获取指定日期范围内各软件类别的下载量分布
// @Tags 管理员/统计
// @Accept json
// @Produce json
// @Param start_date query string false "开始日期 (格式: 2006-01-02)" example("2024-03-01")
// @Param end_date query string false "结束日期 (格式: 2006-01-02)" example("2024-03-12")
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回类别统计数据，包含start_date、end_date和stats字段"
// @Failure 400 {object} response.Response "请求参数错误，日期格式不正确"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/stats/category [get]
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
//
// @Summary 获取统计概览
// @Description 获取系统下载统计的概览信息，包含今日下载量、本月下载量、本年下载量和按类别的下载分布
// @Tags 管理员/统计
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}} "返回统计概览，包含today、monthly、yearly和by_category字段"
// @Failure 401 {object} response.Response "未授权访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/stats/overview [get]
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
