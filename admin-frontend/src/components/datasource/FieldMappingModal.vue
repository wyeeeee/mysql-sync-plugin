<template>
  <a-modal
    v-model:open="visible"
    :title="`字段映射 - ${table?.tableAlias || table?.tableName}`"
    width="900px"
    @ok="handleOk"
    @cancel="visible = false"
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
      :columns="columns"
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
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
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
const fieldMappings = ref<any[]>([])
const hasAnyComment = ref(false)

const columns = [
  { title: 'MySQL字段名', key: 'fieldName', width: '25%' },
  { title: '数据库备注', key: 'comment', width: '25%' },
  { title: '显示别名', key: 'fieldAlias', width: '30%' },
  { title: '操作', key: 'action', width: '20%' }
]

watch(() => props.open, async (val) => {
  visible.value = val
  if (val && props.table) {
    await loadFieldMappings()
  }
})

watch(visible, (val) => {
  emit('update:open', val)
})

const loadFieldMappings = async () => {
  if (!props.table) return

  try {
    const res = await datasourceApi.getFieldMappings(props.table.id)
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

const loadFieldsFromDatabase = async () => {
  if (!props.table || !props.datasource) return

  try {
    let res
    if (props.table.queryMode === 'sql') {
      if (!props.table.customSql) {
        message.error('自定义SQL为空,无法加载字段')
        return
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

const applyComment = (index: number) => {
  const field = fieldMappings.value[index]
  if (field.comment) {
    field.aliasField = field.comment
  }
}

const applyAllComments = () => {
  fieldMappings.value.forEach(field => {
    if (field.comment) {
      field.aliasField = field.comment
    }
  })
  message.success('已应用所有备注')
}

const addFieldMapping = () => {
  fieldMappings.value.push({
    index: fieldMappings.value.length,
    mysqlField: '',
    aliasField: '',
    comment: '',
    fromDatabase: false
  })
}

const removeFieldMapping = (index: number) => {
  fieldMappings.value.splice(index, 1)
  fieldMappings.value.forEach((item, idx) => {
    item.index = idx
  })
}

const handleOk = async () => {
  const validMappings = fieldMappings.value.filter(
    (item) => item.mysqlField && item.aliasField
  )

  try {
    const res = await datasourceApi.updateFieldMappings(
      props.table.id,
      validMappings
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
