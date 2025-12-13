<template>
  <div class="dashboard">
    <div class="page-header">
      <h2 class="page-title">
        <DashboardOutlined class="title-icon" />
        仪表盘
      </h2>
      <a-button type="primary" class="cherry-btn-sm" @click="loadData" :loading="loading">
        <ReloadOutlined />
        刷新数据
      </a-button>
    </div>

    <!-- 统计卡片 -->
    <a-row :gutter="[16, 16]" class="stats-row">
      <a-col :xs="12" :sm="12" :md="6" :lg="6">
        <div class="stat-card stat-card-total">
          <div class="stat-icon">
            <FileTextOutlined />
          </div>
          <div class="stat-content">
            <div class="stat-value">
              <a-spin v-if="loading" size="small" />
              <span v-else>{{ stats.totalCount }}</span>
            </div>
            <div class="stat-label">日志总数</div>
          </div>
        </div>
      </a-col>
      <a-col :xs="12" :sm="12" :md="6" :lg="6">
        <div class="stat-card stat-card-today">
          <div class="stat-icon">
            <CalendarOutlined />
          </div>
          <div class="stat-content">
            <div class="stat-value">
              <a-spin v-if="loading" size="small" />
              <span v-else>{{ stats.todayCount }}</span>
            </div>
            <div class="stat-label">今日日志</div>
          </div>
        </div>
      </a-col>
      <a-col :xs="12" :sm="12" :md="6" :lg="6">
        <div class="stat-card stat-card-error">
          <div class="stat-icon">
            <WarningOutlined />
          </div>
          <div class="stat-content">
            <div class="stat-value" :class="{ 'has-error': stats.errorCount > 0 }">
              <a-spin v-if="loading" size="small" />
              <span v-else>{{ stats.errorCount }}</span>
            </div>
            <div class="stat-label">错误日志</div>
          </div>
        </div>
      </a-col>
      <a-col :xs="12" :sm="12" :md="6" :lg="6">
        <div class="stat-card stat-card-status">
          <div class="stat-icon">
            <CheckCircleOutlined />
          </div>
          <div class="stat-content">
            <div class="stat-value status-running">
              <a-spin v-if="loading" size="small" />
              <span v-else>运行中</span>
            </div>
            <div class="stat-label">服务状态</div>
          </div>
        </div>
      </a-col>
    </a-row>

    <!-- 日志级别分布 -->
    <a-card class="cherry-card" style="margin-top: 20px">
      <template #title>
        <div class="card-title">
          <PieChartOutlined class="card-title-icon" />
          日志级别分布
        </div>
      </template>
      <a-row :gutter="[16, 16]">
        <a-col :xs="12" :sm="6" v-for="(count, level) in stats.levelCounts" :key="level">
          <div class="level-stat">
            <div class="level-badge" :style="{ background: getLevelColor(level as string) }">
              {{ level }}
            </div>
            <div class="level-count">{{ count }}</div>
          </div>
        </a-col>
        <a-col v-if="Object.keys(stats.levelCounts).length === 0" :span="24">
          <a-empty description="暂无数据" />
        </a-col>
      </a-row>
    </a-card>

    <!-- 系统信息 -->
    <a-card class="cherry-card" style="margin-top: 20px">
      <template #title>
        <div class="card-title">
          <InfoCircleOutlined class="card-title-icon" />
          系统信息
        </div>
      </template>
      <a-descriptions :column="{ xs: 1, sm: 2, md: 2 }" class="system-info">
        <a-descriptions-item label="服务名称">
          <span class="info-value">{{ systemInfo.service || '-' }}</span>
        </a-descriptions-item>
        <a-descriptions-item label="版本">
          <a-tag color="processing">{{ systemInfo.version || '-' }}</a-tag>
        </a-descriptions-item>
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
  CheckCircleOutlined,
  DashboardOutlined,
  ReloadOutlined,
  PieChartOutlined,
  InfoCircleOutlined
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
    ERROR: '#ef4444'
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
.dashboard {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
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

/* 统计卡片 */
.stat-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.3s ease;
  border: 1px solid #f0f0f0;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  flex-shrink: 0;
}

.stat-card-total .stat-icon {
  background: linear-gradient(135deg, #1e3a5f 0%, #2d4a6f 100%);
  color: white;
}

.stat-card-today .stat-icon {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
}

.stat-card-error .stat-icon {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.stat-card-status .stat-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white;
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 26px;
  font-weight: 700;
  color: #2c3e50;
  line-height: 1.2;
}

.stat-value.has-error {
  color: #ef4444;
}

.stat-value.status-running {
  color: #27ae60;
  font-size: 20px;
}

.stat-label {
  font-size: 13px;
  color: #7f8c8d;
  margin-top: 4px;
}

/* 卡片样式 */
.cherry-card {
  border-radius: 12px;
  border: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.cherry-card :deep(.ant-card-head) {
  border-bottom: 1px solid #f5f5f5;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #2c3e50;
}

.card-title-icon {
  color: #1e3a5f;
}

/* 日志级别统计 */
.level-stat {
  text-align: center;
  padding: 16px;
  background: #fafafa;
  border-radius: 10px;
  transition: all 0.3s;
}

.level-stat:hover {
  background: #f5f5f5;
}

.level-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 20px;
  color: white;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 8px;
}

.level-count {
  font-size: 24px;
  font-weight: 700;
  color: #2c3e50;
}

/* 系统信息 */
.system-info :deep(.ant-descriptions-item-label) {
  color: #7f8c8d;
}

.info-value {
  font-weight: 500;
  color: #2c3e50;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .page-header {
    margin-bottom: 16px;
  }

  .page-title {
    font-size: 18px;
  }

  .stat-card {
    padding: 16px;
    gap: 12px;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 18px;
    border-radius: 10px;
  }

  .stat-value {
    font-size: 22px;
  }

  .stat-value.status-running {
    font-size: 16px;
  }

  .stat-label {
    font-size: 12px;
  }

  .level-stat {
    padding: 12px;
  }

  .level-count {
    font-size: 20px;
  }
}

@media (max-width: 576px) {
  .stat-card {
    padding: 14px;
    gap: 10px;
  }

  .stat-icon {
    width: 36px;
    height: 36px;
    font-size: 16px;
    border-radius: 8px;
  }

  .stat-value {
    font-size: 18px;
  }

  .stat-label {
    font-size: 11px;
  }

  .cherry-card {
    border-radius: 10px;
  }

  .level-stat {
    padding: 10px;
  }

  .level-badge {
    font-size: 11px;
    padding: 3px 10px;
  }

  .level-count {
    font-size: 18px;
  }
}
</style>
