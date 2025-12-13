<template>
  <div class="users-page">
    <div class="page-header">
      <h2 class="page-title">
        <TeamOutlined class="title-icon" />
        用户管理
      </h2>
      <a-button type="primary" class="cherry-btn-sm" @click="showCreateModal">
        <PlusOutlined />
        新建用户
      </a-button>
    </div>

    <a-card class="cherry-card">
      <!-- 搜索表单 -->
      <div class="search-form-wrapper">
        <a-row :gutter="[12, 12]">
          <a-col :xs="24" :sm="12" :md="6" :lg="5">
            <a-select
              v-model:value="searchForm.role"
              placeholder="选择角色"
              allowClear
              style="width: 100%"
              class="cherry-select"
            >
              <a-select-option value="admin">管理员</a-select-option>
              <a-select-option value="user">普通用户</a-select-option>
            </a-select>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6" :lg="5">
            <a-select
              v-model:value="searchForm.status"
              placeholder="选择状态"
              allowClear
              style="width: 100%"
              class="cherry-select"
            >
              <a-select-option value="active">启用</a-select-option>
              <a-select-option value="disabled">禁用</a-select-option>
            </a-select>
          </a-col>
          <a-col :xs="24" :sm="12" :md="8" :lg="8">
            <a-input
              v-model:value="searchForm.keyword"
              placeholder="搜索用户名或显示名称"
              allowClear
              class="cherry-input"
            >
              <template #prefix>
                <SearchOutlined style="color: #bfbfbf" />
              </template>
            </a-input>
          </a-col>
          <a-col :xs="24" :sm="12" :md="4" :lg="6">
            <a-space class="search-btns">
              <a-button type="primary" class="cherry-btn-sm" @click="handleSearch">
                <SearchOutlined />
                查询
              </a-button>
              <a-button @click="handleReset">重置</a-button>
            </a-space>
          </a-col>
        </a-row>
      </div>

      <!-- 表格 -->
      <div class="table-wrapper">
        <a-table
          :columns="responsiveColumns"
          :data-source="users"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
          :scroll="{ x: 800 }"
          class="cherry-table"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'role'">
              <a-tag :color="record.role === 'admin' ? '#1e3a5f' : 'default'" class="role-tag">
                {{ record.role === 'admin' ? '管理员' : '普通用户' }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'status'">
              <a-badge
                :status="record.status === 'active' ? 'success' : 'error'"
                :text="record.status === 'active' ? '启用' : '禁用'"
              />
            </template>
            <template v-else-if="column.key === 'action'">
              <!-- 桌面端显示按钮 -->
              <a-space v-if="!isMobile" :size="4">
                <a-button type="link" size="small" class="action-btn" @click="showEditModal(record)">
                  编辑
                </a-button>
                <a-button type="link" size="small" class="action-btn" @click="showResetPasswordModal(record)">
                  重置密码
                </a-button>
                <a-button
                  type="link"
                  size="small"
                  class="action-btn"
                  :class="{ 'danger-btn': record.status === 'active' }"
                  @click="toggleUserStatus(record)"
                >
                  {{ record.status === 'active' ? '禁用' : '启用' }}
                </a-button>
                <a-button type="link" size="small" class="action-btn" @click="showPermissionModal(record)">
                  权限
                </a-button>
                <a-button type="link" size="small" class="action-btn danger-btn" @click="handleDelete(record)">
                  删除
                </a-button>
              </a-space>
              <!-- 移动端显示下拉菜单 -->
              <a-dropdown v-else>
                <a-button type="link" size="small">
                  操作 <DownOutlined />
                </a-button>
                <template #overlay>
                  <a-menu>
                    <a-menu-item @click="showEditModal(record)">编辑</a-menu-item>
                    <a-menu-item @click="showResetPasswordModal(record)">重置密码</a-menu-item>
                    <a-menu-item @click="toggleUserStatus(record)">
                      {{ record.status === 'active' ? '禁用' : '启用' }}
                    </a-menu-item>
                    <a-menu-item @click="showPermissionModal(record)">权限管理</a-menu-item>
                    <a-menu-item danger @click="handleDelete(record)">删除</a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </template>
          </template>
        </a-table>
      </div>
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
      :width="isMobile ? '90%' : 420"
      class="cherry-modal"
    >
      <a-form :label-col="{ span: 6 }">
        <a-form-item label="新密码" required>
          <a-input-password
            v-model:value="newPassword"
            placeholder="请输入新密码(至少6位)"
            class="cherry-input"
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
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, TeamOutlined, SearchOutlined, DownOutlined } from '@ant-design/icons-vue'
import { userApi } from '../api'
import UserForm from '../components/user/UserForm.vue'
import PermissionModal from '../components/user/PermissionModal.vue'

