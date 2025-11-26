import axios from 'axios';

// 获取后端服务地址(从环境变量或默认值)
const API_BASE_URL = 'https://xipiapi.moonmark.chat';

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// MySQL配置接口
export interface MySQLConfig {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  table?: string;
}

// 测试MySQL连接
export const testConnection = async (config: MySQLConfig): Promise<boolean> => {
  try {
    const response = await api.post('/api/test_connection', config);
    return response.data.code === 0;
  } catch (error) {
    console.error('连接测试失败:', error);
    return false;
  }
};

// 获取数据库列表
export const getDatabases = async (config: Omit<MySQLConfig, 'database'>): Promise<string[]> => {
  try {
    const response = await api.post('/api/databases', config);
    return response.data.data || [];
  } catch (error) {
    console.error('获取数据库列表失败:', error);
    return [];
  }
};

// 获取数据表列表
export const getTables = async (config: Omit<MySQLConfig, 'table'>): Promise<string[]> => {
  try {
    const response = await api.post('/api/tables', config);
    return response.data.data || [];
  } catch (error) {
    console.error('获取数据表列表失败:', error);
    return [];
  }
};

// 获取表字段列表
export const getTableFields = async (config: MySQLConfig): Promise<any[]> => {
  try {
    const response = await api.post('/api/fields', config);
    return response.data.data || [];
  } catch (error) {
    console.error('获取表字段失败:', error);
    return [];
  }
};
