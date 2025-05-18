declare namespace GenerateResult {

  export interface GenerateResult {
    success: boolean
    message: string
    totalFiles: number
    generatedFiles: number
    skippedFiles: number
  }

  /**
   * strm 成任务
   */
  export interface GenerateTask {
    sourceFilePath: string
    targetFilePath: string
    strmPath: string
    name: string
    sign?: string
  }
}
