<template>
  <div class="flex flex-col min-h-full">
    <!-- 主要内容区域 -->
    <div class="flex-1">
      <router-view />
    </div>

    <!-- 页脚 -->
    <footer class="text-sm text-gray-500 py-4 text-center dark:text-gray-400">
      <p class="mb-2">
        <a
          href="https://github.com/MccRay-s/alist2strm"
          target="_blank"
          class="hover:text-gray-700 dark:hover:text-gray-300"
        >
          GitHub 仓库
        </a>
      </p>
      <p>基于 Vue 3 + TypeScript + Vite 构建</p>
    </footer>
  </div>
</template>

<style scoped>
.footer {
  margin-top: auto;
}
</style>
