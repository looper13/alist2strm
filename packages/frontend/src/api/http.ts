import type { AxiosInstance, AxiosRequestConfig } from 'axios'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { useAuth } from '~/composables'

export class HttpClient {
  private instance: AxiosInstance
  private baseConfig: AxiosRequestConfig = {
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  }

  constructor(config?: AxiosRequestConfig) {
    this.instance = axios.create({
      ...this.baseConfig,
      ...config,
    })

    this.setupInterceptors()
  }

  private setupInterceptors(): void {
    // 请求拦截器
    this.instance.interceptors.request.use(
      (config) => {
        // 添加 token
        const { token } = useAuth()
        if (token.value)
          config.headers.Authorization = `Bearer ${token.value}`
        return config
      },
      (error) => {
        return Promise.reject(new Error(error.message))
      },
    )

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response) => {
        const res = response.data as Api.Common.HttpResponse
        if (res.code === 0)
          return response
        return Promise.reject(res)
      },
      (error) => {
        // 处理 HTTP 错误
        if (error.response) {
          const { data } = error.response as Api.Common.HttpResponse
          if (data.code === 401) {
            const { logout } = useAuth()
            const router = useRouter()
            logout() // 清除 token 和用户信息
            router.push({
              path: '/auth',
              query: {
                redirect: router.currentRoute.value.fullPath,
              },
            })
            return Promise.reject(error)
          }
        }
        return Promise.reject(error)
      },
    )
  }

  // GET 请求
  async get<T = any>(
    url: string,
    config?: AxiosRequestConfig,
  ): Promise<Api.Common.HttpResponse<T>> {
    const response = await this.instance.get<Api.Common.HttpResponse<T>>(url, config)
    return response.data
  }

  // POST 请求
  async post<T = any, D = any>(
    url: string,
    data?: D,
    config?: AxiosRequestConfig,
  ): Promise<Api.Common.HttpResponse<T>> {
    const response = await this.instance.post<Api.Common.HttpResponse<T>>(url, data, config)
    return response.data
  }

  // PUT 请求
  async put<T = any, D = any>(
    url: string,
    data?: D,
    config?: AxiosRequestConfig,
  ): Promise<Api.Common.HttpResponse<T>> {
    const response = await this.instance.put<Api.Common.HttpResponse<T>>(url, data, config)
    return response.data
  }

  // DELETE 请求
  async delete<T = any>(
    url: string,
    config?: AxiosRequestConfig,
  ): Promise<Api.Common.HttpResponse<T>> {
    const response = await this.instance.delete<Api.Common.HttpResponse<T>>(url, config)
    return response.data
  }

  // PATCH 请求
  async patch<T = any, D = any>(
    url: string,
    data?: D,
    config?: AxiosRequestConfig,
  ): Promise<Api.Common.HttpResponse<T>> {
    const response = await this.instance.patch<Api.Common.HttpResponse<T>>(url, data, config)
    return response.data
  }

  // 自定义请求
  async request<T = any>(
    config: AxiosRequestConfig,
  ): Promise<Api.Common.HttpResponse<T>> {
    const response = await this.instance.request<Api.Common.HttpResponse<T>>(config)
    return response.data
  }
}

// 创建默认实例
export const http = new HttpClient({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
})

// 创建带有基础 URL 的实例工厂函数
export function createHttpClient(baseURL: string, config?: AxiosRequestConfig): HttpClient {
  return new HttpClient({
    baseURL,
    ...config,
  })
}
