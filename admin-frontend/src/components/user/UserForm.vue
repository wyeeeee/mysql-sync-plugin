<template>
  <a-modal
    v-model:open="visible"
    :title="title"
    @ok="handleOk"
    @cancel="visible = false"
  >
    <a-form :model="formData" :label-col="{ span: 6 }">
      <a-form-item label="用户名" required>
        <a-input
          v-model:value="formData.username"
          :disabled="isEdit"
          placeholder="请输入用户名"
        />
      </a-form-item>
      <a-form-item v-if="!isEdit" label="密码" required>
        <a-input-password
          v-model:value="formData.password"
          placeholder="请输入密码(至少6位)"
        />
      </a-form-item>
      <a-form-item label="显示名称">
        <a-input
          v-model:value="formData.displayName"
          placeholder="请输入显示名称"
        />
      </a-form-item>
      <a-form-item label="角色" required>
        <a-select v-model:value="formData.role" placeholder="请选择角色">
          <a-select-option value="admin">管理员</a-select-option>
          <a-select-option value="user">普通用户</a-select-option>
        </a-select>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { userApi } from '../../api'

const props = defineProps<{
  open: boolean
  isEdit: boolean
  user?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const visible = ref(props.open)
const title = ref(props.isEdit ? '编辑用户' : '新建用户')
const formData = ref({
  id: 0,
  username: '',
  password: '',
  displayName: '',
  role: 'user'
})

watch(() => props.open, (val) => {
  visible.value = val
  if (val) {
    title.value = props.isEdit ? '编辑用户' : '新建用户'
    if (props.isEdit && props.user) {
      formData.value = {
        id: props.user.id,
        username: props.user.username,
        password: '',
        displayName: props.user.displayName,
        role: props.user.role
      }
    } else {
      formData.value = {
        id: 0,
        username: '',
        password: '',
        displayName: '',
        role: 'user'
      }
    }
  }
})

watch(visible, (val) => {
  emit('update:open', val)
})

const handleOk = async () => {
  if (!formData.value.username) {
    message.error('请输入用户名')
    return
  }
  if (!props.isEdit && !formData.value.password) {
    message.error('请输入密码')
    return
  }
  if (!props.isEdit && formData.value.password.length < 6) {
    message.error('密码长度不能少于6位')
    return
  }
  if (!formData.value.role) {
    message.error('请选择角色')
    return
  }

  try {
    let res
    if (props.isEdit) {
      res = await userApi.updateUser(formData.value.id, {
        displayName: formData.value.displayName,
        role: formData.value.role
      })
    } else {
      res = await userApi.createUser({
        username: formData.value.username,
        password: formData.value.password,
        role: formData.value.role,
        displayName: formData.value.displayName
      })
    }

    if (res.code === 0) {
      message.success(props.isEdit ? '更新成功' : '创建成功')
      visible.value = false
      emit('success')
    } else {
      message.error(res.msg || '操作失败')
    }
  } catch (error) {
    message.error('操作失败')
  }
}
</script>
