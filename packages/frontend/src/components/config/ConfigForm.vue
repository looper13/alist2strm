<script setup lang="ts">
import type { AlistConfig, ConfigItem, StrmConfig } from './config'
import { configAPI } from '~/api/config'
import { useMobile } from '~/composables'
import { CONFIG_ITEMS, defaultConfigs } from './config'
import ConfigPanel from './ConfigPanel.vue'

// 响应式状态
const { isMobile } = useMobile()
const activeTab = ref('ALIST')
const alistConfig = ref<AlistConfig>({ ...defaultConfigs.ALIST })
const strmConfig = ref<StrmConfig>({ ...defaultConfigs.STRM })
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
      if (alistConfigItem?.value) {
        alistConfig.value = JSON.parse(alistConfigItem.value) as AlistConfig
      }
      if (strmConfigItem?.value) {
        strmConfig.value = JSON.parse(strmConfigItem.value) as StrmConfig
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

    if (currentConfig) {
      // 更新现有配置
      await configAPI.update(currentConfig.id, {
        ...currentConfig,
        value: activeTab.value === 'ALIST' ? JSON.stringify(alistConfig.value) : JSON.stringify(strmConfig.value),
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
        value: JSON.stringify(activeTab.value === 'ALIST' ? alistConfig.value : strmConfig.value),
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
            :config-item="item as ConfigItem<AlistConfig>"
          />
          <ConfigPanel
            v-if="item.code === 'STRM'"
            v-model="strmConfig"
            :config-item="item as ConfigItem<StrmConfig>"
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
