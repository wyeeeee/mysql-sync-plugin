<template>
  <div class="settings-page">
    <div class="page-header">
      <h2 class="page-title">
        <SettingOutlined class="title-icon" />
        系统设置
      </h2>
    </div>

    <a-tabs class="cherry-tabs">
      <a-tab-pane key="password">
        <template #tab>
          <span class="tab-label">
            <LockOutlined />
            修改密码
          </span>
        </template>
        <a-card class="cherry-card settings-card">
          <div class="card-header">
            <KeyOutlined class="card-icon" />
            <div class="card-info">
              <h3>账户密码</h3>
              <p>修改您的登录密码，建议定期更换以保障账户安全</p>
            </div>
          </div>
          <a-form
            :model="passwordForm"
            @finish="handleChangePassword"
            layout="vertical"
            class="cherry-form"
          >
            <a-form-item
              label="原密码"
              name="oldPassword"
              :rules="[{ required: true, message: '请输入原密码' }]"
            >
              <a-input-password
                v-model:value="passwordForm.oldPassword"
                placeholder="请输入原密码"
                class="cherry-input"
              />
            </a-form-item>
            <a-form-item
              label="新密码"
              name="newPassword"
              :rules="[
                { required: true, message: '请输入新密码' },
                { min: 6, message: '密码至少6位' }
              ]"
            >
              <a-input-password
                v-model:value="passwordForm.newPassword"
                placeholder="请输入新密码（至少6位）"
                class="cherry-input"
              />
            </a-form-item>
            <a-form-item
              label="确认密码"
              name="confirmPassword"
              :rules="[
                { required: true, message: '请确认新密码' },
                { validator: validateConfirm }
              ]"
            >
              <a-input-password
                v-model:value="passwordForm.confirmPassword"
                placeholder="请再次输入新密码"
                class="cherry-input"
              />
            </a-form-item>
            <a-form-item>
              <a-button type="primary" html-type="submit" :loading="passwordLoading" class="cherry-btn">
                <SaveOutlined />
                保存修改
              </a-button>
            </a-form-item>
          </a-form>
        </a-card>
      </a-tab-pane>

      <a-tab-pane key="logs">
        <template #tab>
          <span class="tab-label">
            <DeleteOutlined />
            日志清理
          </span>
        </template>
        <a-card class="cherry-card settings-card">
          <div class="card-header">
            <ClearOutlined class="card-icon warning" />
            <div class="card-info">
              <h3>日志清理</h3>
              <p>清理历史日志数据，释放存储空间</p>
            </div>
          </div>
          <a-form layout="vertical" class="cherry-form">
            <a-form-item label="清理范围">
              <div class="clean-input-wrapper">
                <span class="clean-label">清理</span>
                <a-input-number
                  v-model:value="cleanDays"
                  :min="1"
                  :max="365"
                  class="clean-input"
                />
                <span class="clean-label">天前的日志</span>
              </div>
              <div class="clean-hint">
                <InfoCircleOutlined />
                将删除 {{ cleanDays }} 天前的所有日志记录
              </div>
            </a-form-item>
            <a-form-item>
              <a-popconfirm
                title="确定要清理日志吗？"
                description="此操作不可恢复，请谨慎操作。"
                @confirm="handleCleanLogs"
                ok-text="确认清理"
                cancel-text="取消"
              >
                <a-button type="primary" danger :loading="cleanLoading" class="cherry-btn-danger">
                  <DeleteOutlined />
                  清理日志
                </a-button>
              </a-popconfirm>
            </a-form-item>
          </a-form>
        </a-card>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { message } from 'ant-design-vue'
import {
  SettingOutlined,
  LockOutlined,
  DeleteOutlined,
  KeyOutlined,
  ClearOutlined,
  SaveOutlined,
  InfoCircleOutlined
} from '@ant-design/icons-vue'
import { authApi, logApi } from '../api'

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})
const passwordLoading = ref(false)

const cleanDays = ref(30)
const cleanLoading = ref(false)

function validateConfirm(_rule: any, value: string) {
  if (value && value !== passwordForm.newPassword) {
    return Promise.reject('两次输入的密码不一致')
  }
  return Promise.resolve()
}

