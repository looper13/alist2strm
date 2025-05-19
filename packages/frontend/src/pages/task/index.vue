<script setup lang="ts">
import type { DataTableColumns, FormRules } from 'naive-ui'
import cronValidate from 'cron-validate'
import { taskAPI } from '~/api/task'

// 状态定义
const loading = ref(false)
const showModal = ref(false)
const showLogModal = ref(false)
const isEdit = ref(false)
const currentId = ref<number | null>(null)
const tasks = ref<Api.Task[]>([])
const taskLogs = ref<Api.TaskLog[]>([])
const logLoading = ref(false)

// 搜索
const searchForm = reactive({
  name: '',
})

// 表单实例和数据
const formRef = ref<any>(null)
const formModel = ref<Api.TaskCreateDto>({
  name: '',
  sourcePath: '',
  targetPath: '',
  fileSuffix: '',
  overwrite: false,
  enabled: true,
  running: false,
  cron: '',
})

// 文件后缀验证规则
function fileSuffixValidator(rule: any, value: string) {
  const suffixes = value.split(',').map((s: string) => s.trim())
  if (suffixes.some((s: string) => !s))
    return new Error('文件后缀不能为空')
  if (suffixes.some((s: string) => s.includes('.')))
    return new Error('文件后缀不需要包含点号')
  return true
}
// 表单规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' },
    { min: 2, max: 50, message: '任务名称长度应在 2-50 个字符之间', trigger: 'blur' },
  ],
  sourcePath: [
    { required: true, message: '请输AList路径', trigger: 'blur' },
    {
      validator: (_, value) => {
        if (!value.startsWith('/'))
          return new Error('路径必须以 / 开头')
        return true
      },
      trigger: 'blur',
    },
  ],
  targetPath: [
    { required: true, message: '请输入目标路径', trigger: 'blur' },
  ],
  fileSuffix: [
    { required: true, message: '请输入文件后缀', trigger: 'blur' },
    { validator: fileSuffixValidator, trigger: 'blur' },
  ],
  cron: [
    { required: true, message: '请输入Cron表达式', trigger: 'blur' },
    {
      validator: (_, value) => {
        if (!value)
          return true
        const result = cronValidate(value)
        if (!result.isValid()) {
          return new Error(`无效的 Cron 表达式: ${result.getError()}`)
        }
        return true
      },
      trigger: 'blur',
    },
  ],
}

// 消息提示
const message = useMessage()

// 加载任务列表
async function loadTasks() {
  try {
    loading.value = true
    const { data } = await taskAPI.findAll({ name: searchForm.name })
    tasks.value = data || []
  }
  catch (error: any) {
    message.error(error.message || '加载失败')
  }
  finally {
    loading.value = false
  }
}

// 打开创建对话框
function handleCreate() {
  isEdit.value = false
  currentId.value = null
  formModel.value = {
    name: '',
    sourcePath: '',
    targetPath: '',
    fileSuffix: '',
    overwrite: false,
    enabled: true,
    running: false,
    cron: '',
  }
  showModal.value = true
}

// 打开编辑对话框
async function handleEdit(row: Api.Task) {
  try {
    isEdit.value = true
    currentId.value = row.id
    formModel.value = {
      name: row.name,
      sourcePath: row.sourcePath,
      targetPath: row.targetPath,
      fileSuffix: row.fileSuffix,
      overwrite: row.overwrite,
      enabled: row.enabled,
      running: row.running,
      cron: row.cron || '',
    }
    showModal.value = true
  }
  catch (error: any) {
    message.error(error.message || '加载失败')
  }
}

// 打开复制新增对话框
function handleCopy(row: Api.Task) {
  isEdit.value = false
  currentId.value = null
  formModel.value = {
    name: `${row.name}_复制`,
    sourcePath: row.sourcePath,
    targetPath: row.targetPath,
    fileSuffix: row.fileSuffix,
    overwrite: row.overwrite,
    enabled: row.enabled,
    running: false,
    cron: row.cron || '',
  }
  showModal.value = true
}

// 保存任务
async function handleSave() {
  try {
    await formRef.value?.validate()
    if (isEdit.value && currentId.value) {
      await taskAPI.update(currentId.value, {
        name: formModel.value.name,
        sourcePath: formModel.value.sourcePath,
        targetPath: formModel.value.targetPath,
        fileSuffix: formModel.value.fileSuffix,
        overwrite: formModel.value.overwrite,
        enabled: formModel.value.enabled,
        running: formModel.value.running,
        cron: formModel.value.cron,
      })
    }
    else {
      await taskAPI.create(formModel.value)
    }
    message.success('保存成功')
    showModal.value = false
    loadTasks()
  }
  catch (error: any) {
    message.error(error.message || '保存失败')
  }
}

