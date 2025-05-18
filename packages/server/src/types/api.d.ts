declare namespace API {
  interface BaseResponse {
    code: number
    message: string
  }

  interface SuccessResponse<T = any> extends BaseResponse {
    code: 0
    data: T
  }

  interface ErrorResponse extends BaseResponse {
    code: number
    error?: string
  }

  interface PageQuery {
    page?: number | string
    pageSize?: number | string
    keyword?: string
  }

  interface PageResponse<T> extends SuccessResponse<{
    list: T[]
    total: number
    page: number
    pageSize: number
  }> {}

  namespace Config {
    interface CreateRequest extends Services.CreateConfigDto {}
    interface UpdateRequest extends Services.UpdateConfigDto {}
    interface QueryRequest extends PageQuery {}
    
    interface DetailResponse extends SuccessResponse<Models.ConfigAttributes> {}
    interface ListResponse extends PageResponse<Models.ConfigAttributes> {}
    interface DeleteResponse extends BaseResponse {}
  }

  namespace Task {
    interface CreateRequest extends Services.CreateTaskDto {}
    interface UpdateRequest extends Services.UpdateTaskDto {}
    interface QueryRequest extends PageQuery {
      enabled?: boolean
      running?: boolean
    }

    interface DetailResponse extends SuccessResponse<Models.TaskAttributes> {}
    interface ListResponse extends PageResponse<Models.TaskAttributes> {}
    interface DeleteResponse extends BaseResponse {}
  }
} 