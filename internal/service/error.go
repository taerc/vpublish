package service

import (
	"context"
	"errors"
	"strings"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

// ErrorReportService 报错记录服务
type ErrorReportService struct {
	recordRepo *repository.ErrorRecordRepository
}

// NewErrorReportService 创建报错记录服务
func NewErrorReportService(recordRepo *repository.ErrorRecordRepository) *ErrorReportService {
	return &ErrorReportService{
		recordRepo: recordRepo,
	}
}

// ErrorReportRequest 报错上报请求
type ErrorReportRequest struct {
	RequestID     string                 `json:"request_id" binding:"required"`
	Timestamp     int64                  `json:"timestamp" binding:"required"`
	Module        string                 `json:"module"`
	AppType       string                 `json:"app_type" binding:"required"`
	Code          string                 `json:"code" binding:"required"`
	ErrorMessage  string                 `json:"error_message" binding:"required"`
	ErrorType     string                 `json:"error_type"`
	RequestParams map[string]interface{} `json:"request_params"`
	DeviceInfo    map[string]interface{} `json:"device_info" binding:"required"`
}

// Report 上报单条报错记录
func (s *ErrorReportService) Report(ctx context.Context, req *ErrorReportRequest) (*model.ErrorRecord, error) {
	record := &model.ErrorRecord{
		RequestID:     req.RequestID,
		Timestamp:     req.Timestamp,
		Module:        req.Module,
		AppType:       req.AppType,
		Code:          req.Code,
		ErrorMessage:  req.ErrorMessage,
		RequestParams: req.RequestParams,
		DeviceInfo:    req.DeviceInfo,
	}

	// 智能识别报错类型
	if req.ErrorType != "" {
		record.ErrorType = req.ErrorType
	} else {
		record.ErrorType = s.detectErrorType(req.Code, req.ErrorMessage)
	}

	if err := s.recordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

// BatchReport 批量上报
func (s *ErrorReportService) BatchReport(ctx context.Context, reqs []*ErrorReportRequest) (int, int, []uint, error) {
	var records []*model.ErrorRecord
	var successCount int
	var recordIDs []uint

	for _, req := range reqs {
		record := &model.ErrorRecord{
			RequestID:     req.RequestID,
			Timestamp:     req.Timestamp,
			Module:        req.Module,
			AppType:       req.AppType,
			Code:          req.Code,
			ErrorMessage:  req.ErrorMessage,
			RequestParams: req.RequestParams,
			DeviceInfo:    req.DeviceInfo,
		}

		// 智能识别报错类型
		if req.ErrorType != "" {
			record.ErrorType = req.ErrorType
		} else {
			record.ErrorType = s.detectErrorType(req.Code, req.ErrorMessage)
		}

		records = append(records, record)
	}

	if err := s.recordRepo.CreateBatch(ctx, records); err != nil {
		return 0, len(reqs), nil, err
	}

	for _, record := range records {
		successCount++
		recordIDs = append(recordIDs, record.ID)
	}

	return successCount, 0, recordIDs, nil
}

// detectErrorType 智能识别报错类型
func (s *ErrorReportService) detectErrorType(code, message string) string {
	// 基于错误码判断
	if strings.HasPrefix(code, "5") {
		return "system_error"
	}

	// 基于消息判断
	lowerMsg := strings.ToLower(message)
	systemKeywords := []string{
		"error", "exception", "timeout", "failed", "connection",
		"internal", "server", "database", "network", "panic",
	}
	for _, keyword := range systemKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return "system_error"
		}
	}

	// 默认为业务错误
	return "business_error"
}

// ErrorRecordQuery 查询参数
type ErrorRecordQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	StartTime int64  `form:"start_time"`
	EndTime   int64  `form:"end_time"`
	AppType   string `form:"app_type"`
	Module    string `form:"module"`
	ErrorType string `form:"error_type"`
	Keyword   string `form:"keyword"`
}

// List 分页查询报错记录
func (s *ErrorReportService) List(ctx context.Context, query *ErrorRecordQuery) ([]model.ErrorRecord, int64, error) {
	repoQuery := &repository.ErrorRecordListQuery{
		Page:      query.Page,
		PageSize:  query.PageSize,
		StartTime: query.StartTime,
		EndTime:   query.EndTime,
		AppType:   query.AppType,
		Module:    query.Module,
		ErrorType: query.ErrorType,
		Keyword:   query.Keyword,
	}
	return s.recordRepo.List(ctx, repoQuery)
}

// GetByID 根据ID获取报错记录
func (s *ErrorReportService) GetByID(ctx context.Context, id uint) (*model.ErrorRecord, error) {
	return s.recordRepo.GetByID(ctx, id)
}

// UpdateRemark 更新备注
func (s *ErrorReportService) UpdateRemark(ctx context.Context, id uint, remark string) error {
	return s.recordRepo.UpdateRemark(ctx, id, remark)
}

// GetModules 获取所有模块列表
func (s *ErrorReportService) GetModules(ctx context.Context) ([]string, error) {
	return s.recordRepo.GetModules(ctx)
}

// TrendData 趋势数据
type TrendData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// TrendStatistics 趋势统计
type TrendStatistics struct {
	TotalCount int          `json:"total_count"`
	Trend      []TrendData `json:"trend"`
}

// GetTrend 获取趋势统计
func (s *ErrorReportService) GetTrend(ctx context.Context, startDate, endDate, appType string) (*TrendStatistics, error) {
	items, err := s.recordRepo.GetTrend(ctx, startDate, endDate, appType)
	if err != nil {
		return nil, err
	}

	result := &TrendStatistics{
		Trend: make([]TrendData, 0, len(items)),
	}

	for _, item := range items {
		result.TotalCount += item.Count
		result.Trend = append(result.Trend, TrendData{
			Date:  item.Date,
			Count: item.Count,
		})
	}

	return result, nil
}

// ModuleStat 模块统计
type ModuleStat struct {
	Module    string `json:"module"`
	ErrorType string `json:"error_type"`
	Count     int    `json:"count"`
}

// GetModuleStats 获取模块统计
func (s *ErrorReportService) GetModuleStats(ctx context.Context, startDate, endDate, appType string) ([]ModuleStat, error) {
	items, err := s.recordRepo.GetModuleStats(ctx, startDate, endDate, appType)
	if err != nil {
		return nil, err
	}

	result := make([]ModuleStat, 0, len(items))
	for _, item := range items {
		result = append(result, ModuleStat{
			Module:    item.Module,
			ErrorType: item.ErrorType,
			Count:     item.Count,
		})
	}

	return result, nil
}