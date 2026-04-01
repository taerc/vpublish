package service

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/taerc/vpublish/internal/model"
	"github.com/taerc/vpublish/internal/repository"
	"github.com/xuri/excelize/v2"
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
	AppType       string                 `json:"app_type"`
	Code          string                 `json:"code" binding:"required"`
	ErrorMessage  string                 `json:"error_message" binding:"required"`
	ErrorType     string                 `json:"error_type"`
	RequestParams map[string]interface{} `json:"request_params"`
	DeviceInfo    map[string]interface{} `json:"device_info" binding:"required"`
}

// Report 上报单条报错记录
func (s *ErrorReportService) Report(ctx context.Context, req *ErrorReportRequest) (*model.ErrorRecord, error) {
	// 检查 request_id 是否已存在
	exists, err := s.recordRepo.ExistsByRequestID(ctx, req.RequestID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("request_id already exists")
	}

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
	var failedCount int
	var recordIDs []uint

	// 收集所有需要检查的 request_id
	requestIDs := make([]string, 0, len(reqs))
	for _, req := range reqs {
		requestIDs = append(requestIDs, req.RequestID)
	}

	// 批量检查哪些 request_id 已存在
	existingIDs, err := s.recordRepo.GetExistingRequestIDs(ctx, requestIDs)
	if err != nil {
		return 0, len(reqs), nil, err
	}

	// 构建已存在的 request_id 的集合，用于快速查找
	existingSet := make(map[string]bool)
	for _, id := range existingIDs {
		existingSet[id] = true
	}

	for _, req := range reqs {
		// 检查 request_id 是否已存在
		if existingSet[req.RequestID] {
			failedCount++
			continue
		}

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

	// 批量创建新记录
	if len(records) > 0 {
		if err := s.recordRepo.CreateBatch(ctx, records); err != nil {
			return 0, len(reqs), nil, err
		}

		for _, record := range records {
			successCount++
			recordIDs = append(recordIDs, record.ID)
		}
	}

	return successCount, failedCount, recordIDs, nil
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
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	StartTime  int64  `form:"start_time"`
	EndTime    int64  `form:"end_time"`
	AppType    string `form:"app_type"`
	Module     string `form:"module"`
	ErrorType  string `form:"error_type"`
	Keyword    string `form:"keyword"`
	IsSemantic *bool  `form:"is_semantic"` // 是否为语义化报错：true=是, false=否, nil=不限
}

// List 分页查询报错记录
func (s *ErrorReportService) List(ctx context.Context, query *ErrorRecordQuery) ([]model.ErrorRecord, int64, error) {
	repoQuery := &repository.ErrorRecordListQuery{
		Page:       query.Page,
		PageSize:   query.PageSize,
		StartTime:  query.StartTime,
		EndTime:    query.EndTime,
		AppType:    query.AppType,
		Module:     query.Module,
		ErrorType:  query.ErrorType,
		Keyword:    query.Keyword,
		IsSemantic: query.IsSemantic,
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

// IsSemanticError 判断报错信息是否为语义化报错
// 如果报错信息包含中文字符，认为是语义化报错；如果是纯英文和符号，则不是语义化
func (s *ErrorReportService) IsSemanticError(message string) bool {
	// 检查是否包含中文字符
	chinese := regexp.MustCompile(`[\p{Han}]`)
	return chinese.MatchString(message)
}

// ExportToExcel 导出报错记录到 Excel
func (s *ErrorReportService) ExportToExcel(ctx context.Context, query *ErrorRecordQuery) (*excelize.File, error) {
	// 创建 Excel 文件
	f := excelize.NewFile()
	sheetName := "报错记录"
	
	// 创建工作表
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	
	// 设置表头
	headers := []string{
		"序号", "Request-ID", "报错时间", "报错应用", "报错模块",
		"报错类型", "错误码", "报错信息", "是否语义化", "备注",
	}
	
	// 写入表头（第一行）
	for i, header := range headers {
		cell := string(rune('A'+i)) + "1"
		f.SetCellValue(sheetName, cell, header)
		
		// 设置表头样式
		style, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
				Size: 12,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"#E6F7FF"},
				Pattern: 1,
			},
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})
		if err != nil {
			return nil, err
		}
		f.SetCellStyle(sheetName, cell, cell, style)
	}
	
	// 查询数据（不分页，获取所有符合条件的记录）
	exportQuery := &repository.ErrorRecordListQuery{
		Page:       1,
		PageSize:   10000, // 导出最多10000条
		StartTime:  query.StartTime,
		EndTime:    query.EndTime,
		AppType:    query.AppType,
		Module:     query.Module,
		ErrorType:  query.ErrorType,
		Keyword:    query.Keyword,
		IsSemantic: query.IsSemantic,
	}
	records, _, err := s.recordRepo.List(ctx, exportQuery)
	if err != nil {
		return nil, err
	}
	
	// 写入数据
	for i, record := range records {
		row := i + 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), i+1)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), record.RequestID)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), formatTimestamp(record.Timestamp))
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), getAppTypeLabel(record.AppType))
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), record.Module)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), getErrorTypeLabel(record.ErrorType))
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), record.Code)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), record.ErrorMessage)
		f.SetCellValue(sheetName, "I"+strconv.Itoa(row), getSemanticLabel(s.IsSemanticError(record.ErrorMessage)))
		f.SetCellValue(sheetName, "J"+strconv.Itoa(row), record.Remark)
	}
	
	// 设置列宽
	setColumnWidths(f, sheetName)
	
	return f, nil
}

// formatTimestamp 格式化时间戳
func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp/1000, 0)
	return t.Format("2006-01-02 15:04:05")
}

// getAppTypeLabel 获取应用类型标签
func getAppTypeLabel(appType string) string {
	switch appType {
	case "app":
		return "App"
	case "platform":
		return "平台"
	default:
		return appType
	}
}

// getErrorTypeLabel 获取报错类型标签
func getErrorTypeLabel(errorType string) string {
	switch errorType {
	case "system_error":
		return "系统错误"
	case "business_error":
		return "业务错误"
	default:
		return errorType
	}
}

// getSemanticLabel 获取语义化标签
func getSemanticLabel(isSemantic bool) string {
	if isSemantic {
		return "是"
	}
	return "否"
}

// setColumnWidths 设置列宽
func setColumnWidths(f *excelize.File, sheetName string) {
	widths := map[string]float64{
		"A": 6,   // 序号
		"B": 25,  // Request-ID
		"C": 20,  // 报错时间
		"D": 12,  // 报错应用
		"E": 15,  // 报错模块
		"F": 12,  // 报错类型
		"G": 12,  // 错误码
		"H": 50,  // 报错信息
		"I": 12,  // 是否语义化
		"J": 30,  // 备注
	}
	
	for col, width := range widths {
		f.SetColWidth(sheetName, col, col, width)
	}
}