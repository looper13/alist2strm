<script setup lang="ts">
import { ref, watch } from 'vue'
import IconDelete from '~/components/icons/IconDelete.vue'
import { useMobile } from '~/composables'

const props = defineProps<{
  config: Api.Config.EmbyConfig
}>()

const emit = defineEmits<{
  (e: 'update:config', value: Api.Config.EmbyConfig): void
}>()

// 深拷贝配置，避免直接修改props
const embyConfig = ref<Api.Config.EmbyConfig>(JSON.parse(JSON.stringify(props.config)))

// 监听外部配置变化
watch(() => props.config, (newVal) => {
  if (JSON.stringify(embyConfig.value) !== JSON.stringify(newVal)) {
    embyConfig.value = JSON.parse(JSON.stringify(newVal))
  }
}, { deep: true })

// 监听内部配置变化，向外发送更新
watch(embyConfig, (newVal) => {
  if (JSON.stringify(props.config) !== JSON.stringify(newVal)) {
    emit('update:config', JSON.parse(JSON.stringify(newVal)))
  }
}, { deep: true })

const { isMobile } = useMobile()

// 添加新的路径映射
function addPathMapping() {
  embyConfig.value.pathMappings.push({
    path: '',
    embyPath: '',
  })
}

// 删除路径映射
function removePathMapping(index: number) {
  embyConfig.value.pathMappings.splice(index, 1)
}
</script>

<template>
  <div class="emby-panel">
    <NCard title="Emby服务器配置" class="mb-4">
      <NForm
        :label-placement="isMobile ? 'top' : 'left'"
        :label-width="isMobile ? 'auto' : 120"
        require-mark-placement="right-hanging"
      >
        <NFormItem label="服务器地址">
          <NInput
            v-model:value="embyConfig.embyServer"
            placeholder="例如: http://emby:8096"
            clearable
          />
        </NFormItem>

        <NFormItem label="API密钥">
          <NInput
            v-model:value="embyConfig.embyToken"
            placeholder="请输入Emby API密钥/Token"
            type="password"
            show-password-on="click"
            clearable
          />
        </NFormItem>
      </NForm>
    </NCard>

    <NCard title="路径映射配置" class="mb-4">
      <div class="path-mapping-container">
        <div v-for="(mapping, index) in embyConfig.pathMappings" :key="index" class="path-mapping-card">
          <div class="path-mapping-header">
            <span class="path-mapping-title">路径映射 {{ index + 1 }}</span>
            <NButton
              type="error"
              quaternary
              size="small"
              :disabled="embyConfig.pathMappings.length <= 1"
              @click="removePathMapping(index)"
            >
              <template #icon>
                <NIcon :component="IconDelete" />
              </template>
              <!-- <span v-if="!isMobile">删除</span> -->
            </NButton>
          </div>

          <div class="path-mapping-content">
            <NInput
              v-model:value="mapping.path"
              placeholder="本地路径，例如: /media/movies"
              clearable
              class="path-input"
            />
            <div class="path-arrow">
              <NIcon>
                <div class="i-carbon-arrow-right" />
              </NIcon>
            </div>
            <NInput
              v-model:value="mapping.embyPath"
              placeholder="Emby路径，例如: /媒体库/电影"
              clearable
              class="emby-input"
            />
          </div>
        </div>

        <div class="add-mapping-btn">
          <NButton type="primary" @click="addPathMapping">
            <template #icon>
              <NIcon>
                <div class="i-carbon-add" />
              </NIcon>
            </template>
            添加路径映射
          </NButton>
        </div>
      </div>
    </NCard>

    <NCard title="配置说明">
      <div class="text-sm text-gray-600 dark:text-gray-400">
        <p>
          <strong>Emby服务器地址</strong>: 您的Emby服务器的完整URL地址，包含协议和端口号。
        </p>
        <p class="mt-2">
          <strong>API密钥</strong>: 在Emby管理界面中生成的API密钥，用于授权访问Emby API。
        </p>
        <p class="mt-2">
          <strong>路径映射</strong>: 设置本地文件系统路径与Emby服务器路径的映射关系，解决Docker环境下的路径差异问题。
        </p>
        <ul class="mt-2 pl-6 list-disc">
          <li>
            <strong>本地路径</strong>: AList2Strm容器内看到的文件路径，即任务目标路径的前缀。
          </li>
          <li>
            <strong>Emby路径</strong>: Emby容器内看到的对应文件路径。
          </li>
        </ul>
        <p class="mt-2">
          <em>例如: 如果您的AList2Strm容器内路径是 /media/movies，而Emby内看到的是 /媒体库/电影，那么添加这对映射后，系统会自动完成路径转换。</em>
        </p>
      </div>
    </NCard>
  </div>
</template>

<style scoped>
.emby-panel :deep(.n-form-item) {
  margin-bottom: 16px;
}

.emby-panel :deep(.n-card + .n-card) {
  margin-top: 16px;
}

.path-mapping-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.path-mapping-card {
  background-color: var(--n-card-color);
  border-radius: 6px;
  padding: 12px;
  border: 1px solid var(--n-border-color);
  transition: all 0.3s ease;
}

.path-mapping-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.09);
}

.path-mapping-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.path-mapping-title {
  font-weight: 500;
  color: var(--n-text-color);
}

.path-mapping-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.path-input,
.emby-input {
  flex: 1;
  min-width: 180px;
}

.path-arrow {
  display: flex;
  justify-content: center;
  align-items: center;
  color: var(--n-text-color-3);
  flex-shrink: 0;
}

.add-mapping-btn {
  display: flex;
  justify-content: center;
  margin-top: 8px;
}

/* 移动端适配 */
@media (max-width: 640px) {
  .path-mapping-content {
    flex-direction: column;
    align-items: stretch;
  }
  .path-arrow {
    transform: rotate(90deg);
    margin: 4px 0;
  }
}
</style>
