import axios from 'axios'
import type { CreateTaskDto, Task, TaskLog, UpdateTaskDto } from '@/types/task'

const api = axios.create({
  baseURL: '/api',
})

export const taskService = {
  async getTasks() {
    const { data } = await api.get<Task[]>('/tasks')
    return data
  },

  async createTask(task: CreateTaskDto) {
    const { data } = await api.post<Task>('/tasks', task)
    return data
  },

  async updateTask(id: number, task: UpdateTaskDto) {
    const { data } = await api.patch<Task>(`/tasks/${id}`, task)
    return data
  },

  async deleteTask(id: number) {
    await api.delete(`/tasks/${id}`)
  },

  async executeTask(id: number) {
    await api.post(`/tasks/${id}/execute`)
  },

  async getTaskLogs(id: number) {
    const { data } = await api.get<TaskLog[]>(`/tasks/${id}/logs`)
    return data
  },
} 