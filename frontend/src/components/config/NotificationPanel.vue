<script setup lang="ts">
import type { NotificationConfig } from '~/types/notification'
import { ref, watch } from 'vue'
import { useMobile } from '~/composables'

const props = defineProps<{
  config: NotificationConfig
}>()

const emit = defineEmits<{
  (e: 'update:config', value: NotificationConfig): void
}>()

// 深拷贝配置，避免直接修改props
const notificationConfig = ref<NotificationConfig>(JSON.parse(JSON.stringify(props.config)))

// 监听外部配置变化
watch(() => props.config, (newVal) => {
  if (JSON.stringify(notificationConfig.value) !== JSON.stringify(newVal)) {
    notificationConfig.value = JSON.parse(JSON.stringify(newVal))
  }
}, { deep: true })

// 监听内部配置变化，向外发送更新
watch(notificationConfig, (newVal) => {
  if (JSON.stringify(props.config) !== JSON.stringify(newVal)) {
    emit('update:config', JSON.parse(JSON.stringify(newVal)))
  }
}, { deep: true })

const { isMobile } = useMobile()

// 通知渠道列表
const channelTypes = [
  { label: 'Telegram', value: 'telegram' },
  { label: '企业微信', value: 'wework' },
]

// 模板类型列表
const templateTypes = [
  { label: '任务完成通知', value: 'taskComplete' },
  { label: '任务失败通知', value: 'taskFailed' },
]

// 获取可用的渠道列表
const availableChannels = computed(() => {
  return Object.keys(notificationConfig.value.channels).map(key => ({
    label: channelTypes.find(item => item.value === key)?.label || key,
    value: key,
  }))
})

// 当前选中的渠道
const currentChannel = ref(Object.keys(notificationConfig.value.channels)[0] || 'telegram')

// 当前选中的模板
const currentTemplate = ref(Object.keys(notificationConfig.value.templates)[0] || 'taskComplete')

// 切换渠道
function handleChannelChange(channel: string) {
  currentChannel.value = channel
}

// 切换模板
function handleTemplateChange(template: string) {
  currentTemplate.value = template
}

// 获取当前渠道的配置字段列表
function getChannelConfigFields(type: string) {
  if (type === 'telegram') {
    return [
      { key: 'botToken', label: 'Bot Token', placeholder: '请输入 Telegram Bot Token' },
      { key: 'chatId', label: 'Chat ID', placeholder: '请输入 Telegram Chat ID' },
      { key: 'parseMode', label: '解析模式', placeholder: 'Markdown 或 HTML' },
    ]
  }
  else if (type === 'wework') {
    return [
      { key: 'corpId', label: '企业ID', placeholder: '请输入企业微信企业ID' },
      { key: 'agentId', label: '应用ID', placeholder: '请输入企业微信应用ID' },
      { key: 'corpSecret', label: '应用密钥', placeholder: '请输入企业微信应用密钥' },
      { key: 'toUser', label: '接收用户', placeholder: '接收者，默认为@all' },
    ]
  }
  return []
}
</script>

<template>
  <div class="notification-panel">
    <NCard title="基本设置" class="mb-4">
      <NForm
        :label-placement="isMobile ? 'top' : 'left'"
        :label-width="isMobile ? 'auto' : 120"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="启用通知">
          <NSwitch v-model:value="notificationConfig.enabled" />
        </NFormItem>

        <NFormItem label="默认通知渠道">
          <NSelect
            v-model:value="notificationConfig.defaultChannel"
            :options="availableChannels"
            placeholder="请选择默认通知渠道"
          />
        </NFormItem>
      </NForm>
    </NCard>

    <NCard title="通知渠道配置" class="mb-4">
      <NTabs v-model:value="currentChannel" type="segment" @update:value="handleChannelChange">
        <NTabPane
          v-for="(_, channelKey) in notificationConfig.channels"
          :key="channelKey"
          :tab="channelTypes.find(item => item.value === channelKey)?.label || channelKey"
          :name="channelKey"
        >
          <NForm
            :label-placement="isMobile ? 'top' : 'left'"
            :label-width="isMobile ? 'auto' : 120"
            require-mark-placement="right-hanging"
          >
            <NFormItem label="启用渠道">
              <NSwitch v-model:value="notificationConfig.channels[channelKey].enabled" />
            </NFormItem>

            <div v-for="field in getChannelConfigFields(channelKey)" :key="field.key">
              <NFormItem :label="field.label">
                <NInput
                  v-model:value="notificationConfig.channels[channelKey].config[field.key]"
                  :placeholder="field.placeholder"
                />
              </NFormItem>
            </div>
          </NForm>
        </NTabPane>
      </NTabs>
    </NCard>

    <NCard title="通知模板配置" class="mb-4">
      <NTabs v-model:value="currentTemplate" type="segment" @update:value="handleTemplateChange">
        <NTabPane
          v-for="(_, templateKey) in notificationConfig.templates"
          :key="templateKey"
          :tab="templateTypes.find(item => item.value === templateKey)?.label || templateKey"
          :name="templateKey"
        >
          <NForm
            :label-placement="isMobile ? 'top' : 'left'"
            :label-width="isMobile ? 'auto' : 120"
            require-mark-placement="right-hanging"
          >
            <NFormItem label="Telegram模板">
              <NInput
                v-model:value="notificationConfig.templates[templateKey].telegram"
                type="textarea"
                :autosize="{ minRows: 3, maxRows: 6 }"
                placeholder="请输入 Telegram 消息模板"
              />
            </NFormItem>

            <NFormItem label="企业微信模板">
              <NInput
                v-model:value="notificationConfig.templates[templateKey].wework"
                type="textarea"
                :autosize="{ minRows: 3, maxRows: 6 }"
                placeholder="请输入企业微信消息模板"
              />
            </NFormItem>
          </NForm>
        </NTabPane>
      </NTabs>
    </NCard>

    <NCard title="队列设置">
      <NForm
        :label-placement="isMobile ? 'top' : 'left'"
        :label-width="isMobile ? 'auto' : 120"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="最大重试次数">
          <NInputNumber
            v-model:value="notificationConfig.queueSettings.maxRetries"
            :min="0"
            :step="1"
            class="w-full"
          />
        </NFormItem>

        <NFormItem label="重试间隔(秒)">
          <NInputNumber
            v-model:value="notificationConfig.queueSettings.retryInterval"
            :min="10"
            :step="10"
            class="w-full"
          />
        </NFormItem>

        <NFormItem label="并发数">
          <NInputNumber
            v-model:value="notificationConfig.queueSettings.concurrency"
            :min="1"
            :step="1"
            class="w-full"
          />
        </NFormItem>
      </NForm>
    </NCard>
  </div>
</template>

<style scoped>
.notification-panel :deep(.n-form-item) {
  margin-bottom: 16px;
}

.notification-panel :deep(.n-input-number) {
  width: 100%;
}

.notification-panel :deep(.n-card + .n-card) {
  margin-top: 16px;
}
</style>
