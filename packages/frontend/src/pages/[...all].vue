<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <div class="error-code">
        404
      </div>
      <h1 class="error-title">
        页面未找到
      </h1>
      <p class="error-message">
        抱歉，您访问的页面不存在
      </p>
      <router-link to="/admin" class="back-home">
        返回首页
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.not-found-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
}

.not-found-content {
  text-align: center;
  padding: 2rem;
}

.error-code {
  font-size: 6rem;
  font-weight: bold;
  color: #dc3545;
  margin-bottom: 1rem;
}

.error-title {
  font-size: 2rem;
  color: #343a40;
  margin-bottom: 1rem;
}

.error-message {
  font-size: 1.2rem;
  color: #6c757d;
  margin-bottom: 2rem;
}

.back-home {
  display: inline-block;
  padding: 0.8rem 1.5rem;
  background-color: #007bff;
  color: white;
  text-decoration: none;
  border-radius: 5px;
  transition: background-color 0.3s;
}

.back-home:hover {
  background-color: #0056b3;
}
</style>

<route lang="yaml">
name: not-found
layout: empty
path: "/:all(.*)*"
</route>
