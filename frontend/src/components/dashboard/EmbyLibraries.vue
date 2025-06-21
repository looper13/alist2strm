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
      <div class="flex items-center space-x-2">
        <NButton v-if="!loading" size="small" quaternary circle class="text-gray-500" @click="loadEmbyLibraries">
          <div class="i-carbon-renew text-base" />
        </NButton>
        <NSpin v-show="loading" size="small" />
      </div>
    </template>

    <NEmpty v-if="!loading && (!embyLibraries || embyLibraries.length === 0)" description="暂无媒体库数据" />

    <div v-else class="pb-4 overflow-x-auto">
      <div class="flex min-w-max space-x-5">
        <div
          v-for="library in embyLibraries"
          :key="library.Id"
          class="group rounded-lg cursor-pointer shadow relative overflow-hidden hover:shadow-md"
          style="width: 280px; height: 160px;"
        >
          <img
            :src="library.PrimaryImageItemId ? embyAPI.getImageUrl(library.PrimaryImageItemId, 'Primary', { maxWidth: 400, quality: 90 }) : '/api/emby/items/library-default/images/Primary'"
            :alt="library.Name"
            class="h-full w-full transition-all duration-300 object-cover object-center group-hover:scale-110"
            @error="($event.target as HTMLImageElement).src = 'https://via.placeholder.com/280x160?text=No+Image'"
          >
          <!-- 半透明标签显示媒体库名称 -->
          <div class="text-sm text-white font-medium px-3 py-1.5 rounded-bl-lg rounded-tr-lg bg-black/60 right-0 top-0 absolute">
            {{ library.Name }}
          </div>
        </div>
      </div>
    </div>
  </NCard>
</template>
