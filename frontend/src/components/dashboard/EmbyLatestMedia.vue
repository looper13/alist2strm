<script setup lang="ts">
import { embyAPI } from '~/api/emby'

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
      <div class="flex items-center">
        <NSelect
          v-if="!loading"
          v-model:value="pageSize"
          :options="pageSizeOptions"
          size="small"
          class="w-20 sm:w-24"
          @update:value="handlePageSizeChange"
        />
        <NSpin v-show="loading" size="small" class="ml-2" />
      </div>
    </template>

    <NEmpty v-if="!loading && (!latestEmbyMedia || latestEmbyMedia.length === 0)" description="暂无最近入库媒体" />

    <div v-else class="pb-4 overflow-x-auto">
      <div class="hide-scrollbar pb-4 flex gap-3 overflow-x-auto snap-x snap-mandatory sm:gap-5">
        <div
          v-for="media in latestEmbyMedia"
          :key="media.Id"
          class="group rounded-lg flex-shrink-0 cursor-pointer shadow relative overflow-hidden snap-center hover:shadow-md"
          style="width: 140px; height: 210px;"
          :style="{ width: 'min(140px, 30vw)', height: 'min(210px, 45vw)' }"
        >
          <img
            :src="embyAPI.getImageUrl(media.Id, 'Primary', { maxWidth: 300, quality: 90 })"
            :alt="media.Name"
            class="h-full w-full transition-all duration-300 object-cover object-center group-hover:scale-110"
            style="position: relative; z-index: 0;"
            @error="($event.target as HTMLImageElement).src = 'https://via.placeholder.com/180x270?text=No+Image'"
          >
          <!-- 右上角半透明标签显示媒体类型 -->
          <div class="text-sm text-white font-medium px-3 py-1.5 rounded-bl-lg rounded-tr-lg bg-black/60 right-0 top-0 absolute">
            {{ media.Type === 'Movie' ? '电影' : media.Type === 'Series' ? '剧集' : '其他' }}
          </div>

          <!-- PC端蒙版（全屏高度，悬停显示） -->
          <div
            class="p-3 bg-black/70 opacity-0 flex-col hidden transition-opacity duration-300 inset-0 justify-end absolute group-hover:opacity-100 sm:flex"
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

          <!-- 移动端蒙版（底部渐变，始终显示） -->
          <div
            class="p-3 flex flex-col bottom-0 left-0 right-0 justify-end absolute sm:hidden"
            style="
              background: linear-gradient(to top, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0.7) 50%, rgba(0,0,0,0) 100%);
              height: 60%;
            "
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
