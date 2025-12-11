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

    <!-- 创建/编辑数据源对话框 -->
    <a-modal
      v-model:open="modalVisible"
      :title="modalTitle"
      width="600px"
      @ok="handleModalOk"
      @cancel="handleModalCancel"
    >
      <a-form :model="formData" :label-col="{ span: 6 }">
        <a-form-item label="数据源名称" required>
          <a-input v-model:value="formData.name" placeholder="请输入数据源名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model:value="formData.description"
            placeholder="请输入描述"
            :rows="2"
          />
        </a-form-item>
        <a-form-item label="主机地址" required>
          <a-input v-model:value="formData.host" placeholder="例如: localhost" />
        </a-form-item>
        <a-form-item label="端口" required>
          <a-input-number
            v-model:value="formData.port"
            :min="1"
            :max="65535"
            style="width: 100%"
            placeholder="例如: 3306"
          />
        </a-form-item>
        <a-form-item label="数据库名" required>
          <a-select
            v-model:value="formData.databaseName"
            placeholder="请选择数据库"
            show-search
            :loading="databasesLoading"
            :disabled="!formData.host || !formData.port || !formData.username || !formData.password"
            @focus="loadDatabasesForForm"
          >
            <a-select-option v-for="db in availableDatabases" :key="db" :value="db">
              {{ db }}
            </a-select-option>
          </a-select>
          <div style="margin-top: 4px; color: #999; font-size: 12px">
            请先填写主机、端口、用户名和密码,然后点击此处加载数据库列表
          </div>
        </a-form-item>
        <a-form-item label="用户名" required>
          <a-input v-model:value="formData.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item label="密码" :required="!isEdit">
          <a-input-password
            v-model:value="formData.password"
            :placeholder="isEdit ? '留空则不修改密码' : '请输入密码'"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 表配置对话框 -->
    <a-modal
      v-model:open="tablesModalVisible"
      :title="`表配置 - ${currentDatasource?.name}`"
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
        :loading="tablesLoading"
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
              <a-button type="link" size="small" @click="showEditTableModal(record)">
                编辑
              </a-button>
              <a-button
                type="link"
                size="small"
                @click="showFieldMappingsModal(record)"
              >
                字段映射
              </a-button>
              <a-button
                type="link"
                size="small"
                danger
                @click="handleDeleteTable(record)"
              >
                删除
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-modal>

    <!-- 添加表配置对话框 -->
    <a-modal
      v-model:open="tableModalVisible"
      :title="tableModalTitle"
      width="700px"
      @ok="handleTableModalOk"
      @cancel="handleTableModalCancel"
    >
      <a-form :model="tableFormData" :label-col="{ span: 6 }">
        <a-form-item label="查询模式" required>
          <a-radio-group v-model:value="tableFormData.queryMode" @change="handleQueryModeChange">
            <a-radio-button value="table">从数据库选择表</a-radio-button>
            <a-radio-button value="sql">自定义SQL查询</a-radio-button>
          </a-radio-group>
        </a-form-item>

        <!-- 表查询模式 -->
        <template v-if="tableFormData.queryMode === 'table'">
          <a-form-item label="选择表" required>
            <a-spin :spinning="tablesLoadingForBatch">
              <div style="margin-bottom: 8px">
                <a-checkbox
                  :indeterminate="selectedTableNames.length > 0 && selectedTableNames.length < availableTables.length"
                  :checked="selectedTableNames.length === availableTables.length && availableTables.length > 0"
                  @change="handleSelectAllTables"
                >
                  全选
                </a-checkbox>
              </div>
              <div style="max-height: 300px; overflow-y: auto; border: 1px solid #d9d9d9; border-radius: 4px; padding: 8px">
                <a-checkbox-group
                  v-model:value="selectedTableNames"
                  style="width: 100%"
                >
                  <a-row>
                    <a-col
                      v-for="table in availableTables"
                      :key="table"
                      :span="24"
                      style="margin-bottom: 8px"
                    >
                      <a-checkbox :value="table">
                        {{ table }}
                      </a-checkbox>
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

        <!-- SQL查询模式 -->
        <template v-if="tableFormData.queryMode === 'sql'">
          <a-form-item label="表名" required>
            <a-input
              v-model:value="tableFormData.tableName"
              placeholder="请输入表名(用于标识)"
            />
          </a-form-item>
          <a-form-item label="表别名">
            <a-input
              v-model:value="tableFormData.tableAlias"
              placeholder="请输入表别名(用于前端显示)"
            />
          </a-form-item>
          <a-form-item label="自定义SQL" required>
            <a-textarea
              v-model:value="tableFormData.customSql"
              placeholder="请输入SQL语句"
              :rows="6"
            />
          </a-form-item>
        </template>
      </a-form>
    </a-modal>

    <!-- 编辑单表配置对话框 -->
    <a-modal
      v-model:open="editSingleTableModalVisible"
      title="编辑表配置"
      width="600px"
      @ok="handleEditSingleTableOk"
      @cancel="handleEditSingleTableCancel"
    >
      <a-form :model="editSingleTableFormData" :label-col="{ span: 6 }">
        <a-form-item label="表名">
          <a-input
            v-model:value="editSingleTableFormData.tableName"
            disabled
            placeholder="表名不可修改"
          />
        </a-form-item>
        <a-form-item label="表别名">
          <a-input
            v-model:value="editSingleTableFormData.tableAlias"
            placeholder="请输入表别名(用于前端显示)"
          />
        </a-form-item>
        <a-form-item label="查询模式">
          <a-tag :color="editSingleTableFormData.queryMode === 'table' ? 'blue' : 'green'">
            {{ editSingleTableFormData.queryMode === 'table' ? '表查询' : 'SQL查询' }}
          </a-tag>
        </a-form-item>
        <a-form-item v-if="editSingleTableFormData.queryMode === 'sql'" label="自定义SQL">
          <a-textarea
            v-model:value="editSingleTableFormData.customSql"
            placeholder="请输入SQL语句"
            :rows="6"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 字段映射对话框 -->
    <a-modal
      v-model:open="fieldMappingsModalVisible"
      :title="`字段映射 - ${currentTable?.tableAlias || currentTable?.tableName}`"
      width="900px"
      @ok="handleFieldMappingsOk"
      @cancel="handleFieldMappingsCancel"
    >
      <div style="margin-bottom: 16px">
        <a-space>
          <a-button type="primary" size="small" @click="loadFieldsFromDatabase">
            <template #icon><PlusOutlined /></template>
            从数据库加载字段
          </a-button>
          <a-button size="small" @click="applyAllComments" v-if="hasAnyComment">
            一键应用备注
          </a-button>
          <a-button size="small" @click="addFieldMapping">
            手动添加映射
          </a-button>
        </a-space>
      </div>

      <a-table
        :columns="fieldMappingColumns"
        :data-source="fieldMappings"
        :pagination="false"
        row-key="index"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'fieldName'">
            <a-input
              v-model:value="record.mysqlField"
              placeholder="MySQL字段名"
              :disabled="record.fromDatabase"
            />
          </template>
          <template v-else-if="column.key === 'comment'">
            <span style="color: #666">{{ record.comment || '-' }}</span>
          </template>
          <template v-else-if="column.key === 'fieldAlias'">
            <a-input v-model:value="record.aliasField" placeholder="显示别名" />
          </template>
          <template v-else-if="column.key === 'action'">
            <a-space>
              <a-button
                v-if="record.comment"
                type="link"
                size="small"
                @click="applyComment(index)"
              >
                应用备注
              </a-button>
              <a-button
                type="link"
                size="small"
                danger
                @click="removeFieldMapping(index)"
              >
                删除
              </a-button>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import { datasourceApi } from '../api'

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

