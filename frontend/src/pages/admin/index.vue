<script setup lang="ts">
import type { DashboardStats } from '~/api/dashboard'
import { dashboardAPI } from '~/api/dashboard'
import { embyAPI } from '~/api/emby'

defineOptions({
  name: 'DashboardPage',
})

// 仪表盘数据
const loading = ref(true)
const stats = ref<DashboardStats | null>(null)

// Emby 媒体库相关数据和状态
const embyLibraries = ref<any[]>([])
const loadingEmbyLibraries = ref(false) // 媒体库加载状态

// Emby 最近入库媒体相关数据和状态
const latestEmbyMedia = ref<any[]>([])
const loadingLatestMedia = ref(false) // 最近入库媒体加载状态

// 分页配置
const pageSize = ref(10)
const pageSizeOptions = [
  { label: '10条/页', value: 10 },
  { label: '20条/页', value: 20 },
  { label: '30条/页', value: 30 },
  { label: '50条/页', value: 50 },
]

// 加载仪表盘统计数据
async function loadDashboardStats() {
  loading.value = true
  try {
    const { code, data } = await dashboardAPI.getStats()
    if (code === 0 && data) {
      stats.value = data
    }
  }
  catch (error) {
    console.error('加载仪表盘统计数据失败:', error)
  }
  finally {
    loading.value = false
  }
}

/**
 * 加载 Emby 媒体库数据
 * 独立的媒体库加载函数，不影响其他组件的状态
 */
async function loadEmbyLibraries() {
  loadingEmbyLibraries.value = true
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
    loadingEmbyLibraries.value = false
  }
}

/**
 * 加载 Emby 最近入库媒体数据
 * 独立的最近入库媒体加载函数，不影响媒体库的状态
 */
async function loadLatestEmbyMedia() {
  loadingLatestMedia.value = true
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
    loadingLatestMedia.value = false
  }
}

/**
 * 处理页面大小变更
 * 仅影响最近入库媒体，不会触发媒体库刷新
 */
function handlePageSizeChange(size: number) {
  pageSize.value = size
  loadLatestEmbyMedia() // 只重新加载最近入库媒体数据
}

/**
 * 单独刷新 Emby 媒体库
 * 独立的刷新按钮事件处理函数
 */
function refreshEmbyLibraries() {
  loadEmbyLibraries()
}

/**
 * 单独刷新最近入库媒体
 * 独立的刷新按钮事件处理函数
 */
function refreshLatestEmbyMedia() {
  loadLatestEmbyMedia()
}

// 页面加载时获取数据，各个模块并行加载
onMounted(async () => {
  await Promise.all([
    loadDashboardStats(),
    loadEmbyLibraries(),
    loadLatestEmbyMedia(),
  ])
})
</script>

