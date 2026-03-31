package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/taerc/vpublish/internal/service"
	"github.com/taerc/vpublish/pkg/response"
)

// ErrorReportHandler 报错记录处理器
type ErrorReportHandler struct {
	reportService *service.ErrorReportService
}

// NewErrorReportHandler 创建报错记录处理器
func NewErrorReportHandler(reportService *service.ErrorReportService) *ErrorReportHandler {
	return &ErrorReportHandler{reportService: reportService}
}

// ErrorReportRequest 报错上报请求参数
type ErrorReportRequest struct {
	// 接口请求唯一标识（必填）
	RequestID string `json:"request_id" binding:"required" example:"req-20260330-001"`
	// 报错发生时间戳（毫秒，必填）
	Timestamp int64 `json:"timestamp" binding:"required" example:"1711737600000"`
	// 报错模块（自动识别，非必填）
	Module string `json:"module" example:"user"`
	// 应用类型：app/platform（必填）
	AppType string `json:"app_type" binding:"required" example:"app"`
	// 接口返回的错误码（必填）
	Code string `json:"code" binding:"required" example:"500"`
	// 接口返回的消息信息（必填）
	ErrorMessage string `json:"error_message" binding:"required" example:"Internal Server Error"`
	// 报错类型：system_error/business_error（非必填，系统自动识别）
	ErrorType string `json:"error_type" example:"system_error"`
	// 接口入参（非必填）
	RequestParams map[string]interface{} `json:"request_params"`
	// 设备信息（必填）
	DeviceInfo map[string]interface{} `json:"device_info" binding:"required"`
}

// BatchErrorReportRequest 批量报错上报请求参数
type BatchErrorReportRequest struct {
	// 报错记录列表（最多100条）
	Records []ErrorReportRequest `json:"records" binding:"required,max=100"`
}

// Report 单条报错上报
//
// @Summary 单条报错上报
// @Description 上报单条接口报错信息到平台，系统会自动识别报错类型（无需JWT认证）
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param request body ErrorReportRequest true "报错上报请求参数"
// @Success 200 {object} response.Response{data=model.ErrorRecord} "上报成功，返回创建的报错记录"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/error/report [post]
func (h *ErrorReportHandler) Report(c *gin.Context) {
	var req ErrorReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 转换为service层请求
	svcReq := &service.ErrorReportRequest{
		RequestID:     req.RequestID,
		Timestamp:     req.Timestamp,
		Module:        req.Module,
		AppType:       req.AppType,
		Code:          req.Code,
		ErrorMessage:  req.ErrorMessage,
		ErrorType:     req.ErrorType,
		RequestParams: req.RequestParams,
		DeviceInfo:    req.DeviceInfo,
	}

	record, err := h.reportService.Report(c.Request.Context(), svcReq)
	if err != nil {
		response.InternalError(c, "failed to report error: "+err.Error())
		return
	}

	response.Success(c, record)
}

// BatchReport 批量报错上报
//
// @Summary 批量报错上报
// @Description 批量上报多条接口报错信息，最多100条（无需JWT认证）
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param request body BatchErrorReportRequest true "批量报错上报请求参数"
// @Success 200 {object} response.Response{data=map[string]interface{}} "上报成功，返回成功数量和失败数量"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /admin/error/report/batch [post]
func (h *ErrorReportHandler) BatchReport(c *gin.Context) {
	var req BatchErrorReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 转换为service层请求
	var svcReqs []*service.ErrorReportRequest
	for _, r := range req.Records {
		svcReqs = append(svcReqs, &service.ErrorReportRequest{
			RequestID:     r.RequestID,
			Timestamp:     r.Timestamp,
			Module:        r.Module,
			AppType:       r.AppType,
			Code:          r.Code,
			ErrorMessage:  r.ErrorMessage,
			ErrorType:     r.ErrorType,
			RequestParams: r.RequestParams,
			DeviceInfo:    r.DeviceInfo,
		})
	}

	successCount, failedCount, recordIDs, err := h.reportService.BatchReport(c.Request.Context(), svcReqs)
	if err != nil {
		response.InternalError(c, "failed to batch report: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"success_count": successCount,
		"failed_count":  failedCount,
		"record_ids":    recordIDs,
	})
}