// 数据库列表相关
const availableDatabases = ref<string[]>([])
const databasesLoading = ref(false)
const tempDatasourceId = ref<number>()

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '数据源名称', dataIndex: 'name', key: 'name' },
  { title: '描述', dataIndex: 'description', key: 'description' },
  { title: '连接信息', key: 'connection' },
  { title: '用户名', dataIndex: 'username', key: 'username' },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt' },
  { title: '操作', key: 'action', width: 280 }
]

// 模态框相关
const modalVisible = ref(false)
const modalTitle = ref('新建数据源')
const isEdit = ref(false)
const formData = reactive({
  id: 0,
  name: '',
  description: '',
  host: '',
  port: 3306,
  databaseName: '',
  username: '',
  password: ''
})

// 表配置相关
const tablesModalVisible = ref(false)
const currentDatasource = ref<any>(null)
const tables = ref<any[]>([])
const tablesLoading = ref(false)

const tableColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '表名', dataIndex: 'tableName', key: 'tableName' },
  { title: '表别名', dataIndex: 'tableAlias', key: 'tableAlias' },
  { title: '查询模式', key: 'queryMode' },
  { title: '操作', key: 'action', width: 220 }
]

// 添加表配置
const tableModalVisible = ref(false)
const tableModalTitle = ref('添加表配置')
const tableFormData = reactive({
  id: 0,
  tableName: '',
  tableAlias: '',
  queryMode: 'table',
  customSql: ''
})

