<script setup lang="ts">
import type { DataTableColumns, FormRules } from 'naive-ui'
import cronValidate from 'cron-validate'
import { configAPI } from '~/api/config'
import { taskAPI } from '~/api/task'
import { taskLogAPI } from '~/api/task-log'

const { isMobile } = useMobile()
// 状态定义
const loading = ref(false)
const showModal = ref(false)
const showLogDrawer = ref(false)
const isEdit = ref(false)
const currentId = ref<number | null>(null)
const tasks = ref<Api.Task.Record[]>([])
const taskLogs = ref<Api.Task.Log[]>([])
const logLoading = ref(false)
const strmConfig = ref<Api.Config.StrmConfig | null>(null)

// 加载 STRM 配置
async function loadStrmConfig() {
  try {
    const { data } = await configAPI.getByCode('STRM')
    if (data && data.value) {
      strmConfig.value = JSON.parse(data.value) as Api.Config.StrmConfig
    }
  }
  catch (error: any) {
    console.error('加载 STRM 配置失败', error)
  }
}

// 日志分页
const logPagination = reactive({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 50],
  onChange: (page: number) => {
    logPagination.page = page
    loadTaskLogs()
  },
  onUpdatePageSize: (pageSize: number) => {
    logPagination.pageSize = pageSize
    logPagination.page = 1
    loadTaskLogs()
  },
})

// 搜索
const searchForm = reactive({
  name: '',
})

