<script setup lang="ts">
import { darkTheme, dateZhCN, zhCN } from 'naive-ui'

const isDark = useDark()
const theme = computed(() => (isDark.value ? darkTheme : null))
</script>

<template>
  <!-- 全局化配置 Config Provider -->
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN" :theme="theme">
    <!-- 全局化 Loading Bar Provider -->
    <n-loading-bar-provider>
      <!-- 全局化 Dialog Provider -->
      <n-dialog-provider>
        <!-- 全局化 Notification Provider -->
        <n-notification-provider>
          <!-- 全局化 Message Provider -->
          <n-message-provider>
            <slot />
          </n-message-provider>
        </n-notification-provider>
      </n-dialog-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>
