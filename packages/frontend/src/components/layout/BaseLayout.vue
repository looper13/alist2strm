<script setup lang="ts">
import type { DropdownOption } from 'naive-ui'
import { useRouter } from 'vue-router'
import { useAuth, useMobile } from '~/composables'
import { toggleTheme } from '~/composables/dark'
import SiderMenu from './SiderMenu.vue'

const { isMobile } = useMobile()
const { logout, userInfo } = useAuth()
const router = useRouter()

// 处理退出登录
function handleLogout() {
  logout()
  router.push('/auth')
}

// 用户菜单选项
const userMenuOptions: DropdownOption[] = [
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h('div', { class: 'i-carbon-logout' }),
  },
]

// 处理菜单选择
function handleSelect(key: string) {
  if (key === 'logout')
    handleLogout()
}
</script>

<template>
  <n-layout position="absolute">
    <!-- 头部导航 -->
    <n-layout-header class="header" bordered>
      <div class="px-4 flex h-full items-center justify-between">
        <div class="flex gap-2 items-center">
          <div class="i-carbon-media-library text-2xl text-green-400 md:text-3xl" />
          <p class="text-xl text-green-400 font-bold md:text-2xl">
            AList2Strm
          </p>
        </div>
        <!-- 主题切换按钮和用户菜单 -->
        <div class="flex gap-4 items-center">
          <button class="btn" @click="toggleTheme">
            <div class="i-carbon-sun dark:i-carbon-moon text-xl" />
          </button>
          <n-dropdown
            trigger="click"
            :options="userMenuOptions"
            @select="handleSelect"
          >
            <div class="flex gap-2 cursor-pointer select-none items-center hover:opacity-80">
              <div class="i-carbon-user-avatar text-xl" />
              <span v-if="!isMobile" class="text-sm">{{ userInfo?.nickname || userInfo?.username }}</span>
            </div>
          </n-dropdown>
        </div>
      </div>
    </n-layout-header>

    <n-layout has-sider position="absolute" style="top: 64px; bottom: 0">
      <!-- PC端侧边栏 -->
      <SiderMenu v-if="!isMobile" />

      <!-- 内容区域 -->
      <n-layout :native-scrollbar="false" class="main-content">
        <n-space vertical size="large" class="p-4 md:p-6">
          <slot />
        </n-space>
      </n-layout>
    </n-layout>

    <!-- 移动端菜单按钮 -->
    <MobileMenu v-if="isMobile" />
  </n-layout>
</template>

<style scoped>
.header {
  height: 64px;
  z-index: 100;
}

.main-content {
  height: 100%;
}
@media (max-width: 768px) {
  .header {
    height: 56px;
  }
}
</style>
