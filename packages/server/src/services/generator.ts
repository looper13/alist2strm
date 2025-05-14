import fs from 'node:fs/promises'
import path from 'node:path'
import type { GenerateResult, AlistStorage } from '../types'
import alistService from './alist'
import config from '../config'
import { logger } from '../utils/logger'

interface GenerateOptions {
  sourcePath?: string
  targetPath: string
  fileSuffix?: string[]
  overwrite?: boolean
  concurrency?: number
  batchSize?: number
}

interface FileTask {
  sourceFilePath: string
  targetFilePath: string
  strmPath: string
  name: string
}

class GeneratorService {
  // 并发处理数量
  private readonly DEFAULT_CONCURRENCY = 5
  // 批量处理任务数量
  private readonly DEFAULT_BATCH_SIZE = 50
  // 重试次数
  private readonly MAX_RETRIES = 3
  // 重试延迟
  private readonly RETRY_DELAY = 1000
  // 文件信息缓存
  private fileInfoCache = new Map<string, any>()

  /**
   * 生成 STRM 文件
   *
   * @param options 配置选项
   * @returns 生成结果
   * @throws 如果目标路径未指定，则抛出错误
   * @throws 如果指定路径未找到存储，则抛出错误
   */
  async generateStrm({
    sourcePath = config.generator.path,
    targetPath,
    fileSuffix = config.generator.fileSuffix,
    overwrite = false,
    concurrency = this.DEFAULT_CONCURRENCY,
    batchSize = this.DEFAULT_BATCH_SIZE,
  }: GenerateOptions): Promise<GenerateResult> {
    // 重置缓存
    this.fileInfoCache.clear()

    logger.info.info('开始生成 STRM 文件', {
      sourcePath,
      targetPath,
      fileSuffix,
      overwrite,
      concurrency,
      batchSize,
    })

    if (!targetPath) {
      logger.error.error('目标路径不能为空')
      throw new Error('目标路径不能为空')
    }

    // 确保目标目录存在
    await fs.mkdir(targetPath, { recursive: true })
    logger.debug.debug('已创建/验证目标目录', { targetPath })

    // 获取存储信息
    const storage = await alistService.findStorage(sourcePath)
    if (!storage) {
      logger.error.error('未找到存储', { sourcePath })
      throw new Error(`未找到路径对应的存储: ${sourcePath}`)
    }

    // 初始化结果统计
    const result: GenerateResult = {
      totalFiles: 0,
      generatedFiles: 0,
      skippedFiles: 0,
    }

    // 收集所有任务
    const allTasks = await this._collectTasks(sourcePath, targetPath, fileSuffix, overwrite, result)

    // 批量处理任务
    await this._processTasks(allTasks, storage, result, concurrency, batchSize)

    // 清理缓存
    this.fileInfoCache.clear()

    logger.info.info('STRM 文件生成完成', { result })
    return result
  }

  /**
   * 收集所有需要处理的文件任务
   */
  private async _collectTasks(
    sourcePath: string,
    targetPath: string,
    fileSuffix: string[],
    overwrite: boolean,
    result: GenerateResult,
  ): Promise<FileTask[]> {
    const allTasks: FileTask[] = []

    const collectInDir = async (currentPath: string, targetBase: string) => {
      const files = await alistService.listFiles(currentPath)
      const dirs = files.filter((f) => f.is_dir)
      const mediaFiles = files.filter(
        (f) => !f.is_dir && fileSuffix.includes(path.extname(f.name).toLowerCase().slice(1)),
      )

      result.totalFiles += mediaFiles.length

      // 收集当前目录的文件任务
      for (const file of mediaFiles) {
        const sourceFilePath = path.join(currentPath, file.name)
        const relativePath = path.relative(sourcePath, currentPath)
        const targetFilePath = path.join(targetBase, relativePath, file.name)
        const strmPath = targetFilePath + '.strm'

        // 如果文件不存在或需要覆盖
        if (overwrite || !(await this._fileExists(strmPath))) {
          allTasks.push({
            sourceFilePath,
            targetFilePath,
            strmPath,
            name: file.name,
          })
        } else {
          result.skippedFiles++
        }
      }

      // 递归处理子目录
      await Promise.all(
        dirs.map(async (dir) => {
          const nextSourcePath = path.join(currentPath, dir.name)
          const nextTargetPath = path.join(
            targetBase,
            path.relative(sourcePath, currentPath),
            dir.name,
          )
          await fs.mkdir(nextTargetPath, { recursive: true })
          await collectInDir(nextSourcePath, targetBase)
        }),
      )
    }

    await collectInDir(sourcePath, targetPath)
    return allTasks
  }

