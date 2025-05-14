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
  sign?: string
}

class GeneratorService {
  // 并发处理数量
  private readonly DEFAULT_CONCURRENCY = 100 // 由于不需要获取文件信息，可以提高并发数
  // 批量处理任务数量
  private readonly DEFAULT_BATCH_SIZE = 100 // 增加批量处理大小
  // 请求间隔（毫秒）
  private readonly REQUEST_DELAY = 200

  /**
   * 生成 STRM 文件
   */
  async generateStrm({
    sourcePath = config.generator.path,
    targetPath,
    fileSuffix = config.generator.fileSuffix,
    overwrite = false,
    concurrency = this.DEFAULT_CONCURRENCY,
    batchSize = this.DEFAULT_BATCH_SIZE,
  }: GenerateOptions): Promise<GenerateResult> {
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

    try {
      // 收集所有任务
      const allTasks = await this._collectTasks(
        sourcePath,
        targetPath,
        fileSuffix,
        overwrite,
        result,
      )
      logger.info.info('任务收集完成', { totalTasks: allTasks.length })

      // 批量处理任务
      await this._processTasks(allTasks, storage, result, concurrency, batchSize)

      logger.info.info('STRM 文件生成完成', { result })
      return result
    } catch (error) {
      logger.error.error('生成过程中发生错误', {
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      throw error
    }
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
      try {
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
              sign: file.sign,
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
      } catch (error) {
        logger.error.error('扫描目录失败', {
          path: currentPath,
          error: (error as Error).message,
          stack: (error as Error).stack,
        })
        throw error
      }
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
        await this._sleep(this.REQUEST_DELAY)
      }

      // 输出进度
      logger.info.info('批次处理进度', {
        processed: Math.min(i + batchSize, tasks.length),
        total: tasks.length,
        success: result.generatedFiles,
        skipped: result.skippedFiles,
        percentage: Math.round((Math.min(i + batchSize, tasks.length) / tasks.length) * 100),
      })
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
    try {
      const alistUrl = `${config.alist.host}/d${task.sourceFilePath}`
      // 如果存储启用了签名并且文件有签名，则添加签名参数
      const finalUrl = storage.enable_sign && task.sign ? `${alistUrl}?sign=${task.sign}` : alistUrl

      await fs.writeFile(task.strmPath, finalUrl)
      result.generatedFiles++
      logger.debug.debug('已生成 STRM 文件', {
        source: task.sourceFilePath,
        target: task.strmPath,
        hasSign: Boolean(storage.enable_sign && task.sign),
      })
    } catch (error) {
      logger.error.error('生成 STRM 文件失败', {
        source: task.sourceFilePath,
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      throw error
    }
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
