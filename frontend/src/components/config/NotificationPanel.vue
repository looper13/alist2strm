<script setup lang="ts">
import type { PropType } from 'vue'
import { NCard, NForm, NFormItem, NInput, NInputNumber, NSelect, NSwitch, NTabPane, NTabs, NTag, useMessage } from 'naive-ui'
import { defineComponent, h, ref, watch } from 'vue'
import { useMobile } from '~/composables'

const props = defineProps<{
  config: Api.Config.NotificationConfig
}>()

const emit = defineEmits<{
  (e: 'update:config', value: Api.Config.NotificationConfig): void
}>()

// 深拷贝配置，避免直接修改props
const notificationConfig = ref<Api.Config.NotificationConfig>(JSON.parse(JSON.stringify(props.config)))

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

// 复制变量到剪贴板
const message = useMessage()
function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(
    () => {
      message.success(`已复制: ${text}`)
    },
    (err) => {
      message.error(`复制失败: ${err}`)
    },
  )
}

// 模板变量项组件
const TemplateVarItem = defineComponent({
  name: 'TemplateVarItem',
  props: {
    var: {
      type: String,
      required: true,
    },
    desc: {
      type: String,
      required: true,
    },
    type: {
      type: String as PropType<'default' | 'error' | 'info' | 'success' | 'warning' | 'primary'>,
      default: 'default',
    },
  },
  setup(props) {
    const handleClick = () => {
      copyToClipboard(props.var)
    }

    return () => h('div', {
      class: 'flex items-center p-1 rounded cursor-pointer transition duration-300 hover:bg-gray-100 dark:hover:bg-gray-800',
      onClick: handleClick,
    }, [
      h(NTag, {
        size: 'small',
        type: props.type,
        class: 'mr-2 cursor-pointer',
      }, { default: () => props.var }),
      h('span', { class: 'text-sm' }, props.desc),
    ])
  },
})
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
          :tab="channelTypes.find(item => item.value === String(channelKey))?.label || channelKey"
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

            <div v-for="field in getChannelConfigFields(String(channelKey))" :key="field.key">
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
          :tab="templateTypes.find(item => item.value === String(templateKey))?.label || templateKey"
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

            <NFormItem label="">
              <NCard title="可用模板变量" size="small" class="mt-2">
                <div class="text-xs text-gray-500 mb-2">
                  点击变量标签可复制到剪贴板
                </div>
                <NTabs type="line" size="small" animated>
                  <NTabPane name="basic" tab="基本信息">
                    <div class="p-2 gap-2 grid grid-cols-1 sm:grid-cols-2">
                      <TemplateVarItem var="&#123;&#123;.TaskName&#125;&#125;" desc="任务名称" type="info" />
                      <TemplateVarItem var="&#123;&#123;.EventTime&#125;&#125;" desc="完成时间" type="info" />
                      <TemplateVarItem var="&#123;&#123;.Duration&#125;&#125;" desc="处理耗时(秒)" type="info" />
                      <TemplateVarItem var="&#123;&#123;.Status&#125;&#125;" desc="任务状态" type="info" />
                    </div>
                  </NTabPane>
                  <NTabPane name="files" tab="文件统计">
                    <div class="p-2 gap-2 grid grid-cols-1 sm:grid-cols-2">
                      <TemplateVarItem var="&#123;&#123;.TotalFile&#125;&#125;" desc="总文件数" type="success" />
                      <TemplateVarItem var="&#123;&#123;.GeneratedFile&#125;&#125;" desc="已生成文件数" type="success" />
                      <TemplateVarItem var="&#123;&#123;.SkipFile&#125;&#125;" desc="已跳过文件数" type="success" />
                      <TemplateVarItem var="&#123;&#123;.OverwriteFile&#125;&#125;" desc="已覆盖文件数" type="success" />
                      <TemplateVarItem var="&#123;&#123;.FailedCount&#125;&#125;" desc="失败文件数" type="success" />
                    </div>
                  </NTabPane>
                  <NTabPane name="extras" tab="元数据与字幕">
                    <div class="p-2 gap-2 grid grid-cols-1 sm:grid-cols-2">
                      <TemplateVarItem var="&#123;&#123;.MetadataCount&#125;&#125;" desc="元数据总数" type="warning" />
                      <TemplateVarItem var="&#123;&#123;.MetadataDownloaded&#125;&#125;" desc="已下载元数据" type="warning" />
                      <TemplateVarItem var="&#123;&#123;.MetadataSkipped&#125;&#125;" desc="已跳过元数据" type="warning" />
                      <TemplateVarItem var="&#123;&#123;.SubtitleCount&#125;&#125;" desc="字幕总数" type="warning" />
                      <TemplateVarItem var="&#123;&#123;.SubtitleDownloaded&#125;&#125;" desc="已下载字幕" type="warning" />
                      <TemplateVarItem var="&#123;&#123;.SubtitleSkipped&#125;&#125;" desc="已跳过字幕" type="warning" />
                    </div>
                  </NTabPane>
                  <NTabPane name="others" tab="其他变量">
                    <div class="p-2 gap-2 grid grid-cols-1 sm:grid-cols-2">
                      <TemplateVarItem var="&#123;&#123;.ErrorMessage&#125;&#125;" desc="错误信息" type="error" />
                      <TemplateVarItem var="&#123;&#123;.OtherSkipped&#125;&#125;" desc="其他跳过文件" type="error" />
                      <TemplateVarItem var="&#123;&#123;.SourcePath&#125;&#125;" desc="源路径" type="default" />
                      <TemplateVarItem var="&#123;&#123;.TargetPath&#125;&#125;" desc="目标路径" type="default" />
                    </div>
                  </NTabPane>
                </NTabs>
              </NCard>
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
