import { get, post, put, type ApiResponse, type PageResponse } from './request'
import service from './request'

// ==================== 类型定义 ====================

// 设备信息
export interface DeviceInfo {
  app_version: string    // 应用版本号
  device_model: string   // 设备型号
  os_version: string     // 操作系统版本
  drone_model?: string   // 无人机型号
  device_id?: string     // 设备标识
}

// 报错记录
export interface ErrorRecord {
  id: number
  request_id: string
  timestamp: number
  module: string
  app_type: string       // app/platform
  path: string           // 接口路径
  code: string           // 错误码
  error_message: string  // 报错信息
  error_type: string     // system_error/business_error
  remark?: string        // 备注
  created_at: string
  updated_at: string
}

// 报错记录详情
export interface ErrorRecordDetail extends ErrorRecord {
  request_params?: Record<string, any>
  device_info?: DeviceInfo
}

// 报错上报请求
export interface ErrorReportRequest {
  request_id: string
  timestamp: number
  module?: string
  code: string
  error_message: string
  error_type?: string
  request_params?: Record<string, any>
  device_info: DeviceInfo
  path: string          // 接口路径
}

// 批量上报请求
export interface BatchErrorReportRequest {
  records: ErrorReportRequest[]
}

// 报错查询参数
export interface ErrorQueryParams {
  page?: number
  size?: number
  start_time?: number
  end_time?: number
  app_type?: string
  module?: string
  error_type?: string
  keyword?: string
  is_semantic?: boolean  // 是否为语义化报错
}

// 统计趋势项
export interface TrendItem {
  date: string
  count: number
}

// 模块统计项
export interface ModuleStatItem {
  module: string
  error_type: string
  count: number
}

// ==================== API 方法 ====================

export const errorApi = {
  // ========== 报错上报 ==========

  // 单条上报
  report(data: ErrorReportRequest): Promise<ApiResponse<{ record_id: number }>> {
    return post('/admin/error/report', data)
  },

  // 批量上报
  reportBatch(data: BatchErrorReportRequest): Promise<ApiResponse<{
    success_count: number
    failed_count: number
    record_ids: number[]
  }>> {
    return post('/admin/error/report/batch', data)
  },

  // ========== 报错查询 ==========

  // 分页查询
  list(params?: ErrorQueryParams): Promise<ApiResponse<PageResponse<ErrorRecord>>> {
    return get('/admin/error/records', { params })
  },

  // 查询详情
  getDetail(id: number): Promise<ApiResponse<ErrorRecordDetail>> {
    return get(`/admin/error/records/${id}`)
  },

  // 更新备注
  updateRemark(id: number, remark: string): Promise<ApiResponse<null>> {
    return put(`/admin/error/records/${id}/remark`, { remark })
  },

  // 获取模块列表
  getModules(): Promise<ApiResponse<string[]>> {
    return get('/admin/error/modules')
  },

  // 导出报错记录到 Excel
  exportExcel(params?: ErrorQueryParams): Promise<Blob> {
    return service.get('/admin/error/export', {
      params,
      responseType: 'blob',
    }).then((response) => response.data)
  },

  // ========== 统计 ==========

  // 报错趋势统计
  trend(params: { start_date: string; end_date: string; app_type?: string }): Promise<ApiResponse<{
    total_count: number
    trend: TrendItem[]
  }>> {
    return get('/admin/error/statistics/trend', { params })
  },

  // 模块统计
  moduleStats(params: { start_date: string; end_date: string; app_type?: string }): Promise<ApiResponse<ModuleStatItem[]>> {
    return get('/admin/error/statistics/module', { params })
  },
}
