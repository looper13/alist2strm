<script setup lang="ts" generic="T">
import type { ConfigItem } from './config'
import { useMobile } from '~/composables'

const props = defineProps<{
  configItem: ConfigItem<T>
}>()

const { configItem } = toRefs(props)

const configInfo = defineModel<T>('modelValue', {
  type: Object as () => T,
  default: () => ({} as T),
})
const { isMobile } = useMobile()
</script>

<template>
  <NForm
    :label-placement="isMobile ? 'top' : 'left'"
    :label-width="isMobile ? 'auto' : 120"
    require-mark-placement="right-hanging"
  >
    <div
      v-for="field in configItem.fields"
      :key="field.key"
    >
      <NFormItem :label="field.label">
        <div class="flex flex-col w-full">
          <div class="w-full">
            <template v-if="field.type === 'text'">
              <NInput
                v-model:value="configInfo[field.key] as string"
                type="text"
                :placeholder="field.placeholder"
              />
            </template>
            <template v-else-if="field.type === 'number'">
              <NInputNumber
                v-model:value="configInfo[field.key] as number"
                :placeholder="field.placeholder"
                :min="field.min"
                :step="field.step"
                class="w-full"
              />
            </template>
            <template v-else-if="field.type === 'boolean'">
              <NSwitch v-model:value="configInfo[field.key] as boolean" />
            </template>
          </div>
          <div v-if="field.describe" class="text-sm text-gray-500 mt-2">
            {{ field.describe }}
          </div>
        </div>
      </NFormItem>
    </div>
  </NForm>
</template>

<style scoped>
:deep(.n-form-item) {
  margin-bottom: 0;
}

:deep(.n-input-number) {
  width: 100%;
}

:deep(.n-input) {
  width: 100%;
}
</style>
