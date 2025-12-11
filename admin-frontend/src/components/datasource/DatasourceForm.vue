<template>
  <a-modal
    v-model:open="visible"
    :title="title"
    width="600px"
    @ok="handleOk"
    @cancel="handleCancel"
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
          @focus="handleLoadDatabases"
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
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { datasourceApi } from '../../api'

const props = defineProps<{
  open: boolean
  isEdit: boolean
  datasource?: any
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'success'): void
}>()

const visible = ref(props.open)
const title = ref(props.isEdit ? '编辑数据源' : '新建数据源')
const formData = ref({
  id: 0,
  name: '',
  description: '',
  host: '',
  port: 3306,
  databaseName: '',
  username: '',
  password: ''
})

const availableDatabases = ref<string[]>([])
const databasesLoading = ref(false)
const tempDatasourceId = ref<number>()

watch(() => props.open, (val) => {
  visible.value = val
  if (val) {
    title.value = props.isEdit ? '编辑数据源' : '新建数据源'
    if (props.isEdit && props.datasource) {
      formData.value = {
        id: props.datasource.id,
        name: props.datasource.name,
        description: props.datasource.description,
        host: props.datasource.host,
        port: props.datasource.port,
        databaseName: props.datasource.databaseName,
        username: props.datasource.username,
        password: ''
      }
      availableDatabases.value = [props.datasource.databaseName]
      tempDatasourceId.value = props.datasource.id
    } else {
      formData.value = {
        id: 0,
        name: '',
        description: '',
        host: '',
        port: 3306,
        databaseName: '',
        username: '',
        password: ''
      }
      availableDatabases.value = []
      tempDatasourceId.value = undefined
    }
  }
})

watch(visible, (val) => {
  emit('update:open', val)
})

const handleLoadDatabases = async () => {
  if (props.isEdit && tempDatasourceId.value) {
    await loadDatabasesFromDatasource(tempDatasourceId.value)
    return
  }

  if (!formData.value.host || !formData.value.port || !formData.value.username || !formData.value.password) {
    message.warning('请先填写主机、端口、用户名和密码')
    return
  }

  databasesLoading.value = true
  try {
    const res = await datasourceApi.createDatasource({
      name: '__temp__',
      host: formData.value.host,
      port: formData.value.port,
      databaseName: 'mysql',
      username: formData.value.username,
      password: formData.value.password
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

const cleanupTempDatasource = async () => {
  if (!props.isEdit && tempDatasourceId.value) {
    try {
      await datasourceApi.deleteDatasource(tempDatasourceId.value)
    } catch (error) {
      // 忽略错误
    }
    tempDatasourceId.value = undefined
  }
}

const handleOk = async () => {
  if (!formData.value.name || !formData.value.host || !formData.value.port || !formData.value.databaseName || !formData.value.username) {
    message.error('请填写所有必填项')
    return
  }
  if (!props.isEdit && !formData.value.password) {
    message.error('请输入密码')
    return
  }

  try {
    let res
    if (props.isEdit) {
      res = await datasourceApi.updateDatasource(formData.value.id, {
        name: formData.value.name,
        description: formData.value.description,
        host: formData.value.host,
        port: formData.value.port,
        databaseName: formData.value.databaseName,
        username: formData.value.username,
        password: formData.value.password || undefined
      })
    } else {
      res = await datasourceApi.createDatasource({
        name: formData.value.name,
        description: formData.value.description,
        host: formData.value.host,
        port: formData.value.port,
        databaseName: formData.value.databaseName,
        username: formData.value.username,
        password: formData.value.password
      })
    }

    if (res.code === 0) {
      message.success(props.isEdit ? '更新成功' : '创建成功')
      await cleanupTempDatasource()
      visible.value = false
      emit('success')
    } else {
      message.error(res.msg || '操作失败')
    }
  } catch (error) {
    message.error('操作失败')
  }
}

const handleCancel = async () => {
  await cleanupTempDatasource()
  visible.value = false
}
</script>
