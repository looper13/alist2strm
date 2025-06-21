<script setup lang="ts">
import { taskLogAPI } from '~/api/task-log'

defineOptions({
  name: 'GenerationOverview',
})

const loading = ref(false)
const selectedTimeRange = ref('day') // 默认选择"日"
const generationStats = ref({
  totalFiles: 0,
  processedFiles: 0,
  skippedFiles: 0,
  strmGenerated: 0,
  metadataDownloaded: 0,
  subtitleDownloaded: 0,
})

// 加载文件处理统计数据
async function loadFileProcessingStats() {
  loading.value = true
  try {
    const { code, data } = await taskLogAPI.getFileProcessingStats(selectedTimeRange.value)
    if (code === 0 && data) {
      generationStats.value = data
    }
  }
  catch (error) {
    console.error('加载文件处理统计数据失败:', error)
  }
  finally {
    loading.value = false
  }
}

// 处理时间范围变化
function handleTimeRangeChange(range: string) {
  selectedTimeRange.value = range
  loadFileProcessingStats()
}

// 组件挂载时加载数据
onMounted(() => {
  loadFileProcessingStats()
})

// 暴露刷新方法给父组件
defineExpose({
  refresh: loadFileProcessingStats,
})
</script>

<template>
  <NCard title="文件处理概览" class="shadow-sm">
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
        <div class="mr-3 p-2 rounded-full bg-green-100 flex-shrink-0 sm:mr-4 sm:p-3 dark:bg-green-800/30">
          <div class="i-carbon-document text-xl text-green-600 sm:text-2xl dark:text-green-400" />
        </div>
        <div>
          <div class="text-2xl text-gray-900 font-bold sm:text-3xl dark:text-white">
            {{ generationStats.totalFiles }}
          </div>
          <div class="text-xs text-gray-500 sm:text-sm dark:text-gray-400">
            文件扫描
          </div>
        </div>
      </div>

      <div class="mb-3 gap-2 grid grid-cols-2 sm:gap-4">
        <div class="p-2 rounded-md bg-green-50 flex flex-col items-center dark:bg-green-900/20 sm:items-start">
          <div class="text-base text-green-600 font-bold sm:text-lg dark:text-green-400">
            {{ generationStats.processedFiles }}
          </div>
          <div class="text-xs text-gray-600 mt-1 dark:text-gray-400">
            已处理
          </div>
        </div>
        <div class="p-2 rounded-md bg-yellow-50 flex flex-col items-center dark:bg-yellow-900/20 sm:items-start">
          <div class="text-base text-yellow-600 font-bold sm:text-lg dark:text-yellow-400">
            {{ generationStats.skippedFiles }}
          </div>
          <div class="text-xs text-gray-600 mt-1 dark:text-gray-400">
            跳过处理
          </div>
        </div>
      </div>

      <div class="gap-2 grid grid-cols-1 sm:gap-4 sm:grid-cols-3">
        <div class="p-2 rounded-md bg-blue-50 flex flex-row items-center justify-between dark:bg-blue-900/20 sm:flex-col sm:items-start sm:justify-start">
          <div class="text-xs text-gray-600 order-first sm:text-sm dark:text-gray-400 sm:mb-1 sm:order-last">
            STRM 文件生成
          </div>
          <div class="text-sm text-blue-600 font-bold dark:text-blue-400">
            {{ generationStats.strmGenerated }}
          </div>
        </div>
        <div class="p-2 rounded-md bg-purple-50 flex flex-row items-center justify-between dark:bg-purple-900/20 sm:flex-col sm:items-start sm:justify-start">
          <div class="text-xs text-gray-600 order-first sm:text-sm dark:text-gray-400 sm:mb-1 sm:order-last">
            元数据下载
          </div>
          <div class="text-sm text-purple-600 font-bold dark:text-purple-400">
            {{ generationStats.metadataDownloaded }}
          </div>
        </div>
        <div class="p-2 rounded-md bg-indigo-50 flex flex-row items-center justify-between dark:bg-indigo-900/20 sm:flex-col sm:items-start sm:justify-start">
          <div class="text-xs text-gray-600 order-first sm:text-sm dark:text-gray-400 sm:mb-1 sm:order-last">
            字幕下载
          </div>
          <div class="text-sm text-indigo-600 font-bold dark:text-indigo-400">
            {{ generationStats.subtitleDownloaded }}
          </div>
        </div>
      </div>
    </template>
  </NCard>
</template>
