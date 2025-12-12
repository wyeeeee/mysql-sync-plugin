import { useState, useEffect } from 'react';
import { message } from 'antd';
import { initView } from 'dingtalk-docs-cool-app';
import { getToken, removeToken, removeUser } from './auth';
import Login from './Login';
import Config from './Config';
import './App.css';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [sdkReady, setSdkReady] = useState(false);

  useEffect(() => {
    // 首先初始化钉钉酷应用SDK
    initView({
      onReady: () => {
        console.log('钉钉酷应用SDK初始化成功');
        setSdkReady(true);

        // SDK 初始化成功后，检查是否已登录
        const token = getToken();
        if (token) {
          setIsLoggedIn(true);
        }
      },
      onError: (e: any) => {
        console.error('钉钉酷应用SDK初始化失败:', e);
        message.error('初始化失败,请刷新重试');
        setSdkReady(true); // 即使失败也设置为 true，允许在开发环境中继续
      },
    });
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
          <p>正在初始化钉钉酷应用 SDK...</p>
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