// 删除任务
async function handleDelete(row: Api.Task) {
  try {
    await taskAPI.delete(row.id)
    message.success('删除成功')
    loadTasks()
  }
  catch (error: any) {
    message.error(error.message || '删除失败')
  }
}

// 执行任务
async function handleExecute(row: Api.Task) {
  try {
    await taskAPI.execute(row.id)
    message.success('执行成功')
    loadTasks()
  }
  catch (error: any) {
    message.error(error.message || '执行失败')
  }
}

// 查看任务日志
async function handleViewLogs(row: Api.Task) {
  try {
    currentId.value = row.id
    logLoading.value = true
    const { data } = await taskAPI.findLogs(row.id)
    taskLogs.value = data || []
    showLogModal.value = true
  }
  catch (error: any) {
    message.error(error.message || '加载日志失败')
  }
  finally {
    logLoading.value = false
  }
}

// 表格列定义
const columns: DataTableColumns<Api.Task> = [
  { title: '任务名称', key: 'name' },
  { title: 'cron', key: 'cron', width: 150 },
  { title: '源路径', key: 'sourcePath', width: 200, ellipsis: { tooltip: true } },
  { title: '目标路径', key: 'targetPath', width: 200, ellipsis: { tooltip: true } },
  { title: '文件后缀', key: 'fileSuffix', width: 200, render: (row: Api.Task) => {
    return h(NSpace, { size: 'small' }, {
      default: () => row.fileSuffix.split(',').map((suffix: string) =>
        h(NTag, { size: 'small' }, { default: () => suffix }),
      ),
    })
  } },
  { title: '覆盖', key: 'overwrite', width: 80, render: (row: Api.Task) => {
    return h('div', [
      h(
        NSwitch,
        {
          value: row.overwrite,
          size: 'small',
          loading: loading.value,
          onUpdateValue: async (value: boolean) => {
            try {
              await taskAPI.update(row.id, {
                ...row,
                overwrite: value,
              })
              message.success(value ? '已开启覆盖' : '已关闭覆盖')
              loadTasks()
            }
            catch (error: any) {
              message.error(error.message || '操作失败')
            }
          },
        },
      ),
    ])
  } },
  { title: '启用', key: 'enabled', width: 80, render: (row: Api.Task) => {
    return h('div', [
      h(
        NSwitch,
        {
          value: row.enabled,
          size: 'small',
          loading: loading.value,
          onUpdateValue: async (value: boolean) => {
            try {
              await taskAPI.update(row.id, {
                ...row,
                enabled: value,
              })
              message.success(value ? '任务已启用' : '任务已停用')
              loadTasks()
            }
            catch (error: any) {
              message.error(error.message || '操作失败')
            }
          },
        },
      ),
    ])
  } },
  { title: '运行状态', key: 'running', render: (row: Api.Task) => {
    return h('div', { class: 'flex items-center gap-2' }, [
      h(
        NTag,
        {
          type: row.running ? 'success' : 'warning',
          size: 'small',
        },
        { default: () => row.running ? '运行中' : '已停止' },
      ),
    ])
  } },
  { title: '最后运行时间', key: 'lastRunAt', width: 150, render: (row: Api.Task) => {
    return row.lastRunAt
      ? h(NTime, { time: new Date(row.lastRunAt), type: 'datetime' })
      : h(NText, { depth: 3 }, { default: () => '从未运行' })
  } },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    render: (row) => {
      return h(NSpace, {}, {
        default: () => [
          h(NButton, {
            size: 'small',
            onClick: () => handleEdit(row),
          }, { default: () => '编辑' }),
          h(NButton, {
            size: 'small',
            type: 'info',
            onClick: () => handleCopy(row),
          }, { default: () => '复制' }),
          h(
            NButton,
            {
              type: 'warning',
              size: 'small',
              onClick: () => handleExecute(row),
              disabled: row.running,
            },
            { default: () => row.running ? '执行中' : '执行' },
          ),
          h(NButton, {
            size: 'small',
            type: 'info',
            onClick: () => handleViewLogs(row),
          }, { default: () => '日志' }),
          h(NPopconfirm, {
            onPositiveClick: () => handleDelete(row),
          }, {
            default: () => '确认删除该任务吗？',
            trigger: () => h(NButton, {
              size: 'small',
              type: 'error',
            }, { default: () => '删除' }),
          }),
        ],
      })
    },
  },
]