  /**
   * 批量处理文件任务
   */
  private async _processTasks(
    tasks: FileTask[],
    storage: AlistStorage,
    result: GenerateResult,
    concurrency: number,
    batchSize: number,
  ) {
    // 将任务分成批次
    for (let i = 0; i < tasks.length; i += batchSize) {
      const batch = tasks.slice(i, i + batchSize)
      const chunks = this._chunkArray(batch, concurrency)

      // 处理每个并发块
      for (const chunk of chunks) {
        await Promise.all(chunk.map((task) => this._processTask(task, storage, result)))
        // 每个并发块处理完后添加短暂延迟
        await this._sleep(200)
      }
    }
  }

  /**
   * 处理单个文件任务
   */
  private async _processTask(
    task: FileTask,
    storage: AlistStorage,
    result: GenerateResult,
  ): Promise<void> {
    for (let retry = 0; retry < this.MAX_RETRIES; retry++) {
      try {
        // 检查缓存
        let fileInfo = this.fileInfoCache.get(task.sourceFilePath)
        if (!fileInfo) {
          // 添加随机延迟，避免同时请求过多
          if (retry > 0) {
            await this._sleep(Math.random() * this.RETRY_DELAY * (retry + 1))
          }
          fileInfo = await alistService.getFileInfo(task.sourceFilePath)
          this.fileInfoCache.set(task.sourceFilePath, fileInfo)
        }

        const alistUrl = `${config.alist.host}/d${task.sourceFilePath}`
        const finalUrl =
          storage.enable_sign && fileInfo.sign ? `${alistUrl}?sign=${fileInfo.sign}` : alistUrl

        await fs.writeFile(task.strmPath, finalUrl)
        result.generatedFiles++
        logger.debug.debug('已生成 STRM 文件', {
          source: task.sourceFilePath,
          target: task.strmPath,
          hasSign: Boolean(storage.enable_sign && fileInfo.sign),
        })

        break // 成功后跳出重试循环
      } catch (error: any) {
        const isRetryable = this._isRetryableError(error)

        if (!isRetryable || retry === this.MAX_RETRIES - 1) {
          logger.error.error('生成 STRM 文件失败', {
            source: task.sourceFilePath,
            error: error.message,
            stack: error.stack,
            retries: retry + 1,
            isRetryable,
          })

          if (!isRetryable) {
            break
          }
        } else {
          logger.warn.warn('重试获取文件信息', {
            source: task.sourceFilePath,
            retry: retry + 1,
            error: error.message,
          })
        }
      }
    }
  }

  private _isRetryableError(error: any): boolean {
    if (!error) return false

    const errorMessage = error.message?.toLowerCase() || ''
    const retryablePatterns = [
      'timeout',
      'failed link',
      'econnrefused',
      'econnreset',
      'socket hang up',
      'network error',
      'code: 0',
      'code: 429', // Too Many Requests
      'code: 500',
      'code: 502',
      'code: 503',
      'code: 504',
    ]

    return retryablePatterns.some((pattern) => errorMessage.includes(pattern))
  }

  private async _sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }

  private _chunkArray<T>(array: T[], size: number): T[][] {
    const chunks: T[][] = []
    for (let i = 0; i < array.length; i += size) {
      chunks.push(array.slice(i, i + size))
    }
    return chunks
  }

  /**
   * 检查文件是否存在
   * @param filePath 文件路径
   * @returns 是否存在
   */
  private async _fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath)
      return true
    } catch {
      return false
    }
  }
}

export default new GeneratorService()
