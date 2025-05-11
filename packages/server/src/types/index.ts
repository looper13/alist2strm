import { Model, Optional } from 'sequelize'

export interface TaskAttributes {
  id: number
  name: string
  sourcePath: string
  targetPath: string
  fileSuffix: string
  overwrite: boolean
  enabled: boolean
  cronExpression: string | null
  lastRunAt: Date | null
  createdAt: Date
  updatedAt: Date
}

export interface TaskCreationAttributes
  extends Optional<TaskAttributes, 'id' | 'createdAt' | 'updatedAt'> {}

export interface TaskLogAttributes {
  id: number
  taskId: number
  status: 'pending' | 'success' | 'error'
  startTime: Date
  endTime: Date | null
  totalFiles: number | null
  generatedFiles: number | null
  skippedFiles: number | null
  error: string | null
  createdAt: Date
  updatedAt: Date
}

export interface TaskLogCreationAttributes
  extends Optional<TaskLogAttributes, 'id' | 'createdAt' | 'updatedAt'> {}

export interface TaskInstance
  extends Model<TaskAttributes, TaskCreationAttributes>,
    TaskAttributes {}
export interface TaskLogInstance
  extends Model<TaskLogAttributes, TaskLogCreationAttributes>,
    TaskLogAttributes {}

export interface GenerateResult {
  totalFiles: number
  generatedFiles: number
  skippedFiles: number
}

export interface Config {
  alist: {
    host: string
    token: string
  }
  generator: {
    path: string
    targetPath: string
    fileSuffix: string[]
  }
  cron: {
    expression: string
    enable: boolean
  }
  server: {
    port: number
  }
}

export interface AlistStorage {
  mount_path: string
  enable_sign: boolean
}
