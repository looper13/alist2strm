<script setup lang="ts">
import { ref } from 'vue'
import { useDialog, useLoadingBar, useMessage, useNotification } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import cronValidate from 'cron-validate'
import { taskService } from '@/services/task'
import type { Task, TaskLog } from '@/types/task'

const tasks = ref<Task[]>([])
const currentTask = ref<Task | null>(null)
const taskLogs = ref<TaskLog[]>([])
const showTaskForm = ref(false)
const showLogs = ref(false)
const formRef = ref<FormInst | null>(null)
const isLoading = ref(false)

const message = useMessage()
const dialog = useDialog()
const notification = useNotification()
const loadingBar = useLoadingBar()

const taskForm = ref({
  name: '',
  sourcePath: '',
  targetPath: '',
  fileSuffix: 'mp4,mkv,avi',
  overwrite: false,
  cronExpression: '',
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

// 表单验证规则
const rules: FormRules = {
  name: [
    { required: true, message: '请输入任务名称', trigger: 'blur' },
    { min: 2, max: 50, message: '任务名称长度应在 2-50 个字符之间', trigger: 'blur' },
  ],
  sourcePath: [
    { required: true, message: '请输入AList路径', trigger: 'blur' },
    {
      validator: (rule, value) => {
        if (!value.startsWith('/'))
          return new Error('路径必须以 / 开头')
        return true
      },
      trigger: 'blur',
    },
  ],
  targetPath: [
    { required: true, message: '请输入目标路径', trigger: 'blur' },
    {
      validator: (rule, value) => {
        if (!value.startsWith('/'))
          return new Error('路径必须以 / 开头')
        return true
      },
      trigger: 'blur',
    },
  ],
  fileSuffix: [
    { required: true, message: '请输入文件后缀', trigger: 'blur' },
    { validator: fileSuffixValidator, trigger: 'blur' },
  ],
  cronExpression: [
    {
      validator: (rule, value) => {
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

// 加载任务列表
async function loadTasks() {
  loadingBar.start()
  isLoading.value = true
  try {
    tasks.value = await taskService.getTasks()
  }
  catch {
    message.error('加载任务列表失败')
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 创建任务
async function createTask() {
  if (!formRef.value)
    return

  try {
    await formRef.value.validate()
    loadingBar.start()
    isLoading.value = true
    await taskService.createTask(taskForm.value)
    showTaskForm.value = false
    await loadTasks()
    taskForm.value = {
      name: '',
      sourcePath: '',
      targetPath: '',
      fileSuffix: 'mp4,mkv,avi',
      overwrite: false,
      cronExpression: '',
    }
    message.success('创建任务成功')
  }
  catch (err) {
    if (err instanceof Error) {
      message.error(err.message)
    }
    else {
      message.error('创建任务失败')
    }
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 删除任务
async function deleteTask(id: number) {
  dialog.warning({
    title: '确认删除',
    content: '确定要删除这个任务吗？此操作不可恢复。',
    positiveText: '确定删除',
    negativeText: '取消',
    type: 'warning',
    onPositiveClick: async () => {
      loadingBar.start()
      isLoading.value = true
      try {
        await taskService.deleteTask(id)
        await loadTasks()
        message.success('删除任务成功')
      }
      catch {
        message.error('删除任务失败')
      }
      finally {
        loadingBar.finish()
        isLoading.value = false
      }
    },
  })
}

// 执行任务
async function executeTask(id: number) {
  loadingBar.start()
  isLoading.value = true
  try {
    await taskService.executeTask(id)
    notification.success({
      title: '执行成功',
      content: '任务已开始执行',
      duration: 3000,
    })
  }
  catch (error) {
    notification.error({
      title: '执行失败',
      content: (error as Error).message,
      duration: 5000,
    })
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 查看日志
async function viewLogs(task: Task) {
  loadingBar.start()
  isLoading.value = true
  try {
    currentTask.value = task
    taskLogs.value = await taskService.getTaskLogs(task.id)
    showLogs.value = true
  }
  catch {
    message.error('加载日志失败')
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 切换任务状态
async function toggleTaskStatus(task: Task) {
  loadingBar.start()
  isLoading.value = true
  try {
    await taskService.updateTask(task.id, { enabled: !task.enabled })
    await loadTasks()
    message.success(task.enabled ? '任务已停用' : '任务已启用')
  }
  catch {
    message.error('更新任务状态失败')
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 初始加载
loadTasks()
</script>

<template>
  <n-card>
    <template #header>
      <div class="flex items-center justify-between">
        <div class="text-2xl font-bold">
          任务管理
        </div>
        <n-button type="primary" :loading="isLoading" @click="showTaskForm = true">
          添加任务
        </n-button>
      </div>
    </template>

    <!-- 任务列表 -->
    <n-space vertical size="large">
      <n-spin :show="isLoading">
        <n-empty v-if="tasks.length === 0" description="暂无任务">
          <template #extra>
            <n-button type="primary" :loading="isLoading" @click="showTaskForm = true">
              创建第一个任务
            </n-button>
          </template>
        </n-empty>

        <n-card v-for="task in tasks" v-else :key="task.id" class="w-full">
          <template #header>
            <div class="flex items-center justify-between">
              <n-space align="center">
                <span class="text-lg font-semibold">{{ task.name }}</span>
                <n-tag :type="task.enabled ? 'success' : 'warning'" size="small">
                  {{ task.enabled ? '已启用' : '已停用' }}
                </n-tag>
              </n-space>
              <n-space>
                <n-button
                  :type="task.enabled ? 'warning' : 'success'"
                  :loading="isLoading"
                  @click="toggleTaskStatus(task)"
                >
                  {{ task.enabled ? '停用' : '启用' }}
                </n-button>
                <n-button type="info" :loading="isLoading" @click="executeTask(task.id)">
                  执行
                </n-button>
                <n-button type="primary" :loading="isLoading" @click="viewLogs(task)">
                  日志
                </n-button>
                <n-popconfirm
                  positive-text="删除"
                  negative-text="取消"
                  @positive-click="deleteTask(task.id)"
                >
                  <template #trigger>
                    <n-button type="error" :loading="isLoading">
                      删除
                    </n-button>
                  </template>
                  确定要删除这个任务吗？
                </n-popconfirm>
              </n-space>
            </div>
          </template>

          <n-descriptions :column="2" bordered>
            <n-descriptions-item label="Cron 表达式">
              <n-tag v-if="task.cronExpression" type="info">
                {{ task.cronExpression }}
              </n-tag>
              <n-text v-else depth="3">
                未设置
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item label="AList路径">
              <n-text code>
                {{ task.sourcePath }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item label="目标路径">
              <n-text code>
                {{ task.targetPath }}
              </n-text>
            </n-descriptions-item>
            <n-descriptions-item label="文件后缀">
              <n-space>
                <n-tag v-for="suffix in task.fileSuffix.split(',')" :key="suffix" size="small">
                  {{ suffix }}
                </n-tag>
              </n-space>
            </n-descriptions-item>
            <n-descriptions-item label="覆盖文件">
              <n-switch v-model:value="task.overwrite" disabled />
            </n-descriptions-item>
            <n-descriptions-item v-if="task.lastRunAt" label="上次运行">
              <n-time :time="new Date(task.lastRunAt)" type="datetime" />
            </n-descriptions-item>
          </n-descriptions>
        </n-card>
      </n-spin>
    </n-space>

    <!-- 添加/编辑任务表单 -->
    <n-modal v-model:show="showTaskForm" preset="card" title="添加任务" style="max-width: 600px">
      <n-spin :show="isLoading">
        <n-form
          ref="formRef"
          :model="taskForm"
          :rules="rules"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
          @submit.prevent="createTask"
        >
          <n-form-item label="名称" path="name">
            <n-input v-model:value="taskForm.name" placeholder="请输入任务名称">
              <template #prefix>
                <div class="i-carbon-task" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="AList路径" path="sourcePath">
            <n-input v-model:value="taskForm.sourcePath" placeholder="请输入AList路径">
              <template #prefix>
                <div class="i-carbon-folder" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="目标路径" path="targetPath">
            <n-input v-model:value="taskForm.targetPath" placeholder="请输入目标路径">
              <template #prefix>
                <div class="i-carbon-folder" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="文件后缀" path="fileSuffix">
            <n-input v-model:value="taskForm.fileSuffix" placeholder="请输入文件后缀，用逗号分隔">
              <template #prefix>
                <div class="i-carbon-document" />
              </template>
            </n-input>
          </n-form-item>
          <n-form-item label="Cron 表达式" path="cronExpression">
            <n-input v-model:value="taskForm.cronExpression" placeholder="*/5 * * * *">
              <template #prefix>
                <div class="i-carbon-time" />
              </template>
            </n-input>
            <template #feedback>
              留空表示不设置定时任务
            </template>
          </n-form-item>
          <n-form-item label="覆盖文件">
            <n-switch v-model:value="taskForm.overwrite">
              <template #checked>
                是
              </template>
              <template #unchecked>
                否
              </template>
            </n-switch>
          </n-form-item>
          <div class="flex justify-end space-x-2">
            <n-button :disabled="isLoading" @click="showTaskForm = false">
              取消
            </n-button>
            <n-button type="primary" attr-type="submit" :loading="isLoading">
              保存
            </n-button>
          </div>
        </n-form>
      </n-spin>
    </n-modal>

    <!-- 日志查看 -->
    <n-modal v-model:show="showLogs" preset="card" style="width: 800px">
      <template #header>
        <div class="flex items-center justify-between">
          <n-space align="center">
            <span class="text-lg font-semibold">任务日志</span>
            <n-tag>{{ currentTask?.name }}</n-tag>
          </n-space>
        </div>
      </template>
      <n-spin :show="isLoading">
        <n-scrollbar style="max-height: 500px">
          <n-empty v-if="taskLogs.length === 0" description="暂无日志" />
          <n-timeline v-else>
            <n-timeline-item
              v-for="log in taskLogs"
              :key="log.id"
              :type="log.status === 'success' ? 'success' : 'error'"
              :title="log.status"
            >
              <n-card size="small" :class="log.status === 'success' ? 'bg-green-50' : 'bg-red-50'">
                <n-space vertical>
                  <n-space align="center">
                    <div class="i-carbon-time" />
                    <n-time :time="new Date(log.startTime)" type="datetime" />
                  </n-space>
                  <n-space v-if="log.endTime" align="center">
                    <div class="i-carbon-checkmark" />
                    <n-time :time="new Date(log.endTime)" type="datetime" />
                  </n-space>
                  <n-space v-if="log.status === 'success'" justify="space-around">
                    <n-statistic label="总文件数" :value="log.totalFiles || 0" />
                    <n-statistic label="生成文件数" :value="log.generatedFiles || 0" />
                    <n-statistic label="跳过文件数" :value="log.skippedFiles || 0" />
                  </n-space>
                  <div v-if="log.error" class="text-red-600">
                    <n-alert type="error" :title="log.error" />
                  </div>
                </n-space>
              </n-card>
            </n-timeline-item>
          </n-timeline>
        </n-scrollbar>
      </n-spin>
    </n-modal>
  </n-card>
</template>
