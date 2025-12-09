<template>
  <div class="settings-page">
    <h2>系统设置</h2>

    <a-tabs>
      <a-tab-pane key="password" tab="修改密码">
        <a-card style="max-width: 400px">
          <a-form
            :model="passwordForm"
            @finish="handleChangePassword"
            layout="vertical"
          >
            <a-form-item
              label="原密码"
              name="oldPassword"
              :rules="[{ required: true, message: '请输入原密码' }]"
            >
              <a-input-password v-model:value="passwordForm.oldPassword" />
            </a-form-item>
            <a-form-item
              label="新密码"
              name="newPassword"
              :rules="[
                { required: true, message: '请输入新密码' },
                { min: 6, message: '密码至少6位' }
              ]"
            >
              <a-input-password v-model:value="passwordForm.newPassword" />
            </a-form-item>
            <a-form-item
              label="确认密码"
              name="confirmPassword"
              :rules="[
                { required: true, message: '请确认新密码' },
                { validator: validateConfirm }
              ]"
            >
              <a-input-password v-model:value="passwordForm.confirmPassword" />
            </a-form-item>
            <a-form-item>
              <a-button type="primary" html-type="submit" :loading="passwordLoading">
                修改密码
              </a-button>
            </a-form-item>
          </a-form>
        </a-card>
      </a-tab-pane>

      <a-tab-pane key="logs" tab="日志清理">
        <a-card style="max-width: 400px">
          <a-form layout="vertical">
            <a-form-item label="清理多少天前的日志">
              <a-input-number
                v-model:value="cleanDays"
                :min="1"
                :max="365"
                style="width: 200px"
              />
              <span style="margin-left: 8px">天</span>
            </a-form-item>
            <a-form-item>
              <a-popconfirm
                title="确定要清理日志吗？此操作不可恢复。"
                @confirm="handleCleanLogs"
              >
                <a-button type="primary" danger :loading="cleanLoading">
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
.settings-page h2 {
  margin-bottom: 24px;
}
</style>
