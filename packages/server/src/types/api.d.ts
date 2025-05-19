export declare namespace API {
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
    interface QueryRequest extends PageQuery {
      name?: string
    }
    
    interface DetailResponse extends SuccessResponse<Models.ConfigAttributes> {}
    interface ListResponse extends PageResponse<Models.ConfigAttributes> {}
    interface DeleteResponse extends BaseResponse {}
  }

  namespace Task {
    interface CreateRequest extends Services.CreateTaskDto {}
    interface UpdateRequest extends Services.UpdateTaskDto {}
    interface QueryRequest extends PageQuery {
      name?: string
      enabled?: boolean
      running?: boolean
    }

    interface DetailResponse extends SuccessResponse<Models.TaskAttributes> {}
    interface ListResponse extends PageResponse<Models.TaskAttributes> {}
    interface DeleteResponse extends BaseResponse {}
  }

  namespace TaskLog {
    interface CreateRequest extends Services.CreateTaskLogDto {}
    interface UpdateRequest extends Services.UpdateTaskLogDto {}
    interface QueryRequest extends PageQuery {
      taskId?: number
      status?: string
      startTime?: string
      endTime?: string
    }

    interface DetailResponse extends SuccessResponse<Models.TaskLogAttributes> {}
    interface ListResponse extends PageResponse<Models.TaskLogAttributes> {}
    interface DeleteResponse extends BaseResponse {}
  }

  namespace FileHistory {
    interface CreateRequest extends Services.CreateFileHistoryDto {}
    interface UpdateRequest extends Services.UpdateFileHistoryDto {}
    interface QueryRequest extends PageQuery {
      fileName?: string
      sourcePath?: string
      startTime?: string
      endTime?: string
    }

    interface DetailResponse extends SuccessResponse<Models.FileHistoryAttributes> {}
    interface ListResponse extends PageResponse<Models.FileHistoryAttributes> {}
    interface DeleteResponse extends BaseResponse {}
    interface CheckResponse extends SuccessResponse<{ exists: boolean }> {}
  }

  namespace Alist {
    interface ListFilesRequest {
      path?: string
    }

    interface FileInfoRequest {
      path: string
    }

    interface ListFilesResponse extends SuccessResponse<Models.AlistFileInfo[]> {}
    interface FileInfoResponse extends SuccessResponse<Models.AlistFileInfo> {}
  }
} 