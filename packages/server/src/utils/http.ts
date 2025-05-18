import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import type { HttpResponse } from '../types'

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
        // 在这里可以添加通用的请求处理，比如添加 token
        return config
      },
      (error) => {
        return Promise.reject(error)
      },
    )

    // 响应拦截器
    this.instance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        // 在这里可以添加通用的错误处理
        if (error.response) {
          // 服务器返回错误
          const status = error.response.status
          const message = error.response.data?.message || '请求失败'
          return Promise.reject({
            code: status,
            message,
            data: error.response.data,
          })
        }
        if (error.request) {
          // 请求发出但没有收到响应
          return Promise.reject({
            code: -1,
            message: '网络错误，请检查网络连接',
          })
        }
        // 请求配置出错
        return Promise.reject({
          code: -2,
          message: error.message || '请求配置错误',
        })
      },
    )
  }

  // GET 请求
  async get<T = any>(
    url: string,
    config?: AxiosRequestConfig,
  ): Promise<HttpResponse<T>> {
    const response = await this.instance.get<HttpResponse<T>>(url, config)
    return response.data
  }

  // POST 请求
  async post<T = any, D = any>(
    url: string,
    data?: D,
    config?: AxiosRequestConfig,
  ): Promise<HttpResponse<T>> {
    const response = await this.instance.post<HttpResponse<T>>(url, data, config)
    return response.data
  }

  // PUT 请求
  async put<T = any, D = any>(
    url: string,
    data?: D,
    config?: AxiosRequestConfig,
  ): Promise<HttpResponse<T>> {
    const response = await this.instance.put<HttpResponse<T>>(url, data, config)
    return response.data
  }

  // DELETE 请求
  async delete<T = any>(
    url: string,
    config?: AxiosRequestConfig,
  ): Promise<HttpResponse<T>> {
    const response = await this.instance.delete<HttpResponse<T>>(url, config)
    return response.data
  }

  // PATCH 请求
  async patch<T = any, D = any>(
    url: string,
    data?: D,
    config?: AxiosRequestConfig,
  ): Promise<HttpResponse<T>> {
    const response = await this.instance.patch<HttpResponse<T>>(url, data, config)
    return response.data
  }

  // 自定义请求
  async request<T = any>(
    config: AxiosRequestConfig,
  ): Promise<HttpResponse<T>> {
    const response = await this.instance.request<HttpResponse<T>>(config)
    return response.data
  }
}

// 创建默认实例
export const http = new HttpClient()

// 创建带有基础 URL 的实例工厂函数
export function createHttpClient(baseURL: string, config?: AxiosRequestConfig): HttpClient {
  return new HttpClient({
    baseURL,
    ...config,
  })
} 