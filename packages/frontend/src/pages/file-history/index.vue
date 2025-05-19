<script setup lang="ts">
import type { DataTableColumns } from 'naive-ui'
import { h, ref } from 'vue'
import { fileHistoryAPI } from '~/api/file-history'

// 表格数据
const fileHistories = ref<Api.FileHistory[]>([])
const loading = ref(false)
const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 30, 40],
})

// 搜索表单
const searchForm = ref({
  keyword: '',
  fileType: '',
  fileSuffix: '',
  startTime: '',
  endTime: '',
})

// 加载数据
async function loadData() {
  try {
    loading.value = true
    const { data } = await fileHistoryAPI.findByPage({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      ...searchForm.value,
    })
    if (data?.data) {
      fileHistories.value = data.data.list
      pagination.value.itemCount = data.data.total
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
    startTime: '',
    endTime: '',
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
  { title: '创建时间', key: 'createdAt', width: 180 },
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
  <div class="p-4">
    <NCard title="文件生成记录">
      <!-- 搜索表单 -->
      <NForm
        inline
        :model="searchForm"
        label-placement="left"
        label-width="auto"
        class="mb-4"
      >
        <NFormItem label="关键字">
          <NInput
            v-model:value="searchForm.keyword"
            placeholder="文件名/路径"
            clearable
            @keyup.enter="handleSearch"
          />
        </NFormItem>
        <NFormItem label="文件类型">
          <NInput
            v-model:value="searchForm.fileType"
            placeholder="文件类型"
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
        <!-- 时间范围
        <NFormItem label="时间范围">
          <NDatePicker
            v-model:value="searchForm.startTime"
            type="datetime"
            clearable
            placeholder="开始时间"
          />
          <span class="mx-2">-</span>
          <NDatePicker
            v-model:value="searchForm.endTime"
            type="datetime"
            clearable
            placeholder="结束时间"
          />
        </NFormItem>
        -->
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
        :pagination="pagination"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </NCard>
  </div>
</template>
