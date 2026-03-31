package repository

import (
	"context"

	"github.com/taerc/vpublish/internal/model"
	"gorm.io/gorm"
)

// ErrorRecordRepository 报错记录仓库
type ErrorRecordRepository struct {
	db *gorm.DB
}

// NewErrorRecordRepository 创建报错记录仓库
func NewErrorRecordRepository(db *gorm.DB) *ErrorRecordRepository {
	return &ErrorRecordRepository{db: db}
}

// Create 创建报错记录
func (r *ErrorRecordRepository) Create(ctx context.Context, record *model.ErrorRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// CreateBatch 批量创建报错记录
func (r *ErrorRecordRepository) CreateBatch(ctx context.Context, records []*model.ErrorRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

// Update 更新报错记录
func (r *ErrorRecordRepository) Update(ctx context.Context, record *model.ErrorRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// GetByID 根据ID获取报错记录
func (r *ErrorRecordRepository) GetByID(ctx context.Context, id uint) (*model.ErrorRecord, error) {
	var record model.ErrorRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetByRequestID 根据RequestID获取报错记录
func (r *ErrorRecordRepository) GetByRequestID(ctx context.Context, requestID string) (*model.ErrorRecord, error) {
	var record model.ErrorRecord
	err := r.db.WithContext(ctx).Where("request_id = ?", requestID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListQuery 查询参数
type ErrorRecordListQuery struct {
	Page        int
	PageSize    int
	StartTime   int64
	EndTime     int64
	AppType     string
	Module      string
	ErrorType   string
	Keyword     string
	IsSemantic  *bool // 是否为语义化报错：true=是, false=否, nil=不限
}

// List 分页查询报错记录
func (r *ErrorRecordRepository) List(ctx context.Context, query *ErrorRecordListQuery) ([]model.ErrorRecord, int64, error) {
	var records []model.ErrorRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&model.ErrorRecord{})

	// 时间范围筛选
	if query.StartTime > 0 {
		db = db.Where("timestamp >= ?", query.StartTime)
	}
	if query.EndTime > 0 {
		db = db.Where("timestamp <= ?", query.EndTime)
	}

	// 应用类型筛选
	if query.AppType != "" {
		db = db.Where("app_type = ?", query.AppType)
	}

	// 模块筛选
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}

	// 报错类型筛选
	if query.ErrorType != "" {
		db = db.Where("error_type = ?", query.ErrorType)
	}

	// 关键词搜索
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("error_message LIKE ? OR request_id LIKE ?", keyword, keyword)
	}

	// 语义化报错筛选
	if query.IsSemantic != nil {
		if *query.IsSemantic {
			// 是语义化报错：包含中文字符
			db = db.Where("error_message REGEXP ?", "[\\u4e00-\\u9fa5]")
		} else {
			// 不是语义化报错：不包含中文字符
			db = db.Where("error_message NOT REGEXP ?", "[\\u4e00-\\u9fa5]")
		}
	}

	// 统计总数
	db.Count(&total)

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&records).Error
	return records, total, err
}

// GetModules 获取所有模块列表
func (r *ErrorRecordRepository) GetModules(ctx context.Context) ([]string, error) {
	var modules []string
	err := r.db.WithContext(ctx).Model(&model.ErrorRecord{}).
		Distinct("module").
		Where("module != ''").
		Pluck("module", &modules).Error
	return modules, err
}

// UpdateRemark 更新备注
func (r *ErrorRecordRepository) UpdateRemark(ctx context.Context, id uint, remark string) error {
	return r.db.WithContext(ctx).Model(&model.ErrorRecord{}).
		Where("id = ?", id).
		Update("remark", remark).Error
}

// CountByDate 统计指定日期的报错数量
func (r *ErrorRecordRepository) CountByDate(ctx context.Context, startTime, endTime int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ErrorRecord{}).
		Where("timestamp >= ? AND timestamp < ?", startTime, endTime).
		Count(&count).Error
	return count, err
}

// TrendItem 趋势项
type TrendItem struct {
	Date  string
	Count int
}

// GetTrend 获取趋势统计（直接从error_records表查询）
func (r *ErrorRecordRepository) GetTrend(ctx context.Context, startDate, endDate, appType string) ([]TrendItem, error) {
	var results []struct {
		Date  string
		Count int
	}

	query := r.db.WithContext(ctx).Model(&model.ErrorRecord{}).
		Select("DATE(FROM_UNIXTIME(timestamp/1000)) as date, COUNT(*) as count").
		Where("DATE(FROM_UNIXTIME(timestamp/1000)) >= ? AND DATE(FROM_UNIXTIME(timestamp/1000)) <= ?", startDate, endDate)

	if appType != "" {
		query = query.Where("app_type = ?", appType)
	}

	err := query.Group("DATE(FROM_UNIXTIME(timestamp/1000))").
		Order("date ASC").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	items := make([]TrendItem, len(results))
	for i, r := range results {
		items[i] = TrendItem{Date: r.Date, Count: r.Count}
	}
	return items, nil
}

// ModuleStatItem 模块统计项
type ModuleStatItem struct {
	Module    string
	ErrorType string
	Count     int
}

// GetModuleStats 获取模块统计（直接从error_records表查询）
func (r *ErrorRecordRepository) GetModuleStats(ctx context.Context, startDate, endDate, appType string) ([]ModuleStatItem, error) {
	var results []struct {
		Module    string
		ErrorType string
		Count     int
	}

	query := r.db.WithContext(ctx).Model(&model.ErrorRecord{}).
		Select("module, error_type, COUNT(*) as count").
		Where("DATE(FROM_UNIXTIME(timestamp/1000)) >= ? AND DATE(FROM_UNIXTIME(timestamp/1000)) <= ?", startDate, endDate)

	if appType != "" {
		query = query.Where("app_type = ?", appType)
	}

	err := query.Group("module, error_type").
		Order("count DESC").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	items := make([]ModuleStatItem, len(results))
	for i, r := range results {
		items[i] = ModuleStatItem{Module: r.Module, ErrorType: r.ErrorType, Count: r.Count}
	}
	return items, nil
}