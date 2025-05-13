<script setup lang="ts">
import { h } from 'vue'
import { NButton, NPopconfirm, NSpace, NTag, NText, NTime } from 'naive-ui'
import type { Task } from '@/types/task'

const props = defineProps<{
  tasks: Task[]
  loading: boolean
  pagination: {
    page: number
    pageSize: number
    itemCount: number
    showSizePicker: boolean
    pageSizes: number[]
  }
}>()

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [pageSize: number]
  'edit': [task: Task]
  'delete': [taskId: number]
  'execute': [taskId: number]
  'viewLogs': [task: Task]
  'toggleStatus': [task: Task]
}>()

const columns = [
  {
    title: '任务名称',
    key: 'name',
    width: 200,
    ellipsis: {
      tooltip: true,
    },
    render: (row: Task) => {
      return h('div', { class: 'flex items-center gap-2' }, [
        h('span', { class: 'font-medium' }, row.name),
        h(
          NTag,
          {
            type: row.enabled ? 'success' : 'warning',
            size: 'small',
          },
          { default: () => row.enabled ? '已启用' : '已停用' },
        ),
      ])
    },
  },
  {
    title: 'AList路径',
    key: 'sourcePath',
    width: 200,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: '目标路径',
    key: 'targetPath',
    width: 200,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: '文件后缀',
    key: 'fileSuffix',
    width: 150,
    render: (row: Task) => {
      return h(NSpace, { size: 'small' }, {
        default: () => row.fileSuffix.split(',').map((suffix: string) =>
          h(NTag, { size: 'small' }, { default: () => suffix }),
        ),
      })
    },
  },
  {
    title: '是否覆盖',
    key: 'overwrite',
    width: 100,
    render: (row: Task) => {
      return h(NTag, {
        type: row.overwrite ? 'warning' : 'info',
        size: 'small',
      }, {
        default: () => row.overwrite ? '是' : '否',
      })
    },
  },
  {
    title: '定时任务',
    key: 'cronExpression',
    width: 120,
    render: (row: Task) => {
      return row.cronExpression
        ? h(NTag, { type: 'info' }, { default: () => row.cronExpression })
        : h(NText, { depth: 3 }, { default: () => '未设置' })
    },
  },
  {
    title: '上次运行',
    key: 'lastRunAt',
    width: 180,
    render: (row: Task) => {
      return row.lastRunAt
        ? h(NTime, { time: new Date(row.lastRunAt), type: 'datetime' })
        : h(NText, { depth: 3 }, { default: () => '从未运行' })
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 300,
    fixed: 'right' as const,
    render: (row: Task) => {
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(
            NButton,
            {
              type: row.enabled ? 'warning' : 'success',
              size: 'small',
              loading: props.loading,
              onClick: () => emit('toggleStatus', row),
            },
            { default: () => row.enabled ? '停用' : '启用' },
          ),
          h(
            NButton,
            {
              type: 'info',
              size: 'small',
              loading: props.loading,
              onClick: () => emit('execute', row.id),
            },
            { default: () => '执行' },
          ),
          h(
            NButton,
            {
              type: 'primary',
              size: 'small',
              loading: props.loading,
              onClick: () => emit('viewLogs', row),
            },
            { default: () => '日志' },
          ),
          h(
            NButton,
            {
              type: 'warning',
              size: 'small',
              loading: props.loading,
              onClick: () => emit('edit', row),
            },
            { default: () => '编辑' },
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => emit('delete', row.id),
            },
            {
              default: () => '确定要删除这个任务吗？',
              trigger: () => h(
                NButton,
                {
                  type: 'error',
                  size: 'small',
                  loading: props.loading,
                },
                { default: () => '删除' },
              ),
            },
          ),
        ],
      })
    },
  },
]
</script>

<template>
  <n-data-table
    :columns="columns"
    :data="tasks"
    :loading="loading"
    remote
    :pagination="{
      page: pagination.page,
      pageSize: pagination.pageSize,
      itemCount: pagination.itemCount,
      showSizePicker: pagination.showSizePicker,
      pageSizes: pagination.pageSizes,
      onUpdatePage: (page: number) => emit('update:page', page),
      onUpdatePageSize: (pageSize: number) => emit('update:pageSize', pageSize),
    }"
    :bordered="false"
    :single-line="false"
  />
</template>
