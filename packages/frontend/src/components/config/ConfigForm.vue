<script setup lang="ts">
import { configAPI } from '~/api/config'
import { useMobile } from '~/composables'

defineOptions({
  name: 'ConfigForm',
})

const { isMobile } = useMobile()

interface ConfigItem {
  name: string
  code: string
  type: 'text' | 'number' | 'password' | 'switch'
  placeholder: string
  describe: string
  min?: number
  step?: number
}

interface ConfigGroup {
  title: string
  prefix: string
  items: ConfigItem[]
}

// 配置项定义
const CONFIG_ITEMS: ConfigGroup[] = [
  {
    title: 'AList 配置',
    prefix: 'ALIST_',
    items: [
      { name: 'Alist地址', code: 'ALIST_HOST', type: 'text', placeholder: 'http://127.0.0.1:5244', describe: 'Alist 服务器地址，建议内网地址' },
      { name: 'Alist Token', code: 'ALIST_TOKEN', type: 'password', placeholder: 'alist-token-xxxx', describe: 'Alist 访问令牌' },
      { name: 'Alist 域名', code: 'ALIST_REPLACE_HOST', type: 'text', placeholder: '将内 strm 内容请求地址替换', describe: '用于替换 strm 文件中的请求地址，优先级高于 Alist地址，建议外网域名或地址' },
      { name: '任务请求间隔', code: 'ALIST_REQ_INTERVAL', type: 'number', placeholder: '请输入任务请求间隔（毫秒）', min: 0, step: 100, describe: '每次请求之间的间隔时间，默认100' },
      { name: '请求重试次数', code: 'ALIST_REQ_RETRY_COUNT', type: 'number', placeholder: '请输入请求重试次数', min: 0, step: 1, describe: '请求失败时的重试次数，默认3次' },
      { name: '请求重试间隔', code: 'ALIST_REQ_RETRY_INTERVAL', type: 'number', placeholder: '请输入请求重试间隔（毫秒）', min: 0, step: 100, describe: '重试请求之间的间隔时间，建议大于任务请求间隔，默认10000' },
    ],
  },
  {
    title: 'strm 配置',
    prefix: 'STRM_',
    items: [
      {
        name: '替换扩展名',
        code: 'STRM_REPLACE_SUFFIX',
        type: 'switch',
        describe: '开启后，生成的 strm 文件则不包含源文件的扩展名，例如：test.mp4 将生成 test.strm',
        placeholder: '替换扩展名',
      },
      {
        name: 'URL编码',
        code: 'STRM_URL_ENCODE',
        type: 'switch',
        describe: '开启后会对 strm 内容进行URL编码，建议开启',
        placeholder: 'URL编码',
      },
    ],
  },
]

// 状态定义
const loading = ref(false)
const saving = ref(false)
const originConfig = ref<Api.Config.Record[]>([])
const configs = ref<Record<string, any>>({})

// 消息提示
const message = useMessage()

// 加载配置
async function loadConfigs() {
  try {
    loading.value = true
    const { data } = await configAPI.findAll()
    originConfig.value = data || []
    if (data) {
      const configMap: Record<string, any> = {}

      data.forEach((item) => {
        CONFIG_ITEMS.forEach((group) => {
          const configItem = group.items.find(c => c.code === item.code)
          if (configItem) {
            if (configItem.type === 'text' || configItem.type === 'password') {
              configMap[item.code] = item.value || ''
            }
            else if (configItem.type === 'number') {
              const numValue = item.value ? Number(item.value) : 0
              configMap[item.code] = Number.isNaN(numValue) ? 0 : numValue
            }
            else if (configItem.type === 'switch') {
              configMap[item.code] = item.value === 'Y'
            }
          }
        })
      })
      configs.value = configMap
    }
  }
  catch (error: any) {
    message.error(error.message || '加载失败')
  }
  finally {
    loading.value = false
  }
}

// 获取配置的当前值的字符串表示
function getConfigValue(code: string, type: string, value: any): string {
  if (value === null || value === undefined)
    return ''

  if (type === 'switch')
    return value ? 'Y' : 'N'
  if (type === 'number')
    return String(value)
  return String(value)
}

// 检查配置是否有变化
function hasConfigChanged(code: string, type: string, value: any): boolean {
  const originalItem = originConfig.value.find(item => item.code === code)
  if (!originalItem)
    return true // 新配置项

  const currentValue = getConfigValue(code, type, value)
  return currentValue !== originalItem.value
}

