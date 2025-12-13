import { useState } from 'react';
import { Form, Input, Button, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { login, saveToken, saveUser } from './auth';
import './App.css';

interface LoginProps {
  onLoginSuccess: () => void;
}

function Login({ onLoginSuccess }: LoginProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const { token, user } = await login(values.username, values.password);

      saveToken(token);
      saveUser(user);

      message.success('登录成功');
      onLoginSuccess();
    } catch (error: any) {
      if (error.errorFields) {
        message.error('请填写完整的登录信息');
      } else {
        message.error(error.message || '登录失败');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-header">
        <div className="login-logo">🍒</div>
        <h1 className="login-title">樱桃表格取数系统</h1>
        <p className="login-subtitle">连接数据，赋能业务</p>
      </div>

      <div className="login-card">
        <Form
          form={form}
          layout="vertical"
          onFinish={handleLogin}
          className="cherry-form"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#bfbfbf' }} />}
              placeholder="请输入用户名"
              size="large"
              className="cherry-input"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#bfbfbf' }} />}
              placeholder="请输入密码"
              size="large"
              className="cherry-input"
            />
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
              className="cherry-btn-primary"
            >
              登 录
            </Button>
          </Form.Item>
        </Form>
      </div>
    </div>
  );
}

export default Login;
