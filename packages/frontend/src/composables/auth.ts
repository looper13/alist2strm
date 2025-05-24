import { authAPI } from '~/api/auth'

export const useAuth = createGlobalState(() => {
  // 持久化存储 token
  const token = useStorage('token', '')

  // 用户信息
  const userInfo = useStorage<Api.Auth.LoginResult['user'] | null>('user-info', null)

  // 计算属性：是否已登录
  const isAuthenticated = computed(() => !!token.value)

  /**
   * 登录
   */
  async function login(username: string, password: string) {
    const response = await authAPI.login({ username, password })
    if (response?.data?.token) {
      token.value = response.data.token
      userInfo.value = response.data.user
      return response.data
    }
    throw new Error('登录失败')
  }

  /**
   * 注册
   */
  async function register(username: string, password: string, nickname?: string) {
    const response = await authAPI.register({ username, password, nickname })
    if (response?.code === 0) {
      return response.data
    }
    throw new Error('注册失败')
  }

  /**
   * 登出
   */
  function logout() {
    token.value = ''
    userInfo.value = null
  }

  /**
   * 获取最新的用户信息
   */
  async function refreshUserInfo() {
    if (!isAuthenticated.value)
      return null

    try {
      const response = await authAPI.getCurrentUser()
      if (response?.data?.user) {
        userInfo.value = response.data.user
        return response.data.user
      }
      throw new Error('获取用户信息失败')
    }
    catch (error) {
      logout()
      throw error
    }
  }

  return {
    token,
    userInfo,
    isAuthenticated,
    login,
    register,
    logout,
    refreshUserInfo,
  }
})
