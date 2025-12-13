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
    initView({
      onReady: () => {
        console.log('钉钉酷应用SDK初始化成功');
        setSdkReady(true);
        const token = getToken();
        if (token) {
          setIsLoggedIn(true);
        }
      },
      onError: (e: any) => {
        console.error('钉钉酷应用SDK初始化失败:', e);
        message.error('初始化失败，请刷新重试');
        setSdkReady(true);
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

  if (!sdkReady) {
    return (
      <div className="loading-container">
        <div className="loading-logo">🍒</div>
        <div className="loading-spinner"></div>
        <p className="loading-text">正在初始化...</p>
      </div>
    );
  }

  if (!isLoggedIn) {
    return <Login onLoginSuccess={handleLoginSuccess} />;
  }

  return <Config onLogout={handleLogout} />;
}

export default App;
