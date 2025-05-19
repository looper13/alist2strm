import { http } from './http'

export const fileHistoryAPI = {
  /**
   * 分页查询文件历史
   */
  findByPage(params: Api.FileHistoryQuery) {
    return http.get<Api.HttpResponse<Api.PaginationResponse<Api.FileHistory>>>('/file-history', { params })
  },
}
