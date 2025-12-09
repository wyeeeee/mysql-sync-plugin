<template>
  <div class="dashboard">
    <h2>仪表盘</h2>
    <a-row :gutter="16">
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="日志总数"
            :value="stats.totalCount"
            :loading="loading"
          >
            <template #prefix>
              <FileTextOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="今日日志"
            :value="stats.todayCount"
            :loading="loading"
          >
            <template #prefix>
              <CalendarOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="错误日志"
            :value="stats.errorCount"
            :loading="loading"
            :value-style="{ color: stats.errorCount > 0 ? '#cf1322' : '#3f8600' }"
          >
            <template #prefix>
              <WarningOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="服务状态"
            value="运行中"
            :loading="loading"
            :value-style="{ color: '#3f8600' }"
          >
            <template #prefix>
              <CheckCircleOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <a-card title="日志级别分布" style="margin-top: 16px">
      <a-row :gutter="16">
        <a-col :span="6" v-for="(count, level) in stats.levelCounts" :key="level">
          <a-statistic
            :title="level"
            :value="count"
            :value-style="{ color: getLevelColor(level as string) }"
          />
        </a-col>
      </a-row>
    </a-card>

    <a-card title="系统信息" style="margin-top: 16px">
      <a-descriptions :column="2">
        <a-descriptions-item label="服务名称">{{ systemInfo.service }}</a-descriptions-item>
        <a-descriptions-item label="版本">{{ systemInfo.version }}</a-descriptions-item>
      </a-descriptions>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  FileTextOutlined,
  CalendarOutlined,
  WarningOutlined,
  CheckCircleOutlined
} from '@ant-design/icons-vue'
import { logApi, systemApi } from '../api'

const loading = ref(true)
const stats = ref({
  totalCount: 0,
  todayCount: 0,
  errorCount: 0,
  levelCounts: {} as Record<string, number>
})
const systemInfo = ref({
  service: '',
  version: ''
})

function getLevelColor(level: string): string {
  const colors: Record<string, string> = {
    DEBUG: '#8c8c8c',
    INFO: '#1890ff',
    WARN: '#faad14',
    ERROR: '#cf1322'
  }
  return colors[level] || '#333'
}

async function loadData() {
  loading.value = true
  try {
    const [statsRes, infoRes] = await Promise.all([
      logApi.getStats(),
      systemApi.getInfo()
    ])
    if (statsRes.code === 0) {
      stats.value = statsRes.data
    }
    if (infoRes.code === 0) {
      systemInfo.value = infoRes.data
    }
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.dashboard h2 {
  margin-bottom: 24px;
}
</style>
