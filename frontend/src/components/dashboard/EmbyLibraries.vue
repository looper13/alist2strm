<script setup lang="ts">
import { embyAPI } from '~/api/emby'

// 媒体库数据和状态
const embyLibraries = ref<any[]>([])
const loading = ref(false)

// 加载 Emby 媒体库数据
async function loadEmbyLibraries() {
  loading.value = true
  try {
    // 获取媒体库列表
    const { code, data } = await embyAPI.getLibraries()
    if (code === 0 && data) {
      embyLibraries.value = data
    }
  }
  catch (error) {
    console.error('加载 Emby 媒体库数据失败:', error)
  }
  finally {
    loading.value = false
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadEmbyLibraries()
})

// 暴露刷新方法给父组件
defineExpose({
  refresh: loadEmbyLibraries,
})
</script>

<template>
  <NCard title="Emby 媒体库" class="shadow-sm">
    <template #header-extra>
      <div class="flex items-center">
        <NButton v-if="!loading" size="small" quaternary circle class="text-gray-500" @click="loadEmbyLibraries">
          <div class="i-carbon-renew text-base" />
        </NButton>
        <NSpin v-show="loading" size="small" class="ml-2" />
      </div>
    </template>

    <NSpin v-if="loading" :show="true" class="flex h-30 w-full items-center justify-center sm:h-40">
      <template #description>
        <div class="text-xs sm:text-sm dark:text-gray-300">
          加载中...
        </div>
      </template>
    </NSpin>

    <NEmpty v-else-if="!embyLibraries || embyLibraries.length === 0" description="暂无媒体库数据" />

    <div v-else class="pb-4 overflow-x-auto">
      <div class="hide-scrollbar pb-4 flex gap-3 overflow-x-auto snap-x snap-mandatory sm:gap-5">
        <div
          v-for="library in embyLibraries"
          :key="library.Id"
          class="group rounded-lg flex-shrink-0 cursor-pointer shadow transition-all duration-300 relative overflow-hidden snap-center sm:h-160px sm:w-280px hover:shadow-md"
          style="width: min(200px, 80vw); height: min(120px, 45vw);"
        >
          <img
            :src="library.PrimaryImageItemId ? embyAPI.getImageUrl(library.PrimaryImageItemId, 'Primary', { maxWidth: 400, quality: 90 }) : '/api/emby/items/library-default/images/Primary'"
            :alt="library.Name"
            class="h-full w-full transition-all duration-300 object-cover object-center group-hover:scale-110"
            @error="($event.target as HTMLImageElement).src = 'https://via.placeholder.com/280x160?text=No+Image'"
          >

          <!-- 顶部半透明标签显示媒体库名称 -->
          <div class="px-3 py-1.5 rounded-bl-lg rounded-tr-lg bg-black/60 flex items-center right-0 top-0 absolute">
            <span class="text-sm text-white font-medium">{{ library.Name }}</span>
          </div>

          <!-- 底部渐变蒙版，显示媒体库详情 -->
          <!-- <div
            class="p-3 opacity-0 flex flex-col transition-opacity duration-300 bottom-0 left-0 right-0 justify-end absolute group-hover:opacity-100"
            style="background: linear-gradient(to top, rgba(0,0,0,0.8) 0%, rgba(0,0,0,0.5) 60%, rgba(0,0,0,0) 100%);"
          >
            <div v-if="library.ItemCount" class="text-xs text-white/90 truncate">
              {{ library.ItemCount }} 个项目
            </div>
            <div v-if="library.CollectionType" class="text-xs text-white/80 truncate">
              {{
                library.CollectionType === 'movies' ? '电影'
                : library.CollectionType === 'tvshows' ? '电视剧'
                  : library.CollectionType === 'music' ? '音乐'
                    : library.CollectionType
              }}
            </div>
          </div> -->
        </div>
      </div>
    </div>
  </NCard>
</template>
