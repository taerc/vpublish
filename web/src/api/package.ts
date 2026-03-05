import { get, post, put, del, downloadFile, type ApiResponse, type PageResponse } from './request'
import type { Category } from './category'

// 类型定义
// 类型定义
export interface Package {
  id: number
  category_id: number
  category?: Category
  name: string
  description: string
  icon: string
  developer: string
  website: string
  is_active: boolean
  created_by: number
  creator?: { id: number; username: string; nickname: string }
  created_at: string
  updated_at: string
  latest_version?: Version // 最新版本信息
}

export interface Version {
  id: number
  package_id: number
  version: string
  version_code: number
  file_name: string
  file_size: number
  file_hash: string
  changelog: string
  release_notes: string
  min_version: string
  force_upgrade: boolean
  is_latest: boolean
  is_stable: boolean
  download_count: number
  published_at: string
  created_at: string
}

export interface CreatePackageRequest {
  category_id: number
  name: string
  description?: string
  icon?: string
  developer?: string
  website?: string
}

export interface UpdatePackageRequest {
  name?: string
  description?: string
  icon?: string
  developer?: string
  website?: string
  is_active?: boolean
}

// API 方法
export const packageApi = {
  // 软件包列表
  list(params: { category_id?: number; page?: number; page_size?: number }): Promise<ApiResponse<PageResponse<Package>>> {
    return get('/admin/packages', { params })
  },

  // 获取单个软件包
  get(id: number): Promise<ApiResponse<Package>> {
    return get(`/admin/packages/${id}`)
  },

  // 创建软件包（带文件上传）
  create(formData: FormData): Promise<ApiResponse<{ pkg: Package; version: Version }>> {
    return post('/admin/packages', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // 更新软件包
  update(id: number, data: UpdatePackageRequest): Promise<ApiResponse<Package>> {
    return put(`/admin/packages/${id}`, data)
  },

  // 删除软件包
  delete(id: number): Promise<ApiResponse<null>> {
    return del(`/admin/packages/${id}`)
  },

  // 版本列表
  listVersions(packageId: number, params?: { page?: number; page_size?: number }): Promise<ApiResponse<PageResponse<Version>>> {
    return get(`/admin/packages/${packageId}/versions`, { params })
  },

  // 上传版本
  uploadVersion(packageId: number, formData: FormData): Promise<ApiResponse<Version>> {
    return post(`/admin/packages/${packageId}/versions`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  // 删除版本
  deleteVersion(versionId: number): Promise<ApiResponse<null>> {
    return del(`/admin/versions/${versionId}`)
  },

  // 下载版本
  async downloadVersion(versionId: number, filename?: string): Promise<void> {
    await downloadFile(`/admin/versions/${versionId}/download`, filename)
  },
}