import { useEffect, useState } from 'react';
import { Form, Button, Select, message } from 'antd';
import { DatabaseOutlined, TableOutlined } from '@ant-design/icons';
import { bitable } from '@lark-base-open/connector-api';
import { getToken, getUserDatasources, getUserTables } from './auth';
import './App.css';

interface Datasource {
  id: number;
  name: string;
  description?: string;
}

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
  const [datasources, setDatasources] = useState<Datasource[]>([]);
  const [tables, setTables] = useState<DatasourceTable[]>([]);
  const [selectedTable, setSelectedTable] = useState<DatasourceTable | null>(null);

  useEffect(() => {
    loadDatasources();
  }, []);

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

  const handleTableChange = (tableId: number) => {
    const table = tables.find(t => t.id === tableId);
    if (table) {
      setSelectedTable(table);
    }
  };

  const handleSaveConfig = async () => {
    try {
      await form.validateFields(['datasourceId', 'tableId']);

      if (!selectedTable) {
        message.error('请选择表');
        return;
      }

      setLoading(true);

      const config = {
        tableId: selectedTable.id,
      };

      await saveConfigToFeishu(config);
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

  const saveConfigToFeishu = async (config: any) => {
    try {
      await bitable.saveConfigAndGoNext(config);
      message.success('配置保存成功，正在跳转...');
    } catch (error: any) {
      console.log('保存配置:', config);
      message.info('开发环境：配置已保存到控制台');
    }
  };

  return (
    <div className="config-container">
      <div className="config-header">
        <div className="config-header-left">
          <span className="config-logo">🍒</span>
          <h1 className="config-title">樱桃表格取数系统</h1>
        </div>
        <Button type="text" onClick={onLogout} className="cherry-btn-text">
          退出
        </Button>
      </div>

      <div className="config-card">
        <Form form={form} layout="vertical" className="cherry-form">
          <Form.Item
            label={<span><DatabaseOutlined style={{ marginRight: 6 }} />选择数据源</span>}
            name="datasourceId"
            rules={[{ required: true, message: '请选择数据源' }]}
          >
            <Select
              placeholder="请选择数据源"
              showSearch
              loading={loading}
              onChange={handleDatasourceChange}
              className="cherry-select"
              filterOption={(input, option) =>
                String(option?.children || '').toLowerCase().includes(input.toLowerCase())
              }
            >
              {datasources.map(ds => (
                <Select.Option key={ds.id} value={ds.id}>
                  {ds.name}{ds.description && ` (${ds.description})`}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            label={<span><TableOutlined style={{ marginRight: 6 }} />选择数据表</span>}
            name="tableId"
            rules={[{ required: true, message: '请选择数据表' }]}
          >
            <Select
              placeholder={tables.length === 0 ? '请先选择数据源' : '请选择数据表'}
              showSearch
              disabled={tables.length === 0}
              onChange={handleTableChange}
              className="cherry-select"
              filterOption={(input, option) =>
                String(option?.children || '').toLowerCase().includes(input.toLowerCase())
              }
            >
              {tables.map(table => (
                <Select.Option key={table.id} value={table.id}>
                  {table.tableAlias || table.tableName}
                  {table.tableAlias && ` (${table.tableName})`}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          {selectedTable && (
            <div className="table-info-card">
              <div className="table-info-title">
                <TableOutlined /> 已选择的数据表
              </div>
              <div className="table-info-item">
                <span className="table-info-label">表名</span>
                <span className="table-info-value">{selectedTable.tableName}</span>
              </div>
              {selectedTable.tableAlias && (
                <div className="table-info-item">
                  <span className="table-info-label">别名</span>
                  <span className="table-info-value">{selectedTable.tableAlias}</span>
                </div>
              )}
              <div className="table-info-item">
                <span className="table-info-label">取数模式</span>
                <span className="table-info-value">
                  {selectedTable.queryMode === 'table' ? '数据表' : '自定义SQL'}
                </span>
              </div>
              {selectedTable.customSql && (
                <div className="table-info-item">
                  <span className="table-info-label">SQL语句</span>
                  <div className="table-info-sql">{selectedTable.customSql}</div>
                </div>
              )}
            </div>
          )}

          <Form.Item style={{ marginTop: 24, marginBottom: 0 }}>
            <Button
              type="primary"
              onClick={handleSaveConfig}
              loading={loading}
              disabled={!selectedTable}
              block
              className="cherry-btn-primary"
            >
              确认配置
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
}

export default Config;
