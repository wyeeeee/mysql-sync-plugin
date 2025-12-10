import axios from 'axios';

// 飞书前端使用 /feishu 基础路径
const API_BASE_URL = '/feishu';

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 字段映射配置
export interface FieldMapping {
  mysqlField: string;  // MySQL原始字段名
  aliasField: string;  // AI表格显示的别名
}

// MySQL配置接口
export interface MySQLConfig {
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
  table?: string;
  queryMode?: 'table' | 'sql';     // 取数模式
  customSQL?: string;              // 自定义SQL语句
  fieldMappings?: FieldMapping[];  // 字段映射配置
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
export const getDatabases = async (config: Omit<MySQLConfig, 'database' | 'table'>): Promise<string[]> => {
  const response = await api.post('/api/databases', config);
  if (response.data.code !== 0) {
    throw new Error(response.data.msg || '获取数据库列表失败');
  }
  return response.data.data || [];
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

// 预览SQL执行结果（获取字段列表）
export const previewSQL = async (config: MySQLConfig): Promise<any[]> => {
  const response = await api.post('/api/preview_sql', config);
  if (response.data.code !== 0) {
    throw new Error(response.data.msg || 'SQL执行失败');
  }
  return response.data.data || [];
};
