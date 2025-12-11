<template>
  <div class="datasources-page">
    <div class="page-header">
      <h2>数据源管理</h2>
      <a-button type="primary" @click="showCreateModal">
        <template #icon><PlusOutlined /></template>
        新建数据源
      </a-button>
    </div>

    <a-card>
      <a-form layout="inline" :model="searchForm" class="search-form">
        <a-form-item label="关键词">
          <a-input
            v-model:value="searchForm.keyword"
            placeholder="数据源名称或描述"
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
        :data-source="datasources"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'connection'">
            {{ record.host }}:{{ record.port }}/{{ record.databaseName }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button type="link" size="small" @click="testConnection(record)">
                测试连接
              </a-button>
              <a-button type="link" size="small" @click="showEditModal(record)">
                编辑
              </a-button>
              <a-button type="link" size="small" @click="showTablesModal(record)">
                表配置
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
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { datasourceApi } from '../api'
import DatasourceForm from '../components/datasource/DatasourceForm.vue'
import TableConfigModal from '../components/datasource/TableConfigModal.vue'
import FieldMappingModal from '../components/datasource/FieldMappingModal.vue'

const loading = ref(false)
const datasources = ref<any[]>([])
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0
})

const searchForm = reactive({
  keyword: ''
})

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '数据源名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '连接信息', key: 'connection' },
  { title: '用户名', dataIndex: 'username', key: 'username' },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt' },
  { title: '操作', key: 'action', width: 280 }
]

const modalVisible = ref(false)
const isEdit = ref(false)
const currentDatasource = ref<any>(null)

const tablesModalVisible = ref(false)
const fieldMappingsModalVisible = ref(false)
const currentTable = ref<any>(null)

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
  try {
    const res = await datasourceApi.testConnection(record.id)
    if (res.code === 0) {
      message.success('连接测试成功')
    } else {
      message.error(res.msg || '连接测试失败')
    }
  } catch (error) {
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

onMounted(() => {
  loadDatasources()
})
</script>

<style scoped>
.datasources-page {
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
