<script setup lang="ts">
import { configAPI } from '~/api/config'

defineOptions({
  name: 'ConfigForm',
})

// 配置项定义
const CONFIG_ITEMS = [
  {
    title: 'AList 配置',
    prefix: 'ALIST_',
    items: [
      { name: 'Alist地址', code: 'ALIST_HOST', type: 'text', placeholder: 'http://127.0.0.1:5244', describe: '' },
      { name: 'Alist Token', code: 'ALIST_TOKEN', type: 'password', placeholder: 'alist-token-xxxx', describe: '' },
      { name: 'Alist 内容地址', code: 'ALIST_REPLACE_HOST', type: 'text', placeholder: '将内 strm 内容请求地址替换', describe: '' },
      { name: '任务请求间隔', code: 'ALIST_REQ_INTERVAL', type: 'number', placeholder: '请输入任务请求间隔（毫秒）', min: 0, step: 100, describe: '' },
      { name: '请求重试次数', code: 'ALIST_REQ_RETRY_COUNT', type: 'number', placeholder: '请输入请求重试次数', min: 0, step: 1, describe: '' },
      { name: '请求重试间隔', code: 'ALIST_REQ_RETRY_INTERVAL', type: 'number', placeholder: '请输入请求重试间隔（毫秒）', min: 0, step: 100, describe: '' },
    ],
  },
  {
    title: 'strm 配置',
    prefix: 'STRM_',
    items: [
      { name: '替换扩展名', code: 'STRM_REPLACE_SUFFIX', type: 'switch', describe: '开启后将文件扩展名替换为 strm', placeholder: '替换扩展名' },
      { name: 'URL编码', code: 'STRM_URL_ENCODE', type: 'switch', describe: '开启后会对 strm 内容进行 URL 编码', placeholder: '替换扩展名' },
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
      const configMap: Record<string, string | number | boolean> = {}

      data.forEach((item) => {
        CONFIG_ITEMS.forEach((group) => {
          const configItem = group.items.find(c => c.code === item.code)
          if (configItem) {
            if (configItem.type === 'text' || configItem.type === 'password') {
              configMap[item.code] = item.value
            }
            else if (configItem.type === 'number') {
              configMap[item.code] = item.value ? Number(item.value) : 0
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

// 保存所有配置
async function handleSaveAll() {
  // try {
  //   saving.value = true
  //   const { data: existingConfigs } = await configAPI.findAll()
  //   const existingConfigMap = new Map(existingConfigs?.map(c => [c.code, c]) || [])

  //   for (const item of CONFIG_ITEMS) {
  //     const value = configs.value[item.code]
  //     const strValue = item.type === 'number' ? String(value) : value || ''

  //     const existingConfig = existingConfigMap.get(item.code)
  //     if (existingConfig) {
  //       await configAPI.update(existingConfig.id, {
  //         value: strValue,
  //       } as Api.Config.Update)
  //     }
  //     else {
  //       await configAPI.create({
  //         name: item.name,
  //         code: item.code,
  //         value: strValue,
  //       } as Api.Config.Create)
  //     }
  //   }

  //   message.success('保存成功')
  //   // 重新加载配置以确保显示最新数据
  //   await loadConfigs()
  // }
  // catch (error: any) {
  //   message.error(error.message || '保存失败')
  // }
  // finally {
  //   saving.value = false
  // }
}

// 初始化加载
onMounted(() => {
  loadConfigs()
})
</script>

<template>
  <div>
    <NSpin :show="loading">
      <NCard
        v-for="group in CONFIG_ITEMS"
        :key="group.prefix"
        :title="group.title"
        class="mb-4"
      >
        <NForm label-placement="left" label-width="150">
          <NFormItem
            v-for="item in group.items"
            :key="item.code"
            :label="item.name"
          >
            <NInput
              v-if="item.type === 'text'"
              v-model:value="configs[item.code]"
              style="width: 100%"
              :placeholder="item.placeholder"
            />
            <NInputNumber
              v-else-if="item.type === 'number'"
              v-model:value="configs[item.code]"
              style="width: 100%"
              :placeholder="item.placeholder"
              :min="item.min"
              :step="item.step"
            />
            <NInput
              v-else-if="item.type === 'password'"
              v-model:value="configs[item.code]"
              style="width: 100%"
              type="password"
              show-password-on="click"
              :placeholder="item.placeholder"
            />
            <!-- 开关 -->
            <NSwitch
              v-else-if="item.type === 'switch'"
              v-model:value="configs[item.code]"
            />
            <!-- 开关的描述信息 -->
            <span
              v-if="item.type === 'switch' && item.describe"
              class="text-sm text-gray-400 ml-2"
            >
              {{ item.describe }}
            </span>
          </NFormItem>
          <div class="flex justify-end">
            <NFormItem>
              <NButton
                type="primary"
                :loading="saving"
                @click="handleSaveAll"
              >
                保存配置
              </NButton>
            </NFormItem>
          </div>
        </NForm>
      </NCard>
    </NSpin>
  </div>
</template>
