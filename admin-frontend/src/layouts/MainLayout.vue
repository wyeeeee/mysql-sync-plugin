<template>
  <a-layout class="main-layout">
    <!-- 移动端遮罩层 -->
    <div
      v-if="mobileMenuVisible"
      class="mobile-overlay"
      @click="mobileMenuVisible = false"
    ></div>

    <!-- 侧边栏 -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :collapsible="!isMobile"
      :trigger="isMobile ? null : undefined"
      :collapsed-width="isMobile ? 0 : 80"
      :width="220"
      :class="['cherry-sider', { 'mobile-visible': mobileMenuVisible }]"
      breakpoint="lg"
      @breakpoint="handleBreakpoint"
    >
      <div class="logo">
        <span class="logo-icon">🍒</span>
        <span v-if="!collapsed" class="logo-text">樱桃取数系统</span>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
        class="cherry-menu"
        @click="handleMenuClick"
      >
        <a-menu-item key="/">
          <router-link to="/">
            <DashboardOutlined />
            <span>仪表盘</span>
          </router-link>
        </a-menu-item>
        <a-menu-item key="/users">
          <router-link to="/users">
            <TeamOutlined />
            <span>用户管理</span>
          </router-link>
        </a-menu-item>
        <a-menu-item key="/datasources">
          <router-link to="/datasources">
            <DatabaseOutlined />
            <span>数据源管理</span>
          </router-link>
        </a-menu-item>
        <a-menu-item key="/logs">
          <router-link to="/logs">
            <FileTextOutlined />
            <span>日志管理</span>
          </router-link>
        </a-menu-item>
        <a-menu-item key="/settings">
          <router-link to="/settings">
            <SettingOutlined />
            <span>系统设置</span>
          </router-link>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-layout class="main-content-layout">
      <a-layout-header class="header">
        <div class="header-left">
          <!-- 移动端菜单按钮 -->
          <a-button
            v-if="isMobile"
            type="text"
            class="mobile-menu-btn"
            @click="mobileMenuVisible = !mobileMenuVisible"
          >
            <MenuOutlined />
          </a-button>
          <span v-if="isMobile" class="mobile-title">🍒 樱桃取数</span>
        </div>
        <div class="header-right">
          <a-dropdown>
            <a class="user-dropdown">
              <a-avatar class="user-avatar" :size="32">
                <template #icon><UserOutlined /></template>
              </a-avatar>
              <span class="username">{{ authStore.username }}</span>
            </a>
            <template #overlay>
              <a-menu>
                <a-menu-item @click="handleLogout">
                  <LogoutOutlined />
                  退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>
      <a-layout-content class="content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  DashboardOutlined,
  TeamOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '../stores/auth'
import { authApi } from '../api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const collapsed = ref(false)
const isMobile = ref(false)
const mobileMenuVisible = ref(false)
const selectedKeys = computed(() => [route.path])

function checkMobile() {
  isMobile.value = window.innerWidth < 992
  if (!isMobile.value) {
    mobileMenuVisible.value = false
  }
}

function handleBreakpoint(broken: boolean) {
  isMobile.value = broken
  if (broken) {
    collapsed.value = true
  }
}

function handleMenuClick() {
  if (isMobile.value) {
    mobileMenuVisible.value = false
  }
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

async function handleLogout() {
  try {
    await authApi.logout()
  } catch (e) {
    // 忽略错误
  }
  authStore.logout()
  message.success('已退出登录')
  router.push('/login')
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}

/* 侧边栏深色主题 */
.cherry-sider {
  background: linear-gradient(180deg, #152a45 0%, #1e3a5f 100%) !important;
}

.cherry-sider :deep(.ant-layout-sider-trigger) {
  background: rgba(0, 0, 0, 0.15);
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  background: rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.logo-icon {
  font-size: 28px;
  flex-shrink: 0;
}

.logo-text {
  color: white;
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
}

/* 菜单样式 */
.cherry-menu {
  background: transparent !important;
  border-right: none !important;
}

.cherry-menu :deep(.ant-menu-item) {
  margin: 4px 8px;
  border-radius: 8px;
}

.cherry-menu :deep(.ant-menu-item a) {
  color: rgba(255, 255, 255, 0.85);
}

.cherry-menu :deep(.ant-menu-item:hover) {
  background: rgba(255, 255, 255, 0.15) !important;
}

.cherry-menu :deep(.ant-menu-item-selected) {
  background: rgba(255, 255, 255, 0.25) !important;
}

.cherry-menu :deep(.ant-menu-item-selected a) {
  color: white;
}

/* 顶部导航 */
.header {
  background: white;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2px 8px rgba(30, 58, 95, 0.08);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.mobile-menu-btn {
  font-size: 18px;
  color: #1e3a5f;
}

.mobile-title {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #2c3e50;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 8px;
  transition: background 0.3s;
}

.user-dropdown:hover {
  background: #e8eef5;
}

.user-avatar {
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
}

.username {
  font-weight: 500;
}

/* 内容区域 */
.main-content-layout {
  background: #f8f9fa;
}

.content {
  margin: 16px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  min-height: calc(100vh - 64px - 32px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 移动端遮罩 */
.mobile-overlay {
  display: none;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
  z-index: 999;
}

/* 响应式适配 */
@media (max-width: 991px) {
  .cherry-sider {
    position: fixed !important;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 1000;
    transform: translateX(-100%);
    transition: transform 0.3s ease;
  }

  .cherry-sider.mobile-visible {
    transform: translateX(0);
  }

  .mobile-overlay {
    display: block;
  }

  .header {
    padding: 0 16px;
  }

  .content {
    margin: 12px;
    padding: 16px;
    border-radius: 10px;
    min-height: calc(100vh - 64px - 24px);
  }

  .username {
    display: none;
  }
}

@media (max-width: 576px) {
  .content {
    margin: 8px;
    padding: 12px;
    border-radius: 8px;
  }

  .header {
    padding: 0 12px;
  }

  .mobile-title {
    font-size: 14px;
  }
}
</style>
