<template>
  <a-modal
    v-model:open="visible"
    :title="`字段映射 - ${table?.tableAlias || table?.tableName}`"
    width="900px"
    :body-style="{ maxHeight: '70vh', overflowY: 'auto' }"
    @ok="handleOk"
    @cancel="visible = false"
  >
    <a-spin :spinning="loading">
      <div style="margin-bottom: 16px">
        <a-space>
          <a-button size="small" @click="toggleAll(true)">
            全部启用
          </a-button>
          <a-button size="small" @click="toggleAll(false)">
            全部禁用
          </a-button>
          <a-button size="small" @click="applyAllComments" v-if="hasAnyComment">
            一键应用备注
          </a-button>
          <a-button type="primary" size="small" @click="refreshFieldsFromDatabase">
            <template #icon><ReloadOutlined /></template>
            刷新字段
          </a-button>
        </a-space>
      </div>

      <a-table
        :columns="columns"
        :data-source="fieldMappings"
        :pagination="false"
        row-key="index"
        size="small"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'enabled'">
            <a-switch
              v-model:checked="record.enabled"
              checked-children="启用"
              un-checked-children="禁用"
            />
          </template>
          <template v-else-if="column.key === 'fieldName'">
            <span>{{ record.mysqlField }}</span>
          </template>
          <template v-else-if="column.key === 'comment'">
            <span style="color: #666">{{ record.comment || '-' }}</span>
          </template>
          <template v-else-if="column.key === 'fieldAlias'">
            <a-input
              v-model:value="record.aliasField"
              placeholder="显示别名"
              :disabled="!record.enabled"
            />
          </template>
          <template v-else-if="column.key === 'action'">
            <a-button
              v-if="record.comment"
              type="link"
              size="small"
              :disabled="!record.enabled"
              @click="applyComment(index)"
            >
              应用备注
            </a-button>
          </template>
        </template>
      </a-table>

      <div v-if="fieldMappings.length === 0" style="text-align: center; padding: 20px; color: #999">
        暂无字段数据，请检查数据源配置
      </div>
    </a-spin>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { datasourceApi } from '../../api'

const props = defineProps<{
  open: boolean
  table?: any
  datasource?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const visible = ref(props.open)
const loading = ref(false)
const fieldMappings = ref<any[]>([])

const hasAnyComment = computed(() => fieldMappings.value.some((f: any) => f.comment))

const columns = [
  { title: '启用', key: 'enabled', width: '80px' },
  { title: 'MySQL字段名', key: 'fieldName', width: '25%' },
  { title: '数据库备注', key: 'comment', width: '25%' },
  { title: '显示别名', key: 'fieldAlias', width: '30%' },
  { title: '操作', key: 'action', width: '100px' }
]

watch(() => props.open, async (val) => {
  visible.value = val
  if (val && props.table) {
    await loadFieldMappingsWithDbFields()
  }
})

watch(visible, (val) => {
  emit('update:open', val)
})

// 加载字段映射，同时获取数据库字段信息
const loadFieldMappingsWithDbFields = async () => {
  if (!props.table || !props.datasource) return

  loading.value = true
  try {
    // 并行获取已保存的映射和数据库字段
    const [mappingsRes, dbFieldsRes] = await Promise.all([
      datasourceApi.getFieldMappings(props.table.id),
      loadDbFields()
    ])

    const savedMappings = mappingsRes.code === 0 ? (mappingsRes.data || []) : []
    const dbFields = dbFieldsRes || []

    // 创建已保存映射的查找表
    const savedMap = new Map<string, any>()
    savedMappings.forEach((item: any) => {
      savedMap.set(item.fieldName, item)
    })

    // 创建数据库字段的查找表
    const dbFieldMap = new Map<string, any>()
    dbFields.forEach((field: any) => {
      dbFieldMap.set(field.name, field)
    })

    // 合并字段列表：以数据库字段为基准
    if (dbFields.length > 0) {
      fieldMappings.value = dbFields.map((field: any, index: number) => {
        const saved = savedMap.get(field.name)
        return {
          index,
          mysqlField: field.name,
          aliasField: saved?.fieldAlias || field.name,
          comment: field.comment || '',
          enabled: saved ? saved.enabled : true
        }
      })
    } else if (savedMappings.length > 0) {
      // 如果无法获取数据库字段，使用已保存的映射
      fieldMappings.value = savedMappings.map((item: any, index: number) => ({
        index,
        mysqlField: item.fieldName,
        aliasField: item.fieldAlias,
        comment: '',
        enabled: item.enabled
      }))
    } else {
      fieldMappings.value = []
    }
  } catch (error) {
    message.error('加载字段映射失败')
  } finally {
    loading.value = false
  }
}

// 从数据库加载字段列表
const loadDbFields = async (): Promise<any[]> => {
  if (!props.table || !props.datasource) return []

  try {
    let res
    if (props.table.queryMode === 'sql') {
      if (!props.table.customSql) {
        return []
      }
      res = await datasourceApi.getFieldListFromSQL(
        props.datasource.id,
        props.datasource.databaseName,
        props.table.customSql
      )
    } else {
      res = await datasourceApi.getFieldList(
        props.datasource.id,
        props.datasource.databaseName,
        props.table.tableName
      )
    }

    if (res.code === 0) {
      return res.data || []
    }
    return []
  } catch (error) {
    return []
  }
}

// 刷新字段（重新从数据库加载）
const refreshFieldsFromDatabase = async () => {
  await loadFieldMappingsWithDbFields()
  message.success('字段已刷新')
}

// 全部启用/禁用
const toggleAll = (enabled: boolean) => {
  fieldMappings.value.forEach(field => {
    field.enabled = enabled
  })
  message.success(enabled ? '已全部启用' : '已全部禁用')
}

// 应用单个备注
const applyComment = (index: number) => {
  const field = fieldMappings.value[index]
  if (field.comment) {
    field.aliasField = field.comment
  }
}

// 一键应用所有备注
const applyAllComments = () => {
  fieldMappings.value.forEach(field => {
    if (field.comment && field.enabled) {
      field.aliasField = field.comment
    }
  })
  message.success('已应用所有备注')
}

// 保存
const handleOk = async () => {
  // 转换为后端需要的格式，包含所有字段
  const allMappings = fieldMappings.value.map(item => ({
    mysqlField: item.mysqlField,
    aliasField: item.aliasField || item.mysqlField,
    enabled: item.enabled
  }))

  try {
    const res = await datasourceApi.updateFieldMappings(
      props.table.id,
      allMappings
    )
    if (res.code === 0) {
      message.success('更新成功')
      visible.value = false
      emit('success')
    } else {
      message.error(res.msg || '更新失败')
    }
  } catch (error) {
    message.error('更新失败')
  }
}
</script>
