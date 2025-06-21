<script setup lang="ts">
defineOptions({
  name: 'GenerationOverview',
})

// 静态数据
const generationStats = ref({
  total: 0,
  processed: 0,
  skipped: 0,
  strmGenerated: 0,
  metadataDownloaded: 0,
  subtitleDownloaded: 0,
})

// 暴露刷新方法给父组件（此处只是一个占位，没有实际功能）
defineExpose({
  refresh: () => {
    // 实际应用中这里会有API请求
    // 静态组件，刷新方法仅做占位符
  },
})
</script>

<template>
  <NCard title="文件处理概览" class="shadow-sm">
    <div class="mb-4 flex items-center justify-center sm:justify-start">
      <div class="mr-3 p-2 rounded-full bg-green-100 flex-shrink-0 sm:mr-4 sm:p-3 dark:bg-green-800/30">
        <div class="i-carbon-document text-xl text-green-600 sm:text-2xl dark:text-green-400" />
      </div>
      <div>
        <div class="text-2xl text-gray-900 font-bold sm:text-3xl dark:text-white">
          {{ generationStats.total }}
        </div>
        <div class="text-xs text-gray-500 sm:text-sm dark:text-gray-400">
          文件扫描
        </div>
      </div>
    </div>

    <div class="mb-3 gap-2 grid grid-cols-2 sm:gap-4">
      <div class="p-2 rounded-md bg-green-50 flex flex-col items-center dark:bg-green-900/20 sm:items-start">
        <div class="text-base text-green-600 font-bold sm:text-lg dark:text-green-400">
          {{ generationStats.processed }}
        </div>
        <div class="text-xs text-gray-600 mt-1 dark:text-gray-400">
          已处理
        </div>
      </div>
      <div class="p-2 rounded-md bg-yellow-50 flex flex-col items-center dark:bg-yellow-900/20 sm:items-start">
        <div class="text-base text-yellow-600 font-bold sm:text-lg dark:text-yellow-400">
          {{ generationStats.skipped }}
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
  </NCard>
</template>
