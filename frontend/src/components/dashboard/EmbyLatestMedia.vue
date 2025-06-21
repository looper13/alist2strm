<script setup lang="ts">
import { embyAPI } from '~/api/emby'

// Emby 最近入库媒体相关数据和状态
const latestEmbyMedia = ref<any[]>([])
const loading = ref(false)

// 分页配置
const pageSize = ref(10)
const pageSizeOptions = [
  { label: '10条', value: 10 },
  { label: '20条', value: 20 },
  { label: '30条', value: 30 },
  { label: '50条', value: 50 },
]

/**
 * 加载 Emby 最近入库媒体数据
 */
async function loadLatestEmbyMedia() {
  loading.value = true
  try {
    // 获取最新入库媒体
    const { code, data } = await embyAPI.getLatestMedia(pageSize.value) // 使用动态页面大小
    if (code === 0 && data) {
      latestEmbyMedia.value = data
    }
  }
  catch (error) {
    console.error('加载 Emby 最近入库媒体数据失败:', error)
  }
  finally {
    loading.value = false
  }
}

/**
 * 处理页面大小变更
 */
function handlePageSizeChange(size: number) {
  pageSize.value = size
  loadLatestEmbyMedia() // 重新加载最近入库媒体数据
}

// 组件挂载时加载数据
onMounted(() => {
  loadLatestEmbyMedia()
})

// 暴露刷新方法给父组件
defineExpose({
  refresh: loadLatestEmbyMedia,
})
</script>

<template>
  <NCard title="Emby 近期入库媒体" class="shadow-sm">
    <template #header-extra>
      <div class="flex items-center space-x-2">
        <NSelect
          v-if="!loading"
          v-model:value="pageSize"
          :options="pageSizeOptions"
          size="small"
          class="w-24"
          @update:value="handlePageSizeChange"
        />
        <NButton v-if="!loading" size="small" quaternary circle class="text-gray-500" @click="loadLatestEmbyMedia">
          <div class="i-carbon-renew text-base" />
        </NButton>
        <NSpin v-show="loading" size="small" />
      </div>
    </template>

    <NEmpty v-if="!loading && (!latestEmbyMedia || latestEmbyMedia.length === 0)" description="暂无最近入库媒体" />

    <div v-else class="pb-4 overflow-x-auto">
      <div class="flex min-w-max space-x-5">
        <div
          v-for="media in latestEmbyMedia"
          :key="media.Id"
          class="group rounded-lg cursor-pointer shadow relative overflow-hidden hover:shadow-md"
          style="width: 180px; height: 270px;"
        >
          <img
            :src="embyAPI.getImageUrl(media.Id, 'Primary', { maxWidth: 300, quality: 90 })"
            :alt="media.Name"
            class="h-full w-full transition-all duration-300 object-cover object-center group-hover:scale-110"
            @error="($event.target as HTMLImageElement).src = 'https://via.placeholder.com/180x270?text=No+Image'"
          >
          <!-- 右上角半透明标签显示媒体类型 -->
          <div class="text-sm text-white font-medium px-3 py-1.5 rounded-bl-lg rounded-tr-lg bg-black/60 right-0 top-0 absolute">
            {{ media.Type === 'Movie' ? '电影' : media.Type === 'Series' ? '剧集' : '其他' }}
          </div>

          <!-- 悬停时显示的半透明蒙版，包含年份、剧集名称、集数 -->
          <div
            class="p-3 bg-black/70 opacity-0 flex flex-col transition-opacity duration-300 inset-0 justify-end absolute group-hover:opacity-100"
          >
            <div class="text-sm text-white font-medium truncate">
              {{ media.Name }}
            </div>
            <div v-if="media.ProductionYear" class="text-xs text-white/80 mt-1 truncate">
              {{ media.ProductionYear }}
            </div>
            <div v-if="media.SeriesName" class="text-xs text-white/80 mt-0.5 truncate">
              {{ media.SeriesName }}
            </div>
            <div v-if="media.IndexNumber" class="text-xs text-white/80 truncate">
              第 {{ media.IndexNumber }} 集
            </div>
          </div>
        </div>
      </div>
    </div>
  </NCard>
</template>
