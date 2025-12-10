import { useEffect, useState } from 'react';
import { Form, Input, Button, Select, message, Steps, Space, Table, Divider, Radio } from 'antd';
import { bitable } from '@lark-base-open/connector-api';
import { MySQLConfig, FieldMapping, getDatabases, getTables, getTableFields, previewSQL } from './api';
import './App.css';

const { Step } = Steps;
const { Option } = Select;
const { TextArea } = Input;

// localStorage 存储键名
const STORAGE_KEY = 'mysql_sync_history';

// 历史记录类型（不包含密码）
interface ConnectionHistory {
  host: string;
  port: number;
  username: string;
}

// 从 localStorage 加载历史记录
const loadHistory = (): ConnectionHistory | null => {
  try {
    const data = localStorage.getItem(STORAGE_KEY);
    if (data) {
      return JSON.parse(data);
    }
  } catch (e) {
    console.error('加载历史记录失败:', e);
  }
  return null;
};

// 保存历史记录到 localStorage（不包含密码）
const saveHistory = (config: { host: string; port: number; username: string }) => {
  try {
    const history: ConnectionHistory = {
      host: config.host,
      port: config.port,
      username: config.username,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(history));
  } catch (e) {
    console.error('保存历史记录失败:', e);
  }
};

// 字段信息类型
interface FieldInfo {
  id: string;
  name: string;
  type: string;
  isPrimary: boolean;
  description?: string;
}

// 取数模式类型
type QueryMode = 'table' | 'sql';

