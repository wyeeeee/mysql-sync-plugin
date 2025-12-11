<template>
  <div class="users-page">
    <div class="page-header">
      <h2>用户管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <template #icon><PlusOutlined /></template>
        新建用户
      </a-button>
    </div>

    <a-card>
      <a-form layout="inline" :model="searchForm" class="search-form">
        <a-form-item label="角色">
          <a-select
            v-model:value="searchForm.role"
            style="width: 120px"
            placeholder="全部"
            allowClear
          >
            <a-select-option value="admin">管理员</a-select-option>
            <a-select-option value="user">普通用户</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="状态">
          <a-select
            v-model:value="searchForm.status"
            style="width: 120px"
            placeholder="全部"
            allowClear
          >
            <a-select-option value="active">启用</a-select-option>
            <a-select-option value="disabled">禁用</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="关键词">
          <a-input
            v-model:value="searchForm.keyword"
            placeholder="用户名或显示名称"
            style="width: 200px"
          />
        </a-form-item>
        <a-form-item>
          <a-button type="primary" @click="handleSearch">查询</a-button>
          <a-button style="margin-left: 8px" @click="handleReset">重置</a-button>
        </a-form-item>
      </a-form>

      <a-table
        :columns="columns"
        :data-source="users"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'role'">
            <a-tag :color="record.role === 'admin' ? 'blue' : 'default'">
              {{ record.role === 'admin' ? '管理员' : '普通用户' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.status === 'active' ? 'success' : 'error'">
              {{ record.status === 'active' ? '启用' : '禁用' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="showEditModal(record)">
                编辑
              </a-button>
              <a-button
                type="link"
                size="small"
                @click="showResetPasswordModal(record)"
              >
                重置密码
              </a-button>
              <a-button
                type="link"
                size="small"
                :danger="record.status === 'active'"
                @click="toggleUserStatus(record)"
              >
                {{ record.status === 'active' ? '禁用' : '启用' }}
              </a-button>
              <a-button
                type="link"
                size="small"
                @click="showPermissionModal(record)"
              >
                权限管理
              </a-button>
              <a-button
                type="link"
                size="small"
                danger
                @click="handleDelete(record)"
              >
                删除
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- 用户表单组件 -->
    <UserForm
      v-model:open="modalVisible"
      :is-edit="isEdit"
      :user="currentUser"
      @success="loadUsers"
    />

    <!-- 重置密码对话框 -->
    <a-modal
      v-model:open="resetPasswordVisible"
      title="重置密码"
      @ok="handleResetPassword"
      @cancel="resetPasswordVisible = false"
    >
      <a-form :label-col="{ span: 6 }">
        <a-form-item label="新密码" required>
          <a-input-password
            v-model:value="newPassword"
            placeholder="请输入新密码(至少6位)"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 权限管理组件 -->
    <PermissionModal
      v-model:open="permissionVisible"
      :user="currentPermissionUser"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { userApi } from '../api'
import UserForm from '../components/user/UserForm.vue'
import PermissionModal from '../components/user/PermissionModal.vue'

const loading = ref(false)
const users = ref<any[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0
})

const searchForm = reactive({
  role: undefined,
  status: undefined,
  keyword: ''
})

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '用户名', dataIndex: 'username', key: 'username' },
  { title: '显示名称', dataIndex: 'displayName', key: 'displayName' },
  { title: '角色', dataIndex: 'role', key: 'role' },
  { title: '状态', dataIndex: 'status', key: 'status' },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt' },
  { title: '操作', key: 'action', width: 350 }
]

const modalVisible = ref(false)
const isEdit = ref(false)
const currentUser = ref<any>(null)

const resetPasswordVisible = ref(false)
const currentUserId = ref(0)
const newPassword = ref('')

const permissionVisible = ref(false)
const currentPermissionUser = ref<any>(null)

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await userApi.listUsers({
      page: pagination.current,
      pageSize: pagination.pageSize,
      ...searchForm
    })
    if (res.code === 0) {
      users.value = res.data.list || []
      pagination.total = res.data.total
    } else {
      message.error(res.msg || '加载用户列表失败')
    }
  } catch (error) {
    message.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  loadUsers()
}

const handleReset = () => {
  searchForm.role = undefined
  searchForm.status = undefined
  searchForm.keyword = ''
  handleSearch()
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadUsers()
}

const showCreateModal = () => {
  isEdit.value = false
  currentUser.value = null
  modalVisible.value = true
}

const showEditModal = (record: any) => {
  isEdit.value = true
  currentUser.value = record
  modalVisible.value = true
}

const showResetPasswordModal = (record: any) => {
  currentUserId.value = record.id
  newPassword.value = ''
  resetPasswordVisible.value = true
}

const handleResetPassword = async () => {
  if (!newPassword.value) {
    message.error('请输入新密码')
    return
  }
  if (newPassword.value.length < 6) {
    message.error('密码长度不能少于6位')
    return
  }

  try {
    const res = await userApi.resetPassword(currentUserId.value, newPassword.value)
    if (res.code === 0) {
      message.success('重置密码成功')
      resetPasswordVisible.value = false
    } else {
      message.error(res.msg || '重置密码失败')
    }
  } catch (error) {
    message.error('重置密码失败')
  }
}

const toggleUserStatus = (record: any) => {
  const newStatus = record.status === 'active' ? 'disabled' : 'active'
  const action = newStatus === 'active' ? '启用' : '禁用'

  Modal.confirm({
    title: `确认${action}用户?`,
    content: `确定要${action}用户 "${record.username}" 吗?`,
    onOk: async () => {
      try {
        const res = await userApi.updateUserStatus(record.id, newStatus)
        if (res.code === 0) {
          message.success(`${action}成功`)
          loadUsers()
        } else {
          message.error(res.msg || `${action}失败`)
        }
      } catch (error) {
        message.error(`${action}失败`)
      }
    }
  })
}

const handleDelete = (record: any) => {
  Modal.confirm({
    title: '确认删除?',
    content: `确定要删除用户 "${record.username}" 吗? 此操作不可恢复。`,
    okText: '确认',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        const res = await userApi.deleteUser(record.id)
        if (res.code === 0) {
          message.success('删除成功')
          loadUsers()
        } else {
          message.error(res.msg || '删除失败')
        }
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

const showPermissionModal = (record: any) => {
  currentPermissionUser.value = record
  permissionVisible.value = true
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.users-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
}

.search-form {
  margin-bottom: 16px;
}
</style>
