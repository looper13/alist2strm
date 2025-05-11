export interface Task {
  id: number
  name: string
  sourcePath: string
  targetPath: string
  fileSuffix: string
  overwrite: boolean
  enabled: boolean
  cronExpression: string | null
  lastRunAt: string | null
  createdAt: string
  updatedAt: string
}

export interface TaskLog {
  id: number
  taskId: number
  status: 'pending' | 'success' | 'error'
  startTime: string
  endTime: string | null
  totalFiles: number | null
  generatedFiles: number | null
  skippedFiles: number | null
  error: string | null
  createdAt: string
  updatedAt: string
}

export interface CreateTaskDto {
  name: string
  sourcePath: string
  targetPath: string
  fileSuffix?: string
  overwrite?: boolean
  cronExpression?: string
}

export interface UpdateTaskDto extends Partial<CreateTaskDto> {
  enabled?: boolean
}
