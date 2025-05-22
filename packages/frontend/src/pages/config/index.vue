<script setup lang="ts">
import { configAPI } from '~/api/config'

defineOptions({
  name: 'ConfigIndexPage',
})
// 配置项定义
const CONFIG_ITEMS = [
  { name: 'Alist地址', code: 'ALIST_HOST', type: 'text', placeholder: 'http://127.0.0.1:5244' },
  { name: 'Alist Token', code: 'ALIST_TOKEN', type: 'password', placeholder: 'alist-token-xxxx' },
  { name: 'AList请求并发', code: 'ALIST_REQ_CONCURRENCY', type: 'number', placeholder: '设置后会并发多个请求同时进行', min: 0, step: 1 },
  { name: '任务请求间隔', code: 'ALIST_REQ_INTERVAL', type: 'number', placeholder: '请输入任务请求间隔（毫秒）', min: 0, step: 100 },
  { name: '请求重试次数', code: 'ALIST_REQ_RETRY_COUNT', type: 'number', placeholder: '请输入请求重试次数', min: 0, step: 1 },
  { name: '请求重试间隔', code: 'ALIST_REQ_RETRY_INTERVAL', type: 'number', placeholder: '请输入请求重试间隔（毫秒）', min: 0, step: 100 },
]
// 状态定义
const loading = ref(false)
const saving = ref(false)
const configs = ref<Record<string, any>>({})

// 消息提示
const message = useMessage()

// 加载配置
async function loadConfigs() {
  try {
    loading.value = true
    const { data } = await configAPI.findAll()
    if (data) {
      const configMap: Record<string, any> = {}
      data.forEach((item) => {
        // 根据配置项类型转换值
        const configItem = CONFIG_ITEMS.find(c => c.code === item.code)
        if (configItem?.type === 'number')
          configMap[item.code] = item.value ? Number(item.value) : 0
        else
          configMap[item.code] = item.value
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
  try {
    saving.value = true
    const { data: existingConfigs } = await configAPI.findAll()
    const existingConfigMap = new Map(existingConfigs?.map(c => [c.code, c]) || [])

    for (const item of CONFIG_ITEMS) {
      const value = configs.value[item.code]
      const strValue = item.type === 'number' ? String(value) : value || ''

      const existingConfig = existingConfigMap.get(item.code)
      if (existingConfig) {
        await configAPI.update(existingConfig.id, {
          value: strValue,
        } as Api.Config.Update)
      }
      else {
        await configAPI.create({
          name: item.name,
          code: item.code,
          value: strValue,
        } as Api.Config.Create)
      }
    }

    message.success('保存成功')
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
  <div>
    <NSpin :show="loading">
      <NCard title="系统配置" class="mb-4">
        <NForm label-placement="left" label-width="200">
          <NFormItem
            v-for="item in CONFIG_ITEMS"
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
