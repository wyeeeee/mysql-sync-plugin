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

    <!-- 创建/编辑用户对话框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
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

    <!-- 权限管理对话框 -->
    <a-modal
      v-model:open="permissionVisible"
      title="权限管理"
      width="800px"
      @ok="permissionVisible = false"
      @cancel="permissionVisible = false"
    >
      <a-tabs v-model:activeKey="permissionTab">
        <a-tab-pane key="datasource" tab="数据源权限">
          <a-checkbox-group
            v-model:value="selectedDatasources"
            style="width: 100%"
            @change="handleDatasourcePermissionChange"
          >
            <a-row>
              <a-col
                v-for="ds in datasourcesWithPermission"
                :key="ds.id"
                :span="24"
                style="margin-bottom: 8px"
              >
                <a-checkbox :value="ds.id">
                  {{ ds.name }}
                  <span style="color: #999; margin-left: 8px">
                    {{ ds.description }}
                  </span>
                </a-checkbox>
              </a-col>
            </a-row>
          </a-checkbox-group>
        </a-tab-pane>
        <a-tab-pane key="table" tab="表权限" :disabled="!currentDatasourceId">
          <a-form-item label="选择数据源">
            <a-select
              v-model:value="currentDatasourceId"
              placeholder="请选择数据源"
              style="width: 100%"
              @change="loadTablesWithPermission"
            >
              <a-select-option
                v-for="ds in userDatasources"
                :key="ds.id"
                :value="ds.id"
              >
                {{ ds.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
          <a-checkbox-group
            v-model:value="selectedTables"
            style="width: 100%"
            @change="handleTablePermissionChange"
          >
            <a-row>
              <a-col
                v-for="table in tablesWithPermission"
                :key="table.id"
                :span="24"
                style="margin-bottom: 8px"
              >
                <a-checkbox :value="table.id">
                  {{ table.tableAlias || table.tableName }}
                  <span style="color: #999; margin-left: 8px">
                    ({{ table.tableName }})
                  </span>
                </a-checkbox>
              </a-col>
            </a-row>
          </a-checkbox-group>
        </a-tab-pane>
      </a-tabs>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { userApi, datasourceApi, permissionApi } from '../api'

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

// 模态框相关
const modalVisible = ref(false)
const modalTitle = ref('新建用户')
const isEdit = ref(false)
const formData = reactive({
  id: 0,
  username: '',
  password: '',
  displayName: '',
  role: 'user'
})

// 重置密码相关
const resetPasswordVisible = ref(false)
const currentUserId = ref(0)
const newPassword = ref('')

// 权限管理相关
const permissionVisible = ref(false)
const permissionTab = ref('datasource')
const currentPermissionUser = ref<any>(null)
const datasourcesWithPermission = ref<any[]>([])
const selectedDatasources = ref<number[]>([])
const userDatasources = ref<any[]>([])
const currentDatasourceId = ref<number>()
const tablesWithPermission = ref<any[]>([])
const selectedTables = ref<number[]>([])

// 加载用户列表
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

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadUsers()
}

// 重置搜索
const handleReset = () => {
  searchForm.role = undefined
  searchForm.status = undefined
  searchForm.keyword = ''
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadUsers()
}

// 显示创建对话框
const showCreateModal = () => {
  modalTitle.value = '新建用户'
  isEdit.value = false
  formData.id = 0
  formData.username = ''
  formData.password = ''
  formData.displayName = ''
  formData.role = 'user'
  modalVisible.value = true
}

// 显示编辑对话框
const showEditModal = (record: any) => {
  modalTitle.value = '编辑用户'
  isEdit.value = true
  formData.id = record.id
  formData.username = record.username
  formData.displayName = record.displayName
  formData.role = record.role
  modalVisible.value = true
}

// 对话框确认
const handleModalOk = async () => {
  if (!formData.username) {
    message.error('请输入用户名')
    return
  }
  if (!isEdit.value && !formData.password) {
    message.error('请输入密码')
    return
  }
  if (!isEdit.value && formData.password.length < 6) {
    message.error('密码长度不能少于6位')
    return
  }
  if (!formData.role) {
    message.error('请选择角色')
    return
  }

  try {
    let res
    if (isEdit.value) {
      res = await userApi.updateUser(formData.id, {
        displayName: formData.displayName,
        role: formData.role
      })
    } else {
      res = await userApi.createUser({
        username: formData.username,
        password: formData.password,
        role: formData.role,
        displayName: formData.displayName
      })
    }

    if (res.code === 0) {
      message.success(isEdit.value ? '更新成功' : '创建成功')
      modalVisible.value = false
      loadUsers()
    } else {
      message.error(res.msg || '操作失败')
    }
  } catch (error) {
    message.error('操作失败')
  }
}

// 对话框取消
const handleModalCancel = () => {
  modalVisible.value = false
}

// 显示重置密码对话框
const showResetPasswordModal = (record: any) => {
  currentUserId.value = record.id
  newPassword.value = ''
  resetPasswordVisible.value = true
}

// 重置密码
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

// 切换用户状态
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

// 删除用户
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

// 显示权限管理对话框
const showPermissionModal = async (record: any) => {
  currentPermissionUser.value = record
  permissionTab.value = 'datasource'
  permissionVisible.value = true
  await loadDatasourcesWithPermission()
}

// 加载数据源权限
const loadDatasourcesWithPermission = async () => {
  try {
    const res = await permissionApi.listAllDatasourcesWithPermission(
      currentPermissionUser.value.id
    )
    if (res.code === 0) {
      datasourcesWithPermission.value = res.data || []
      selectedDatasources.value = datasourcesWithPermission.value
        .filter((ds: any) => ds.hasPermission)
        .map((ds: any) => ds.id)

      // 同时加载用户已有权限的数据源列表
      const userDsRes = await permissionApi.listUserDatasources(
        currentPermissionUser.value.id
      )
      if (userDsRes.code === 0) {
        userDatasources.value = userDsRes.data || []
      }
    }
  } catch (error) {
    message.error('加载数据源权限失败')
  }
}

// 数据源权限变化
const handleDatasourcePermissionChange = async (checkedValues: number[]) => {
  const userId = currentPermissionUser.value.id
  const oldValues = datasourcesWithPermission.value
    .filter((ds: any) => ds.hasPermission)
    .map((ds: any) => ds.id)

  // 找出新增和删除的
  const added = checkedValues.filter((id) => !oldValues.includes(id))
  const removed = oldValues.filter((id) => !checkedValues.includes(id))

  try {
    if (added.length > 0) {
      await permissionApi.grantDatasourcePermissions(userId, added)
    }
    for (const dsId of removed) {
      await permissionApi.revokeDatasourcePermission(userId, dsId)
    }
    message.success('权限更新成功')
    await loadDatasourcesWithPermission()
  } catch (error) {
    message.error('权限更新失败')
  }
}

// 加载表权限
const loadTablesWithPermission = async () => {
  if (!currentDatasourceId.value) return

  try {
    const res = await permissionApi.listAllTablesWithPermission(
      currentPermissionUser.value.id,
      currentDatasourceId.value
    )
    if (res.code === 0) {
      tablesWithPermission.value = res.data || []
      selectedTables.value = tablesWithPermission.value
        .filter((table: any) => table.hasPermission)
        .map((table: any) => table.id)
    }
  } catch (error) {
    message.error('加载表权限失败')
  }
}

// 表权限变化
const handleTablePermissionChange = async (checkedValues: number[]) => {
  const userId = currentPermissionUser.value.id
  const oldValues = tablesWithPermission.value
    .filter((table: any) => table.hasPermission)
    .map((table: any) => table.id)

  // 找出新增和删除的
  const added = checkedValues.filter((id) => !oldValues.includes(id))
  const removed = oldValues.filter((id) => !checkedValues.includes(id))

  try {
    if (added.length > 0) {
      await permissionApi.grantTablePermissions(userId, added)
    }
    for (const tableId of removed) {
      await permissionApi.revokeTablePermission(userId, tableId)
    }
    message.success('权限更新成功')
    await loadTablesWithPermission()
  } catch (error) {
    message.error('权限更新失败')
  }
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
