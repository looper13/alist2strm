<script setup lang="ts">
import type { ConfigItem } from './config'
import { configAPI } from '~/api/config'
import { useMobile } from '~/composables'
import { CONFIG_ITEMS, defaultConfigs } from './config'
import ConfigPanel from './ConfigPanel.vue'
import EmbyPanel from './EmbyPanel.vue'
import NotificationPanel from './NotificationPanel.vue'

// 响应式状态
const { isMobile } = useMobile()
const activeTab = ref('ALIST')
const alistConfig = ref<Api.Config.AlistConfig>({ ...defaultConfigs.ALIST })
const strmConfig = ref<Api.Config.StrmConfig>({ ...defaultConfigs.STRM })
const embyConfig = ref<Api.Config.EmbyConfig>({ ...defaultConfigs.EMBY })
const notificationConfig = ref<Api.Config.NotificationConfig>({ ...defaultConfigs.NOTIFICATION_SETTINGS })
const cloudDriveConfig = ref<Api.Config.CloudDriveConfig>({ ...defaultConfigs.CLOUD_DRIVE })
const loading = ref(false)
const saving = ref(false)
const notification = useNotification()

onMounted(async () => {
  loading.value = true
  try {
    const { data, code } = await configAPI.configs()
    if (code !== 0) {
      console.error('加载配置失败:', code)
      return
    }
    if (data) {
      const alistConfigItem = data.find(item => item.code === 'ALIST')
      const strmConfigItem = data.find(item => item.code === 'STRM')
      const embyConfigItem = data.find(item => item.code === 'EMBY')
      const notificationConfigItem = data.find(item => item.code === 'NOTIFICATION_SETTINGS')
      const cloudDriveConfigItem = data.find(item => item.code === 'CLOUD_DRIVE')

      if (alistConfigItem?.value) {
        alistConfig.value = JSON.parse(alistConfigItem.value) as Api.Config.AlistConfig
      }
      if (strmConfigItem?.value) {
        strmConfig.value = JSON.parse(strmConfigItem.value) as Api.Config.StrmConfig
      }
      if (embyConfigItem?.value) {
        embyConfig.value = JSON.parse(embyConfigItem.value) as Api.Config.EmbyConfig
      }
      if (notificationConfigItem?.value) {
        notificationConfig.value = JSON.parse(notificationConfigItem.value) as Api.Config.NotificationConfig
      }
      if (cloudDriveConfigItem?.value) {
        cloudDriveConfig.value = JSON.parse(cloudDriveConfigItem.value) as Api.Config.CloudDriveConfig
      }
    }
  }
  finally {
    loading.value = false
  }
})

// 保存配置
async function handleSave() {
  saving.value = true
  try {
    const { data } = await configAPI.configs()
    const currentConfig = data?.find(item => item.code === activeTab.value)

    let configValue = ''
    if (activeTab.value === 'ALIST') {
      configValue = JSON.stringify(alistConfig.value)
    }
    else if (activeTab.value === 'STRM') {
      configValue = JSON.stringify(strmConfig.value)
    }
    else if (activeTab.value === 'EMBY') {
      configValue = JSON.stringify(embyConfig.value)
    }
    else if (activeTab.value === 'NOTIFICATION_SETTINGS') {
      configValue = JSON.stringify(notificationConfig.value)
    }
    else if (activeTab.value === 'CLOUD_DRIVE') {
      configValue = JSON.stringify(cloudDriveConfig.value)
    }

    if (currentConfig) {
      // 更新现有配置
      await configAPI.update(currentConfig.id, {
        ...currentConfig,
        value: configValue,
      })
    }
    else {
      const defaultConfig = CONFIG_ITEMS.find(item => item.code === activeTab.value)
      if (!defaultConfig) {
        console.error('未找到默认配置项:', activeTab.value)
        return
      }
      const { name, code } = defaultConfig
      // 创建新配置
      await configAPI.create({
        name,
        code,
        value: configValue,
      })
    }
    notification.success({
      title: '操作提示',
      description: '配置已成功保存。',
      duration: 1500,
    })
  }
  finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="config-form">
    <NSpin :show="loading">
      <NTabs v-model:value="activeTab" type="line" justify-content="space-evenly">
        <NTabPane
          v-for="(item, index) in CONFIG_ITEMS"
          :key="index"
          :tab="item.name"
          :name="item.code"
        >
          <ConfigPanel
            v-if="item.code === 'ALIST'"
            v-model="alistConfig"
            :config-item="item as ConfigItem<Api.Config.AlistConfig>"
          />
          <ConfigPanel
            v-if="item.code === 'CLOUD_DRIVE'"
            v-model="cloudDriveConfig"
            :config-item="item as ConfigItem<Api.Config.CloudDriveConfig>"
          />
          <ConfigPanel
            v-if="item.code === 'STRM'"
            v-model="strmConfig"
            :config-item="item as ConfigItem<Api.Config.StrmConfig>"
          />
          <EmbyPanel
            v-if="item.code === 'EMBY'"
            :config="embyConfig"
            :config-item="item as ConfigItem<Api.Config.EmbyConfig>"
            @update:config="(val) => embyConfig = val"
          />
          <NotificationPanel
            v-if="item.code === 'NOTIFICATION_SETTINGS'"
            :config="notificationConfig"
            @update:config="(val) => notificationConfig = val"
          />
        </NTabPane>
      </NTabs>
      <div class="flex justify-end" :class="{ 'mb-12': isMobile }">
        <NButton
          type="primary"
          :loading="saving"
          @click="handleSave"
        >
          保存
        </NButton>
      </div>
    </NSpin>
  </div>
</template>

<style scoped>
.config-form :deep(.n-tab-pane) {
  padding: 16px 0;
}

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
}
</style>