// ErrorRecordQuery 报错记录查询参数
type ErrorRecordQuery struct {
	// 页码
	Page int `form:"page" example:"1"`
	// 每页数量
	PageSize int `form:"page_size" example:"20"`
	// 开始时间戳（毫秒）
	StartTime int64 `form:"start_time" example:"1711737600000"`
	// 结束时间戳（毫秒）
	EndTime int64 `form:"end_time" example:"1711824000000"`
	// 应用类型：app/platform
	AppType string `form:"app_type" example:"app"`
	// 报错模块
	Module string `form:"module" example:"user"`
	// 报错类型：system_error/business_error
	ErrorType string `form:"error_type" example:"system_error"`
	// 关键词搜索
	Keyword string `form:"keyword" example:"Internal"`
	// 是否为语义化报错：true=是, false=否
	IsSemantic *bool `form:"is_semantic" example:"true"`
}

// List 报错记录列表
//
// @Summary 获取报错记录分页列表
// @Description 分页查询报错记录列表，支持按时间、应用类型、模块、报错类型、是否语义化等条件筛选
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页数量" minimum(1) maximum(100) default(20)
// @Param start_time query int false "开始时间戳（毫秒）"
// @Param end_time query int false "结束时间戳（毫秒）"
// @Param app_type query string false "应用类型：app/platform"
// @Param module query string false "报错模块"
// @Param error_type query string false "报错类型：system_error/business_error"
// @Param keyword query string false "关键词搜索"
// @Param is_semantic query bool false "是否为语义化报错：true=是, false=否"
// @Success 200 {object} response.Response{data=response.PageData{list=[]model.ErrorRecord}} "获取成功，返回报错记录分页列表"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/records [get]
func (h *ErrorReportHandler) List(c *gin.Context) {
	var query ErrorRecordQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query: "+err.Error())
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// 转换为service层查询参数
	svcQuery := &service.ErrorRecordQuery{
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

	records, total, err := h.reportService.List(c.Request.Context(), svcQuery)
	if err != nil {
		response.InternalError(c, "failed to get error records")
		return
	}

	response.Page(c, records, total, query.Page, query.PageSize)
}

// Get 获取报错记录详情
//
// @Summary 获取报错记录详情
// @Description 根据ID获取报错记录的详细信息
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param id path int true "报错记录ID" minimum(1) example(1)
// @Success 200 {object} response.Response{data=model.ErrorRecord} "返回报错记录详情"
// @Failure 400 {object} response.Response "请求参数错误，无效的ID"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "报错记录不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/records/{id} [get]
func (h *ErrorReportHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	record, err := h.reportService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "record not found")
		return
	}

	response.Success(c, record)
}

// UpdateRemarkRequest 更新备注请求参数
type UpdateRemarkRequest struct {
	// 备注信息
	Remark string `json:"remark" binding:"required" example:"已处理，问题已解决"`
}

// UpdateRemark 更新备注
//
// @Summary 更新报错记录备注
// @Description 更新指定报错记录的备注信息
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param id path int true "报错记录ID" minimum(1) example(1)
// @Param request body UpdateRemarkRequest true "更新备注请求参数"
// @Success 200 {object} response.Response "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 404 {object} response.Response "报错记录不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/records/{id}/remark [put]
func (h *ErrorReportHandler) UpdateRemark(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var req UpdateRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.reportService.UpdateRemark(c.Request.Context(), uint(id), req.Remark); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetModules 获取模块列表
//
// @Summary 获取所有报错模块列表
// @Description 获取系统中所有的报错模块列表，用于筛选条件
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]string} "返回模块列表"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/modules [get]
func (h *ErrorReportHandler) GetModules(c *gin.Context) {
	modules, err := h.reportService.GetModules(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get modules")
		return
	}

	response.Success(c, modules)
}