// 表单实例和数据
const formRef = ref<any>(null)
const formModel = ref<Api.Task.Create>({
  name: '',
  mediaType: 'movie',
  sourcePath: '',
  targetPath: '',
  fileSuffix: '',
  overwrite: false,
  enabled: true,
  cron: '',

  downloadMetadata: false,
  metadataExtensions: '',
  downloadSubtitle: false,
  subtitleExtensions: '',
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
    // { required: true, message: '请输入Cron表达式', trigger: 'blur' },
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
async function handleCreate() {
  isEdit.value = false
  currentId.value = null
  
  // 如果还没有加载 STRM 配置，先加载一次
  if (!strmConfig.value) {
    await loadStrmConfig()
  }
  
  formModel.value = {
    name: '',
    mediaType: 'movie',
    sourcePath: '',
    targetPath: '',
    // 如果有 STRM 配置，自动填充默认后缀
    fileSuffix: strmConfig.value?.defaultSuffix || '',
    overwrite: false,
    enabled: true,
    cron: '',
    downloadMetadata: false,
    metadataExtensions: '',
    downloadSubtitle: false,
    subtitleExtensions: '',
  }
  showModal.value = true
}

// 打开编辑对话框
async function handleEdit(row: Api.Task.Record) {
  try {
    isEdit.value = true
    currentId.value = row.id
    formModel.value = {
      mediaType: row.mediaType || 'movie',
      name: row.name,
      sourcePath: row.sourcePath,
      targetPath: row.targetPath,
      fileSuffix: row.fileSuffix,
      overwrite: row.overwrite,
      enabled: row.enabled,
      cron: row.cron || '',
      downloadMetadata: row.downloadMetadata || false,
      metadataExtensions: row.metadataExtensions || '',
      downloadSubtitle: row.downloadSubtitle || false,
      subtitleExtensions: row.subtitleExtensions || '',
    }
    showModal.value = true
  }
  catch (error: any) {
    message.error(error.message || '加载失败')
  }
}

// 打开复制新增对话框
function handleCopy(row: Api.Task.Record) {
  isEdit.value = false
  currentId.value = null
  formModel.value = {
    mediaType: row.mediaType || 'movie',
    name: `${row.name}_复制`,
    sourcePath: row.sourcePath,
    targetPath: row.targetPath,
    fileSuffix: row.fileSuffix,
    overwrite: row.overwrite,
    enabled: row.enabled,
    cron: row.cron || '',
    downloadMetadata: row.downloadMetadata || false,
    metadataExtensions: row.metadataExtensions || '',
    downloadSubtitle: row.downloadSubtitle || false,
    subtitleExtensions: row.subtitleExtensions || '',
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
        cron: formModel.value.cron,
        mediaType: formModel.value.mediaType,
        downloadMetadata: formModel.value.downloadMetadata,
        metadataExtensions: formModel.value.metadataExtensions,
        downloadSubtitle: formModel.value.downloadSubtitle,
        subtitleExtensions: formModel.value.subtitleExtensions,
      } as Api.Task.Update)
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
async function handleDelete(row: Api.Task.Record) {
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
async function handleExecute(row: Api.Task.Record) {
  try {
    await taskAPI.execute(row.id)
    message.success('执行成功')
    loadTasks()
  }
  catch (error: any) {
    message.error(error.message || '执行失败')
  }
}

async function handleUpdateEnabled(item: Api.Task.Record) {
  try {
    await taskAPI.update(item.id, {
      ...item,
    })
    message.success(item.enabled ? '任务已启用' : '任务已停用')
    loadTasks()
  }
  catch (error: any) {
    message.error(error.message || '操作失败')
  }
}

async function handleReset(row: Api.Task.Record) {
  try {
    await taskAPI.resetStatus(row.id)
    message.success('重置成功')
    loadTasks()
  }
  catch (error: any) {
    message.error(error.message || '重置失败')
  }
}

// 查看任务日志
async function handleViewLogs(row: Api.Task.Record) {
  try {
    currentId.value = row.id
    logLoading.value = true
    logPagination.page = 1
    await loadTaskLogs()
    showLogDrawer.value = true
  }
  catch (error: any) {
    message.error(error.message || '加载日志失败')
  }
  finally {
    logLoading.value = false
  }
}

// 加载任务日志
async function loadTaskLogs() {
  if (!currentId.value)
    return
  try {
    logLoading.value = true
    const { data } = await taskLogAPI.findLogs({
      taskId: currentId.value,
      page: logPagination.page,
      pageSize: logPagination.pageSize,
      sortBy: 'updatedAt',
      sortOrder: 'desc',
    })
    if (data && data.list) {
      taskLogs.value = data.list || []
      logPagination.itemCount = data.total || 0
    }
  }
  catch (error: any) {
    message.error(error.message || '加载日志失败')
  }
  finally {
    logLoading.value = false
  }
}
// 日志表格列定义
const logColumns: DataTableColumns<Api.Task.Log> = [
  {
    title: '状态',
    key: 'status',
    width: isMobile ? 70 : 100,
    render: (row) => {
      const statusMap = {
        running: { type: 'info', text: '运行中' },
        completed: { type: 'success', text: '已完成' },
        failed: { type: 'error', text: '失败' },
        stopped: { type: 'warning', text: '已停止' },
      }
      const status = statusMap[row.status as keyof typeof statusMap] || { type: 'default', text: row.status }
      return h(NTag, { type: status.type as any, size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => status.text })
    },
  },
  {
    title: '开始时间',
    key: 'startTime',
    width: isMobile ? 150 : 180,
    render: row => h(NTime, { time: new Date(row.startTime), type: 'datetime', format: isMobile ? 'MM-dd HH:mm' : 'yyyy-MM-dd HH:mm:ss' }),
  },
  {
    title: '结束时间',
    key: 'endTime',
    width: isMobile ? 150 : 180,
    render: row => row.endTime
      ? h(NTime, { time: new Date(row.endTime), type: 'datetime', format: isMobile ? 'MM-dd HH:mm' : 'yyyy-MM-dd HH:mm:ss' })
      : h(NText, { depth: 3 }, { default: () => '-' }),
  },
  {
    title: '耗时(秒)',
    key: 'duration',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NText, { depth: 2 }, { default: () => row.duration || '-' })
    },
  },
  {
    title: '总数',
    key: 'totalFile',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NTag, { type: 'info', size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => row.totalFile })
    },
  },
  {
    title: '已生成',
    key: 'generatedFile',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NTag, { type: 'success', size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => row.generatedFile })
    },
  },
  {
    title: '已跳过',
    key: 'skipFile',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NTag, { type: 'warning', size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => row.skipFile })
    },
  },
  {
    title: '元数据',
    key: 'metadataCount',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NTag, { type: 'info', size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => row.metadataCount || 0 })
    },
  },
  {
    title: '字幕',
    key: 'subtitleCount',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NTag, { type: 'info', size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => row.subtitleCount || 0 })
    },
  },
  {
    title: () => {
      return h(
        'div',
        {
          style: 'display: flex; align-items: center; justify-content: end; width: 100%;',
        },
        [
          '失败',
          h(
            NTooltip,
            { trigger: 'hover' },
            {
              trigger: () => h('div', { class: 'i-ri:question-line ml-1', style: 'cursor: pointer; font-size: 14px;' }),
              default: () => '失败文件数仅供参考，特别是在任务执行中途失败时可能不准确',
            },
          ),
        ],
      )
    },
    key: 'failedCount',
    width: isMobile ? 70 : 100,
    align: 'right',
    render: (row) => {
      return h(NTag, { type: 'error', size: 'small', style: isMobile ? 'padding: 0 4px;' : '' }, { default: () => row.failedCount || 0 })
    },
  },
  {
    title: '消息',
    key: 'message',
    width: isMobile ? 200 : 300,
    ellipsis: {
      tooltip: true,
    },
    render: (row) => {
      const msgText = row.message || '-'
      return h(NText, {
        type: row.status === 'failed' ? 'error' : undefined,
        depth: msgText === '-' ? 3 : undefined,
      }, { default: () => msgText })
    },
  },
]

