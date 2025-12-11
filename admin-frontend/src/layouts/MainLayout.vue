<template>
  <a-layout class="main-layout">
    <a-layout-sider v-model:collapsed="collapsed" collapsible>
      <div class="logo">
        <span v-if="!collapsed">MySQL同步插件</span>
        <span v-else>MS</span>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
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
    <a-layout>
      <a-layout-header class="header">
        <div class="header-right">
          <a-dropdown>
            <a class="user-dropdown">
              <UserOutlined />
              <span>{{ authStore.username }}</span>
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
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  DashboardOutlined,
  TeamOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '../stores/auth'
import { authApi } from '../api'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const collapsed = ref(false)
const selectedKeys = computed(() => [route.path])

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

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 16px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}

.header {
  background: white;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #333;
  cursor: pointer;
}

.content {
  margin: 24px;
  padding: 24px;
  background: white;
  border-radius: 8px;
  min-height: 280px;
}
</style>
