<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import { h, ref } from 'vue'
import { fileHistoryAPI } from '~/api/file-history'

// 计算表格可用高度
const tableHeight = ref(0)
function updateTableHeight() {
  // 窗口高度 - (顶部header + padding) - (底部footer + padding) - (内容区padding) - (搜索表单高度) - (分页器高度)
  const headerHeight = 64 + 24
  const footerHeight = 64 + 24
  const contentPadding = 24 * 2
  const cardTitleHeight = 68 // 卡片标题高度
  const searchFormHeight = 58 // 搜索表单预估高度
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
const fileHistories = ref<Api.FileHistory[]>([])
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
  fileType: '',
  fileSuffix: '',
  dateRange: [
    new Date(new Date().setHours(0, 0, 0, 0) - 3 * 24 * 60 * 60 * 1000).getTime(),
    new Date(new Date().setHours(23, 59, 59, 999)).getTime(),
  ] as [number, number],
})

// 加载数据
async function loadData() {
  try {
    loading.value = true
    const { data } = await fileHistoryAPI.findByPage({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      ...searchForm.value,
      startTime: searchForm.value.dateRange[0] ? new Date(searchForm.value.dateRange[0]).toISOString() : undefined,
      endTime: searchForm.value.dateRange[1] ? new Date(searchForm.value.dateRange[1]).toISOString() : undefined,
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
    fileType: '',
    fileSuffix: '',
    dateRange: [
      new Date(new Date().setHours(0, 0, 0, 0) - 3 * 24 * 60 * 60 * 1000).getTime(),
      new Date(new Date().setHours(23, 59, 59, 999)).getTime(),
    ] as [number, number],
  }
  handleSearch()
}

// 表格列定义
const columns: DataTableColumns<Api.FileHistory> = [
  { title: '文件名', key: 'fileName', width: 200, ellipsis: { tooltip: true } },
  { title: '源路径', key: 'sourcePath', width: 200, ellipsis: { tooltip: true } },
  { title: '目标路径', key: 'targetFilePath', width: 200, ellipsis: { tooltip: true } },
  { title: '文件大小', key: 'fileSize', width: 100, render: (row) => {
    return h('span', {}, formatFileSize(row.fileSize))
  } },
  { title: '文件类型', key: 'fileType', width: 100 },
  { title: '文件后缀', key: 'fileSuffix', width: 100 },
  { title: '创建时间', key: 'createdAt', width: 180, render: (row) => {
    return h('span', {}, new Date(row.createdAt).toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: false,
    }))
  } },
]

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
      <NForm
        inline
        :model="searchForm"
        label-placement="left"
        label-width="auto"
      >
        <NFormItem label="关键字">
          <NInput
            v-model:value="searchForm.keyword"
            placeholder="文件名/路径"
            clearable
            @keyup.enter="handleSearch"
          />
        </NFormItem>
        <NFormItem label="文件后缀">
          <NInput
            v-model:value="searchForm.fileSuffix"
            placeholder="文件后缀"
            clearable
            @keyup.enter="handleSearch"
          />
        </NFormItem>
        <NFormItem label="时间范围">
          <NDatePicker
            v-model:value="searchForm.dateRange"
            type="daterange"
            clearable
            placeholder="选择日期范围"
          />
        </NFormItem>
        <NFormItem>
          <NSpace>
            <NButton type="primary" @click="handleSearch">
              搜索
            </NButton>
            <NButton @click="handleReset">
              重置
            </NButton>
          </NSpace>
        </NFormItem>
      </NForm>

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
