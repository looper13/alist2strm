import { http } from './http'

export const fileHistoryAPI = {
  /**
   * 分页查询文件历史
   */
  findByPage(params: Api.FileHistory.Query & { keyword?: string }) {
    return http.get<Api.Common.PaginationResponse<Api.FileHistory.Record>>('/file-history', { params })
  },

  /**
   * 批量删除文件历史记录
   */
  bulkDelete(ids: number[]) {
    return http.delete('/file-history', { data: { ids } })
  },

  /**
   * 清空所有文件历史记录
   */
  clearAll() {
    return http.delete('/file-history/clear')
  },
}
