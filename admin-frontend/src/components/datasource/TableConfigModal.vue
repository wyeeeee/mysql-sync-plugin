<template>
  <a-modal
    v-model:open="visible"
    :title="`表配置 - ${datasource?.name}`"
    width="900px"
    :footer="null"
  >
    <div style="margin-bottom: 16px">
      <a-button type="primary" @click="showAddTableModal">
        <template #icon><PlusOutlined /></template>
        添加表配置
      </a-button>
    </div>

    <a-table
      :columns="tableColumns"
      :data-source="tables"
      :loading="loading"
      row-key="id"
      :pagination="false"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'queryMode'">
          <a-tag :color="record.queryMode === 'table' ? 'blue' : 'green'">
            {{ record.queryMode === 'table' ? '表查询' : 'SQL查询' }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'action'">
          <a-space>
            <a-button type="link" size="small" @click="handleEdit(record)">
              编辑
            </a-button>
            <a-button type="link" size="small" @click="handleFieldMapping(record)">
              字段映射
            </a-button>
            <a-button type="link" size="small" danger @click="handleDelete(record)">
              删除
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- 添加/编辑表配置对话框 -->
    <a-modal
      v-model:open="tableModalVisible"
      :title="tableModalTitle"
      width="700px"
      @ok="handleTableModalOk"
      @cancel="tableModalVisible = false"
    >
      <a-form :model="tableFormData" :label-col="{ span: 6 }">
        <a-form-item label="查询模式" required>
          <a-radio-group v-model:value="tableFormData.queryMode" @change="handleQueryModeChange">
            <a-radio-button value="table">从数据库选择表</a-radio-button>
            <a-radio-button value="sql">自定义SQL查询</a-radio-button>
          </a-radio-group>
        </a-form-item>

        <template v-if="tableFormData.queryMode === 'table'">
          <a-form-item label="选择表" required>
            <a-spin :spinning="tablesLoading">
              <div style="margin-bottom: 8px">
                <a-checkbox
                  :indeterminate="selectedTableNames.length > 0 && selectedTableNames.length < availableTables.length"
                  :checked="selectedTableNames.length === availableTables.length && availableTables.length > 0"
                  @change="handleSelectAll"
                >
                  全选
                </a-checkbox>
              </div>
              <div style="max-height: 300px; overflow-y: auto; border: 1px solid #d9d9d9; border-radius: 4px; padding: 8px">
                <a-checkbox-group v-model:value="selectedTableNames" style="width: 100%">
                  <a-row>
                    <a-col v-for="table in availableTables" :key="table" :span="24" style="margin-bottom: 8px">
                      <a-checkbox :value="table">{{ table }}</a-checkbox>
                    </a-col>
                  </a-row>
                </a-checkbox-group>
                <div v-if="availableTables.length === 0" style="text-align: center; padding: 20px; color: #999">
                  暂无可用的表
                </div>
              </div>
            </a-spin>
            <div style="margin-top: 8px; color: #999; font-size: 12px">
              已选择 {{ selectedTableNames.length }} 个表
            </div>
          </a-form-item>
        </template>

        <template v-if="tableFormData.queryMode === 'sql'">
          <a-form-item label="表名" required>
            <a-input v-model:value="tableFormData.tableName" placeholder="请输入表名(用于标识)" />
          </a-form-item>
          <a-form-item label="表别名">
            <a-input v-model:value="tableFormData.tableAlias" placeholder="请输入表别名(用于前端显示)" />
          </a-form-item>
          <a-form-item label="自定义SQL" required>
            <a-textarea v-model:value="tableFormData.customSql" placeholder="请输入SQL语句" :rows="6" />
          </a-form-item>
        </template>
      </a-form>
    </a-modal>

    <!-- 编辑单表配置对话框 -->
    <a-modal
      v-model:open="editModalVisible"
      title="编辑表配置"
      width="600px"
      @ok="handleEditOk"
      @cancel="editModalVisible = false"
    >
      <a-form :model="editFormData" :label-col="{ span: 6 }">
        <a-form-item label="表名">
          <a-input v-model:value="editFormData.tableName" disabled placeholder="表名不可修改" />
        </a-form-item>
        <a-form-item label="表别名">
          <a-input v-model:value="editFormData.tableAlias" placeholder="请输入表别名(用于前端显示)" />
        </a-form-item>
        <a-form-item label="查询模式">
          <a-tag :color="editFormData.queryMode === 'table' ? 'blue' : 'green'">
            {{ editFormData.queryMode === 'table' ? '表查询' : 'SQL查询' }}
          </a-tag>
        </a-form-item>
        <a-form-item v-if="editFormData.queryMode === 'sql'" label="自定义SQL">
          <a-textarea v-model:value="editFormData.customSql" placeholder="请输入SQL语句" :rows="6" />
        </a-form-item>
      </a-form>
    </a-modal>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { datasourceApi } from '../../api'

const props = defineProps<{
  open: boolean
  datasource?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'fieldMapping', table: any): void
}>()

const visible = ref(props.open)
const loading = ref(false)
const tables = ref<any[]>([])

const tableColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '表名', dataIndex: 'tableName', key: 'tableName' },
  { title: '表别名', dataIndex: 'tableAlias', key: 'tableAlias' },
  { title: '查询模式', key: 'queryMode' },
  { title: '操作', key: 'action', width: 220 }
]

