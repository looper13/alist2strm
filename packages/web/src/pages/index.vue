<script setup lang="ts">
import { ref } from 'vue'
import { useDialog, useLoadingBar, useMessage, useNotification } from 'naive-ui'
import type { Task, TaskLog } from '@/types/task'
import { taskService } from '@/services/task'
import TaskTable from '@/components/task/TaskTable.vue'
import TaskForm from '@/components/task/TaskForm.vue'
import TaskLogs from '@/components/task/TaskLogs.vue'

const tasks = ref<Task[]>([])
const currentTask = ref<Task | null>(null)
const taskLogs = ref<TaskLog[]>([])
const showTaskForm = ref(false)
const showLogs = ref(false)
const isLoading = ref(false)
const isEditing = ref(false)
const editingTaskId = ref<number | null>(null)

const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 50],
})

const message = useMessage()
const dialog = useDialog()
const notification = useNotification()
const loadingBar = useLoadingBar()

// 加载任务列表
async function loadTasks() {
  loadingBar.start()
  isLoading.value = true
  try {
    const { records, total } = await taskService.getTasksWithPagination(
      pagination.value.page,
      pagination.value.pageSize,
    )
    tasks.value = records
    pagination.value.itemCount = total
  }
  catch (error) {
    message.error('加载任务列表失败')
    console.error('加载任务列表失败:', error)
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 处理表单提交
async function handleFormSubmit(formData: any) {
  loadingBar.start()
  isLoading.value = true
  try {
    if (isEditing.value && editingTaskId.value) {
      await taskService.updateTask(editingTaskId.value, formData)
      message.success('更新任务成功')
    }
    else {
      await taskService.createTask(formData)
      message.success('创建任务成功')
    }
    showTaskForm.value = false
    resetForm()
    await loadTasks()
  }
  catch (err) {
    if (err instanceof Error) {
      message.error(err.message)
    }
    else {
      message.error(isEditing.value ? '更新任务失败' : '创建任务失败')
    }
  }
  finally {
    loadingBar.finish()
    isLoading.value = false
  }
}

// 重置表单状态
function resetForm() {
  isEditing.value = false
  editingTaskId.value = null
  showTaskForm.value = false
}

// 编辑任务
function handleEdit(task: Task) {
  isEditing.value = true
  editingTaskId.value = task.id
  showTaskForm.value = true
}

// 删除任务
async function handleDelete(id: number) {
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
async function handleExecute(id: number) {
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
async function handleViewLogs(task: Task) {
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
async function handleToggleStatus(task: Task) {
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
    <n-spin :show="isLoading">
      <n-empty v-if="tasks.length === 0" description="暂无任务">
        <template #extra>
          <n-button type="primary" :loading="isLoading" @click="showTaskForm = true">
            创建第一个任务
          </n-button>
        </template>
      </n-empty>

      <TaskTable
        v-else
        :tasks="tasks"
        :loading="isLoading"
        :pagination="pagination"
        @update:page="(page) => { pagination.page = page; loadTasks() }"
        @update:page-size="(pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; loadTasks() }"
        @edit="handleEdit"
        @delete="handleDelete"
        @execute="handleExecute"
        @view-logs="handleViewLogs"
        @toggle-status="handleToggleStatus"
      />
    </n-spin>

    <!-- 添加/编辑任务表单 -->
    <n-modal v-model:show="showTaskForm" preset="card" :title="isEditing ? '编辑任务' : '添加任务'" style="max-width: 600px">
      <TaskForm
        :loading="isLoading"
        :is-editing="isEditing"
        :editing-task="editingTaskId ? tasks.find(t => t.id === editingTaskId) : undefined"
        @submit="handleFormSubmit"
        @cancel="resetForm"
      />
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
      <TaskLogs
        :task="currentTask"
        :logs="taskLogs"
        :loading="isLoading"
      />
    </n-modal>
  </n-card>
</template>