// 编辑单表配置
const editSingleTableModalVisible = ref(false)
const editSingleTableFormData = reactive({
  id: 0,
  tableName: '',
  tableAlias: '',
  queryMode: 'table',
  customSql: ''
})

// 可用表列表(用于批量选择)
const availableTables = ref<string[]>([])
const selectedTableNames = ref<string[]>([])
const tablesLoadingForBatch = ref(false)

// 字段映射
const fieldMappingsModalVisible = ref(false)
const currentTable = ref<any>(null)
const fieldMappings = ref<any[]>([])
const hasAnyComment = ref(false)

const fieldMappingColumns = [
  { title: 'MySQL字段名', key: 'fieldName', width: '25%' },
  { title: '数据库备注', key: 'comment', width: '25%' },
  { title: '显示别名', key: 'fieldAlias', width: '30%' },
  { title: '操作', key: 'action', width: '20%' }
]

// 加载数据源列表
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

// 搜索
const handleSearch = () => {
  pagination.current = 1
  loadDatasources()
}

// 重置搜索
const handleReset = () => {
  searchForm.keyword = ''
  handleSearch()
}

// 表格变化
const handleTableChange = (pag: any) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadDatasources()
}

// 显示创建对话框
const showCreateModal = () => {
  modalTitle.value = '新建数据源'
  isEdit.value = false
  formData.id = 0
  formData.name = ''
  formData.description = ''
  formData.host = ''
  formData.port = 3306
  formData.databaseName = ''
  formData.username = ''
  formData.password = ''
  availableDatabases.value = []
  tempDatasourceId.value = undefined
  modalVisible.value = true
}

// 显示编辑对话框
const showEditModal = (record: any) => {
  modalTitle.value = '编辑数据源'
  isEdit.value = true
  formData.id = record.id
  formData.name = record.name
  formData.description = record.description
  formData.host = record.host
  formData.port = record.port
  formData.databaseName = record.databaseName
  formData.username = record.username
  formData.password = ''
  availableDatabases.value = [record.databaseName]
  tempDatasourceId.value = record.id
  modalVisible.value = true
}

