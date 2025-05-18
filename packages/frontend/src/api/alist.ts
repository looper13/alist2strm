import type { HttpResponse } from '~/types'
import { http } from './http'

class AlistAPI {
  private baseUrl = '/alist'

  async listDirs(path: string): Promise<HttpResponse<Api.AlistDir[]>> {
    return http.get(`${this.baseUrl}/dirs`, { params: { path } })
  }
}

export const alistAPI = new AlistAPI()
