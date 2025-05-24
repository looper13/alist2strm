<script setup lang="ts">
defineOptions({
  name: 'ConfigIndexPage',
})
</script>

<template>
  <ConfigForm />
</template>

<route lang="yaml">
name: config
layout: default
path: /admin/config
</route>
