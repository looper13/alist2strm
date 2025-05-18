declare namespace Models {
  interface BaseAttributes {
    id: number
    createdAt: Date
    updatedAt: Date
  }

  interface ConfigAttributes extends BaseAttributes {
    name: string
    code: string
    value: string
  }

  interface TaskAttributes extends BaseAttributes {
    name: string
    sourcePath: string
    targetPath: string
    fileSuffix: string
    overwrite: boolean
    enabled: boolean
    cron: string | null
    running: boolean
    lastRunAt: Date | null
  }

  interface TaskLogAttributes extends BaseAttributes {
    taskId: number
    status: string
    message: string | null
    startTime: Date
    endTime: Date | null
    totalFile: number
    generatedFile: number
    skipFile: number
  }

  interface FileHistoryAttributes extends BaseAttributes {
    fileName: string
    sourcePath: string
    targetFilePath: string
    fileSize: number
    fileType: string
    fileSuffix: string
  }
} 