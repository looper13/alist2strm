<template>
  <a
    class="btn"
    rel="noreferrer"
    href="https://github.com/MccRay-s/alist2strm"
    target="_blank"
    title="GitHub"
  >
    <div class="i-carbon-logo-github text-xl" />
  </a>
</template>

<style scoped>
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  transition: background-color 0.2s;
}

.btn:hover {
  background-color: rgba(0, 0, 0, 0.06);
}

:deep(.dark) .btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}
</style>
