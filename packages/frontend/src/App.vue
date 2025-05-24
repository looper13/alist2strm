<script setup lang="ts">
const route = useRoute()
const isLoginPage = computed(() => route.path === '/auth')
</script>

<template>
  <AppProvider>
    <template v-if="isLoginPage">
      <RouterView />
    </template>
    <BaseLayout v-else>
      <RouterView />
    </BaseLayout>
  </AppProvider>
</template>