// 对话框确认
const handleModalOk = async () => {
  if (!formData.name || !formData.host || !formData.port || !formData.databaseName || !formData.username) {
    message.error('请填写所有必填项')
    return
  }
  if (!isEdit.value && !formData.password) {
    message.error('请输入密码')
    return
  }

  try {
    let res
    if (isEdit.value) {
      res = await datasourceApi.updateDatasource(formData.id, {
        name: formData.name,
        description: formData.description,
        host: formData.host,
        port: formData.port,
        databaseName: formData.databaseName,
        username: formData.username,
        password: formData.password || undefined
      })
    } else {
      res = await datasourceApi.createDatasource({
        name: formData.name,
        description: formData.description,
        host: formData.host,
        port: formData.port,
        databaseName: formData.databaseName,
        username: formData.username,
        password: formData.password
      })
    }

    if (res.code === 0) {
      message.success(isEdit.value ? '更新成功' : '创建成功')
      await cleanupTempDatasource()
      modalVisible.value = false
      loadDatasources()
    } else {
      message.error(res.msg || '操作失败')
    }
  } catch (error) {
    message.error('操作失败')
  }
}

// 对话框取消
const handleModalCancel = async () => {
  await cleanupTempDatasource()
  modalVisible.value = false
}

// 测试连接
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

// 删除数据源
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

// 显示表配置对话框
const showTablesModal = async (record: any) => {
  currentDatasource.value = record
  tablesModalVisible.value = true
  await loadTables()
}

// 加载表配置列表
const loadTables = async () => {
  if (!currentDatasource.value) return

  tablesLoading.value = true
  try {
    const res = await datasourceApi.listTables(currentDatasource.value.id)
    if (res.code === 0) {
      tables.value = res.data || []
    } else {
      message.error(res.msg || '加载表配置失败')
    }
  } catch (error) {
    message.error('加载表配置失败')
  } finally {
    tablesLoading.value = false
  }
}

// 显示添加表配置对话框
const showAddTableModal = async () => {
  tableModalTitle.value = '添加表配置'
  tableFormData.id = 0
  tableFormData.tableName = ''
  tableFormData.tableAlias = ''
  tableFormData.queryMode = 'table'
  tableFormData.customSql = ''
  selectedTableNames.value = []
  tableModalVisible.value = true

  // 如果是表查询模式,加载可用表列表
  if (tableFormData.queryMode === 'table') {
    await loadAvailableTables()
  }
}

// 查询模式变化
const handleQueryModeChange = async () => {
  if (tableFormData.queryMode === 'table') {
    // 切换到表查询模式,加载表列表
    await loadAvailableTables()
    tableFormData.tableName = ''
    tableFormData.tableAlias = ''
    tableFormData.customSql = ''
  } else {
    // 切换到SQL模式,清空选择
    selectedTableNames.value = []
  }
}

// 全选/取消全选表
const handleSelectAllTables = (e: any) => {
  if (e.target.checked) {
    selectedTableNames.value = [...availableTables.value]
  } else {
    selectedTableNames.value = []
  }
}

// 加载可用的表列表
const loadAvailableTables = async () => {
  if (!currentDatasource.value) return

  tablesLoadingForBatch.value = true
  try {
    const res = await datasourceApi.getTableList(
      currentDatasource.value.id,
      currentDatasource.value.databaseName
    )
    if (res.code === 0) {
      availableTables.value = res.data || []
    } else {
      message.error(res.msg || '获取表列表失败')
    }
  } catch (error) {
    message.error('获取表列表失败')
  } finally {
    tablesLoadingForBatch.value = false
  }
}

// 显示编辑单表配置对话框
const showEditTableModal = (record: any) => {
  editSingleTableFormData.id = record.id
  editSingleTableFormData.tableName = record.tableName
  editSingleTableFormData.tableAlias = record.tableAlias || record.tableName
  editSingleTableFormData.queryMode = record.queryMode
  editSingleTableFormData.customSql = record.customSql || ''
  editSingleTableModalVisible.value = true
}

