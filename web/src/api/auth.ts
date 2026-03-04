import { get, post, put, del, type ApiResponse, type PageResponse } from './request'

// 类型定义
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  refresh_token: string
  expires_in: number
  user: {
    id: number
    username: string
    nickname: string
    role: string
  }
}

export interface User {
  id: number
  username: string
  nickname: string
  email: string
  role: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateUserRequest {
  username: string
  password: string
  nickname?: string
  email?: string
  role?: string
}

export interface UpdateUserRequest {
  nickname?: string
  email?: string
  role?: string
  is_active?: boolean
}

// API 方法
export const authApi = {
  // 登录
  login(data: LoginRequest): Promise<ApiResponse<LoginResponse>> {
    return post('/admin/auth/login', data)
  },

  // 刷新令牌
  refreshToken(refreshToken: string): Promise<ApiResponse<{ token: string; expires_in: number }>> {
    return post('/admin/auth/refresh', {}, {
      headers: { 'X-Refresh-Token': refreshToken }
    })
  },

  // 登出
  logout(): Promise<ApiResponse<null>> {
    return post('/admin/auth/logout')
  },

  // 获取当前用户信息
  getProfile(): Promise<ApiResponse<User>> {
    return get('/admin/auth/profile')
  },

  // 修改密码
  changePassword(data: { old_password: string; new_password: string }): Promise<ApiResponse<null>> {
    return put('/admin/auth/password', data)
  },
}

export const userApi = {
  // 用户列表
  list(params: { page?: number; page_size?: number }): Promise<ApiResponse<PageResponse<User>>> {
    return get('/admin/users', { params })
  },

  // 获取单个用户
  get(id: number): Promise<ApiResponse<User>> {
    return get(`/admin/users/${id}`)
  },

  // 创建用户
  create(data: CreateUserRequest): Promise<ApiResponse<User>> {
    return post('/admin/users', data)
  },

  // 更新用户
  update(id: number, data: UpdateUserRequest): Promise<ApiResponse<User>> {
    return put(`/admin/users/${id}`, data)
  },

  // 删除用户
  delete(id: number): Promise<ApiResponse<null>> {
    return del(`/admin/users/${id}`)
  },

  // 重置密码
  resetPassword(id: number, password: string): Promise<ApiResponse<null>> {
    return put(`/admin/users/${id}/password`, { password })
  },
}