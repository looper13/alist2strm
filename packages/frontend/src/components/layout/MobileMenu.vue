<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import IconDashboard from '../icons/IconDashboard.vue'
import IconFiles from '../icons/IconFiles.vue'
import IconSchedule from '../icons/IconSchedule.vue'
import IconSettings from '../icons/IconSettings.vue'

const router = useRouter()
const menuOptions = ref<MenuOption[]>([
  {
    label: '介绍',
    key: '/admin/',
    icon: renderIcon(IconDashboard),
  },
  {
    label: '基础配置',
    key: '/admin/config',
    icon: renderIcon(IconSettings),
  },
  {
    label: '任务管理',
    key: '/admin/task',
    icon: renderIcon(IconSchedule),
  },
  {
    label: '生成记录',
    key: '/admin/file-history',
    icon: renderIcon(IconFiles),
  },
])

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

function handleUpdateValue(key: string) {
  router.push(key)
}
</script>

<template>
  <NFloatButton
    right="16px"
    bottom="16px"
    menu-trigger="click"
  >
    <div class="i-carbon-menu text-xl" />
    <template #menu>
      <NFloatButton v-for="option in menuOptions" :key="option.key" @click="handleUpdateValue(option.key as string)">
        <NIcon :component="option.icon" />
      </NFloatButton>
    </template>
  </NFloatButton>
</template>