const tableModalVisible = ref(false)
const tableModalTitle = ref('添加表配置')
const tableFormData = ref({
  queryMode: 'table',
  tableName: '',
  tableAlias: '',
  customSql: ''
})

const editModalVisible = ref(false)
const editFormData = ref({
  id: 0,
  tableName: '',
  tableAlias: '',
  queryMode: 'table',
  customSql: ''
})

const availableTables = ref<string[]>([])
const selectedTableNames = ref<string[]>([])
const tablesLoading = ref(false)

watch(() => props.open, async (val) => {
  visible.value = val
  if (val && props.datasource) {
    await loadTables()
  }
})

watch(visible, (val) => {
  emit('update:open', val)
})

const loadTables = async () => {
  if (!props.datasource) return

  loading.value = true
  try {
    const res = await datasourceApi.listTables(props.datasource.id)
    if (res.code === 0) {
      tables.value = res.data || []
    } else {
      message.error(res.msg || '加载表配置失败')
    }
  } catch (error) {
    message.error('加载表配置失败')
  } finally {
    loading.value = false
  }
}

const showAddTableModal = async () => {
  tableModalTitle.value = '添加表配置'
  tableFormData.value = {
    queryMode: 'table',
    tableName: '',
    tableAlias: '',
    customSql: ''
  }
  selectedTableNames.value = []
  tableModalVisible.value = true
  await loadAvailableTables()
}

const loadAvailableTables = async () => {
  if (!props.datasource) return

  tablesLoading.value = true
  try {
    const res = await datasourceApi.getTableList(props.datasource.id, props.datasource.databaseName)
    if (res.code === 0) {
      availableTables.value = res.data || []
    } else {
      message.error(res.msg || '获取表列表失败')
    }
  } catch (error) {
    message.error('获取表列表失败')
  } finally {
    tablesLoading.value = false
  }
}

const handleQueryModeChange = async () => {
  if (tableFormData.value.queryMode === 'table') {
    await loadAvailableTables()
    tableFormData.value.tableName = ''
    tableFormData.value.tableAlias = ''
    tableFormData.value.customSql = ''
  } else {
    selectedTableNames.value = []
  }
}

const handleSelectAll = (e: any) => {
  if (e.target.checked) {
    selectedTableNames.value = [...availableTables.value]
  } else {
    selectedTableNames.value = []
  }
}

const handleTableModalOk = async () => {
  if (tableFormData.value.queryMode === 'table') {
    if (selectedTableNames.value.length === 0) {
      message.error('请至少选择一个表')
      return
    }
  } else {
    if (!tableFormData.value.tableName) {
      message.error('请输入表名')
      return
    }
    if (!tableFormData.value.customSql) {
      message.error('请输入自定义SQL')
      return
    }
  }

  try {
    if (tableFormData.value.queryMode === 'table') {
      for (const tableName of selectedTableNames.value) {
        await datasourceApi.createTable(props.datasource.id, {
          tableName,
          tableAlias: tableName,
          queryMode: 'table'
        })
      }
      message.success(`成功添加 ${selectedTableNames.value.length} 个表配置`)
    } else {
      const res = await datasourceApi.createTable(props.datasource.id, {
        tableName: tableFormData.value.tableName,
        tableAlias: tableFormData.value.tableAlias,
        queryMode: 'sql',
        customSql: tableFormData.value.customSql
      })
      if (res.code === 0) {
        message.success('创建成功')
      } else {
        message.error(res.msg || '创建失败')
        return
      }
    }
    tableModalVisible.value = false
    await loadTables()
  } catch (error) {
    message.error('操作失败')
  }
}

const handleEdit = (record: any) => {
  editFormData.value = {
    id: record.id,
    tableName: record.tableName,
    tableAlias: record.tableAlias || record.tableName,
    queryMode: record.queryMode,
    customSql: record.customSql || ''
  }
  editModalVisible.value = true
}

const handleEditOk = async () => {
  if (!editFormData.value.tableAlias) {
    message.error('请输入表别名')
    return
  }

  if (editFormData.value.queryMode === 'sql' && !editFormData.value.customSql) {
    message.error('请输入自定义SQL')
    return
  }

  try {
    const res = await datasourceApi.updateTable(editFormData.value.id, {
      tableAlias: editFormData.value.tableAlias,
      queryMode: editFormData.value.queryMode,
      customSql: editFormData.value.customSql
    })
    if (res.code === 0) {
      message.success('更新成功')
      editModalVisible.value = false
      await loadTables()
    } else {
      message.error(res.msg || '更新失败')
    }
  } catch (error) {
    message.error('操作失败')
  }
}

const handleDelete = (record: any) => {
  Modal.confirm({
    title: '确认删除?',
    content: `确定要删除表配置 "${record.tableName}" 吗? 此操作不可恢复。`,
    okText: '确认',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        const res = await datasourceApi.deleteTable(record.id)
        if (res.code === 0) {
          message.success('删除成功')
          await loadTables()
        } else {
          message.error(res.msg || '删除失败')
        }
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

const handleFieldMapping = (record: any) => {
  emit('fieldMapping', record)
}
</script>
