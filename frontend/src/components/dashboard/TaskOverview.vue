<script setup lang="ts">
import { taskAPI } from '~/api/task'

const loading = ref(false)
const taskStats = ref<Api.Task.Stats | null>(null)
const selectedTimeRange = ref('day') // 默认选择"日"

// 加载任务统计数据
async function loadTaskStats() {
  loading.value = true
  try {
    const { code, data } = await taskAPI.getStats(selectedTimeRange.value)
    if (code === 0 && data) {
      taskStats.value = data
    }
  }
  catch (error) {
    console.error('加载任务统计数据失败:', error)
  }
  finally {
    loading.value = false
  }
}

// 处理时间范围变化
function handleTimeRangeChange(range: string) {
  selectedTimeRange.value = range
  loadTaskStats()
}

// 组件挂载时加载数据
onMounted(() => {
  loadTaskStats()
})

// 暴露刷新方法给父组件
defineExpose({
  refresh: loadTaskStats,
})
</script>

<template>
  <NCard title="任务概览" class="shadow-sm">
    <template #header-extra>
      <NButtonGroup size="small" class="lg:m-0 -mr-1">
        <NButton
          size="tiny"
          class="sm:text-xs"
          :type="selectedTimeRange === 'day' ? 'primary' : 'default'"
          @click="handleTimeRangeChange('day')"
        >
          日
        </NButton>
        <NButton
          size="tiny"
          class="sm:text-xs"
          :type="selectedTimeRange === 'month' ? 'primary' : 'default'"
          @click="handleTimeRangeChange('month')"
        >
          月
        </NButton>
        <NButton
          size="tiny"
          class="sm:text-xs"
          :type="selectedTimeRange === 'year' ? 'primary' : 'default'"
          @click="handleTimeRangeChange('year')"
        >
          年
        </NButton>
      </NButtonGroup>
    </template>
    <NSpin v-if="loading" :show="true" class="flex h-30 w-full items-center justify-center sm:h-40">
      <template #description>
        <div class="text-xs sm:text-sm dark:text-gray-300">
          加载中...
        </div>
      </template>
    </NSpin>

    <template v-else>
      <div class="mb-4 flex items-center justify-center sm:justify-start">
        <div class="mr-3 p-2 rounded-full bg-blue-100 flex-shrink-0 sm:mr-4 sm:p-3 dark:bg-blue-800/30">
          <div class="i-carbon-task text-xl text-blue-600 sm:text-2xl dark:text-blue-400" />
        </div>
        <div>
          <div class="text-2xl text-gray-900 font-bold sm:text-3xl dark:text-white">
            {{ taskStats?.total || 0 }}
          </div>
          <div class="text-xs text-gray-500 sm:text-sm dark:text-gray-400">
            总任务数
          </div>
        </div>
      </div>

      <div class="mb-3 gap-2 grid grid-cols-2 sm:gap-4">
        <div class="p-2 rounded-md bg-green-50 flex flex-col items-center dark:bg-green-900/20 sm:items-start">
          <div class="text-base text-green-600 font-bold sm:text-lg dark:text-green-400">
            {{ taskStats?.enabled || 0 }}
          </div>
          <div class="text-xs text-gray-600 mt-1 dark:text-gray-400">
            开启
          </div>
        </div>
        <div class="p-2 rounded-md bg-red-50 flex flex-col items-center dark:bg-red-900/20 sm:items-start">
          <div class="text-base text-red-600 font-bold sm:text-lg dark:text-red-400">
            {{ taskStats?.disabled || 0 }}
          </div>
          <div class="text-xs text-gray-600 mt-1 dark:text-gray-400">
            关闭
          </div>
        </div>
      </div>

      <div class="gap-2 grid grid-cols-1 sm:gap-4 sm:grid-cols-3">
        <div class="p-2 rounded-md bg-gray-50 flex flex-row items-center justify-between dark:bg-gray-800/50 sm:flex-col sm:items-start sm:justify-start">
          <div class="text-xs text-gray-600 order-first sm:text-sm dark:text-gray-400 sm:mb-1 sm:order-last">
            总执行次数
          </div>
          <div class="text-sm text-gray-700 font-bold dark:text-gray-300">
            {{ taskStats?.totalExecutions || 0 }}
          </div>
        </div>
        <div class="p-2 rounded-md bg-gray-50 flex flex-row items-center justify-between dark:bg-gray-800/50 sm:flex-col sm:items-start sm:justify-start">
          <div class="text-xs text-gray-600 order-first sm:text-sm dark:text-gray-400 sm:mb-1 sm:order-last">
            {{ selectedTimeRange === 'day' ? '今日成功' : selectedTimeRange === 'month' ? '本月成功' : '今年成功' }}
          </div>
          <div class="text-sm text-green-600 font-bold dark:text-green-400">
            {{ taskStats?.successCount || 0 }}
          </div>
        </div>
        <div class="p-2 rounded-md bg-gray-50 flex flex-row items-center justify-between dark:bg-gray-800/50 sm:flex-col sm:items-start sm:justify-start">
          <div class="text-xs text-gray-600 order-first sm:text-sm dark:text-gray-400 sm:mb-1 sm:order-last">
            {{ selectedTimeRange === 'day' ? '今日失败' : selectedTimeRange === 'month' ? '本月失败' : '今年失败' }}
          </div>
          <div class="text-sm text-red-600 font-bold dark:text-red-400">
            {{ taskStats?.failedCount || 0 }}
          </div>
        </div>
      </div>
    </template>
  </NCard>
</template>