// 添加表配置对话框确认
const handleTableModalOk = async () => {
  // 验证
  if (tableFormData.queryMode === 'table') {
    if (selectedTableNames.value.length === 0) {
      message.error('请至少选择一个表')
      return
    }
  } else if (tableFormData.queryMode === 'sql') {
    if (!tableFormData.tableName) {
      message.error('请输入表名')
      return
    }
    if (!tableFormData.customSql) {
      message.error('请输入自定义SQL')
      return
    }
  }

  try {
    // 新建模式
    if (tableFormData.queryMode === 'table') {
      // 批量创建表配置
      for (const tableName of selectedTableNames.value) {
        await datasourceApi.createTable(currentDatasource.value.id, {
          tableName,
          tableAlias: tableName,
          queryMode: 'table'
        })
      }
      message.success(`成功添加 ${selectedTableNames.value.length} 个表配置`)
    } else {
      // 创建SQL查询配置
      const res = await datasourceApi.createTable(currentDatasource.value.id, {
        tableName: tableFormData.tableName,
        tableAlias: tableFormData.tableAlias,
        queryMode: 'sql',
        customSql: tableFormData.customSql
      })
      if (res.code === 0) {
        message.success('创建成功')
      } else {
        message.error(res.msg || '创建失败')
        return
      }
    }
    tableModalVisible.value = false
    loadTables()
  } catch (error) {
    message.error('操作失败')
  }
}

// 添加表配置对话框取消
const handleTableModalCancel = () => {
  tableModalVisible.value = false
}

// 编辑单表配置对话框确认
const handleEditSingleTableOk = async () => {
  // 验证
  if (!editSingleTableFormData.tableAlias) {
    message.error('请输入表别名')
    return
  }

  if (editSingleTableFormData.queryMode === 'sql' && !editSingleTableFormData.customSql) {
    message.error('请输入自定义SQL')
    return
  }

  try {
    const res = await datasourceApi.updateTable(editSingleTableFormData.id, {
      tableAlias: editSingleTableFormData.tableAlias,
      queryMode: editSingleTableFormData.queryMode,
      customSql: editSingleTableFormData.customSql
    })
    if (res.code === 0) {
      message.success('更新成功')
      editSingleTableModalVisible.value = false
      loadTables()
    } else {
      message.error(res.msg || '更新失败')
    }
  } catch (error) {
    message.error('操作失败')
  }
}

// 编辑单表配置对话框取消
const handleEditSingleTableCancel = () => {
  editSingleTableModalVisible.value = false
}

// 删除表配置
const handleDeleteTable = (record: any) => {
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
          loadTables()
        } else {
          message.error(res.msg || '删除失败')
        }
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

// 显示字段映射对话框
const showFieldMappingsModal = async (record: any) => {
  currentTable.value = record
  fieldMappingsModalVisible.value = true
  await loadFieldMappings()
}

// 加载字段映射
const loadFieldMappings = async () => {
  if (!currentTable.value) return

  try {
    const res = await datasourceApi.getFieldMappings(currentTable.value.id)
    if (res.code === 0) {
      fieldMappings.value = (res.data || []).map((item: any, index: number) => ({
        index,
        mysqlField: item.fieldName,
        aliasField: item.fieldAlias,
        comment: '',
        fromDatabase: false
      }))
      hasAnyComment.value = false
    }
  } catch (error) {
    message.error('加载字段映射失败')
  }
}

// 从数据库加载字段
const loadFieldsFromDatabase = async () => {
  if (!currentTable.value || !currentDatasource.value) return

  try {
    let res
    // 根据查询模式调用不同的API
    if (currentTable.value.queryMode === 'sql') {
      // SQL查询模式:使用自定义SQL获取字段
      if (!currentTable.value.customSql) {
        message.error('自定义SQL为空,无法加载字段')
        return
      }
      res = await datasourceApi.getFieldListFromSQL(
        currentDatasource.value.id,
        currentDatasource.value.databaseName,
        currentTable.value.customSql
      )
    } else {
      // 表查询模式:使用表名获取字段
      res = await datasourceApi.getFieldList(
        currentDatasource.value.id,
        currentDatasource.value.databaseName,
        currentTable.value.tableName
      )
    }

    if (res.code === 0) {
      const fields = res.data || []
      fieldMappings.value = fields.map((field: any, index: number) => ({
        index,
        mysqlField: field.name,
        aliasField: field.name,
        comment: field.comment || '',
        fromDatabase: true
      }))
      hasAnyComment.value = fields.some((f: any) => f.comment)
      message.success(`成功加载 ${fields.length} 个字段`)
    } else {
      message.error(res.msg || '加载字段失败')
    }
  } catch (error) {
    message.error('加载字段失败')
  }
}