const loading = ref(false)
const users = ref<any[]>([])
const isMobile = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`
})

const searchForm = reactive({
  role: undefined,
  status: undefined,
  keyword: ''
})

const allColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
  { title: '显示名称', dataIndex: 'displayName', key: 'displayName', width: 120 },
  { title: '角色', dataIndex: 'role', key: 'role', width: 100 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 80 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 280, fixed: 'right' as const }
]

const responsiveColumns = computed(() => {
  if (isMobile.value) {
    return allColumns.filter(col => ['username', 'role', 'status', 'action'].includes(col.key))
      .map(col => col.key === 'action' ? { ...col, width: 80, fixed: undefined } : col)
  }
  return allColumns
})

const modalVisible = ref(false)
const isEdit = ref(false)
const currentUser = ref<any>(null)

const resetPasswordVisible = ref(false)
const currentUserId = ref(0)
const newPassword = ref('')

const permissionVisible = ref(false)
const currentPermissionUser = ref<any>(null)

function checkMobile() {
  isMobile.value = window.innerWidth < 768
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  loadUsers()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

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
    okButtonProps: { class: 'cherry-btn-sm' },
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
</script>

<style scoped>
.users-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
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

.cherry-btn-sm {
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  border: none;
  border-radius: 8px;
  height: 36px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.cherry-btn-sm:hover {
  background: linear-gradient(135deg, #2d4a6f 0%, #3d5a7f 100%);
}

.cherry-card {
  border-radius: 12px;
  border: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.search-form-wrapper {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f5f5f5;
}

.search-btns {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.cherry-input :deep(.ant-input) {
  border-radius: 8px;
}

.cherry-input :deep(.ant-input:focus),
.cherry-input :deep(.ant-input-affix-wrapper:focus),
.cherry-input :deep(.ant-input-affix-wrapper-focused) {
  border-color: #1e3a5f;
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1);
}

.cherry-select :deep(.ant-select-selector) {
  border-radius: 8px !important;
}

.cherry-select :deep(.ant-select-focused .ant-select-selector) {
  border-color: #1e3a5f !important;
  box-shadow: 0 0 0 2px rgba(30, 58, 95, 0.1) !important;
}

.table-wrapper {
  overflow-x: auto;
}

.cherry-table :deep(.ant-table) {
  border-radius: 8px;
}

.cherry-table :deep(.ant-table-thead > tr > th) {
  background: #fafafa;
  font-weight: 600;
  color: #2c3e50;
}

.role-tag {
  border-radius: 4px;
}

.action-btn {
  padding: 0 6px;
  font-size: 13px;
  color: #1e3a5f;
}

.action-btn:hover {
  color: #2d4a6f;
}

.action-btn.danger-btn {
  color: #ef4444;
}

.action-btn.danger-btn:hover {
  color: #dc2626;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .page-header {
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
  }

  .search-form-wrapper {
    margin-bottom: 16px;
    padding-bottom: 16px;
  }

  .search-btns {
    width: 100%;
    justify-content: flex-start;
  }

  .cherry-card {
    border-radius: 10px;
  }

  .cherry-card :deep(.ant-card-body) {
    padding: 16px;
  }
}

@media (max-width: 576px) {
  .page-title {
    font-size: 16px;
  }

  .cherry-card :deep(.ant-card-body) {
    padding: 12px;
  }

  .search-form-wrapper {
    margin-bottom: 12px;
    padding-bottom: 12px;
  }
}
</style>
