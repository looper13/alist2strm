import axios from 'axios'
import type { CreateTaskDto, Task, TaskLog, UpdateTaskDto } from '@/types/task'

const api = axios.create({
  baseURL: '/api',
})

// 添加服务端统一响应格式的接口定义
interface ApiResponse<T> {
  code: number
  msg: string
  data: T
}

interface PaginatedData<T> {
  records: T[]
  total: number
}

interface PaginatedResponse<T> {
  records: T[]
  total: number
  page: number
  pageSize: number
}

export const taskService = {
  async getTasks() {
    const { data } = await api.get<ApiResponse<Task[]>>('/tasks')
    return data.data
  },

  async getTasksWithPagination(page: number, pageSize: number): Promise<PaginatedResponse<Task>> {
    const { data } = await api.get<ApiResponse<PaginatedData<Task>>>('/tasks', {
      params: { page, pageSize },
    })
    return {
      records: data.data.records,
      total: data.data.total,
      page,
      pageSize,
    }
  },

  async createTask(task: CreateTaskDto) {
    const { data } = await api.post<ApiResponse<Task>>('/tasks', task)
    return data.data
  },

  async updateTask(id: number, task: UpdateTaskDto) {
    const { data } = await api.put<ApiResponse<Task>>(`/tasks/${id}`, task)
    return data.data
  },

  async deleteTask(id: number) {
    const { data } = await api.delete<ApiResponse<null>>(`/tasks/${id}`)
    return data.data
  },

  async executeTask(id: number) {
    const { data } = await api.post<ApiResponse<null>>(`/tasks/${id}/execute`)
    return data.data
  },

  async getTaskLogs(id: number) {
    const { data } = await api.get<ApiResponse<TaskLog[]>>(`/tasks/${id}/logs`)
    return data.data
  },
}