function App() {
  const [sdkReady, setSdkReady] = useState(false);
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

  // 取数模式
  const [queryMode, setQueryMode] = useState<QueryMode>('table');
  const [customSQL, setCustomSQL] = useState('');

  useEffect(() => {
    // 初始化飞书SDK
    console.log('飞书多维表格SDK初始化成功');
    setSdkReady(true);

    // 加载历史记录并填充表单
    const history = loadHistory();
    if (history) {
      form.setFieldsValue({
        host: history.host,
        port: history.port,
        username: history.username,
      });
      setMysqlConfig(prev => ({
        ...prev,
        host: history.host,
        port: history.port,
        username: history.username,
      }));
    }
  }, [form]);

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

      // 保存连接信息到历史记录（不包含密码）
      saveHistory(config);
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

  // 预览SQL，获取字段列表
  const handlePreviewSQL = async () => {
    if (!customSQL.trim()) {
      message.warning('请输入SQL语句');
      return;
    }

    try {
      setLoading(true);

      const config = {
        ...mysqlConfig,
        customSQL: customSQL.trim(),
      } as MySQLConfig;

      const fieldList = await previewSQL(config);
      setFields(fieldList);

      // 初始化字段映射
      const mappings: FieldMapping[] = fieldList.map((f: FieldInfo) => ({
        mysqlField: f.name,
        aliasField: f.name,
      }));
      setFieldMappings(mappings);

      message.success(`SQL执行成功，共${fieldList.length}个字段`);
    } catch (error: any) {
      message.error('SQL执行失败: ' + (error.message || '未知错误'));
      setFields([]);
      setFieldMappings([]);
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

  // 取数模式切换
  const handleQueryModeChange = (mode: QueryMode) => {
    setQueryMode(mode);
    setFields([]);
    setFieldMappings([]);
    if (mode === 'table') {
      setCustomSQL('');
    }
  };

  // 步骤2: 保存配置
  const handleSaveConfig = async () => {
    try {
      setLoading(true);

      // 根据取数模式验证
      if (queryMode === 'table') {
        const values = await form.validateFields(['database', 'table']);
        const config: MySQLConfig = {
          ...mysqlConfig as Omit<MySQLConfig, 'table' | 'database'>,
          database: values.database,
          table: values.table,
          queryMode: 'table',
          fieldMappings: fieldMappings,
        };

        await saveConfigToFeishu(config);
      } else {
        const values = await form.validateFields(['database']);
        if (!customSQL.trim()) {
          message.error('请输入SQL语句');
          setLoading(false);
          return;
        }
        if (fields.length === 0) {
          message.error('请先预览SQL获取字段列表');
          setLoading(false);
          return;
        }

        const config: MySQLConfig = {
          ...mysqlConfig as Omit<MySQLConfig, 'table' | 'database'>,
          database: values.database,
          queryMode: 'sql',
          customSQL: customSQL.trim(),
          fieldMappings: fieldMappings,
        };

        await saveConfigToFeishu(config);
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

  // 保存配置到飞书
  const saveConfigToFeishu = async (config: MySQLConfig) => {
    try {
      await bitable.saveConfigAndGoNext(config);
      message.success('配置保存成功,正在跳转...');
      console.log('保存配置:', config);
    } catch (error: any) {
      console.log('保存配置:', config);
      message.info('开发环境: 配置已保存到控制台');
    }
  };

  // 返回上一步
  const handlePrevious = () => {
    setCurrentStep(0);
    setDatabases([]);
    setTables([]);
    setFields([]);
    setFieldMappings([]);
    setQueryMode('table');
    setCustomSQL('');
    form.setFieldsValue({ database: undefined, table: undefined });
  };

  // 获取字段的备注
  const getFieldDescription = (mysqlField: string): string => {
    const field = fields.find(f => f.name === mysqlField);
    return field?.description || '';
  };

  // 应用单个字段的备注到别名
  const applyDescriptionToAlias = (mysqlField: string) => {
    const description = getFieldDescription(mysqlField);
    if (description) {
      handleFieldMappingChange(mysqlField, description);
    }
  };

  // 应用所有有备注的字段
  const applyAllDescriptions = () => {
    setFieldMappings(prev =>
      prev.map(m => {
        const description = getFieldDescription(m.mysqlField);
        return description ? { ...m, aliasField: description } : m;
      })
    );
    message.success('已应用所有备注');
  };

  // 检查是否有任何字段有备注
  const hasAnyDescription = fields.some(f => f.description);

  // 字段映射表格列定义
  const mappingColumns = [
    {
      title: 'MySQL字段名',
      dataIndex: 'mysqlField',
      key: 'mysqlField',
      width: '25%',
    },
    {
      title: '数据库备注',
      key: 'description',
      width: '25%',
      render: (_: any, record: FieldMapping) => {
        const description = getFieldDescription(record.mysqlField);
        return description ? (
          <span style={{ color: '#666' }}>{description}</span>
        ) : (
          <span style={{ color: '#ccc' }}>无备注</span>
        );
      },
    },
    {
      title: '多维表格显示名',
      dataIndex: 'aliasField',
      key: 'aliasField',
      width: '35%',
      render: (text: string, record: FieldMapping) => (
        <Input
          value={text}
          onChange={e => handleFieldMappingChange(record.mysqlField, e.target.value)}
          placeholder="输入显示名称"
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: '15%',
      render: (_: any, record: FieldMapping) => {
        const description = getFieldDescription(record.mysqlField);
        return description ? (
          <Button
            type="link"
            size="small"
            onClick={() => applyDescriptionToAlias(record.mysqlField)}
          >
            应用备注
          </Button>
        ) : null;
      },
    },
  ];

  // SDK初始化中
  if (!sdkReady) {
    return (
      <div className="app-container">
        <div style={{ textAlign: 'center', padding: '60px 20px' }}>
          <p>正在初始化飞书多维表格 SDK...</p>
        </div>
      </div>
    );
  }

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

            {/* 取数模式选择 */}
            <Form.Item label="取数模式">
              <Radio.Group
                value={queryMode}
                onChange={e => handleQueryModeChange(e.target.value)}
              >
                <Radio.Button value="table">选择数据表</Radio.Button>
                <Radio.Button value="sql">自定义SQL</Radio.Button>
              </Radio.Group>
            </Form.Item>

            {/* 选择数据表模式 */}
            {queryMode === 'table' && (
              <Form.Item
                label="选择数据表"
                name="table"
                rules={[{ required: queryMode === 'table', message: '请选择数据表' }]}
              >
                <Select
                  placeholder="请先选择数据库"
                  showSearch
                  disabled={tables.length === 0}
                  onChange={handleTableChange}
                  filterOption={(input, option) =>
                    String(option?.label || option?.children || '').toLowerCase().includes(input.toLowerCase())
                  }
                >
                  {tables.map(table => (
                    <Option key={table} value={table}>
                      {table}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            )}

            {/* 自定义SQL模式 */}
            {queryMode === 'sql' && (
              <Form.Item label="自定义SQL">
                <TextArea
                  value={customSQL}
                  onChange={e => setCustomSQL(e.target.value)}
                  placeholder="请输入SELECT查询语句，例如：SELECT id, name, age FROM users WHERE status = 1"
                  rows={4}
                  style={{ fontFamily: 'monospace' }}
                />
                <Button
                  type="default"
                  onClick={handlePreviewSQL}
                  loading={loading}
                  style={{ marginTop: 8 }}
                  disabled={!mysqlConfig.database}
                >
                  预览SQL（获取字段）
                </Button>
              </Form.Item>
            )}

            {/* 字段映射配置 */}
            {fields.length > 0 && (
              <>
                <Divider>字段映射配置</Divider>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                  <span style={{ color: '#666' }}>
                    配置字段在多维表格中的显示名称，留空或保持原名则使用MySQL字段名
                  </span>
                  {hasAnyDescription && (
                    <Button type="primary" size="small" onClick={applyAllDescriptions}>
                      全部应用备注
                    </Button>
                  )}
                </div>
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
