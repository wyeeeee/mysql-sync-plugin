import { useEffect, useState } from 'react';
import { Form, Button, Select, message } from 'antd';
import { getToken, getUserDatasources, getUserTables } from './auth';
import './App.css';

// 钉钉全局对象类型定义
declare global {
  interface Window {
    Dingdocs: any;
  }
}

const { Option } = Select;

// 数据源类型
interface Datasource {
  id: number;
  name: string;
  description?: string;
}

// 表类型
interface DatasourceTable {
  id: number;
  tableName: string;
  tableAlias?: string;
  queryMode: string;
  customSql?: string;
}

interface ConfigProps {
  onLogout: () => void;
}

function Config({ onLogout }: ConfigProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // 数据状态
  const [datasources, setDatasources] = useState<Datasource[]>([]);
  const [tables, setTables] = useState<DatasourceTable[]>([]);
  const [selectedDatasource, setSelectedDatasource] = useState<Datasource | null>(null);
  const [selectedTable, setSelectedTable] = useState<DatasourceTable | null>(null);

  useEffect(() => {
    // 加载用户数据源列表
    loadDatasources();
  }, []);

  // 加载用户可访问的数据源列表
  const loadDatasources = async () => {
    try {
      setLoading(true);
      const token = getToken();
      if (!token) {
        message.error('未登录，请重新登录');
        onLogout();
        return;
      }

      const list = await getUserDatasources(token);
      setDatasources(list);

      if (list.length === 0) {
        message.warning('您还没有可访问的数据源，请联系管理员授权');
      }
    } catch (error: any) {
      message.error('加载数据源失败: ' + (error.message || '未知错误'));
      if (error.message?.includes('认证') || error.message?.includes('登录')) {
        onLogout();
      }
    } finally {
      setLoading(false);
    }
  };

  // 数据源选择变化时，获取表列表
  const handleDatasourceChange = async (datasourceId: number) => {
    try {
      setLoading(true);
      setTables([]);
      setSelectedTable(null);
      form.setFieldsValue({ tableId: undefined });

      const token = getToken();
      if (!token) {
        message.error('未登录，请重新登录');
        onLogout();
        return;
      }

      const ds = datasources.find(d => d.id === datasourceId);
      setSelectedDatasource(ds || null);

      const tableList = await getUserTables(token, datasourceId);
      setTables(tableList);

      if (tableList.length === 0) {
        message.warning('该数据源下没有可访问的表');
      }
    } catch (error: any) {
      message.error('获取表列表失败: ' + (error.message || '未知错误'));
      if (error.message?.includes('认证') || error.message?.includes('登录')) {
        onLogout();
      }
    } finally {
      setLoading(false);
    }
  };

  // 表选择变化时
  const handleTableChange = (tableId: number) => {
    const table = tables.find(t => t.id === tableId);
    if (table) {
      setSelectedTable(table);
    }
  };

  // 保存配置
  const handleSaveConfig = async () => {
    try {
      await form.validateFields(['datasourceId', 'tableId']);

      if (!selectedTable) {
        message.error('请选择表');
        return;
      }

      setLoading(true);

      // 构建配置对象（新方案：使用 tableId）
      const config = {
        tableId: selectedTable.id,
      };

      await saveConfig(config);
    } catch (error: any) {
      if (error.errorFields) {
        message.error('请完成数据源和表的选择');
      } else {
        message.error('保存失败: ' + (error.message || '未知错误'));
      }
    } finally {
      setLoading(false);
    }
  };

  // 保存配置到钉钉
  const saveConfig = async (config: any) => {
    if (window.Dingdocs?.base?.host?.saveConfigAndGoNext) {
      await window.Dingdocs.base.host.saveConfigAndGoNext(config);
      message.success('配置保存成功,正在跳转...');
      console.log('保存配置:', config);
    } else {
      console.log('保存配置:', config);
      message.info('开发环境: 配置已保存到控制台');
    }
  };

  return (
    <div className="app-container">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <h2 style={{ margin: 0 }}>MySQL 同步配置</h2>
        <Button onClick={onLogout}>退出登录</Button>
      </div>

      <Form
        form={form}
        layout="vertical"
      >
        <Form.Item
          label="选择数据源"
          name="datasourceId"
          rules={[{ required: true, message: '请选择数据源' }]}
        >
          <Select
            placeholder="请选择数据源"
            showSearch
            loading={loading}
            onChange={handleDatasourceChange}
            filterOption={(input, option) =>
              String(option?.children || '').toLowerCase().includes(input.toLowerCase())
            }
          >
            {datasources.map(ds => (
              <Option key={ds.id} value={ds.id}>
                {ds.name} {ds.description && `(${ds.description})`}
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          label="选择表"
          name="tableId"
          rules={[{ required: true, message: '请选择表' }]}
        >
          <Select
            placeholder="请先选择数据源"
            showSearch
            disabled={tables.length === 0}
            onChange={handleTableChange}
            filterOption={(input, option) =>
              String(option?.children || '').toLowerCase().includes(input.toLowerCase())
            }
          >
            {tables.map(table => (
              <Option key={table.id} value={table.id}>
                {table.tableAlias || table.tableName}
                {table.tableAlias && ` (${table.tableName})`}
              </Option>
            ))}
          </Select>
        </Form.Item>

        {selectedTable && (
          <div style={{ marginTop: 16, padding: 12, background: '#f5f5f5', borderRadius: 4 }}>
            <p style={{ margin: 0, color: '#666' }}>
              <strong>表名:</strong> {selectedTable.tableName}
            </p>
            {selectedTable.tableAlias && (
              <p style={{ margin: '4px 0 0 0', color: '#666' }}>
                <strong>别名:</strong> {selectedTable.tableAlias}
              </p>
            )}
            <p style={{ margin: '4px 0 0 0', color: '#666' }}>
              <strong>取数模式:</strong> {selectedTable.queryMode === 'table' ? '数据表' : '自定义SQL'}
            </p>
            {selectedTable.customSql && (
              <p style={{ margin: '4px 0 0 0', color: '#666', fontFamily: 'monospace', fontSize: '12px' }}>
                <strong>SQL:</strong> {selectedTable.customSql}
              </p>
            )}
          </div>
        )}

        <Form.Item style={{ marginTop: 24 }}>
          <Button
            type="primary"
            onClick={handleSaveConfig}
            loading={loading}
            disabled={!selectedTable}
            block
          >
            确认
          </Button>
        </Form.Item>
      </Form>
    </div>
  );
}

export default Config;
