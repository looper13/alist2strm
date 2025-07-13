import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { routes } from 'vue-router/auto-routes'
import { http } from './api/http'
import App from './App.vue'
import { setupRouterGuard } from './router/guard'

import './styles/main.css'
import 'uno.css'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// 设置路由守卫
setupRouterGuard(router)

// 设置 HTTP 客户端的路由实例
http.setRouter(router)

const app = createApp(App)
app.use(router)
app.mount('#app')
