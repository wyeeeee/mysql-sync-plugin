<template>
  <div class="logs-page">
    <h2>日志管理</h2>

    <!-- 搜索筛选 -->
    <a-card class="filter-card">
      <a-form layout="inline" :model="filters">
        <a-form-item label="日志级别">
          <a-select
            v-model:value="filters.level"
            placeholder="全部"
            style="width: 120px"
            allowClear
          >
            <a-select-option value="DEBUG">DEBUG</a-select-option>
            <a-select-option value="INFO">INFO</a-select-option>
            <a-select-option value="WARN">WARN</a-select-option>
            <a-select-option value="ERROR">ERROR</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="模块">
          <a-input
            v-model:value="filters.module"
            placeholder="模块名称"
            style="width: 120px"
            allowClear
          />
        </a-form-item>
        <a-form-item label="关键词">
          <a-input
            v-model:value="filters.keyword"
            placeholder="搜索内容"
            style="width: 160px"
            allowClear
          />
        </a-form-item>
        <a-form-item label="时间范围">
          <a-range-picker
            v-model:value="dateRange"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
          />
        </a-form-item>
        <a-form-item>
          <a-space>
            <a-button type="primary" @click="handleSearch">
              <SearchOutlined /> 搜索
            </a-button>
            <a-button @click="handleReset">重置</a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 日志表格 -->
    <a-table
      :columns="columns"
      :data-source="logs"
      :loading="loading"
      :pagination="pagination"
      @change="handleTableChange"
      row-key="id"
      style="margin-top: 16px"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'level'">
          <a-tag :color="getLevelColor(record.level)">
            {{ record.level }}
          </a-tag>
        </template>
        <template v-if="column.key === 'createdAt'">
          {{ formatTime(record.createdAt) }}
        </template>
        <template v-if="column.key === 'detail_action'">
          <a-button type="link" size="small" @click="showDetail(record)">
            详情
          </a-button>
        </template>
      </template>
    </a-table>

    <!-- 详情弹窗 -->
    <a-modal
      v-model:open="detailVisible"
      title="日志详情"
      width="600px"
      :footer="null"
    >
      <a-descriptions :column="1" bordered v-if="currentLog">
        <a-descriptions-item label="ID">{{ currentLog.id }}</a-descriptions-item>
        <a-descriptions-item label="级别">
          <a-tag :color="getLevelColor(currentLog.level)">{{ currentLog.level }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="模块">{{ currentLog.module }}</a-descriptions-item>
        <a-descriptions-item label="操作">{{ currentLog.action }}</a-descriptions-item>
        <a-descriptions-item label="消息">{{ currentLog.message }}</a-descriptions-item>
        <a-descriptions-item label="详情" v-if="currentLog.detail">
          <pre class="detail-pre">{{ currentLog.detail }}</pre>
        </a-descriptions-item>
        <a-descriptions-item label="IP" v-if="currentLog.ip">{{ currentLog.ip }}</a-descriptions-item>
        <a-descriptions-item label="耗时" v-if="currentLog.duration">{{ currentLog.duration }}ms</a-descriptions-item>
        <a-descriptions-item label="时间">{{ formatTime(currentLog.createdAt) }}</a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { SearchOutlined } from '@ant-design/icons-vue'
import dayjs, { Dayjs } from 'dayjs'
import { logApi } from '../api'

interface LogEntry {
  id: number
  level: string
  module: string
  action: string
  message: string
  detail?: string
  ip?: string
  duration?: number
  createdAt: string
}

const loading = ref(false)
const logs = ref<LogEntry[]>([])
const detailVisible = ref(false)
const currentLog = ref<LogEntry | null>(null)
const dateRange = ref<[Dayjs, Dayjs] | null>(null)

const filters = reactive({
  level: undefined as string | undefined,
  module: '',
  keyword: ''
})

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `共 ${total} 条`
})

const columns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
  { title: '级别', dataIndex: 'level', key: 'level', width: 100 },
  { title: '模块', dataIndex: 'module', key: 'module', width: 100 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 120 },
  { title: '消息', dataIndex: 'message', key: 'message', ellipsis: true },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '详情', key: 'detail_action', width: 80 }
]

function getLevelColor(level: string): string {
  const colors: Record<string, string> = {
    DEBUG: 'default',
    INFO: 'blue',
    WARN: 'orange',
    ERROR: 'red'
  }
  return colors[level] || 'default'
}

function formatTime(time: string): string {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function showDetail(record: LogEntry) {
  currentLog.value = record
  detailVisible.value = true
}

async function loadLogs() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.current,
      pageSize: pagination.pageSize
    }
    if (filters.level) params.level = filters.level
    if (filters.module) params.module = filters.module
    if (filters.keyword) params.keyword = filters.keyword
    if (dateRange.value) {
      params.startTime = dateRange.value[0].format('YYYY-MM-DD HH:mm:ss')
      params.endTime = dateRange.value[1].format('YYYY-MM-DD HH:mm:ss')
    }

    const res = await logApi.getLogs(params)
    if (res.code === 0) {
      logs.value = res.data.list || []
      pagination.total = res.data.total
    }
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  pagination.current = 1
  loadLogs()
}

function handleReset() {
  filters.level = undefined
  filters.module = ''
  filters.keyword = ''
  dateRange.value = null
  pagination.current = 1
  loadLogs()
}

function handleTableChange(pag: any) {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  loadLogs()
}

onMounted(loadLogs)
</script>

<style scoped>
.logs-page h2 {
  margin-bottom: 24px;
}

.filter-card {
  margin-bottom: 16px;
}

.detail-pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow: auto;
  background: #f5f5f5;
  padding: 8px;
  border-radius: 4px;
}
</style>
