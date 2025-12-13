<template>
  <a-modal
    v-model:open="visible"
    title="权限管理"
    width="800px"
    @ok="visible = false"
    @cancel="visible = false"
  >
    <a-tabs v-model:activeKey="activeTab">
      <a-tab-pane key="datasource" tab="数据源权限">
        <a-input
          v-model:value="datasourceSearchKeyword"
          placeholder="搜索数据源名称或描述"
          allow-clear
          style="margin-bottom: 12px"
        >
          <template #prefix>
            <SearchOutlined style="color: #bfbfbf" />
          </template>
        </a-input>
        <div style="max-height: 350px; overflow-y: auto;">
          <a-checkbox-group
            v-model:value="selectedDatasources"
            style="width: 100%"
            @change="handleDatasourcePermissionChange"
          >
            <a-row>
              <a-col
                v-for="ds in filteredDatasources"
                :key="ds.id"
                :span="24"
                style="margin-bottom: 8px"
              >
                <a-checkbox :value="ds.id">
                  {{ ds.name }}
                  <span style="color: #999; margin-left: 8px">
                    {{ ds.description }}
                  </span>
                </a-checkbox>
              </a-col>
            </a-row>
          </a-checkbox-group>
          <div v-if="filteredDatasources.length === 0" style="text-align: center; padding: 20px; color: #999">
            {{ datasourceSearchKeyword ? '未找到匹配的数据源' : '暂无数据源' }}
          </div>
        </div>
      </a-tab-pane>
      <a-tab-pane key="table" tab="表权限">
        <a-form-item label="选择数据源">
          <a-select
            v-model:value="currentDatasourceId"
            placeholder="请选择数据源"
            style="width: 100%"
            @change="loadTablesWithPermission"
          >
            <a-select-option
              v-for="ds in userDatasources"
              :key="ds.id"
              :value="ds.id"
            >
              {{ ds.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <div v-if="!currentDatasourceId" style="text-align: center; padding: 20px; color: #999">
          请先选择一个数据源
        </div>
        <div v-else>
          <a-input
            v-model:value="tableSearchKeyword"
            placeholder="搜索表名或别名"
            allow-clear
            style="margin-bottom: 12px"
          >
            <template #prefix>
              <SearchOutlined style="color: #bfbfbf" />
            </template>
          </a-input>
          <div style="margin-bottom: 8px">
            <a-checkbox
              :indeterminate="selectedTables.length > 0 && selectedTables.length < filteredTables.length"
              :checked="selectedTables.length === filteredTables.length && filteredTables.length > 0"
              @change="handleSelectAllTables"
            >
              全选
            </a-checkbox>
            <span style="color: #999; margin-left: 12px; font-size: 12px">
              (已选 {{ selectedTables.length }} / {{ tablesWithPermission.length }})
            </span>
          </div>
          <div style="max-height: 300px; overflow-y: auto;">
            <a-checkbox-group
              v-model:value="selectedTables"
              style="width: 100%"
              @change="handleTablePermissionChange"
            >
              <a-row>
                <a-col
                  v-for="table in filteredTables"
                  :key="table.id"
                  :span="24"
                  style="margin-bottom: 8px"
                >
                  <a-checkbox :value="table.id">
                    {{ table.tableAlias || table.tableName }}
                    <span style="color: #999; margin-left: 8px">
                      ({{ table.tableName }})
                    </span>
                  </a-checkbox>
                </a-col>
              </a-row>
            </a-checkbox-group>
            <div v-if="filteredTables.length === 0" style="text-align: center; padding: 20px; color: #999">
              {{ tableSearchKeyword ? '未找到匹配的表' : '暂无表配置' }}
            </div>
          </div>
        </div>
      </a-tab-pane>
    </a-tabs>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { message } from 'ant-design-vue'
import { SearchOutlined } from '@ant-design/icons-vue'
import { permissionApi } from '../../api'

const props = defineProps<{
  open: boolean
  user?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
}>()

const visible = ref(props.open)
const activeTab = ref('datasource')
const datasourcesWithPermission = ref<any[]>([])
const selectedDatasources = ref<number[]>([])
const userDatasources = ref<any[]>([])
const currentDatasourceId = ref<number>()
const tablesWithPermission = ref<any[]>([])
const selectedTables = ref<number[]>([])
let isUpdatingPermission = false
let datasourcePermissionTimer: number | null = null
let tablePermissionTimer: number | null = null

// 搜索关键字
const datasourceSearchKeyword = ref('')
const tableSearchKeyword = ref('')

// 过滤后的数据源列表
const filteredDatasources = computed(() => {
  if (!datasourceSearchKeyword.value) {
    return datasourcesWithPermission.value
  }
  const keyword = datasourceSearchKeyword.value.toLowerCase()
  return datasourcesWithPermission.value.filter((ds: any) =>
    ds.name?.toLowerCase().includes(keyword) ||
    ds.description?.toLowerCase().includes(keyword)
  )
})

// 过滤后的表列表
const filteredTables = computed(() => {
  if (!tableSearchKeyword.value) {
    return tablesWithPermission.value
  }
  const keyword = tableSearchKeyword.value.toLowerCase()
  return tablesWithPermission.value.filter((table: any) =>
    table.tableName?.toLowerCase().includes(keyword) ||
    table.tableAlias?.toLowerCase().includes(keyword)
  )
})

watch(() => props.open, async (val) => {
  visible.value = val
  if (val && props.user) {
    activeTab.value = 'datasource'
    currentDatasourceId.value = undefined
    tablesWithPermission.value = []
    selectedTables.value = []
    datasourceSearchKeyword.value = ''
    tableSearchKeyword.value = ''
    await loadDatasourcesWithPermission()
  }
})

watch(visible, (val) => {
  emit('update:open', val)
})

const loadDatasourcesWithPermission = async () => {
  if (!props.user) return

  try {
    const res = await permissionApi.listAllDatasourcesWithPermission(props.user.id)
    if (res.code === 0) {
      datasourcesWithPermission.value = res.data || []
      selectedDatasources.value = datasourcesWithPermission.value
        .filter((ds: any) => ds.hasPermission)
        .map((ds: any) => ds.id)

      const userDsRes = await permissionApi.listUserDatasources(props.user.id)
      if (userDsRes.code === 0) {
        userDatasources.value = userDsRes.data || []
        if (userDatasources.value.length > 0) {
          currentDatasourceId.value = userDatasources.value[0].id
          await loadTablesWithPermission()
        }
      }
    }
  } catch (error) {
    message.error('加载数据源权限失败')
  }
}

const handleDatasourcePermissionChange = async (checkedValues: number[]) => {
  if (!props.user || isUpdatingPermission) return

  // 清除之前的定时器
  if (datasourcePermissionTimer !== null) {
    clearTimeout(datasourcePermissionTimer)
  }

  // 使用防抖,500ms后执行批量更新
  datasourcePermissionTimer = window.setTimeout(async () => {
    isUpdatingPermission = true
    const hideLoading = message.loading('正在更新数据源权限...', 0)

    try {
      const userId = props.user.id
      const oldValues = datasourcesWithPermission.value
        .filter((ds: any) => ds.hasPermission)
        .map((ds: any) => ds.id)

      const added = checkedValues.filter((id) => !oldValues.includes(id))
      const removed = oldValues.filter((id) => !checkedValues.includes(id))

      // 使用批量接口一次性处理所有变更
      if (added.length > 0) {
        await permissionApi.grantDatasourcePermissions(userId, added)
      }

      if (removed.length > 0) {
        await permissionApi.revokeDatasourcePermissions(userId, removed)
      }

      hideLoading()
      if (added.length > 0 || removed.length > 0) {
        message.success('数据源权限更新成功')
        await loadDatasourcesWithPermission()
      }
    } catch (error) {
      hideLoading()
      message.error('数据源权限更新失败')
      await loadDatasourcesWithPermission()
    } finally {
      isUpdatingPermission = false
      datasourcePermissionTimer = null
    }
  }, 500)
}

const loadTablesWithPermission = async () => {
  if (!currentDatasourceId.value || !props.user) return

  try {
    const res = await permissionApi.listAllTablesWithPermission(
      props.user.id,
      currentDatasourceId.value
    )
    if (res.code === 0) {
      tablesWithPermission.value = res.data || []
      selectedTables.value = tablesWithPermission.value
        .filter((table: any) => table.hasPermission)
        .map((table: any) => table.id)
    }
  } catch (error) {
    message.error('加载表权限失败')
  }
}

const handleSelectAllTables = async (e: any) => {
  if (isUpdatingPermission || !props.user) return

  isUpdatingPermission = true
  const hideLoading = message.loading('正在更新权限...', 0)

  try {
    const userId = props.user.id
    const oldValues = tablesWithPermission.value
      .filter((table: any) => table.hasPermission)
      .map((table: any) => table.id)

    if (e.target.checked) {
      const allTableIds = tablesWithPermission.value.map((table: any) => table.id)
      const added = allTableIds.filter((id) => !oldValues.includes(id))

      if (added.length > 0) {
        await permissionApi.grantTablePermissions(userId, added)
      }
      selectedTables.value = allTableIds
    } else {
      if (oldValues.length > 0) {
        await permissionApi.revokeTablePermissions(userId, oldValues)
      }
      selectedTables.value = []
    }

    hideLoading()
    message.success('权限更新成功')
    await loadTablesWithPermission()
  } catch (error) {
    hideLoading()
    message.error('权限更新失败')
    await loadTablesWithPermission()
  } finally {
    isUpdatingPermission = false
  }
}

const handleTablePermissionChange = async (checkedValues: number[]) => {
  if (isUpdatingPermission || !props.user) return

  // 清除之前的定时器
  if (tablePermissionTimer !== null) {
    clearTimeout(tablePermissionTimer)
  }

  // 使用防抖,500ms后执行批量更新
  tablePermissionTimer = window.setTimeout(async () => {
    isUpdatingPermission = true
    const hideLoading = message.loading('正在更新表权限...', 0)

    try {
      const userId = props.user.id
      const oldValues = tablesWithPermission.value
        .filter((table: any) => table.hasPermission)
        .map((table: any) => table.id)

      const added = checkedValues.filter((id) => !oldValues.includes(id))
      const removed = oldValues.filter((id) => !checkedValues.includes(id))

      // 使用批量接口一次性处理所有变更
      if (added.length > 0) {
        await permissionApi.grantTablePermissions(userId, added)
      }

      if (removed.length > 0) {
        await permissionApi.revokeTablePermissions(userId, removed)
      }

      hideLoading()
      if (added.length > 0 || removed.length > 0) {
        message.success('表权限更新成功')
        await loadTablesWithPermission()
      }
    } catch (error) {
      hideLoading()
      message.error('表权限更新失败')
      await loadTablesWithPermission()
    } finally {
      isUpdatingPermission = false
      tablePermissionTimer = null
    }
  }, 500)
}
</script>
