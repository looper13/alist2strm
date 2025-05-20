<template>
  <n-layout position="absolute">
    <n-layout-header style="height: 64px; padding: 24px" bordered>
      <p class="text-2xl text-green-400 font-bold text-center">
        AList2Strm
      </p>
    </n-layout-header>
    <n-layout has-sider position="absolute" style="top: 64px; bottom: 64px">
      <n-layout-sider bordered content-style="padding: 24px;">
        <SiderMenu />
      </n-layout-sider>
      <n-layout content-style="padding: 24px;">
        <n-space vertical size="large">
          <slot />
        </n-space>
      </n-layout>
    </n-layout>
    <n-layout-footer
      bordered
      position="absolute"
      style="height: 64px; padding: 24px"
    >
      <Footer />
    </n-layout-footer>
  </n-layout>
</template>
