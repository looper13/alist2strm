import { http } from './http'

class AuthAPI {
  private baseUrl = '/users'

  /**
   * 用户登录
   */
  async login(params: Api.Auth.LoginParams) {
    return http.post<Api.Auth.LoginResult>(`${this.baseUrl}/login`, params)
  }

  /**
   * 用户注册
   */
  async register(params: Api.Auth.RegisterParams) {
    return http.post<Api.Auth.LoginResult>(`${this.baseUrl}/register`, params)
  }

  /**
   * 获取当前用户信息
   */
  async getCurrentUser(): Promise<Api.Common.HttpResponse<Api.Auth.LoginResult>> {
    return http.get<Api.Auth.LoginResult>(`${this.baseUrl}/me`)
  }
}

export const authAPI = new AuthAPI()
