import { http } from './http'

class AuthAPI {
  private baseUrl = '/users'

  /**
   * 用户登录
   */
  async login(params: Api.Auth.LoginParams) {
    return http.post<Api.Auth.LoginResult>(`/login`, params)
  }

  /**
   * 用户注册
   */
  async register(params: Api.Auth.RegisterParams) {
    return http.post<Api.Auth.LoginResult>(`/register`, params)
  }

  /**
   * 获取当前用户信息
   */
  async getCurrentUser(): Promise<Api.Common.HttpResponse<Api.Auth.UserInfo>> {
    return http.get<Api.Auth.UserInfo>(`${this.baseUrl}/info`)
  }

  /**
   * 更新当前用户信息
   */
  async updateUser(id: number, params: Api.Auth.UpdateUserParams): Promise<Api.Common.HttpResponse<Api.Auth.LoginResult['user']>> {
    return http.put<Api.Auth.LoginResult['user']>(`${this.baseUrl}/${id}`, params)
  }
}

export const authAPI = new AuthAPI()
