<template>
  <a-config-provider :locale="zhCN" :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import zhCN from 'ant-design-vue/es/locale/zh_CN'

const themeConfig = {
  token: {
    colorPrimary: '#1e3a5f',
    colorLink: '#1e3a5f',
    colorLinkHover: '#2d4a6f',
    borderRadius: 6,
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif"
  }
}
</script>

<style>
:root {
  --cherry-primary: #1e3a5f;
  --cherry-primary-dark: #152a45;
  --cherry-primary-light: #e1e8f0;
  --cherry-gradient: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  --cherry-bg: linear-gradient(135deg, #f0f2f5 0%, #f5f7fa 50%, #ffffff 100%);
  --text-primary: #1a1a2e;
  --text-secondary: #5a6a7a;
  --border-color: #d0d7de;
  --card-shadow: 0 2px 8px rgba(30, 58, 95, 0.08);
  --card-shadow-hover: 0 4px 16px rgba(30, 58, 95, 0.12);
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: #f8f9fa;
}

#app {
  min-height: 100vh;
}

/* 全局滚动条美化 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: #ddd;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #ccc;
}

/* Ant Design 全局样式覆盖 */
.ant-btn-primary {
  background: var(--cherry-gradient);
  border: none;
}

.ant-btn-primary:hover {
  background: linear-gradient(135deg, #2d4a6f 0%, #3d5a7f 100%);
}

.ant-input:focus,
.ant-input-affix-wrapper:focus,
.ant-input-affix-wrapper-focused {
  border-color: var(--cherry-primary);
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1);
}

.ant-select-focused .ant-select-selector {
  border-color: var(--cherry-primary) !important;
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1) !important;
}

.ant-picker-focused {
  border-color: var(--cherry-primary);
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1);
}

.ant-checkbox-checked .ant-checkbox-inner {
  background-color: var(--cherry-primary);
  border-color: var(--cherry-primary);
}

.ant-switch-checked {
  background: var(--cherry-primary);
}

.ant-tabs-ink-bar {
  background: var(--cherry-primary);
}

.ant-tabs-tab.ant-tabs-tab-active .ant-tabs-tab-btn {
  color: var(--cherry-primary);
}

.ant-pagination-item-active {
  border-color: var(--cherry-primary);
}

.ant-pagination-item-active a {
  color: var(--cherry-primary);
}

.ant-spin-dot-item {
  background-color: var(--cherry-primary);
}

.ant-message-notice-content {
  border-radius: 8px;
}

.ant-modal-content {
  border-radius: 12px;
  overflow: hidden;
}

.ant-modal-header {
  border-radius: 12px 12px 0 0;
}

.ant-card {
  border-radius: 12px;
}

.ant-table {
  border-radius: 8px;
}

.ant-table-thead > tr > th {
  background: #fafafa;
}

/* 响应式工具类 */
@media (max-width: 576px) {
  .ant-modal {
    max-width: calc(100vw - 32px);
    margin: 16px auto;
  }

  .ant-modal-body {
    padding: 16px;
  }

  .ant-card-body {
    padding: 12px;
  }

  .ant-table-wrapper {
    overflow-x: auto;
  }

  .ant-form-item {
    margin-bottom: 16px;
  }
}

@media (max-width: 768px) {
  .ant-modal {
    max-width: calc(100vw - 48px);
  }

  .ant-descriptions-item-label {
    width: 80px !important;
  }
}
</style>
