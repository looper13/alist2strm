<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import IconDashboard from '../icons/IconDashboard.vue'
import IconFiles from '../icons/IconFiles.vue'
import IconSchedule from '../icons/IconSchedule.vue'
import IconSettings from '../icons/IconSettings.vue'

const router = useRouter()
const route = useRoute()
const isCollapsed = ref(false)
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

const activeKey = computed(() => route.path)
</script>

<template>
  <n-layout-sider
    bordered
    collapse-mode="width"
    :collapsed-width="64"
    :width="200"
    show-trigger
    @collapse="isCollapsed = true"
    @expand="isCollapsed = false"
  >
    <NMenu
      :value="activeKey"
      :collapsed="isCollapsed"
      :collapsed-width="64"
      :collapsed-icon-size="22"
      :options="menuOptions"
      @update:value="handleUpdateValue"
    />
  </n-layout-sider>
</template>
