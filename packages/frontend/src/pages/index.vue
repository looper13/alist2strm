<script setup lang="ts" generic="T extends any, O extends any">
defineOptions({
  name: 'IndexPage',
})
</script>

<template>
  <div class="mx-auto px-4 py-8 container">
    <!-- 标题区域 -->
    <div class="mb-12 text-center">
      <h1 class="text-4xl font-bold mb-4">
        AList2Strm
      </h1>
      <p class="text-lg text-gray-600">
        将 AList 媒体库转换为 Strm 格式的流媒体服务
      </p>
    </div>

    <!-- 功能特点 -->
    <div class="mb-12 gap-8 grid grid-cols-1 md:grid-cols-3">
      <div class="p-6 rounded-lg bg-white shadow-md">
        <div class="i-carbon-media-library text-2xl mb-4" />
        <h3 class="text-xl font-semibold mb-2">
          AList 集成
        </h3>
        <p class="text-gray-600">
          无缝对接 AList 媒体库，支持多种存储源，支持Alist sign
        </p>
      </div>
      <div class="p-6 rounded-lg bg-white shadow-md">
        <div class="i-carbon-video text-2xl mb-4" />
        <h3 class="text-xl font-semibold mb-2">
          Strm 转换
        </h3>
        <p class="text-gray-600">
          自动将媒体文件转换为 Strm 格式，支持主流播放器
        </p>
      </div>
      <div class="p-6 rounded-lg bg-white shadow-md">
        <div class="i-carbon-settings text-2xl mb-4" />
        <h3 class="text-xl font-semibold mb-2">
          简单配置
        </h3>
        <p class="text-gray-600">
          直观的配置界面，轻松管理媒体库和转换规则
        </p>
      </div>
    </div>

    <!-- 快速开始 -->
    <div class="mb-12 p-8 rounded-lg bg-gray-50">
      <h2 class="text-2xl font-bold mb-4">
        快速开始
      </h2>
      <div class="space-y-4">
        <p class="text-gray-600">
          1. 配置 AList 服务器地址和认证信息
        </p>
        <p class="text-gray-600">
          2. 选择需要转换的媒体库目录
        </p>
        <p class="text-gray-600">
          3. 设置转换规则和输出目录
        </p>
        <p class="text-gray-600">
          4. 开始转换，享受流媒体服务
        </p>
      </div>
    </div>

    <!-- 项目信息 -->
    <div class="text-sm text-gray-500 text-center">
      <p class="mb-2">
        <a href="https://github.com/MccRay-s/alist2strm" target="_blank" class="hover:text-gray-700">
          GitHub 仓库
        </a>
      </p>
      <p>基于 Vue 3 + TypeScript + Vite 构建</p>
    </div>
  </div>
</template>

<style scoped>
</style>
