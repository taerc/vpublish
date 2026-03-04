import { get, type ApiResponse } from './request'

// 类型定义
export interface CategoryStat {
  category_name: string
  category_code: string
  total_count: number
}

export interface StatsOverview {
  today: number
  monthly: number
  yearly: number
  by_category: CategoryStat[]
}

export interface DailyStats {
  date: string
  count: number
}

export interface MonthlyStats {
  year: number
  month: number
  count: number
  category_id?: number
}

export interface YearlyStats {
  year: number
  count: number
  category_id?: number
}

export interface DailyTrendItem {
  date: string
  count: number
}

// API 方法
export const statsApi = {
  // 统计概览
  overview(): Promise<ApiResponse<StatsOverview>> {
    return get('/admin/stats/overview')
  },

  // 每日统计
  daily(params: { date?: string; category_id?: number }): Promise<ApiResponse<DailyStats>> {
    return get('/admin/stats/daily', { params })
  },

  // 每日下载趋势（多日数据）
  trend(params: { start_date?: string; end_date?: string; category_id?: number }): Promise<ApiResponse<{
    start_date: string
    end_date: string
    trend: DailyTrendItem[]
  }>> {
    return get('/admin/stats/trend', { params })
  },

  // 月度统计
  monthly(params: { year?: number; month?: number; category_id?: number }): Promise<ApiResponse<MonthlyStats>> {
    return get('/admin/stats/monthly', { params })
  },

  // 年度统计
  yearly(params: { year?: number; category_id?: number }): Promise<ApiResponse<YearlyStats>> {
    return get('/admin/stats/yearly', { params })
  },

  // 按类别统计
  byCategory(params: { start_date?: string; end_date?: string }): Promise<ApiResponse<{
    start_date: string
    end_date: string
    stats: CategoryStat[]
  }>> {
    return get('/admin/stats/category', { params })
  },
}