<template>
  <div class="mx-auto px-4 py-6 container">
    <!-- <h1 class="text-2xl font-bold mb-6 dark:text-white">
      AList2Strm 仪表盘
    </h1> -->

    <!-- 数据加载中状态显示 -->
    <NSpin v-if="loading" :show="true" class="flex h-60 w-full items-center justify-center">
      <template #description>
        <div class="text-base dark:text-gray-300">
          加载仪表盘数据中...
        </div>
      </template>
    </NSpin>

    <template v-else>
      <!-- 概览统计卡片区域 -->
      <div class="mb-8 gap-6 grid grid-cols-1 lg:grid-cols-3">
        <!-- 任务概览卡片 -->
        <NCard title="任务概览" class="shadow-sm">
          <div class="mb-4 flex items-center">
            <div class="mr-4 p-3 rounded-full bg-blue-100 dark:bg-blue-800/30">
              <div class="i-carbon-task text-2xl text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <div class="text-3xl text-gray-900 font-bold dark:text-white">
                {{ stats?.taskStats.total || 0 }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                总任务数
              </div>
            </div>
          </div>

          <div class="mb-2 gap-4 grid grid-cols-2">
            <div class="p-2 rounded-md bg-green-50 dark:bg-green-900/20">
              <div class="text-lg text-green-600 font-bold dark:text-green-400">
                {{ stats?.taskStats.enabled || 0 }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                已启用
              </div>
            </div>
            <div class="p-2 rounded-md bg-red-50 dark:bg-red-900/20">
              <div class="text-lg text-red-600 font-bold dark:text-red-400">
                {{ stats?.taskStats.disabled || 0 }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                已禁用
              </div>
            </div>
          </div>

          <div class="gap-4 grid grid-cols-3">
            <div class="p-2 rounded-md bg-gray-50 dark:bg-gray-800/50">
              <div class="text-sm text-gray-700 font-bold dark:text-gray-300">
                {{ stats?.taskStats.totalExecutions || 0 }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                总执行次数
              </div>
            </div>
            <div class="p-2 rounded-md bg-gray-50 dark:bg-gray-800/50">
              <div class="text-sm text-green-600 font-bold dark:text-green-400">
                {{ stats?.taskStats.todaySuccess || 0 }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                今日成功
              </div>
            </div>
            <div class="p-2 rounded-md bg-gray-50 dark:bg-gray-800/50">
              <div class="text-sm text-red-600 font-bold dark:text-red-400">
                {{ stats?.taskStats.todayFailed || 0 }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                今日失败
              </div>
            </div>
          </div>
        </NCard>

        <!-- 生成数据概览 -->
        <NCard title="生成数据概览" class="shadow-sm">
          <div class="mb-4 flex items-center">
            <div class="mr-4 p-3 rounded-full bg-green-100 dark:bg-green-800/30">
              <div class="i-carbon-document text-2xl text-green-600 dark:text-green-400" />
            </div>
            <div>
              <div class="text-3xl text-gray-900 font-bold dark:text-white">
                {{ stats?.strmStats.totalGenerated || 0 }}
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                STRM 总生成数
              </div>
            </div>
          </div>

          <div class="gap-4 grid grid-cols-2">
            <div class="p-3 rounded-md bg-green-50 dark:bg-green-900/20">
              <div class="text-lg text-green-600 font-bold dark:text-green-400">
                {{ stats?.strmStats.todayGenerated || 0 }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                今日生成数
              </div>
            </div>
            <div class="p-3 rounded-md bg-blue-50 dark:bg-blue-900/20">
              <div class="text-lg text-blue-600 font-bold dark:text-blue-400">
                {{ Math.floor((stats?.strmStats.downloadSize || 0) / 1024) }} GB
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                原数据下载量
              </div>
            </div>
          </div>
        </NCard>

        <!-- Emby 概览 -->
        <NCard title="Emby 概览" class="shadow-sm">
          <div class="mb-4 flex items-center">
            <div class="mr-4 p-3 rounded-full bg-purple-100 dark:bg-purple-800/30">
              <div class="i-carbon-media-library text-2xl text-purple-600 dark:text-purple-400" />
            </div>
            <div class="flex-1">
              <div class="flex items-center">
                <div class="text-lg text-gray-900 font-bold mr-2 dark:text-white">
                  Emby 服务器
                </div>
                <NTag :type="stats?.embyStats.serverStatus === 'online' ? 'success' : 'error'" size="small">
                  {{ stats?.embyStats.serverStatus === 'online' ? '在线' : '离线' }}
                </NTag>
              </div>
              <div class="text-sm text-gray-500 dark:text-gray-400">
                用户数: <span class="text-purple-600 font-medium dark:text-purple-400">{{ stats?.embyStats.userCount || 0 }}</span>
              </div>
            </div>
          </div>

          <div class="gap-4 grid grid-cols-2">
            <div class="p-3 rounded-md bg-purple-50 dark:bg-purple-900/20">
              <div class="text-lg text-purple-600 font-bold dark:text-purple-400">
                {{ embyLibraries.length }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                媒体库数量
              </div>
            </div>
            <div class="p-3 rounded-md bg-purple-50 dark:bg-purple-900/20">
              <div class="text-lg text-purple-600 font-bold dark:text-purple-400">
                {{ latestEmbyMedia.length }} / {{ pageSize }}
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-400">
                近期入库数量
              </div>
            </div>
          </div>
        </NCard>
      </div>

      <!-- Emby 媒体库横向列表 -->
      <NCard title="Emby 媒体库" class="mb-6 shadow-sm">
        <template #header-extra>
          <div class="flex items-center space-x-2">
            <NButton v-if="!loadingEmbyLibraries" size="small" quaternary circle class="text-gray-500" @click="refreshEmbyLibraries">
              <div class="i-carbon-renew text-base" />
            </NButton>
            <NSpin v-show="loadingEmbyLibraries" size="small" />
          </div>
        </template>

        <NEmpty v-if="!loadingEmbyLibraries && (!embyLibraries || embyLibraries.length === 0)" description="暂无媒体库数据" />

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

      <!-- 近期 Emby 入库媒体横向列表 -->
      <NCard title="Emby 近期入库媒体" class="mb-6 shadow-sm">
        <template #header-extra>
          <div class="flex items-center space-x-2">
            <NSelect
              v-if="!loadingLatestMedia"
              v-model:value="pageSize"
              :options="pageSizeOptions"
              size="small"
              class="w-24"
              @update:value="handlePageSizeChange"
            />
            <NButton v-if="!loadingLatestMedia" size="small" quaternary circle class="text-gray-500" @click="refreshLatestEmbyMedia">
              <div class="i-carbon-renew text-base" />
            </NButton>
            <NSpin v-show="loadingLatestMedia" size="small" />
          </div>
        </template>

        <NEmpty v-if="!loadingLatestMedia && (!latestEmbyMedia || latestEmbyMedia.length === 0)" description="暂无最近入库媒体" />

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
  </div>
</template>

<route lang="yaml">
name: home
layout: default
path: "/admin"
</route>