// 处理输入值更新
function handleValueUpdate(code: string, type: string, value: any) {
  if (type === 'number') {
    configs.value[code] = value === null ? 0 : Number(value)
  }
  else {
    configs.value[code] = value
  }
}

// 保存所有配置
async function handleSaveAll() {
  try {
    saving.value = true
    const changedConfigs: Array<{ code: string, value: string, type: string, name: string, id?: number }> = []

    // 收集变更的配置
    CONFIG_ITEMS.forEach((group) => {
      group.items.forEach((item) => {
        const value = configs.value[item.code]
        if (hasConfigChanged(item.code, item.type, value)) {
          const originalItem = originConfig.value.find(c => c.code === item.code)
          changedConfigs.push({
            code: item.code,
            value: getConfigValue(item.code, item.type, value),
            type: item.type,
            name: item.name,
            id: originalItem?.id,
          })
        }
      })
    })

    if (changedConfigs.length === 0) {
      message.info('没有配置发生变化')
      return
    }

    // 保存变更的配置
    for (const config of changedConfigs) {
      if (config.id) {
        await configAPI.update(config.id, {
          value: config.value,
        } as Api.Config.Update)
      }
      else {
        await configAPI.create({
          name: config.name,
          code: config.code,
          value: config.value,
        } as Api.Config.Create)
      }
    }

    message.success(`成功保存 ${changedConfigs.length} 项配置`)
    // 重新加载配置以确保显示最新数据
    await loadConfigs()
  }
  catch (error: any) {
    message.error(error.message || '保存失败')
  }
  finally {
    saving.value = false
  }
}

// 初始化加载
onMounted(() => {
  loadConfigs()
})
</script>

<template>
  <div class="mx-auto max-w-4xl">
    <NSpin :show="loading">
      <NCard
        v-for="group in CONFIG_ITEMS"
        :key="group.prefix"
        :title="group.title"
        class="mb-4 shadow-sm transition-shadow duration-300 hover:shadow-md"
      >
        <NForm
          :label-placement="isMobile ? 'top' : 'left'"
          :label-width="isMobile ? 'auto' : 120"
          require-mark-placement="right-hanging"
          size="medium"
          :show-feedback="false"
        >
          <NGrid :cols="1" :x-gap="16" :y-gap="16">
            <NGridItem
              v-for="item in group.items"
              :key="item.code"
            >
              <NFormItem
                :label="item.name"
                :show-require-mark="false"
              >
                <div class="flex flex-col w-full">
                  <div class="w-full">
                    <NInput
                      v-if="item.type === 'text'"
                      :value="configs[item.code] as string"
                      type="text"
                      :placeholder="item.placeholder"
                      @update:value="val => handleValueUpdate(item.code, item.type, val)"
                    />
                    <NInputNumber
                      v-else-if="item.type === 'number'"
                      :value="configs[item.code] as number"
                      :placeholder="item.placeholder"
                      :min="item.min"
                      :step="item.step"
                      class="w-full"
                      @update:value="val => handleValueUpdate(item.code, item.type, val)"
                    />
                    <NInput
                      v-else-if="item.type === 'password'"
                      :value="configs[item.code] as string"
                      type="password"
                      show-password-on="click"
                      :placeholder="item.placeholder"
                      @update:value="val => handleValueUpdate(item.code, item.type, val)"
                    />
                    <NSwitch
                      v-else-if="item.type === 'switch'"
                      :value="configs[item.code] as boolean"
                      @update:value="val => handleValueUpdate(item.code, item.type, val)"
                    />
                  </div>
                  <div v-if="item.describe" class="text-sm text-gray-500 mt-2">
                    {{ item.describe }}
                  </div>
                </div>
              </NFormItem>
            </NGridItem>
          </NGrid>
        </NForm>
      </NCard>

      <!-- 保存按钮 -->
      <div class="mt-4 flex justify-end">
        <NButton
          type="primary"
          :loading="saving"
          class="w-full md:w-auto"
          @click="handleSaveAll"
        >
          保存配置
        </NButton>
      </div>
    </NSpin>
  </div>
</template>

<style scoped>
.n-form-item {
  margin-bottom: 0;
}

:deep(.n-input-number) {
  width: 100%;
}

:deep(.n-input) {
  width: 100%;
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

  :deep(.n-input),
  :deep(.n-input-number),
  :deep(.n-input-number-input) {
    width: 100% !important;
    max-width: 100% !important;
  }
}
</style>
