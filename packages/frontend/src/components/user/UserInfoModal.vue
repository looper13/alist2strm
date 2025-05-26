<script setup lang="ts">
import { useMessage } from 'naive-ui'
import { authAPI } from '~/api/auth'
import { useAuth } from '~/composables/auth'

defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'update:show', value: boolean): void
}>()

const message = useMessage()
const { userInfo, refreshUserInfo } = useAuth()

const formRef = ref()
const loading = ref(false)
const formValue = ref({
  nickname: '',
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// 监听用户信息变化
watch(() => userInfo.value, (newUserInfo) => {
  if (newUserInfo) {
    formValue.value.nickname = newUserInfo.nickname || ''
  }
}, { immediate: true })

// 表单验证规则
const rules = {
  nickname: {
    required: false,
    message: '请输入昵称',
    trigger: 'blur',
  },
  oldPassword: {
    required: true,
    message: '请输入原密码',
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (formValue.value.newPassword && !value)
        return new Error('修改密码时需要输入原密码')
      return true
    },
  },
  newPassword: {
    required: false,
    message: '请输入新密码',
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (value && value.length < 6)
        return new Error('密码长度不能小于6位')
      return true
    },
  },
  confirmPassword: {
    required: false,
    message: '请确认新密码',
    trigger: 'blur',
    validator: (_rule: any, value: string) => {
      if (formValue.value.newPassword && !value)
        return new Error('请确认新密码')
      if (value !== formValue.value.newPassword)
        return new Error('两次输入的密码不一致')
      return true
    },
  },
}

// 提交表单
async function handleSubmit() {
  try {
    loading.value = true
    await formRef.value?.validate()

    // 构建更新参数
    const params: Api.Auth.UpdateUserParams = {}
    if (formValue.value.nickname !== userInfo.value?.nickname)
      params.nickname = formValue.value.nickname
    if (formValue.value.newPassword) {
      params.oldPassword = formValue.value.oldPassword
      params.password = formValue.value.newPassword
    }

    // 如果没有任何修改，直接关闭
    if (Object.keys(params).length === 0) {
      emit('update:show', false)
      return
    }

    await authAPI.updateCurrentUser(params)
    await refreshUserInfo() // 更新成功后刷新用户信息
    message.success('更新成功')
    emit('update:show', false)
  }
  catch (error: any) {
    message.error((error?.response?.data?.message || error?.message) || '更新失败')
  }
  finally {
    loading.value = false
  }
}

// 关闭弹窗时重置表单
function handleClose() {
  formRef.value?.restoreValidation()
  formValue.value = {
    nickname: userInfo.value?.nickname || '',
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  }
}
</script>

<template>
  <NModal
    :show="show"
    preset="card"
    title="个人信息"
    :style="{ width: '90%', maxWidth: '500px' }"
    @update:show="emit('update:show', $event)"
    @after-leave="handleClose"
  >
    <NForm
      ref="formRef"
      :model="formValue"
      :rules="rules"
      label-placement="left"
      label-width="80"
      require-mark-placement="right-hanging"
    >
      <NFormItem label="用户名" path="username">
        <NInput
          :value="userInfo?.username"
          disabled
          placeholder="用户名"
        />
      </NFormItem>

      <NFormItem label="昵称" path="nickname">
        <NInput
          v-model:value="formValue.nickname"
          placeholder="请输入昵称"
        />
      </NFormItem>

      <NDivider>修改密码（选填）</NDivider>

      <NFormItem label="原密码" path="oldPassword">
        <NInput
          v-model:value="formValue.oldPassword"
          type="password"
          show-password-on="click"
          placeholder="请输入原密码"
        />
      </NFormItem>

      <NFormItem label="新密码" path="newPassword">
        <NInput
          v-model:value="formValue.newPassword"
          type="password"
          show-password-on="click"
          placeholder="请输入新密码"
        />
      </NFormItem>

      <NFormItem label="确认密码" path="confirmPassword">
        <NInput
          v-model:value="formValue.confirmPassword"
          type="password"
          show-password-on="click"
          placeholder="请再次输入新密码"
        />
      </NFormItem>
    </NForm>

    <template #footer>
      <div class="flex gap-2 justify-end">
        <NButton @click="emit('update:show', false)">
          取消
        </NButton>
        <NButton
          type="primary"
          :loading="loading"
          @click="handleSubmit"
        >
          保存
        </NButton>
      </div>
    </template>
  </NModal>
</template>
