import { http } from './http'

export const fileHistoryAPI = {
  /**
   * 分页查询文件历史
   */
  findByPage(params: Api.FileHistory.Query) {
    return http.get<Api.Common.PaginationResponse<Api.FileHistory.Record>>('/file-histories', { params })
  },
}
