import { useEffect, useState } from 'react';
import { Form, Input, Button, Select, message, Steps, Space } from 'antd';
import { initView } from 'dingtalk-docs-cool-app';
import { MySQLConfig, getTables } from './api';
import './App.css';

// 钉钉全局对象类型定义
declare global {
  interface Window {
    Dingdocs: any;
  }
}

const { Step } = Steps;
const { Option } = Select;

function App() {
  const [currentStep, setCurrentStep] = useState(0);
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // 数据状态
  const [mysqlConfig, setMysqlConfig] = useState<Partial<MySQLConfig>>({
    port: 3306,
  });
  const [tables, setTables] = useState<string[]>([]);

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

  // 步骤1: 连接数据库
  const handleConnectDB = async () => {
    try {
      const values = await form.validateFields(['host', 'port', 'username', 'password', 'database']);
      setLoading(true);

      const config: Omit<MySQLConfig, 'table'> = {
        host: values.host,
        port: values.port,
        username: values.username,
        password: values.password,
        database: values.database,
      };

      // 获取数据表列表
      const tableList = await getTables(config);

      if (tableList.length === 0) {
        message.warning('该数据库中没有数据表');
        return;
      }

      setTables(tableList);
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

  // 步骤2: 选择数据表并保存配置
  const handleSelectTable = async () => {
    try {
      const values = await form.validateFields(['table']);
      setLoading(true);

      const config: MySQLConfig = {
        ...mysqlConfig as Omit<MySQLConfig, 'table'>,
        table: values.table,
      };

      // 调用钉钉SDK保存配置并跳转到下一步
      // 钉钉AI表格会接管后续的字段选择流程
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
        message.error('请选择数据表');
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
  };

  return (
    <div className="app-container">
      <Steps current={currentStep} style={{ marginBottom: 24 }}>
        <Step title="连接数据库" />
        <Step title="选择数据表" />
      </Steps>

      <Form
        form={form}
        layout="vertical"
        initialValues={{ port: 3306 }}
      >
        {/* 步骤1: 数据库连接配置 */}
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
              label="数据库名"
              name="database"
              rules={[{ required: true, message: '请输入数据库名' }]}
            >
              <Input placeholder="例如: my_database" />
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
              <Button type="primary" onClick={handleConnectDB} loading={loading} block>
                连接数据库
              </Button>
            </Form.Item>
          </div>
        )}

        {/* 步骤2: 选择数据表 */}
        {currentStep === 1 && (
          <div>
            <Form.Item
              label="选择数据表"
              name="table"
              rules={[{ required: true, message: '请选择数据表' }]}
            >
              <Select
                placeholder="请选择要同步的数据表"
                showSearch
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

            <Form.Item>
              <Space>
                <Button onClick={handlePrevious}>上一步</Button>
                <Button type="primary" onClick={handleSelectTable} loading={loading}>
                  下一步
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
