import { useEffect, useState } from 'react';
import { Form, Input, Button, Select, message, Steps, Space, Table, Divider } from 'antd';
import { initView } from 'dingtalk-docs-cool-app';
import { MySQLConfig, FieldMapping, getDatabases, getTables, getTableFields } from './api';
import './App.css';

// 钉钉全局对象类型定义
declare global {
  interface Window {
    Dingdocs: any;
  }
}

const { Step } = Steps;
const { Option } = Select;

// 字段信息类型
interface FieldInfo {
  id: string;
  name: string;
  type: string;
  isPrimary: boolean;
}

function App() {
  const [currentStep, setCurrentStep] = useState(0);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // 数据状态
  const [mysqlConfig, setMysqlConfig] = useState<Partial<MySQLConfig>>({
    port: 3306,
  });
  const [databases, setDatabases] = useState<string[]>([]);
  const [tables, setTables] = useState<string[]>([]);
  const [fields, setFields] = useState<FieldInfo[]>([]);
  const [fieldMappings, setFieldMappings] = useState<FieldMapping[]>([]);

  useEffect(() => {
    // 初始化钉钉酷应用SDK
    initView({
      onReady: () => {
        console.log('钉钉酷应用SDK初始化成功');
      },
      onError: (e: any) => {
        console.error('钉钉酷应用SDK初始化失败:', e);
        message.error('初始化失败,请刷新重试');
      },
    });
  }, []);

  // 步骤1: 连接数据库服务器，获取数据库列表
  const handleConnectServer = async () => {
    try {
      const values = await form.validateFields(['host', 'port', 'username', 'password']);
      setLoading(true);

      const config = {
        host: values.host,
        port: Number(values.port),
        username: values.username,
        password: values.password,
      };

      // 获取数据库列表
      const dbList = await getDatabases(config);

      if (dbList.length === 0) {
        message.warning('该服务器上没有可用的数据库');
        return;
      }

      setDatabases(dbList);
      setMysqlConfig(config);
      setCurrentStep(1);
      message.success('连接成功');
    } catch (error: any) {
      if (error.errorFields) {
        message.error('请填写完整的连接信息');
      } else {
        message.error('连接失败: ' + (error.message || '未知错误'));
      }
    } finally {
      setLoading(false);
    }
  };

  // 数据库选择变化时，获取数据表列表
  const handleDatabaseChange = async (database: string) => {
    try {
      setLoading(true);
      setTables([]);
      setFields([]);
      setFieldMappings([]);
      form.setFieldsValue({ table: undefined });

      const config = {
        ...mysqlConfig,
        database,
      } as Omit<MySQLConfig, 'table'>;

      const tableList = await getTables(config);
      setTables(tableList);
      setMysqlConfig(prev => ({ ...prev, database }));

      if (tableList.length === 0) {
        message.warning('该数据库中没有数据表');
      }
    } catch (error: any) {
      message.error('获取数据表失败: ' + (error.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  // 数据表选择变化时，获取字段列表
  const handleTableChange = async (table: string) => {
    try {
      setLoading(true);

      const config = {
        ...mysqlConfig,
        table,
      } as MySQLConfig;

      const fieldList = await getTableFields(config);
      setFields(fieldList);
      setMysqlConfig(prev => ({ ...prev, table }));

      // 初始化字段映射，默认别名等于原字段名
      const mappings: FieldMapping[] = fieldList.map((f: FieldInfo) => ({
        mysqlField: f.name,
        aliasField: f.name,
      }));
      setFieldMappings(mappings);
    } catch (error: any) {
      message.error('获取字段列表失败: ' + (error.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  // 更新字段映射
  const handleFieldMappingChange = (mysqlField: string, aliasField: string) => {
    setFieldMappings(prev =>
      prev.map(m =>
        m.mysqlField === mysqlField ? { ...m, aliasField } : m
      )
    );
  };

  // 步骤2: 保存配置
  const handleSaveConfig = async () => {
    try {
      const values = await form.validateFields(['database', 'table']);
      setLoading(true);

      const config: MySQLConfig = {
        ...mysqlConfig as Omit<MySQLConfig, 'table' | 'database'>,
        database: values.database,
        table: values.table,
        fieldMappings: fieldMappings,
      };

      // 调用钉钉SDK保存配置并跳转到下一步
      if (window.Dingdocs?.base?.host?.saveConfigAndGoNext) {
        await window.Dingdocs.base.host.saveConfigAndGoNext(config);
        message.success('配置保存成功,正在跳转...');
        console.log('保存配置:', config);
      } else {
        // 开发环境模拟
        console.log('保存配置:', config);
        message.info('开发环境: 配置已保存到控制台');
      }
    } catch (error: any) {
      if (error.errorFields) {
        message.error('请完成所有配置');
      } else {
        message.error('保存失败: ' + (error.message || '未知错误'));
      }
    } finally {
      setLoading(false);
    }
  };

  // 返回上一步
  const handlePrevious = () => {
    setCurrentStep(0);
    setDatabases([]);
    setTables([]);
    setFields([]);
    setFieldMappings([]);
    form.setFieldsValue({ database: undefined, table: undefined });
  };

  // 字段映射表格列定义
  const mappingColumns = [
    {
      title: 'MySQL字段名',
      dataIndex: 'mysqlField',
      key: 'mysqlField',
      width: '40%',
    },
    {
      title: 'AI表格显示名',
      dataIndex: 'aliasField',
      key: 'aliasField',
      width: '60%',
      render: (text: string, record: FieldMapping) => (
        <Input
          value={text}
          onChange={e => handleFieldMappingChange(record.mysqlField, e.target.value)}
          placeholder="输入显示名称"
        />
      ),
    },
  ];

  return (
    <div className="app-container">
      <Steps current={currentStep} style={{ marginBottom: 24 }}>
        <Step title="连接服务器" />
        <Step title="选择数据源" />
      </Steps>

      <Form
        form={form}
        layout="vertical"
        initialValues={{ port: 3306 }}
      >
        {/* 步骤1: 数据库服务器连接配置 */}
        {currentStep === 0 && (
          <div>
            <Form.Item
              label="数据库地址"
              name="host"
              rules={[{ required: true, message: '请输入数据库地址' }]}
            >
              <Input placeholder="例如: 192.168.1.100 或 db.example.com" />
            </Form.Item>

            <Form.Item
              label="端口"
              name="port"
              rules={[{ required: true, message: '请输入端口号' }]}
            >
              <Input type="number" placeholder="3306" />
            </Form.Item>

            <Form.Item
              label="用户名"
              name="username"
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input placeholder="数据库用户名" />
            </Form.Item>

            <Form.Item
              label="密码"
              name="password"
              rules={[{ required: true, message: '请输入密码' }]}
            >
              <Input.Password placeholder="数据库密码" />
            </Form.Item>

            <Form.Item>
              <Button type="primary" onClick={handleConnectServer} loading={loading} block>
                连接服务器
              </Button>
            </Form.Item>
          </div>
        )}

        {/* 步骤2: 选择数据库、数据表和配置字段映射 */}
        {currentStep === 1 && (
          <div>
            <Form.Item
              label="选择数据库"
              name="database"
              rules={[{ required: true, message: '请选择数据库' }]}
            >
              <Select
                placeholder="请选择数据库"
                showSearch
                onChange={handleDatabaseChange}
                filterOption={(input, option) =>
                  (option?.children as string).toLowerCase().includes(input.toLowerCase())
                }
              >
                {databases.map(db => (
                  <Option key={db} value={db}>
                    {db}
                  </Option>
                ))}
              </Select>
            </Form.Item>

            <Form.Item
              label="选择数据表"
              name="table"
              rules={[{ required: true, message: '请选择数据表' }]}
            >
              <Select
                placeholder="请先选择数据库"
                showSearch
                disabled={tables.length === 0}
                onChange={handleTableChange}
                filterOption={(input, option) =>
                  (option?.children as string).toLowerCase().includes(input.toLowerCase())
                }
              >
                {tables.map(table => (
                  <Option key={table} value={table}>
                    {table}
                  </Option>
                ))}
              </Select>
            </Form.Item>

            {/* 字段映射配置 */}
            {fields.length > 0 && (
              <>
                <Divider>字段映射配置</Divider>
                <p style={{ color: '#666', marginBottom: 12 }}>
                  配置字段在AI表格中的显示名称，留空或保持原名则使用MySQL字段名
                </p>
                <Table
                  dataSource={fieldMappings}
                  columns={mappingColumns}
                  rowKey="mysqlField"
                  pagination={false}
                  size="small"
                  style={{ marginBottom: 24 }}
                />
              </>
            )}

            <Form.Item>
              <Space>
                <Button onClick={handlePrevious}>上一步</Button>
                <Button
                  type="primary"
                  onClick={handleSaveConfig}
                  loading={loading}
                  disabled={fields.length === 0}
                >
                  保存配置
                </Button>
              </Space>
            </Form.Item>
          </div>
        )}
      </Form>
    </div>
  );
}

export default App;
