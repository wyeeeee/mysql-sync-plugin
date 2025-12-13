<template>
  <div class="logs-page">
    <div class="page-header">
      <h2 class="page-title">
        <FileTextOutlined class="title-icon" />
        日志管理
      </h2>
    </div>

    <!-- 搜索筛选 -->
    <a-card class="cherry-card filter-card">
      <a-row :gutter="[12, 12]">
        <a-col :xs="12" :sm="8" :md="6" :lg="4">
          <a-select
            v-model:value="filters.level"
            placeholder="日志级别"
            style="width: 100%"
            allowClear
            class="cherry-select"
          >
            <a-select-option value="DEBUG">DEBUG</a-select-option>
            <a-select-option value="INFO">INFO</a-select-option>
            <a-select-option value="WARN">WARN</a-select-option>
            <a-select-option value="ERROR">ERROR</a-select-option>
          </a-select>
        </a-col>
        <a-col :xs="12" :sm="8" :md="6" :lg="4">
          <a-input
            v-model:value="filters.module"
            placeholder="模块名称"
            allowClear
            class="cherry-input"
          />
        </a-col>
        <a-col :xs="24" :sm="8" :md="6" :lg="5">
          <a-input
            v-model:value="filters.keyword"
            placeholder="搜索内容"
            allowClear
            class="cherry-input"
          >
            <template #prefix>
              <SearchOutlined style="color: #bfbfbf" />
            </template>
          </a-input>
        </a-col>
        <a-col :xs="24" :sm="16" :md="12" :lg="7">
          <a-range-picker
            v-model:value="dateRange"
            show-time
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
            class="cherry-picker"
            :placeholder="['开始时间', '结束时间']"
          />
        </a-col>
        <a-col :xs="24" :sm="8" :md="6" :lg="4">
          <a-space class="search-btns">
            <a-button type="primary" class="cherry-btn-sm" @click="handleSearch">
              <SearchOutlined />
              搜索
            </a-button>
            <a-button @click="handleReset">重置</a-button>
          </a-space>
        </a-col>
      </a-row>
    </a-card>

    <!-- 日志表格 -->
    <a-card class="cherry-card table-card">
      <div class="table-wrapper">
        <a-table
          :columns="responsiveColumns"
          :data-source="logs"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
          :scroll="{ x: 800 }"
          class="cherry-table"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'level'">
              <a-tag :color="getLevelColor(record.level)" class="level-tag">
                {{ record.level }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'createdAt'">
              <span class="time-text">{{ formatTime(record.createdAt) }}</span>
            </template>
            <template v-else-if="column.key === 'detail_action'">
              <a-button type="link" size="small" class="action-btn" @click="showDetail(record)">
                <EyeOutlined />
                详情
              </a-button>
            </template>
          </template>
        </a-table>
      </div>
    </a-card>

    <!-- 详情弹窗 -->
    <a-modal
      v-model:open="detailVisible"
      title="日志详情"
      :width="isFullscreen ? '100%' : (isMobile ? '95%' : '800px')"
      :style="isFullscreen ? { top: 0, paddingBottom: 0, maxWidth: '100%' } : {}"
      :bodyStyle="isFullscreen ? { height: 'calc(100vh - 110px)', overflow: 'auto' } : {}"
      :footer="null"
      class="cherry-modal"
    >
      <template #title>
        <div class="modal-title">
          <span><FileTextOutlined /> 日志详情</span>
          <a-button type="text" size="small" class="fullscreen-btn" @click="isFullscreen = !isFullscreen">
            <FullscreenOutlined v-if="!isFullscreen" />
            <FullscreenExitOutlined v-else />
          </a-button>
        </div>
      </template>
      <a-descriptions :column="{ xs: 1, sm: 1, md: 1 }" bordered v-if="currentLog" size="small" class="log-detail">
        <a-descriptions-item label="ID">{{ currentLog.id }}</a-descriptions-item>
        <a-descriptions-item label="级别">
          <a-tag :color="getLevelColor(currentLog.level)">{{ currentLog.level }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="模块">{{ currentLog.module }}</a-descriptions-item>
        <a-descriptions-item label="操作">{{ currentLog.action }}</a-descriptions-item>
        <a-descriptions-item label="消息">{{ currentLog.message }}</a-descriptions-item>
        <a-descriptions-item label="详情" v-if="currentLog.detail">
          <pre class="detail-pre" :class="{ 'detail-pre-fullscreen': isFullscreen }">{{ currentLog.detail }}</pre>
        </a-descriptions-item>
        <a-descriptions-item label="IP" v-if="currentLog.ip">{{ currentLog.ip }}</a-descriptions-item>
        <a-descriptions-item label="耗时" v-if="currentLog.duration">
          <a-tag color="blue">{{ currentLog.duration }}ms</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="时间">{{ formatTime(currentLog.createdAt) }}</a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue'
import { SearchOutlined, FullscreenOutlined, FullscreenExitOutlined, FileTextOutlined, EyeOutlined } from '@ant-design/icons-vue'
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
const isFullscreen = ref(false)
const isMobile = ref(false)

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

const allColumns = [
  { title: 'ID', dataIndex: 'id', key: 'id', width: 70 },
  { title: '级别', dataIndex: 'level', key: 'level', width: 90 },
  { title: '模块', dataIndex: 'module', key: 'module', width: 100 },
  { title: '操作', dataIndex: 'action', key: 'action', width: 100 },
  { title: '消息', dataIndex: 'message', key: 'message', ellipsis: true },
  { title: '时间', dataIndex: 'createdAt', key: 'createdAt', width: 160 },
  { title: '详情', key: 'detail_action', width: 80, fixed: 'right' as const }
]

const responsiveColumns = computed(() => {
  if (isMobile.value) {
    return allColumns.filter(col => ['level', 'message', 'detail_action'].includes(col.key))
      .map(col => col.key === 'detail_action' ? { ...col, fixed: undefined } : col)
  }
  return allColumns
})

function checkMobile() {
  isMobile.value = window.innerWidth < 768
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  loadLogs()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

function getLevelColor(level: string): string {
  const colors: Record<string, string> = {
    DEBUG: 'default',
    INFO: 'processing',
    WARN: 'warning',
    ERROR: 'error'
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
</script>

<style scoped>
.logs-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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

.cherry-card {
  border-radius: 12px;
  border: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.filter-card {
  margin-bottom: 16px;
}

.table-card {
  margin-top: 0;
}

.search-btns {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.cherry-btn-sm {
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  border: none;
  border-radius: 8px;
  height: 32px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.cherry-btn-sm:hover {
  background: linear-gradient(135deg, #2d4a6f 0%, #3d5a7f 100%);
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

.cherry-picker :deep(.ant-picker) {
  border-radius: 8px;
}

.cherry-picker :deep(.ant-picker-focused) {
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

.level-tag {
  border-radius: 4px;
  font-weight: 500;
}

.time-text {
  font-size: 13px;
  color: #666;
}

.action-btn {
  padding: 0 6px;
  font-size: 13px;
  color: #1e3a5f;
  display: flex;
  align-items: center;
  gap: 4px;
}

.action-btn:hover {
  color: #2d4a6f;
}

.modal-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-right: 32px;
  font-weight: 600;
  color: #2c3e50;
}

.modal-title span {
  display: flex;
  align-items: center;
  gap: 8px;
}

.fullscreen-btn {
  color: #666;
}

.fullscreen-btn:hover {
  color: #1e3a5f;
}

.log-detail :deep(.ant-descriptions-item-label) {
  font-weight: 500;
  color: #666;
  width: 80px;
}

.detail-pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 300px;
  overflow: auto;
  background: #f8f9fa;
  padding: 12px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.6;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  border: 1px solid #eee;
}

.detail-pre-fullscreen {
  max-height: none;
  height: auto;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .page-header {
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
  }

  .cherry-card {
    border-radius: 10px;
  }

  .cherry-card :deep(.ant-card-body) {
    padding: 16px;
  }

  .search-btns {
    width: 100%;
    justify-content: flex-start;
  }

  .filter-card {
    margin-bottom: 12px;
  }
}

@media (max-width: 576px) {
  .page-title {
    font-size: 16px;
  }

  .cherry-card :deep(.ant-card-body) {
    padding: 12px;
  }

  .detail-pre {
    font-size: 12px;
    padding: 10px;
  }

  .log-detail :deep(.ant-descriptions-item-label) {
    width: 60px;
    font-size: 13px;
  }
}
</style>
