<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <div class="login-logo">🍒</div>
        <h1 class="login-title">樱桃取数系统</h1>
        <p class="login-subtitle">连接数据，赋能业务</p>
      </div>
      <a-form
        :model="form"
        @finish="handleLogin"
        layout="vertical"
        class="login-form"
      >
        <a-form-item
          name="username"
          :rules="[{ required: true, message: '请输入用户名' }]"
        >
          <a-input
            v-model:value="form.username"
            placeholder="请输入用户名"
            size="large"
            class="cherry-input"
          >
            <template #prefix>
              <UserOutlined class="input-icon" />
            </template>
          </a-input>
        </a-form-item>
        <a-form-item
          name="password"
          :rules="[{ required: true, message: '请输入密码' }]"
        >
          <a-input-password
            v-model:value="form.password"
            placeholder="请输入密码"
            size="large"
            class="cherry-input"
          >
            <template #prefix>
              <LockOutlined class="input-icon" />
            </template>
          </a-input-password>
        </a-form-item>
        <a-form-item style="margin-bottom: 16px;">
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
            class="cherry-btn"
          >
            登 录
          </a-button>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { authApi } from '../api'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  username: '',
  password: ''
})

const loading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    const res = await authApi.login(form.username, form.password)
    if (res.code === 0) {
      authStore.setAuth(res.data, form.username)
      message.success('登录成功')
      router.push('/')
    } else {
      message.error(res.msg || '登录失败')
    }
  } catch (e) {
    message.error('登录失败，请检查网络')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f0f2f5 0%, #f5f7fa 50%, #ffffff 100%);
  padding: 20px;
  box-sizing: border-box;
}

.login-box {
  width: 100%;
  max-width: 400px;
  padding: 40px 32px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(30, 58, 95, 0.12);
  border: 1px solid #d0d7de;
  transition: box-shadow 0.3s ease, transform 0.3s ease;
}

.login-box:hover {
  box-shadow: 0 12px 48px rgba(30, 58, 95, 0.18);
  transform: translateY(-2px);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo {
  font-size: 56px;
  margin-bottom: 16px;
  animation: bounce 2s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-8px); }
}

.login-title {
  margin: 0 0 8px 0;
  font-size: 26px;
  font-weight: 600;
  color: #2c3e50;
}

.login-subtitle {
  margin: 0;
  font-size: 14px;
  color: #7f8c8d;
}

.login-form :deep(.ant-form-item) {
  margin-bottom: 20px;
}

.cherry-input :deep(.ant-input),
.cherry-input :deep(.ant-input-password) {
  height: 48px;
  border-radius: 10px;
  border-color: #f0e6e6;
  font-size: 15px;
}

.cherry-input :deep(.ant-input-affix-wrapper) {
  height: 48px;
  border-radius: 10px;
  border-color: #f0e6e6;
  padding: 0 16px;
}

.cherry-input :deep(.ant-input-affix-wrapper:hover),
.cherry-input :deep(.ant-input-affix-wrapper:focus),
.cherry-input :deep(.ant-input-affix-wrapper-focused) {
  border-color: #1e3a5f;
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1);
}

.input-icon {
  color: #bfbfbf;
  font-size: 16px;
}

.cherry-btn {
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 10px;
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  border: none;
  transition: all 0.3s ease;
}

.cherry-btn:hover {
  background: linear-gradient(135deg, #2d4a6f 0%, #3d5a7f 100%);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(30, 58, 95, 0.35);
}

.login-hint {
  text-align: center;
  margin: 20px 0 0 0;
  color: #bdc3c7;
  font-size: 13px;
}

/* 响应式适配 */
@media (max-width: 480px) {
  .login-box {
    padding: 32px 24px;
    border-radius: 12px;
  }

  .login-logo {
    font-size: 48px;
  }

  .login-title {
    font-size: 22px;
  }

  .cherry-input :deep(.ant-input-affix-wrapper) {
    height: 44px;
  }

  .cherry-btn {
    height: 44px;
    font-size: 15px;
  }
}

@media (max-width: 360px) {
  .login-box {
    padding: 24px 20px;
  }

  .login-logo {
    font-size: 40px;
  }

  .login-title {
    font-size: 20px;
  }
}
</style>
