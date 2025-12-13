import axios from 'axios';

// 钉钉前端使用 /dingtalk 基础路径
const API_BASE_URL = '/dingtalk';

// 创建认证专用的 axios 实例（不需要 token）
const authApi = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器：添加时间戳防止缓存
authApi.interceptors.request.use(
  (config) => {
    config.params = {
      ...config.params,
      _t: Date.now(),
    };
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 用户登录
export const login = async (username: string, password: string): Promise<{ token: string; user: any }> => {
  const response = await authApi.post('/api/auth/login', { username, password });
  if (response.data.code !== 0) {
    throw new Error(response.data.msg || '登录失败');
  }
  return response.data.data;
};

// 获取当前用户信息
export const getCurrentUser = async (token: string): Promise<any> => {
  const response = await authApi.get('/api/auth/current', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (response.data.code !== 0) {
    throw new Error(response.data.msg || '获取用户信息失败');
  }
  return response.data.data;
};

// 用户登出
export const logout = async (token: string): Promise<void> => {
  try {
    await authApi.post('/api/auth/logout', {}, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
  } catch (error) {
    // 忽略登出错误
    console.error('登出失败:', error);
  }
};

// 获取用户可访问的数据源列表
export const getUserDatasources = async (token: string): Promise<any[]> => {
  const response = await authApi.get('/api/user/datasources', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (response.data.code !== 0) {
    throw new Error(response.data.msg || '获取数据源列表失败');
  }
  return response.data.data || [];
};

// 获取数据源下可访问的表列表
export const getUserTables = async (token: string, datasourceId: number): Promise<any[]> => {
  const response = await authApi.get(`/api/user/datasources/${datasourceId}/tables`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });
  if (response.data.code !== 0) {
    throw new Error(response.data.msg || '获取表列表失败');
  }
  return response.data.data || [];
};

// Token 管理
const TOKEN_KEY = 'dingtalk_user_token';
const USER_KEY = 'dingtalk_user_info';

export const saveToken = (token: string) => {
  localStorage.setItem(TOKEN_KEY, token);
};

export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY);
};

export const removeToken = () => {
  localStorage.removeItem(TOKEN_KEY);
};

export const saveUser = (user: any) => {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
};

export const getUser = (): any | null => {
  const userStr = localStorage.getItem(USER_KEY);
  if (userStr) {
    try {
      return JSON.parse(userStr);
    } catch (e) {
      return null;
    }
  }
  return null;
};

export const removeUser = () => {
  localStorage.removeItem(USER_KEY);
};
