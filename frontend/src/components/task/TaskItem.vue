<script setup lang="ts" generic="T">
import { useMobile } from '~/composables'

defineProps<{
  item: Api.Task.Record
}>()
defineEmits<{
  (e: 'edit', task: Api.Task.Record): void
  (e: 'copy', task: Api.Task.Record): void
  (e: 'execute', task: Api.Task.Record): void
  (e: 'logs', task: Api.Task.Record): void
  (e: 'delete', task: Api.Task.Record): void
  (e: 'update:enabled', task: Api.Task.Record): void
  (e: 'reset', task: Api.Task.Record): void
}>()
const { isMobile } = useMobile()
</script>

<template>
  <NCard size="small" class="min-w-240px" :title="item.name">
    <template #header-extra>
      <NTooltip trigger="hover">
        <template #trigger>
          <NSwitch
            :value="item.enabled"
            :loading="item.running"
            @update:value="$emit('update:enabled', { ...item, enabled: !item.enabled })"
          />
        </template>
        {{ item.enabled ? '已启用' : '已禁用' }}
      </NTooltip>
    </template>

    <!-- 任务基本信息 -->
    <div class="mb-4 mt-6">
      <div class="text-sm text-gray-500 space-y-1">
        <div class="flex gap-1 items-center">
          <div class="i-ri:time-line" mr-1 />
          <span>{{ item.cron || '未设置定时' }}</span>
        </div>
        <div class="flex gap-1 items-center">
          <div class="i-material-symbols:movie" mr-1 />
          <span>{{ item.mediaType === 'movie' ? '电影' : '电视剧' }}</span>
        </div>
        <div class="flex gap-1 items-center">
          <div class="i-ri:folder-line" mr-1 />
          <span class="truncate" :title="item.sourcePath">{{ item.sourcePath }}</span>
        </div>
        <div class="flex gap-1 items-center">
          <div class="i-ri:folder-transfer-line" mr-1 />
          <span class="truncate" :title="item.targetPath">{{ item.targetPath }}</span>
        </div>
      </div>

      <!-- 扩展功能信息 -->
      <div class="mt-2 flex flex-wrap gap-2">
        <NTooltip :trigger="isMobile ? 'click' : 'hover'">
          <template #trigger>
            <NTag
              size="small"
              :type="item.downloadMetadata ? 'success' : 'default'"
              :style="!item.downloadMetadata ? 'text-decoration: line-through; opacity: 0.6;' : ''"
            >
              <template #icon>
                <div :class="item.downloadMetadata ? 'i-ri:information-line' : 'i-ri:forbid-line'" />
              </template>
              <span class="flex items-center gap-1">
                元数据
                <div v-if="item.downloadMetadata" class="i-ri:check-line text-green-500" />
                <div v-else class="i-ri:close-line text-red-500" />
              </span>
            </NTag>
          </template>
          <div class="text-center">
            {{ item.downloadMetadata ? '已开启元数据下载' : '已关闭元数据下载' }}
            <template v-if="item.downloadMetadata">
              <br />支持格式：{{ item.metadataExtensions || '.nfo,.jpg,.png' }}
            </template>
          </div>
        </NTooltip>

        <NTooltip :trigger="isMobile ? 'click' : 'hover'">
          <template #trigger>
            <NTag
              size="small"
              :type="item.downloadSubtitle ? 'success' : 'default'"
              :style="!item.downloadSubtitle ? 'text-decoration: line-through; opacity: 0.6;' : ''"
            >
              <template #icon>
                <div :class="item.downloadSubtitle ? 'i-ri:subtitle-line' : 'i-ri:forbid-line'" />
              </template>
              <span class="flex items-center gap-1">
                字幕
                <div v-if="item.downloadSubtitle" class="i-ri:check-line text-green-500" />
                <div v-else class="i-ri:close-line text-red-500" />
              </span>
            </NTag>
          </template>
          <div class="text-center">
            {{ item.downloadSubtitle ? '已开启字幕下载' : '已关闭字幕下载' }}
            <template v-if="item.downloadSubtitle">
              <br />支持格式：{{ item.subtitleExtensions || '.srt,.ass,.ssa' }}
            </template>
          </div>
        </NTooltip>
      </div>
    </div>

    <!-- 文件类型标签 -->
    <div class="mb-4 py-2 flex gap-1 overflow-x-auto">
      <NTag v-for="suffix in item.fileSuffix.split(',')" :key="suffix" size="small">
        {{ suffix }}
      </NTag>
    </div>

    <!-- 最后运行时间 -->
    <div class="mb-4 flex items-center justify-between">
      <div class="text-sm text-gray-500 inline-flex items-center">
        <div class="i-ri:history-line" mr-1 />
        <!-- <span>最后运行：</span> -->
        <NTime v-if="item.lastRunAt" :time="new Date(item.lastRunAt)" type="datetime" />
        <span v-else>从未运行</span>
      </div>
      <NTooltip trigger="hover">
        <template #trigger>
          <NTag :type="item.overwrite ? 'warning' : 'success'" size="small">
            {{ item.overwrite ? '覆盖' : '跳过' }}
          </NTag>
        </template>
        {{ item.overwrite ? '文件存在时覆盖' : '文件存在时跳过' }}
      </NTooltip>
    </div>

    <!-- 操作按钮 -->
    <div class="pt-4 border-t border-gray-200 flex justify-between dark:border-gray-700">
      <!-- 左侧重置按钮 -->
      <div>
        <NPopconfirm @positive-click="$emit('reset', item)">
          <template #trigger>
            <NTooltip trigger="hover">
              <template #trigger>
                <NButton
                  :size="isMobile ? 'tiny' : 'small'"
                >
                  <template #icon>
                    <div class="i-tdesign:rollback" />
                  </template>
                </NButton>
              </template>
              重置状态
            </NTooltip>
          </template>
          确认重置该任务吗？这将还原任务的执行状态。
        </NPopconfirm>
      </div>

      <!-- 右侧按钮组 -->
      <div class="flex gap-2">
        <NTooltip trigger="hover">
          <template #trigger>
            <NButton
              :size="isMobile ? 'tiny' : 'small'"
              type="primary"
              @click="$emit('edit', item)"
            >
              <template #icon>
                <div class="i-ri:edit-line" />
              </template>
            </NButton>
          </template>
          编辑
        </NTooltip>

        <NTooltip trigger="hover">
          <template #trigger>
            <NButton
              :size="isMobile ? 'tiny' : 'small'"
              type="info"
              @click="$emit('copy', item)"
            >
              <template #icon>
                <div class="i-ri:file-copy-line" />
              </template>
            </NButton>
          </template>
          复制
        </NTooltip>

        <NTooltip trigger="hover">
          <template #trigger>
            <NButton
              :size="isMobile ? 'tiny' : 'small'"
              type="warning"
              :disabled="item.running"
              @click="$emit('execute', item)"
            >
              <template #icon>
                <div :class="item.running ? 'i-ri:loader-4-line animate-spin' : 'i-ri:play-line'" />
              </template>
            </NButton>
          </template>
          {{ item.running ? '执行中' : '执行' }}
        </NTooltip>

        <NTooltip trigger="hover">
          <template #trigger>
            <NButton
              :size="isMobile ? 'tiny' : 'small'"
              type="info"
              @click="$emit('logs', item)"
            >
              <template #icon>
                <div class="i-ri:file-list-line" />
              </template>
            </NButton>
          </template>
          查看日志
        </NTooltip>

        <NPopconfirm @positive-click="$emit('delete', item)">
          <template #trigger>
            <NTooltip trigger="hover">
              <template #trigger>
                <NButton
                  :size="isMobile ? 'tiny' : 'small'"
                  type="error"
                >
                  <template #icon>
                    <div class="i-ri:delete-bin-line" />
                  </template>
                </NButton>
              </template>
              删除
            </NTooltip>
          </template>
          确认删除该任务吗？
        </NPopconfirm>
      </div>
    </div>
  </NCard>
</template>
