import { useState, useEffect } from 'react';
import { message } from 'antd';
import { bitable } from '@lark-base-open/connector-api';
import { getToken, removeToken, removeUser } from './auth';
import Login from './Login';
import Config from './Config';
import './App.css';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [sdkReady, setSdkReady] = useState(false);

  useEffect(() => {
    // 首先初始化飞书多维表格 SDK
    const initSDK = async () => {
      try {
        // 飞书 SDK 不需要显式初始化，直接标记为就绪
        console.log('飞书多维表格 SDK 初始化成功');
        setSdkReady(true);

        // SDK 初始化成功后，检查是否已登录
        const token = getToken();
        if (token) {
          setIsLoggedIn(true);
        }
      } catch (e) {
        console.error('飞书多维表格 SDK 初始化失败:', e);
        message.error('初始化失败,请刷新重试');
        setSdkReady(true); // 即使失败也设置为 true，允许在开发环境中继续
      }
    };

    initSDK();
  }, []);

  const handleLoginSuccess = () => {
    setIsLoggedIn(true);
  };

  const handleLogout = () => {
    removeToken();
    removeUser();
    setIsLoggedIn(false);
  };

  // SDK 初始化中
  if (!sdkReady) {
    return (
      <div className="app-container">
        <div style={{ textAlign: 'center', padding: '60px 20px' }}>
          <p>正在初始化飞书多维表格 SDK...</p>
        </div>
      </div>
    );
  }

  if (!isLoggedIn) {
    return <Login onLoginSuccess={handleLoginSuccess} />;
  }

  return <Config onLogout={handleLogout} />;
}

export default App;
