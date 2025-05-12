export interface Task {
  id: string
  name: string
  sourcePath: string
  targetPath: string
  fileSuffix: string
  cronExpression?: string
  enabled: boolean
  lastRunAt?: string
  overwrite: boolean
}

export interface TaskLog {
  id: string
  taskId: string
  status: 'success' | 'error' | 'running'
  startTime: string
  endTime?: string
  totalFiles?: number
  generatedFiles?: number
  skippedFiles?: number
  error?: string
}

export interface TableColumn<T = any> {
  title: string
  key: string
  width?: number | string
  fixed?: 'left' | 'right'
  ellipsis?: boolean | { tooltip: boolean }
  render?: (row: T) => any
}