// 应用单个字段的备注
const applyComment = (index: number) => {
  const field = fieldMappings.value[index]
  if (field.comment) {
    field.aliasField = field.comment
  }
}

// 一键应用所有备注
const applyAllComments = () => {
  fieldMappings.value.forEach(field => {
    if (field.comment) {
      field.aliasField = field.comment
    }
  })
  message.success('已应用所有备注')
}

// 添加字段映射
const addFieldMapping = () => {
  fieldMappings.value.push({
    index: fieldMappings.value.length,
    mysqlField: '',
    aliasField: '',
    comment: '',
    fromDatabase: false
  })
}

// 删除字段映射
const removeFieldMapping = (index: number) => {
  fieldMappings.value.splice(index, 1)
  // 重新设置索引
  fieldMappings.value.forEach((item, idx) => {
    item.index = idx
  })
}

// 字段映射对话框确认
const handleFieldMappingsOk = async () => {
  // 过滤掉空的映射
  const validMappings = fieldMappings.value.filter(
    (item) => item.mysqlField && item.aliasField
  )

  try {
    const res = await datasourceApi.updateFieldMappings(
      currentTable.value.id,
      validMappings
    )
    if (res.code === 0) {
      message.success('更新成功')
      fieldMappingsModalVisible.value = false
    } else {
      message.error(res.msg || '更新失败')
    }
  } catch (error) {
    message.error('更新失败')
  }
}

// 字段映射对话框取消
const handleFieldMappingsCancel = () => {
  fieldMappingsModalVisible.value = false
}

// 加载数据库列表(用于表单)
const loadDatabasesForForm = async () => {
  // 如果是编辑模式且已有数据源ID,直接加载
  if (isEdit.value && tempDatasourceId.value) {
    await loadDatabasesFromDatasource(tempDatasourceId.value)
    return
  }

  // 如果是新建模式,需要先验证连接信息
  if (!formData.host || !formData.port || !formData.username || !formData.password) {
    message.warning('请先填写主机、端口、用户名和密码')
    return
  }

  // 创建临时数据源以获取数据库列表
  databasesLoading.value = true
  try {
    const res = await datasourceApi.createDatasource({
      name: '__temp__',
      host: formData.host,
      port: formData.port,
      databaseName: 'mysql',
      username: formData.username,
      password: formData.password
    })
    if (res.code === 0) {
      tempDatasourceId.value = res.data.id
      await loadDatabasesFromDatasource(res.data.id)
    } else {
      message.error(res.msg || '连接失败')
    }
  } catch (error) {
    message.error('连接失败')
  } finally {
    databasesLoading.value = false
  }
}

// 从数据源加载数据库列表
const loadDatabasesFromDatasource = async (datasourceId: number) => {
  databasesLoading.value = true
  try {
    const res = await datasourceApi.getDatabaseList(datasourceId)
    if (res.code === 0) {
      availableDatabases.value = res.data || []
      if (availableDatabases.value.length === 0) {
        message.warning('未找到可用的数据库')
      }
    } else {
      message.error(res.msg || '获取数据库列表失败')
    }
  } catch (error) {
    message.error('获取数据库列表失败')
  } finally {
    databasesLoading.value = false
  }
}

// 对话框取消时清理临时数据源
const cleanupTempDatasource = async () => {
  if (!isEdit.value && tempDatasourceId.value) {
    try {
      await datasourceApi.deleteDatasource(tempDatasourceId.value)
    } catch (error) {
      // 忽略错误
    }
    tempDatasourceId.value = undefined
  }
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
