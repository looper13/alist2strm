import type { NavigationGuardNext, RouteLocationNormalized, Router } from 'vue-router'
import { useAuth } from '~/composables'

// 需要登录才能访问的路由
const AUTH_ROUTES = [
  '/admin',
  '/admin/task',
  '/admin/config',
  '/admin/file-history',
]

/**
 * 检查路由是否需要认证
 */
function checkAuthRoute(route: RouteLocationNormalized) {
  return AUTH_ROUTES.some(path => route.path.startsWith(path))
}

export function setupRouterGuard(router: Router) {
  const { isAuthenticated } = useAuth()

  router.beforeEach((to: RouteLocationNormalized, _from: RouteLocationNormalized, next: NavigationGuardNext) => {
    // 检查是否需要认证
    const requiresAuth = checkAuthRoute(to)

    // 如果需要认证且未登录，重定向到登录页
    if (requiresAuth && !isAuthenticated.value) {
      next({
        path: '/auth',
        query: { redirect: to.fullPath },
      })
      return
    }

    // 如果已登录且访问登录页，重定向到首页
    if (isAuthenticated.value && to.path === '/auth') {
      next('/admin')
      return
    }
    next()
  })
}
