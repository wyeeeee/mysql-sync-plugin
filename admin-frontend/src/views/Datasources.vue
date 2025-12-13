<template>
  <div class="datasources-page">
    <div class="page-header">
      <h2 class="page-title">
        <DatabaseOutlined class="title-icon" />
        数据源管理
      </h2>
      <a-button type="primary" class="cherry-btn-sm" @click="showCreateModal">
        <PlusOutlined />
        新建数据源
      </a-button>
    </div>

    <a-card class="cherry-card">
      <!-- 搜索表单 -->
      <div class="search-form-wrapper">
        <a-row :gutter="[12, 12]">
          <a-col :xs="24" :sm="16" :md="12" :lg="10">
            <a-input
              v-model:value="searchForm.keyword"
              placeholder="搜索数据源名称或描述"
              allowClear
              class="cherry-input"
            >
              <template #prefix>
                <SearchOutlined style="color: #bfbfbf" />
              </template>
            </a-input>
          </a-col>
          <a-col :xs="24" :sm="8" :md="6" :lg="6">
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
          :data-source="datasources"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
          :scroll="{ x: 900 }"
          class="cherry-table"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'name'">
              <div class="ds-name">
                <DatabaseOutlined class="ds-icon" />
                <span>{{ record.name }}</span>
              </div>
            </template>
            <template v-else-if="column.key === 'connection'">
              <a-tooltip :title="`${record.host}:${record.port}/${record.databaseName}`">
                <code class="connection-info">
                  {{ record.host }}:{{ record.port }}/{{ record.databaseName }}
                </code>
              </a-tooltip>
            </template>
            <template v-else-if="column.key === 'action'">
              <!-- 桌面端显示按钮 -->
              <a-space v-if="!isMobile" :size="4">
                <a-button type="link" size="small" class="action-btn test-btn" @click="testConnection(record)">
                  <ApiOutlined />
                  测试
                </a-button>
                <a-button type="link" size="small" class="action-btn" @click="showEditModal(record)">
                  编辑
                </a-button>
                <a-button type="link" size="small" class="action-btn" @click="showTablesModal(record)">
                  表配置
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
                    <a-menu-item @click="testConnection(record)">测试连接</a-menu-item>
                    <a-menu-item @click="showEditModal(record)">编辑</a-menu-item>
                    <a-menu-item @click="showTablesModal(record)">表配置</a-menu-item>
                    <a-menu-item danger @click="handleDelete(record)">删除</a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </template>
          </template>
        </a-table>
      </div>
    </a-card>

    <!-- 数据源表单组件 -->
    <DatasourceForm
      v-model:open="modalVisible"
      :is-edit="isEdit"
      :datasource="currentDatasource"
      @success="loadDatasources"
    />

    <!-- 表配置组件 -->
    <TableConfigModal
      v-model:open="tablesModalVisible"
      :datasource="currentDatasource"
      @field-mapping="showFieldMappingsModal"
    />

    <!-- 字段映射组件 -->
    <FieldMappingModal
      v-model:open="fieldMappingsModalVisible"
      :table="currentTable"
      :datasource="currentDatasource"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, DatabaseOutlined, SearchOutlined, DownOutlined, ApiOutlined } from '@ant-design/icons-vue'
import { datasourceApi } from '../api'
import DatasourceForm from '../components/datasource/DatasourceForm.vue'
import TableConfigModal from '../components/datasource/TableConfigModal.vue'
import FieldMappingModal from '../components/datasource/FieldMappingModal.vue'

const loading = ref(false)
const datasources = ref<any[]>([])
const isMobile = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`
})

const searchForm = reactive({
  keyword: ''
})

const allColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: '数据源名称', dataIndex: 'name', key: 'name', width: 150 },
  { title: '描述', dataIndex: 'description', key: 'description', width: 150, ellipsis: true },
  { title: '连接信息', key: 'connection', width: 200 },
  { title: '用户名', dataIndex: 'username', key: 'username', width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '操作', key: 'action', width: 220, fixed: 'right' as const }
]

const responsiveColumns = computed(() => {
  if (isMobile.value) {
    return allColumns.filter(col => ['name', 'connection', 'action'].includes(col.key))
      .map(col => col.key === 'action' ? { ...col, width: 80, fixed: undefined } : col)
  }
  return allColumns
})

const modalVisible = ref(false)
const isEdit = ref(false)
const currentDatasource = ref<any>(null)

const tablesModalVisible = ref(false)
const fieldMappingsModalVisible = ref(false)
const currentTable = ref<any>(null)

function checkMobile() {
  isMobile.value = window.innerWidth < 768
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  loadDatasources()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

const loadDatasources = async () => {
  loading.value = true
  try {
    const res = await datasourceApi.listDatasources({
      page: pagination.current,
      pageSize: pagination.pageSize,
      ...searchForm
    })
    if (res.code === 0) {
      datasources.value = res.data.list || []
      pagination.total = res.data.total
    } else {
      message.error(res.msg || '加载数据源列表失败')
    }
  } catch (error) {
    message.error('加载数据源列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  loadDatasources()
}

const handleReset = () => {
  searchForm.keyword = ''
  handleSearch()
}

const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadDatasources()
}

const showCreateModal = () => {
  isEdit.value = false
  currentDatasource.value = null
  modalVisible.value = true
}

const showEditModal = (record: any) => {
  isEdit.value = true
  currentDatasource.value = record
  modalVisible.value = true
}

const testConnection = async (record: any) => {
  const hide = message.loading('正在测试连接...', 0)
  try {
    const res = await datasourceApi.testConnection(record.id)
    hide()
    if (res.code === 0) {
      message.success('连接测试成功')
    } else {
      message.error(res.msg || '连接测试失败')
    }
  } catch (error) {
    hide()
    message.error('连接测试失败')
  }
}

const handleDelete = (record: any) => {
  Modal.confirm({
    title: '确认删除?',
    content: `确定要删除数据源 "${record.name}" 吗? 此操作不可恢复。`,
    okText: '确认',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        const res = await datasourceApi.deleteDatasource(record.id)
        if (res.code === 0) {
          message.success('删除成功')
          loadDatasources()
        } else {
          message.error(res.msg || '删除失败')
        }
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

const showTablesModal = (record: any) => {
  currentDatasource.value = record
  tablesModalVisible.value = true
}

const showFieldMappingsModal = (table: any) => {
  currentTable.value = table
  fieldMappingsModalVisible.value = true
}
</script>

<style scoped>
.datasources-page {
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

.ds-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.ds-icon {
  color: #1e3a5f;
  font-size: 16px;
}

.connection-info {
  background: #f5f5f5;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
  max-width: 180px;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-btn {
  padding: 0 6px;
  font-size: 13px;
  color: #1e3a5f;
}

.action-btn:hover {
  color: #2d4a6f;
}

.action-btn.test-btn {
  color: #27ae60;
}

.action-btn.test-btn:hover {
  color: #1e8449;
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

  .connection-info {
    max-width: 120px;
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

  .connection-info {
    max-width: 100px;
    font-size: 11px;
  }
}
</style>
