<script setup lang="ts">
import { useWindowSize } from '@vueuse/core'
import SiderMenu from './SiderMenu.vue'

const { width } = useWindowSize()
const isMobile = computed(() => width.value <= 768)

function toggleTheme(event: MouseEvent) {
  const x = event.clientX
  const y = event.clientY
  const endRadius = Math.hypot(
    Math.max(x, innerWidth - x),
    Math.max(y, innerHeight - y),
  )
  if (!document.startViewTransition) {
    toggleDark()
    return
  }
  const transition = document.startViewTransition(async () => {
    toggleDark()
  })
  transition.ready.then(() => {
    const clipPath = [
      `circle(0px at ${x}px ${y}px)`,
      `circle(${endRadius}px at ${x}px ${y}px)`,
    ]
    document.documentElement.animate(
      {
        clipPath: isDark.value ? [...clipPath].reverse() : clipPath,
      },
      {
        duration: 500,
        easing: 'ease-in',
        pseudoElement: isDark.value
          ? '::view-transition-old(root)'
          : '::view-transition-new(root)',
      },
    )
  })
}
</script>

<template>
  <n-layout position="absolute">
    <!-- 头部导航 -->
    <n-layout-header class="header" bordered>
      <div class="px-4 flex h-full items-center justify-between">
        <p class="text-xl text-green-400 font-bold md:text-2xl">
          AList2Strm
        </p>
        <!-- 主题切换按钮 -->
        <div class="flex items-center">
          <button class="btn" @click="toggleTheme">
            <div class="i-carbon-sun dark:i-carbon-moon text-xl" />
          </button>
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
