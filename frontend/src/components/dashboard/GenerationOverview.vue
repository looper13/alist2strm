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
  <NCard title="生成数据概览" class="shadow-sm">
    <div class="mb-4 flex items-center">
      <div class="mr-4 p-3 rounded-full bg-green-100 dark:bg-green-800/30">
        <div class="i-carbon-document text-2xl text-green-600 dark:text-green-400" />
      </div>
      <div>
        <div class="text-3xl text-gray-900 font-bold dark:text-white">
          {{ generationStats.total }}
        </div>
        <div class="text-sm text-gray-500 dark:text-gray-400">
          文件扫描总数
        </div>
      </div>
    </div>

    <div class="mb-2 gap-4 grid grid-cols-2">
      <div class="p-2 rounded-md bg-green-50 dark:bg-green-900/20">
        <div class="text-lg text-green-600 font-bold dark:text-green-400">
          {{ generationStats.processed }}
        </div>
        <div class="text-xs text-gray-600 dark:text-gray-400">
          已处理数
        </div>
      </div>
      <div class="p-2 rounded-md bg-yellow-50 dark:bg-yellow-900/20">
        <div class="text-lg text-yellow-600 font-bold dark:text-yellow-400">
          {{ generationStats.skipped }}
        </div>
        <div class="text-xs text-gray-600 dark:text-gray-400">
          已跳过数
        </div>
      </div>
    </div>

    <div class="gap-4 grid grid-cols-3">
      <div class="p-2 rounded-md bg-blue-50 dark:bg-blue-900/20">
        <div class="text-sm text-blue-600 font-bold dark:text-blue-400">
          {{ generationStats.strmGenerated }}
        </div>
        <div class="text-xs text-gray-600 dark:text-gray-400">
          STRM生成数
        </div>
      </div>
      <div class="p-2 rounded-md bg-purple-50 dark:bg-purple-900/20">
        <div class="text-sm text-purple-600 font-bold dark:text-purple-400">
          {{ generationStats.metadataDownloaded }}
        </div>
        <div class="text-xs text-gray-600 dark:text-gray-400">
          元数据下载
        </div>
      </div>
      <div class="p-2 rounded-md bg-indigo-50 dark:bg-indigo-900/20">
        <div class="text-sm text-indigo-600 font-bold dark:text-indigo-400">
          {{ generationStats.subtitleDownloaded }}
        </div>
        <div class="text-xs text-gray-600 dark:text-gray-400">
          字幕下载
        </div>
      </div>
    </div>
  </NCard>
</template>
