/// <reference path="./api.d.ts" />

declare namespace App {
  // 基础响应类型
  type HttpResponse<T = any> = Api.HttpResponse<T>
  type PaginationQuery = Api.PaginationQuery
  type PaginationResponse<T> = Api.PaginationResponse<T>

  // 业务类型
  type Config = Api.Config
  type Task = Api.Task
  type TaskLog = Api.TaskLog
  type FileHistory = Api.FileHistory
  type TaskProgress = Api.TaskProgress

  // 错误类型
  interface AppError extends Error {
    code: number
    status?: number
    details?: Record<string, any>
  }

}

export type HttpResponse<T = any> = App.HttpResponse<T>
export type PaginationQuery = App.PaginationQuery
export type PaginationResponse<T> = App.PaginationResponse<T>
export type Config = App.Config
export type Task = App.Task
export type TaskLog = App.TaskLog
export type FileHistory = App.FileHistory
export type TaskProgress = App.TaskProgress
export type AppError = App.AppError

export as namespace App
export = App
