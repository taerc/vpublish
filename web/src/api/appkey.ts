import { get, post, put, del, type ApiResponse, type PageResponse } from './request'

// AppKey 数据类型
export interface AppKey {
  id: number
  app_name: string
  app_key: string
  description: string
  is_active: boolean
  created_at: string
  updated_at: string
}

// 创建 AppKey 响应（包含 Secret）
export interface AppKeyWithSecret extends AppKey {
  app_secret: string
}

// 创建 AppKey 请求
export interface CreateAppKeyRequest {
  app_name: string
  description?: string
}

// 更新 AppKey 请求
export interface UpdateAppKeyRequest {
  app_name?: string
  description?: string
  is_active?: boolean
}

// AppKey API
export const appKeyApi = {
  // 获取 AppKey 列表
  list(params: { page?: number; page_size?: number }): Promise<ApiResponse<PageResponse<AppKey>>> {
    return get('/admin/appkeys', { params })
  },

  // 获取单个 AppKey
  get(id: number): Promise<ApiResponse<AppKey>> {
    return get(`/admin/appkeys/${id}`)
  },

  // 创建 AppKey
  create(data: CreateAppKeyRequest): Promise<ApiResponse<AppKeyWithSecret>> {
    return post('/admin/appkeys', data)
  },

  // 更新 AppKey
  update(id: number, data: UpdateAppKeyRequest): Promise<ApiResponse<AppKey>> {
    return put(`/admin/appkeys/${id}`, data)
  },

  // 删除 AppKey
  delete(id: number): Promise<ApiResponse<null>> {
    return del(`/admin/appkeys/${id}`)
  },

  // 重新生成 Secret
  regenerateSecret(id: number): Promise<ApiResponse<AppKeyWithSecret>> {
    return post(`/admin/appkeys/${id}/regenerate`)
  },
}