// 日志表格列定义
const logColumns: DataTableColumns<Api.TaskLog> = [
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const statusMap = {
        running: { type: 'info', text: '运行中' },
        completed: { type: 'success', text: '已完成' },
        failed: { type: 'error', text: '失败' },
        stopped: { type: 'warning', text: '已停止' },
      }
      const status = statusMap[row.status as keyof typeof statusMap] || { type: 'default', text: row.status }
      return h(NTag, { type: status.type as any, size: 'small' }, { default: () => status.text })
    },
  },
  {
    title: '开始时间',
    key: 'startTime',
    width: 180,
    render: row => h(NTime, { time: new Date(row.startTime), type: 'datetime' }),
  },
  {
    title: '结束时间',
    key: 'endTime',
    width: 180,
    render: row => row.endTime
      ? h(NTime, { time: new Date(row.endTime), type: 'datetime' })
      : h(NText, { depth: 3 }, { default: () => '-' }),
  },
  {
    title: '消息',
    key: 'message',
    render: row => h(NText, { type: row.status === 'failed' ? 'error' : undefined }, { default: () => row.message || '-' }),
  },
]

// 定时刷新日志
let logTimer: NodeJS.Timeout | number | null = null
watch(showLogModal, (show) => {
  if (show && currentId.value) {
    // 每5秒刷新一次日志
    logTimer = setInterval(async () => {
      try {
        const { data } = await taskAPI.findLogs(currentId.value!)
        taskLogs.value = data || []
      }
      catch (error) {
        console.error('刷新日志失败:', error)
      }
    }, 5000)
  }
  else {
    if (logTimer) {
      clearInterval(logTimer)
      logTimer = null
    }
  }
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (logTimer) {
    clearInterval(logTimer)
    logTimer = null
  }
})

// 初始化加载
onMounted(() => {
  loadTasks()
})
</script>

<template>
  <NSpin :show="loading">
    <NCard title="任务管理">
      <!-- 搜索工具栏 -->
      <NSpace vertical :size="12">
        <NSpace>
          <NInput
            v-model:value="searchForm.name"
            placeholder="请输入任务名称搜索"
            @keydown.enter="loadTasks"
          />
          <NButton type="primary" @click="loadTasks">
            搜索
          </NButton>
          <NButton @click="handleCreate">
            新建任务
          </NButton>
        </NSpace>

        <!-- 任务列表 -->
        <NDataTable
          :columns="columns"
          :data="tasks"
        />
      </NSpace>
    </NCard>

    <!-- 创建/编辑对话框 -->
    <NModal
      v-model:show="showModal"
      :title="isEdit ? '编辑任务' : '创建任务'"
      preset="card"
      :close-on-esc="false"
      :mask-closable="false"
      :style="{ width: '800px' }"
    >
      <NForm
        ref="formRef"
        :model="formModel"
        :rules="rules"
        label-placement="left"
        label-width="100"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="任务名称" path="name">
          <NInput v-model:value="formModel.name" placeholder="请输入任务名称">
            <template #prefix>
              <div class="i-ri:calendar-schedule-fill" />
            </template>
          </NInput>
        </NFormItem>
        <NFormItem label="源路径" path="sourcePath">
          <NInput v-model:value="formModel.sourcePath" placeholder="请输入源路径">
            <template #prefix>
              <div class="i-material-symbols:folder-open-outline-sharp" />
            </template>
          </NInput>
        </NFormItem>
        <NFormItem label="目标路径" path="targetPath">
          <NInput v-model:value="formModel.targetPath" placeholder="请输入目标路径">
            <template #prefix>
              <div class="i-material-symbols:folder-open-outline-sharp" />
            </template>
          </NInput>
        </NFormItem>
        <NFormItem label="文件后缀" path="fileSuffix">
          <NInput v-model:value="formModel.fileSuffix" placeholder="请输入文件后缀,多个用逗号分隔">
            <template #prefix>
              <div class="i-material-symbols:video-file-rounded" />
            </template>
          </NInput>
        </NFormItem>
        <NFormItem label="cron表达式" path="cron">
          <NInput v-model:value="formModel.cron" placeholder="*/5 * * * *">
            <template #prefix>
              <div class="i-carbon-time" />
            </template>
          </NInput>
        </NFormItem>

        <div class="flex justify-start">
          <NFormItem label="覆盖生成" path="overwrite">
            <NSwitch v-model:value="formModel.overwrite" />
          </NFormItem>
          <NFormItem label="是否启用" path="enabled">
            <NSwitch v-model:value="formModel.enabled" />
          </NFormItem>
        </div>
      </NForm>
      <template #footer>
        <NSpace justify="end">
          <NButton @click="showModal = false">
            取消
          </NButton>
          <NButton type="primary" @click="handleSave">
            确定
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 日志查看对话框 -->
    <NModal
      v-model:show="showLogModal"
      title="任务日志"
      preset="card"
      :style="{ width: '800px' }"
    >
      <NSpin :show="logLoading">
        <NDataTable
          :columns="logColumns"
          :data="taskLogs"
          :pagination="{ pageSize: 10 }"
        />
      </NSpin>
    </NModal>
  </NSpin>
</template>
