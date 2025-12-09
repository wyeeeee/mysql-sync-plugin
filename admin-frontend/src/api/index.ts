import axios from 'axios'
import { useAuthStore } from '../stores/auth'
import router from '../router'

const api = axios.create({
  baseURL: '/data/admin/api',
  timeout: 30000
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    // 禁用浏览器缓存
    config.headers['Cache-Control'] = 'no-cache, no-store, must-revalidate'
    config.headers['Pragma'] = 'no-cache'
    config.headers['Expires'] = '0'
    // GET 请求添加时间戳参数防止缓存
    if (config.method?.toLowerCase() === 'get') {
      config.params = { ...config.params, _t: Date.now() }
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    const data = response.data
    if (data.code === 10004) {
      // 认证失败，跳转登录
      const authStore = useAuthStore()
      authStore.logout()
      router.push('/login')
    }
    return data
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 认证相关
export const authApi = {
  login: (username: string, password: string) =>
    api.post('/login', { username, password }),
  logout: () => api.post('/logout'),
  getCurrentUser: () => api.get('/user/current'),
  changePassword: (oldPassword: string, newPassword: string) =>
    api.post('/user/password', { oldPassword, newPassword })
}

// 日志相关
export const logApi = {
  getLogs: (params: Record<string, any>) => api.get('/logs', { params }),
  getStats: () => api.get('/logs/stats'),
  cleanLogs: (days: number) => api.post('/logs/clean', { days })
}

// 系统相关
export const systemApi = {
  getInfo: () => api.get('/system/info')
}

export default api
