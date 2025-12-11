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
        <a-checkbox-group
          v-model:value="selectedDatasources"
          style="width: 100%"
          @change="handleDatasourcePermissionChange"
        >
          <a-row>
            <a-col
              v-for="ds in datasourcesWithPermission"
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
          <div style="margin-bottom: 8px">
            <a-checkbox
              :indeterminate="selectedTables.length > 0 && selectedTables.length < tablesWithPermission.length"
              :checked="selectedTables.length === tablesWithPermission.length && tablesWithPermission.length > 0"
              @change="handleSelectAllTables"
            >
              全选
            </a-checkbox>
          </div>
          <a-checkbox-group
            v-model:value="selectedTables"
            style="width: 100%"
            @change="handleTablePermissionChange"
          >
            <a-row>
              <a-col
                v-for="table in tablesWithPermission"
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
        </div>
      </a-tab-pane>
    </a-tabs>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
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

watch(() => props.open, async (val) => {
  visible.value = val
  if (val && props.user) {
    activeTab.value = 'datasource'
    currentDatasourceId.value = undefined
    tablesWithPermission.value = []
    selectedTables.value = []
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
  if (!props.user) return

  const userId = props.user.id
  const oldValues = datasourcesWithPermission.value
    .filter((ds: any) => ds.hasPermission)
    .map((ds: any) => ds.id)

  const added = checkedValues.filter((id) => !oldValues.includes(id))
  const removed = oldValues.filter((id) => !checkedValues.includes(id))

  try {
    if (added.length > 0) {
      await permissionApi.grantDatasourcePermissions(userId, added)
    }

    if (removed.length > 0) {
      await permissionApi.revokeDatasourcePermissions(userId, removed)
    }

    message.success('权限更新成功')
    await loadDatasourcesWithPermission()
  } catch (error) {
    message.error('权限更新失败')
    await loadDatasourcesWithPermission()
  }
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

  isUpdatingPermission = true
  try {
    const userId = props.user.id
    const oldValues = tablesWithPermission.value
      .filter((table: any) => table.hasPermission)
      .map((table: any) => table.id)

    const added = checkedValues.filter((id) => !oldValues.includes(id))
    const removed = oldValues.filter((id) => !checkedValues.includes(id))

    if (added.length > 0) {
      await permissionApi.grantTablePermissions(userId, added)
    }

    if (removed.length > 0) {
      await permissionApi.revokeTablePermissions(userId, removed)
    }

    if (added.length > 0 || removed.length > 0) {
      message.success('权限更新成功')
      await loadTablesWithPermission()
    }
  } catch (error) {
    message.error('权限更新失败')
    await loadTablesWithPermission()
  } finally {
    isUpdatingPermission = false
  }
}
</script>
