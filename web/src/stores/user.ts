import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type User, type LoginRequest } from '@/api/auth'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref<string>(localStorage.getItem('token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refresh_token') || '')
  const userInfo = ref<User | null>(null)

  // Getters
  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')

  // Actions
  async function login(data: LoginRequest) {
    const res = await authApi.login(data)
    token.value = res.data.token
    refreshToken.value = res.data.refresh_token
    userInfo.value = {
      id: res.data.user.id,
      username: res.data.user.username,
      nickname: res.data.user.nickname,
      role: res.data.user.role,
    } as User
    
    // 保存到 localStorage
    localStorage.setItem('token', token.value)
    localStorage.setItem('refresh_token', refreshToken.value)
    localStorage.setItem('user_info', JSON.stringify(userInfo.value))
    
    return res
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch (e) {
      // ignore
    }
    
    // 清除状态
    token.value = ''
    refreshToken.value = ''
    userInfo.value = null
    
    // 清除 localStorage
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user_info')
    
    // 跳转到登录页
    router.push('/login')
  }

  async function getProfile() {
    const res = await authApi.getProfile()
    userInfo.value = res.data
    localStorage.setItem('user_info', JSON.stringify(userInfo.value))
    return res
  }

  function initFromStorage() {
    const storedToken = localStorage.getItem('token')
    const storedUser = localStorage.getItem('user_info')
    
    if (storedToken) {
      token.value = storedToken
    }
    
    if (storedUser) {
      try {
        userInfo.value = JSON.parse(storedUser)
      } catch (e) {
        // ignore
      }
    }
  }

  // 初始化
  initFromStorage()

  return {
    token,
    refreshToken,
    userInfo,
    isLoggedIn,
    isAdmin,
    login,
    logout,
    getProfile,
  }
})