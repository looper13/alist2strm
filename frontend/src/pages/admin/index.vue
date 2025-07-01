<script setup lang="ts">
import EmbyLatestMedia from '~/components/dashboard/EmbyLatestMedia.vue'
import EmbyLibraries from '~/components/dashboard/EmbyLibraries.vue'
import GenerationOverview from '~/components/dashboard/GenerationOverview.vue'
import TaskOverview from '~/components/dashboard/TaskOverview.vue'

defineOptions({
  name: 'DashboardPage',
})

// 获取组件引用
const taskOverviewRef = ref(null)
const embyLibrariesRef = ref(null)
const embyLatestMediaRef = ref(null)

// 页面加载时不再需要手动加载数据，各组件内部会自动处理
onMounted(() => {
  // 组件内部会自动加载数据
})
</script>

<template>
  <div class="mx-auto p-4 container">
    <!-- 概览统计卡片区域：PC端并排，移动端堆叠 -->
    <div class="mb-6 gap-6 grid grid-cols-1 lg:grid-cols-2">
      <!-- 任务概览卡片 -->
      <TaskOverview ref="taskOverviewRef" />

      <!-- 生成数据概览 -->
      <GenerationOverview />
    </div>

    <!-- Emby 媒体库横向列表 -->
    <EmbyLibraries ref="embyLibrariesRef" class="mb-6" />

    <!-- 近期 Emby 入库媒体横向列表 -->
    <EmbyLatestMedia ref="embyLatestMediaRef" class="mb-6" />
  </div>
</template>

<route lang="yaml">
name: home
layout: default
path: "/admin"
</route>