async function handleChangePassword() {
  passwordLoading.value = true
  try {
    const res = await authApi.changePassword(
      passwordForm.oldPassword,
      passwordForm.newPassword
    )
    if (res.code === 0) {
      message.success('密码修改成功')
      passwordForm.oldPassword = ''
      passwordForm.newPassword = ''
      passwordForm.confirmPassword = ''
    } else {
      message.error(res.msg || '修改失败')
    }
  } catch (e) {
    message.error('修改失败')
  } finally {
    passwordLoading.value = false
  }
}

async function handleCleanLogs() {
  cleanLoading.value = true
  try {
    const res = await logApi.cleanLogs(cleanDays.value)
    if (res.code === 0) {
      message.success(`成功清理 ${res.data.affected} 条日志`)
    } else {
      message.error(res.msg || '清理失败')
    }
  } catch (e) {
    message.error('清理失败')
  } finally {
    cleanLoading.value = false
  }
}
</script>

<style scoped>
.settings-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 10px;
}

.title-icon {
  color: #1e3a5f;
}

.cherry-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 20px;
}

.cherry-tabs :deep(.ant-tabs-tab) {
  padding: 12px 16px;
  font-size: 14px;
}

.cherry-tabs :deep(.ant-tabs-tab-active) {
  font-weight: 600;
}

.cherry-tabs :deep(.ant-tabs-ink-bar) {
  background: #1e3a5f;
}

.cherry-tabs :deep(.ant-tabs-tab.ant-tabs-tab-active .ant-tabs-tab-btn) {
  color: #1e3a5f;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cherry-card {
  border-radius: 12px;
  border: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.settings-card {
  max-width: 500px;
}

.card-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f5f5f5;
}

.card-icon {
  font-size: 32px;
  color: #1e3a5f;
  background: linear-gradient(135deg, #e8eef5 0%, #d0dbe8 100%);
  padding: 12px;
  border-radius: 12px;
}

.card-icon.warning {
  color: #faad14;
  background: linear-gradient(135deg, #fffbe6 0%, #fff1b8 100%);
}

.card-info h3 {
  margin: 0 0 4px 0;
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.card-info p {
  margin: 0;
  font-size: 13px;
  color: #7f8c8d;
}

.cherry-form :deep(.ant-form-item-label > label) {
  font-weight: 500;
  color: #2c3e50;
}

.cherry-input :deep(.ant-input),
.cherry-input :deep(.ant-input-password) {
  border-radius: 8px;
  height: 40px;
}

.cherry-input :deep(.ant-input-affix-wrapper) {
  border-radius: 8px;
}

.cherry-input :deep(.ant-input-affix-wrapper:hover),
.cherry-input :deep(.ant-input-affix-wrapper:focus),
.cherry-input :deep(.ant-input-affix-wrapper-focused) {
  border-color: #1e3a5f;
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1);
}

.cherry-btn {
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  border: none;
  border-radius: 8px;
  height: 40px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
}

.cherry-btn:hover {
  background: linear-gradient(135deg, #2d4a6f 0%, #3d5a7f 100%);
}

.cherry-btn-danger {
  border-radius: 8px;
  height: 40px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
}

.clean-input-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.clean-label {
  color: #666;
  font-size: 14px;
}

.clean-input {
  width: 100px;
}

.clean-input :deep(.ant-input-number) {
  border-radius: 8px;
}

.clean-hint {
  margin-top: 12px;
  padding: 10px 14px;
  background: #fff7e6;
  border-radius: 8px;
  font-size: 13px;
  color: #d48806;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .page-header {
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
  }

  .settings-card {
    max-width: 100%;
  }

  .cherry-card :deep(.ant-card-body) {
    padding: 16px;
  }

  .card-header {
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;
    padding-bottom: 16px;
  }

  .card-icon {
    font-size: 28px;
    padding: 10px;
  }

  .cherry-tabs :deep(.ant-tabs-tab) {
    padding: 10px 12px;
    font-size: 13px;
  }
}

@media (max-width: 576px) {
  .page-title {
    font-size: 16px;
  }

  .cherry-card :deep(.ant-card-body) {
    padding: 12px;
  }

  .card-info h3 {
    font-size: 15px;
  }

  .card-info p {
    font-size: 12px;
  }

  .clean-input-wrapper {
    gap: 8px;
  }

  .clean-input {
    width: 80px;
  }

  .clean-hint {
    font-size: 12px;
    padding: 8px 12px;
  }
}
</style>
