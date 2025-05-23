<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import { fileHistoryAPI } from '~/api/file-history'

const { width } = useWindowSize()
const isMobile = computed(() => width.value <= 768)

// 计算表格可用高度
const tableHeight = ref(0)
function updateTableHeight() {
  // 窗口高度 - (顶部header + padding) - (底部footer + padding) - (内容区padding) - (搜索表单高度) - (分页器高度)
  const headerHeight = 64 + 24
  const footerHeight = 64 + 24
  const contentPadding = 24 * 2
  const cardTitleHeight = 68 // 卡片标题高度
  const searchFormHeight = isMobile.value ? 10 : 58 // 搜索表单预估高度
  const paginationHeight = 40 // 分页器预估高度
  const margin = 20 // 额外边距

  tableHeight.value = window.innerHeight - headerHeight - footerHeight - contentPadding - cardTitleHeight - searchFormHeight - paginationHeight - margin
}

// 监听窗口大小变化
onMounted(() => {
  updateTableHeight()
  window.addEventListener('resize', updateTableHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTableHeight)
})

// 表格数据
const fileHistories = ref<Api.FileHistory.Record[]>([])
const loading = ref(false)
const pagination = ref({
  page: 1,
  pageSize: 50,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 50, 100, 200, 500, 1000],
})

// 搜索表单
const searchForm = ref({
  keyword: '',
})

// 加载数据
async function loadData() {
  try {
    loading.value = true
    const { data } = await fileHistoryAPI.findByPage({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      keyword: searchForm.value.keyword,
    })
    if (data) {
      fileHistories.value = data.list
      pagination.value.itemCount = data.total
    }
  }
  catch (error: any) {
    console.error(error)
    // window.$message?.error(error.message || '加载失败')
  }
  finally {
    loading.value = false
  }
}

// 处理分页变化
function handlePageChange(page: number) {
  pagination.value.page = page
  loadData()
}

// 处理每页条数变化
function handlePageSizeChange(pageSize: number) {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  loadData()
}

// 处理搜索
function handleSearch() {
  pagination.value.page = 1
  loadData()
}

// 重置搜索
function handleReset() {
  searchForm.value = {
    keyword: '',
  }
  handleSearch()
}

// 表格列定义
const columns = computed<DataTableColumns<Api.FileHistory.Record>>(() => [
  {
    title: '文件名',
    key: 'fileName',
    width: isMobile.value ? 180 : 200,
    ellipsis: { tooltip: true },
  },
  {
    title: '源路径',
    key: 'sourcePath',
    width: 200,
    ellipsis: { tooltip: true },
    hidden: isMobile.value,
  },
  {
    title: '目标路径',
    key: 'targetFilePath',
    width: 200,
    ellipsis: { tooltip: true },
    hidden: isMobile.value,
  },
  {
    title: '文件大小',
    key: 'fileSize',
    width: isMobile.value ? 80 : 100,
    render: (row) => {
      return h('span', {}, formatFileSize(row.fileSize))
    },
  },
  {
    title: '文件后缀',
    key: 'fileSuffix',
    width: isMobile.value ? 80 : 100,
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: isMobile.value ? 150 : 180,
    render: (row) => {
      return h('span', {}, new Date(row.createdAt).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: isMobile.value ? undefined : '2-digit',
        hour12: false,
      }))
    },
  },
])

// 格式化文件大小
function formatFileSize(size: number): string {
  if (size < 1024)
    return `${size} B`
  if (size < 1024 * 1024)
    return `${(size / 1024).toFixed(2)} KB`
  if (size < 1024 * 1024 * 1024)
    return `${(size / 1024 / 1024).toFixed(2)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`
}

// 初始加载
loadData()
</script>

<template>
  <div class="h-full">
    <NCard title="文件生成记录">
      <!-- 搜索表单 -->
      <div class="mb-4 flex gap-2">
        <NInput
          v-model:value="searchForm.keyword"
          size="small"
          placeholder="搜索文件名或路径"
          :style="isMobile ? 'width: 100%' : 'width: 300px'"
          clearable
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <div class="i-ri:search-line" />
          </template>
        </NInput>
        <NSpace :wrap="false">
          <NButton
            size="small"
            type="primary"
            @click="handleSearch"
          >
            <template #icon>
              <div class="i-ri:search-line" />
            </template>
          </NButton>
          <NButton
            size="small"
            @click="handleReset"
          >
            <template #icon>
              <div class="i-ri:refresh-line" />
            </template>
          </NButton>
        </NSpace>
      </div>

      <!-- 数据表格 -->
      <NDataTable
        :columns="columns"
        :data="fileHistories"
        :loading="loading"
        :remote="true"
        :pagination="pagination"
        :scroll-x="1200"
        :virtual-scroll="pagination.pageSize > 100"
        :max-height="tableHeight"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </NCard>
  </div>
</template>

<style scoped>
@media (max-width: 768px) {
  :deep(.n-form-item-label) {
    padding-bottom: 6px;
  }
  :deep(.n-card-header) {
    padding: 12px 16px;
  }
  :deep(.n-card__content) {
    padding: 12px 16px;
  }
  :deep(.n-data-table .n-data-table-td) {
    padding: 8px;
  }
}
</style>
