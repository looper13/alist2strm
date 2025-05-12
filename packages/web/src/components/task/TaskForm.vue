<script setup lang="ts">
import { ref } from 'vue'
import type { FormInst, FormRules } from 'naive-ui'
import cronValidate from 'cron-validate'
import type { Task } from '@/types/task'

const props = defineProps<{
  loading?: boolean
  isEditing?: boolean
  editingTask?: Task
}>()

const emit = defineEmits<{
  submit: [formData: any]
  cancel: []
}>()

const formRef = ref<FormInst | null>(null)

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

// 初始化编辑数据
if (props.isEditing && props.editingTask) {
  taskForm.value = {
    name: props.editingTask.name,
    sourcePath: props.editingTask.sourcePath,
    targetPath: props.editingTask.targetPath,
    fileSuffix: props.editingTask.fileSuffix,
    overwrite: props.editingTask.overwrite,
    cronExpression: props.editingTask.cronExpression || '',
  }
}

// 提交表单
async function handleSubmit() {
  if (!formRef.value)
    return

  try {
    await formRef.value.validate()
    emit('submit', taskForm.value)
  }
  catch (err) {
    // 验证失败由父组件处理
    console.error('表单验证失败:', err)
  }
}
</script>

<template>
  <n-form
    ref="formRef"
    :model="taskForm"
    :rules="rules"
    label-placement="left"
    label-width="auto"
    require-mark-placement="right-hanging"
    @submit.prevent="handleSubmit"
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
      <n-button :disabled="loading" @click="emit('cancel')">
        取消
      </n-button>
      <n-button type="primary" attr-type="submit" :loading="loading">
        {{ isEditing ? '更新' : '保存' }}
      </n-button>
    </div>
  </n-form>
</template>
