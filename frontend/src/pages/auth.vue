<!-- 登录/注册页面 -->
<script setup lang="ts">
import { useMessage } from 'naive-ui'
import { useRoute } from 'vue-router'
import { useAuth, useMobile } from '~/composables'
import { isDark, toggleTheme } from '~/composables/dark'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const { isMobile } = useMobile()
const { login, register } = useAuth()

// 当前模式：login 或 register
const mode = ref<'login' | 'register'>('login')

const formRef = ref()
const loading = ref(false)
const formValue = ref({
  username: '',
  password: '',
  nickname: '',
  confirmPassword: '',
})

// 登录表单规则
const loginRules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur',
  },
  password: {
    required: true,
    message: '请输入密码',
    trigger: 'blur',
  },
}

// 注册表单规则
const registerRules = {
  ...loginRules,
  nickname: {
    required: false,
    message: '请输入昵称',
    trigger: 'blur',
  },
  confirmPassword: {
    required: true,
    message: '请确认密码',
    trigger: 'blur',
    validator: (rule: any, value: string) => {
      if (value !== formValue.value.password)
        return new Error('两次输入的密码不一致')
      return true
    },
  },
}

// 切换模式
/* 暂时注释掉切换模式功能
function toggleMode() {
  mode.value = mode.value === 'login' ? 'register' : 'login'
  formRef.value?.restoreValidation()
}
*/

// 提交表单
async function handleSubmit() {
  try {
    loading.value = true
    await formRef.value?.validate()

    if (mode.value === 'login') {
      await login(formValue.value.username, formValue.value.password)
      message.success('登录成功')
      // 如果有重定向地址，则跳转到重定向地址
      const redirect = route.query.redirect as string
      router.push(redirect || '/admin')
    }
    else {
      await register(
        formValue.value.username,
        formValue.value.password,
        formValue.value.nickname || undefined,
      )
      message.success('注册成功')
      // 切换登录，并清空账号密码
      mode.value = 'login'
      formValue.value = {
        username: '',
        password: '',
        nickname: '',
        confirmPassword: '',
      }
    }
  }
  catch (error: any) {
    message.error(error.response?.data?.message || `${mode.value === 'login' ? '登录' : '注册'}失败`)
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="px-4 bg-[#f5f7fa] flex flex-col min-h-screen w-full transition-colors items-center justify-center dark:bg-[#101014]">
    <NCard :class="isMobile ? 'w-full' : 'w-[420px]'" :bordered="false">
      <!-- Logo -->
      <div :class="isMobile ? 'mb-4' : 'mb-8'" class="text-center">
        <div class="i-carbon-media-library text-primary text-4xl mx-auto mb-2 sm:text-5xl sm:mb-4" />
        <h1 class="text-xl font-bold sm:text-2xl">
          AList2Strm
        </h1>
        <p class="text-sm text-gray-500 mt-1 sm:text-base sm:mt-2">
          {{ mode === 'login' ? '欢迎回来' : '创建新账号' }}
        </p>
      </div>

      <!-- 表单 -->
      <NForm
        ref="formRef"
        :model="formValue"
        :rules="mode === 'login' ? loginRules : registerRules"
        :size="isMobile ? 'small' : 'large'"
        label-placement="top"
      >
        <NFormItem :class="isMobile ? 'mb-2' : 'mb-4'" label="用户名" path="username">
          <NInput
            v-model:value="formValue.username"
            placeholder="请输入用户名"
            :autofocus="!isMobile"
          >
            <template #prefix>
              <div class="i-carbon-user" />
            </template>
          </NInput>
        </NFormItem>

        <template v-if="mode === 'register'">
          <NFormItem :class="isMobile ? 'mb-2' : 'mb-4'" label="昵称" path="nickname">
            <NInput
              v-model:value="formValue.nickname"
              placeholder="请输入昵称（选填）"
            >
              <template #prefix>
                <div class="i-carbon-identification" />
              </template>
            </NInput>
          </NFormItem>
        </template>

        <NFormItem :class="isMobile ? 'mb-2' : 'mb-4'" label="密码" path="password">
          <NInput
            v-model:value="formValue.password"
            type="password"
            placeholder="请输入密码"
            show-password-on="click"
            @keydown.enter="mode === 'login' && handleSubmit()"
          >
            <template #prefix>
              <div class="i-carbon-password" />
            </template>
          </NInput>
        </NFormItem>

        <template v-if="mode === 'register'">
          <NFormItem :class="isMobile ? 'mb-2' : 'mb-4'" label="确认密码" path="confirmPassword">
            <NInput
              v-model:value="formValue.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              show-password-on="click"
              @keydown.enter="handleSubmit"
            >
              <template #prefix>
                <div class="i-carbon-password" />
              </template>
            </NInput>
          </NFormItem>
        </template>

        <div :class="isMobile ? 'mt-4 space-y-2' : 'mt-6 space-y-4'">
          <NButton
            type="primary"
            block
            :loading="loading"
            :size="isMobile ? 'small' : 'large'"
            @click="handleSubmit"
          >
            {{ loading ? '登录中...' : '登录' }}
          </NButton>

          <!-- 注释掉注册入口按钮
          <div class="text-center">
            <NButton text :size="isMobile ? 'small' : 'medium'" @click="toggleMode">
              {{ mode === 'login' ? '没有账号？立即注册' : '已有账号？立即登录' }}
            </NButton>
          </div>
          -->
        </div>
      </NForm>
    </NCard>

    <!-- 页脚 -->
    <div :class="isMobile ? 'mt-4' : 'mt-8'" class="text-sm text-gray-500 flex gap-2 items-center">
      <NButton
        text
        :size="isMobile ? 'small' : 'medium'"
        class="flex gap-1 items-center !text-gray-500"
        @click="toggleTheme"
      >
        <div :class="isDark ? 'i-carbon-sun text-base sm:text-lg' : 'i-carbon-moon text-base sm:text-lg'" />
      </NButton>
    </div>
  </div>
</template>

<style scoped>
:deep(.n-card) {
  @apply shadow-sm transition-all;
}

@media (max-width: 768px) {
  :deep(.n-card-header) {
    padding: 12px;
  }

  :deep(.n-card__content) {
    padding: 12px;
  }

  :deep(.n-form-item-label) {
    padding-bottom: 4px;
    font-size: 14px;
  }

  :deep(.n-input) {
    height: 34px;
  }

  :deep(.n-button) {
    height: 34px;
  }
}

::view-transition-old(root),
::view-transition-new(root) {
  animation: none;
  mix-blend-mode: normal;
}

::view-transition-old(root) {
  z-index: 1;
}
::view-transition-new(root) {
  z-index: 9999;
}

.dark::view-transition-old(root) {
  z-index: 9999;
}
.dark::view-transition-new(root) {
  z-index: 1;
}
</style>

<route lang="yaml">
name: auth
layout: empty
</route>
