<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import { usePreferredDark } from '@vueuse/core'

const isDark = usePreferredDark()
const theme = computed(() => (isDark.value ? darkTheme : null))
</script>

<template>
  <n-config-provider :theme="theme">
    <n-loading-bar-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <n-message-provider>
            <slot />
          </n-message-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>