// 初始化加载
onMounted(() => {
  loadTasks()
  loadStrmConfig()
})
</script>

<template>
  <NSpin :show="loading">
    <NCard title="任务管理">
      <!-- 搜索工具栏 -->
      <NSpace vertical :size="12">
        <div class="flex flex-col gap-2 sm:flex-row">
          <NSpace :wrap="false" class="w-full sm:w-auto">
            <NInput
              v-model:value="searchForm.name"
              size="small"
              placeholder="请输入任务名称搜索"
              class="w-full sm:w-[200px]"
              @keydown.enter="loadTasks"
            >
              <template #prefix>
                <div class="i-ri:search-line" />
              </template>
            </NInput>
            <NSpace :wrap="false">
              <!-- <NButton @click="loadTasks">
                搜索
              </NButton> -->
              <NButton
                size="small"
                type="primary"
                @click="loadTasks"
              >
                <template #icon>
                  <div class="i-ri:search-line" />
                </template>
              </NButton>
              <NButton size="small" type="info" @click="handleCreate">
                <template #icon>
                  <div class="i-ri:add-line" />
                </template>
              </NButton>
            </NSpace>
          </NSpace>
        </div>

        <!-- 任务列表 -->
        <div class="gap-4 grid grid-cols-1 lg:grid-cols-2 sm:grid-cols-1 xl:grid-cols-4">
          <template
            v-for="task in tasks"
            :key="task.id"
          >
            <TaskItem
              :item="task"
              @edit="handleEdit"
              @copy="handleCopy"
              @delete="handleDelete"
              @execute="handleExecute"
              @logs="handleViewLogs"
              @update:enabled="handleUpdateEnabled"
              @reset="handleReset"
            />
          </template>
        </div>
      </NSpace>
    </NCard>

    <!-- 创建/编辑对话框 -->
    <NModal
      v-model:show="showModal"
      :title="isEdit ? '编辑任务' : '创建任务'"
      preset="card"
      :close-on-esc="false"
      :mask-closable="false"
      :style="{ width: isMobile ? '100%' : '800px' }"
    >
      <NForm
        ref="formRef"
        :model="formModel"
        :rules="rules"
        :label-placement="isMobile ? 'top' : 'left'"
        :label-width="isMobile ? 'auto' : '120'"
        require-mark-placement="right-hanging"
        :size="isMobile ? 'small' : 'medium'"
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
        <NFormItem label="媒体类型" path="mediaType">
          <NRadioGroup v-model:value="formModel.mediaType">
            <NSpace>
              <NRadio value="movie">
                电影
              </NRadio>
              <NRadio value="tv">
                电视剧
              </NRadio>
            </NSpace>
          </NRadioGroup>
        </NFormItem>

        <NFormItem label="cron表达式" path="cron">
          <NInput v-model:value="formModel.cron" placeholder="*/5 * * * *">
            <template #prefix>
              <div class="i-carbon-time" />
            </template>
          </NInput>
        </NFormItem>

        <div class="space-y-4">
          <div :class="isMobile ? 'space-y-4' : 'flex justify-start space-x-8'">
            <NFormItem :label-width="isMobile ? 'auto' : '120'" label="覆盖生成" path="overwrite">
              <div class="flex items-center space-x-2">
                <NSwitch v-model:value="formModel.overwrite" />
                <span class="text-sm text-gray-500">{{ formModel.overwrite ? '是' : '否' }}</span>
              </div>
            </NFormItem>
            <NFormItem :label-width="isMobile ? 'auto' : '100'" label="是否启用" path="enabled">
              <div class="flex items-center space-x-2">
                <NSwitch v-model:value="formModel.enabled" />
                <span class="text-sm text-gray-500">{{ formModel.enabled ? '是' : '否' }}</span>
              </div>
            </NFormItem>
          </div>

          <div :class="isMobile ? 'space-y-4' : 'flex justify-start space-x-8'">
            <NFormItem :label-width="isMobile ? 'auto' : '120'" label="下载元数据" path="downloadMetadata">
              <div class="flex items-center space-x-2">
                <NSwitch v-model:value="formModel.downloadMetadata" />
                <span class="text-sm text-gray-500">{{ formModel.downloadMetadata ? '是' : '否' }}</span>
              </div>
            </NFormItem>
            <NFormItem :label-width="isMobile ? 'auto' : '100'" label="下载字幕" path="downloadSubtitle">
              <div class="flex items-center space-x-2">
                <NSwitch v-model:value="formModel.downloadSubtitle" />
                <span class="text-sm text-gray-500">{{ formModel.downloadSubtitle ? '是' : '否' }}</span>
              </div>
            </NFormItem>
          </div>

          <NFormItem
            v-if="formModel.downloadMetadata"
            label="元数据扩展名"
            path="metadataExtensions"
          >
            <NInput
              v-model:value="formModel.metadataExtensions"
              placeholder=".nfo,.jpg,.png"
            >
              <template #prefix>
                <div class="i-ri:information-line" />
              </template>
            </NInput>
          </NFormItem>

          <NFormItem
            v-if="formModel.downloadSubtitle"
            label="字幕扩展名"
            path="subtitleExtensions"
          >
            <NInput
              v-model:value="formModel.subtitleExtensions"
              placeholder=".srt,.ass,.ssa"
            >
              <template #prefix>
                <div class="i-ri:subtitle-line" />
              </template>
            </NInput>
          </NFormItem>
        </div>
      </NForm>
      <template #footer>
        <NSpace :justify="isMobile ? 'center' : 'end'" :size="isMobile ? 'large' : 'medium'">
          <NButton :block="isMobile" @click="showModal = false">
            取消
          </NButton>
          <NButton :block="isMobile" type="primary" @click="handleSave">
            确定
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 日志查看抽屉 -->
    <NDrawer
      v-model:show="showLogDrawer"
      :placement="isMobile ? 'right' : 'bottom'"
      :height="isMobile ? '100%' : 600"
      :width="isMobile ? '100%' : 'calc(100% - 48px)'"
      :trap-focus="false"
      :block-scroll="false"
    >
      <NDrawerContent title="任务日志" closable>
        <template #header>
          <div class="flex items-center justify-between">
            <span>任务日志</span>
            <NButton
              type="primary"
              size="small"
              :loading="logLoading"
              @click="loadTaskLogs"
            >
              <template #icon>
                <div class="i-material-symbols:refresh" />
              </template>
              刷新
            </NButton>
          </div>
        </template>
        <div class="flex flex-col h-full">
          <NSpin :show="logLoading">
            <NDataTable
              :columns="logColumns"
              remote
              :scroll-x="1000"
              :max-height="isMobile ? 'calc(100vh - 180px)' : 420"
              :data="taskLogs"
              :pagination="logPagination"
              :row-class-name="() => 'text-sm'"
              size="small"
            />
          </NSpin>
        </div>
      </NDrawerContent>
    </NDrawer>
  </NSpin>
</template>

<route lang="yaml">
name: task
layout: default
path: /admin/task
</route>
