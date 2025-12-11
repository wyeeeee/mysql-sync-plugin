import axios from 'axios'
import { useAuthStore } from '../stores/auth'
import router from '../router'

const api = axios.create({
  baseURL: '/admin/api',
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

// 用户管理相关
export const userApi = {
  createUser: (data: {
    username: string
    password: string
    role: string
    displayName?: string
  }) => api.post('/users', data),
  listUsers: (params: {
    page?: number
    pageSize?: number
    role?: string
    status?: string
    keyword?: string
  }) => api.get('/users', { params }),
  getUser: (id: number) => api.get(`/users/${id}`),
  updateUser: (id: number, data: { displayName?: string; role?: string }) =>
    api.put(`/users/${id}`, data),
  deleteUser: (id: number) => api.delete(`/users/${id}`),
  updateUserStatus: (id: number, status: string) =>
    api.put(`/users/${id}/status`, { status }),
  resetPassword: (id: number, newPassword: string) =>
    api.put(`/users/${id}/password`, { newPassword })
}

// 数据源管理相关
export const datasourceApi = {
  createDatasource: (data: {
    name: string
    description?: string
    host: string
    port: number
    databaseName: string
    username: string
    password: string
  }) => api.post('/datasources', data),
  listDatasources: (params: {
    page?: number
    pageSize?: number
    keyword?: string
  }) => api.get('/datasources', { params }),
  getDatasource: (id: number) => api.get(`/datasources/${id}`),
  updateDatasource: (
    id: number,
    data: {
      name?: string
      description?: string
      host?: string
      port?: number
      databaseName?: string
      username?: string
      password?: string
    }
  ) => api.put(`/datasources/${id}`, data),
  deleteDatasource: (id: number) => api.delete(`/datasources/${id}`),
  testConnection: (id: number) => api.post(`/datasources/${id}/test`),

  // 表配置管理
  createTable: (
    datasourceId: number,
    data: {
      tableName: string
      tableAlias?: string
      queryMode: string
      customSql?: string
      fieldMappings?: Array<{ mysqlField: string; aliasField: string }>
    }
  ) => api.post(`/datasources/${datasourceId}/tables`, data),
  listTables: (datasourceId: number) =>
    api.get(`/datasources/${datasourceId}/tables`),
  getTable: (id: number) => api.get(`/datasource-tables/${id}`),
  updateTable: (
    id: number,
    data: {
      tableAlias?: string
      queryMode?: string
      customSql?: string
      fieldMappings?: Array<{ mysqlField: string; aliasField: string }>
    }
  ) => api.put(`/datasource-tables/${id}`, data),
  deleteTable: (id: number) => api.delete(`/datasource-tables/${id}`),
  getFieldMappings: (tableId: number) =>
    api.get(`/datasource-tables/${tableId}/fields`),
  updateFieldMappings: (
    tableId: number,
    fieldMappings: Array<{ mysqlField: string; aliasField: string }>
  ) => api.post(`/datasource-tables/${tableId}/fields`, { fieldMappings })
}

// 权限管理相关
export const permissionApi = {
  // 数据源权限
  grantDatasourcePermissions: (userId: number, datasourceIds: number[]) =>
    api.post(`/users/${userId}/datasources`, { datasourceIds }),
  revokeDatasourcePermission: (userId: number, datasourceId: number) =>
    api.delete(`/users/${userId}/datasources/${datasourceId}`),
  listUserDatasources: (userId: number) =>
    api.get(`/users/${userId}/datasources`),
  listAllDatasourcesWithPermission: (userId: number) =>
    api.get(`/users/${userId}/datasources-with-permission`),

  // 表权限
  grantTablePermissions: (userId: number, tableIds: number[]) =>
    api.post(`/users/${userId}/tables`, { tableIds }),
  revokeTablePermission: (userId: number, tableId: number) =>
    api.delete(`/users/${userId}/tables/${tableId}`),
  listUserTables: (userId: number, datasourceId: number) =>
    api.get(`/users/${userId}/tables`, { params: { datasourceId } }),
  listAllTablesWithPermission: (userId: number, datasourceId: number) =>
    api.get(`/users/${userId}/tables-with-permission`, {
      params: { datasourceId }
    })
}

export default api