// TrendQuery 趋势统计查询参数
type TrendQuery struct {
	// 开始日期，格式：YYYY-MM-DD
	StartDate string `form:"start_date" binding:"required" example:"2026-03-01"`
	// 结束日期，格式：YYYY-MM-DD
	EndDate string `form:"end_date" binding:"required" example:"2026-03-30"`
	// 应用类型：app/platform
	AppType string `form:"app_type" example:"app"`
}

// GetTrend 获取趋势统计
//
// @Summary 获取报错趋势统计
// @Description 查询指定时间范围内的报错趋势统计
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string true "结束日期，格式：YYYY-MM-DD"
// @Param app_type query string false "应用类型：app/platform"
// @Success 200 {object} response.Response{data=service.TrendStatistics} "返回趋势统计数据"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/statistics/trend [get]
func (h *ErrorReportHandler) GetTrend(c *gin.Context) {
	var query TrendQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query: "+err.Error())
		return
	}

	trend, err := h.reportService.GetTrend(c.Request.Context(), query.StartDate, query.EndDate, query.AppType)
	if err != nil {
		response.InternalError(c, "failed to get trend")
		return
	}

	response.Success(c, trend)
}

// GetModuleStats 获取模块统计
//
// @Summary 获取报错模块统计
// @Description 查询指定时间范围内各模块的报错统计
// @Tags 管理员/报错管理
// @Accept json
// @Produce json
// @Param start_date query string true "开始日期，格式：YYYY-MM-DD"
// @Param end_date query string true "结束日期，格式：YYYY-MM-DD"
// @Param app_type query string false "应用类型：app/platform"
// @Success 200 {object} response.Response{data=[]service.ModuleStat} "返回模块统计数据"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/statistics/module [get]
func (h *ErrorReportHandler) GetModuleStats(c *gin.Context) {
	var query TrendQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query: "+err.Error())
		return
	}

	stats, err := h.reportService.GetModuleStats(c.Request.Context(), query.StartDate, query.EndDate, query.AppType)
	if err != nil {
		response.InternalError(c, "failed to get module stats")
		return
	}

	response.Success(c, stats)
}

// ExportExcel 导出报错记录到 Excel
//
// @Summary 导出报错记录到 Excel
// @Description 根据筛选条件将报错记录导出为 Excel 文件
// @Tags 管理员/报错管理
// @Accept json
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param start_time query int false "开始时间戳（毫秒）"
// @Param end_time query int false "结束时间戳（毫秒）"
// @Param app_type query string false "应用类型：app/platform"
// @Param module query string false "报错模块"
// @Param error_type query string false "报错类型：system_error/business_error"
// @Param keyword query string false "关键词搜索"
// @Param is_semantic query bool false "是否为语义化报错：true=是, false=否"
// @Success 200 {file} file "Excel 文件"
// @Failure 401 {object} response.Response "未认证访问"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Security BearerAuth []
// @Router /admin/error/export [get]
func (h *ErrorReportHandler) ExportExcel(c *gin.Context) {
	var query ErrorRecordQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "invalid query: "+err.Error())
		return
	}

	// 转换为service层查询参数
	svcQuery := &service.ErrorRecordQuery{
		StartTime:  query.StartTime,
		EndTime:    query.EndTime,
		AppType:    query.AppType,
		Module:     query.Module,
		ErrorType:  query.ErrorType,
		Keyword:    query.Keyword,
		IsSemantic: query.IsSemantic,
	}

	// 生成 Excel 文件
	file, err := h.reportService.ExportToExcel(c.Request.Context(), svcQuery)
	if err != nil {
		response.InternalError(c, "failed to export excel: "+err.Error())
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=error_records.xlsx")

	// 写入文件
	if _, err := file.WriteTo(c.Writer); err != nil {
		response.InternalError(c, "failed to write excel: "+err.Error())
		return
	}
}