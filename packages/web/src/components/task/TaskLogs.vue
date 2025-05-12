<script setup lang="ts">
import type { Task, TaskLog } from '@/types/task'

defineProps<{
  task: Task | null
  logs: TaskLog[]
  loading: boolean
}>()
</script>

<template>
  <n-spin :show="loading">
    <n-scrollbar style="max-height: 500px">
      <n-empty v-if="logs.length === 0" description="暂无日志" />
      <n-timeline v-else>
        <n-timeline-item
          v-for="log in logs"
          :key="log.id"
          :type="log.status === 'success' ? 'success' : 'error'"
          :title="log.status"
        >
          <n-card size="small" :class="log.status === 'success' ? 'bg-green-50' : 'bg-red-50'">
            <n-space vertical>
              <n-space align="center">
                <div class="i-carbon-time" />
                <n-time :time="new Date(log.startTime)" type="datetime" />
              </n-space>
              <n-space v-if="log.endTime" align="center">
                <div class="i-carbon-checkmark" />
                <n-time :time="new Date(log.endTime)" type="datetime" />
              </n-space>
              <n-space v-if="log.status === 'success'" justify="space-around">
                <n-statistic label="总文件数" :value="log.totalFiles || 0" />
                <n-statistic label="生成文件数" :value="log.generatedFiles || 0" />
                <n-statistic label="跳过文件数" :value="log.skippedFiles || 0" />
              </n-space>
              <div v-if="log.error" class="text-red-600">
                <n-alert type="error" :title="log.error" />
              </div>
            </n-space>
          </n-card>
        </n-timeline-item>
      </n-timeline>
    </n-scrollbar>
  </n-spin>
</template>
