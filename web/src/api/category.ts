import { get, post, put, del, type ApiResponse, type PageResponse } from './request'

// 类型定义
export interface Category {
  id: number
  name: string
  code: string
  description: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateCategoryRequest {
  name: string
  description?: string
  sort_order?: number
}

export interface UpdateCategoryRequest {
  name?: string
  description?: string
  sort_order?: number
  is_active?: boolean
}

// API 方法
export const categoryApi = {
  // 类别列表
  list(params: { page?: number; page_size?: number }): Promise<ApiResponse<PageResponse<Category>>> {
    return get('/admin/categories', { params })
  },

  // 启用的类别列表
  listActive(): Promise<ApiResponse<Category[]>> {
    return get('/admin/categories/active')
  },
  // 获取单个类别
  get(id: number): Promise<ApiResponse<Category>> {
    return get(`/admin/categories/${id}`)
  },

  // 创建类别
  create(data: CreateCategoryRequest): Promise<ApiResponse<Category>> {
    return post('/admin/categories', data)
  },

  // 更新类别
  update(id: number, data: UpdateCategoryRequest): Promise<ApiResponse<Category>> {
    return put(`/admin/categories/${id}`, data)
  },

  // 删除类别
  delete(id: number): Promise<ApiResponse<null>> {
    return del(`/admin/categories/${id}`)
  },
}