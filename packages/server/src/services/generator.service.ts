import { alistService } from '@/services/alist.service.js'
import { logger } from '@/utils/logger.js'
import fs from 'node:fs/promises'
import path from 'node:path'
import { getConfigCache } from './config-cache.service.js'
import { FileHistoryService } from './file-history.service.js'
import { ALIST_CONFIG, GENERATOR_CONFIG } from '@/constant/index.js'

export class GeneratorService {
  private fileHistoryService: FileHistoryService
  private result: App.Generate.Result

  constructor() {
    this.fileHistoryService = new FileHistoryService()
    this.result = {
      success: true,
      message: '生成成功',
      totalFiles: 0,
      generatedFiles: 0,
      skippedFiles: 0,
    }
  }

  // 统一路径分隔符为正斜杠
  private _normalizePath(filePath: string): string {
    return filePath.replace(/\\/g, '/')
  }

  // 检查文件是否存在
  private async _fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath)
      return true
    } catch {
      return false
    }
  }

  // 分组
  private _chunkArray<T>(array: T[], size: number): T[][] {
    const chunks: T[][] = []
    for (let i = 0; i < array.length; i += size) {
      chunks.push(array.slice(i, i + size))
    }
    return chunks
  }

  /**
   * 处理单个文件，生成 strm 文件
   */
  private async _processFile(
    file: App.AList.AlistFile,
    sourcePath: string,
    currentPath: string,
    targetBase: string,
    overwrite: boolean,
  ): Promise<App.Generate.Task | null> {
    // 源文件路径
    const sourceFilePath = this._normalizePath(path.join(currentPath, file.name))
    // 相对路径
    const relativePath = this._normalizePath(path.relative(sourcePath, currentPath))

    // 目标文件路径（不含扩展名）
    const targetFilePath = this._normalizePath(path.join(targetBase, relativePath, path.parse(file.name).name))
  
    // strm 文件路径（替换原扩展名）
    const strmPath = `${targetFilePath}.strm`

    if (overwrite || !(await this._fileExists(strmPath))) {
      return {
        sourceFilePath,
        targetFilePath,
        strmPath,
        name: file.name,
        sign: file.sign,
        type: `${file.type}`,
        fileSize: file.size || 0,
      }
    }
    return null
  }

  /**
   * 记录文件历史
   */
  private async _recordFileHistory(
    task: App.Generate.Task,
  ): Promise<void> {
    try {
      const fileSuffix = path.extname(task.name).toLowerCase().slice(1)
      await this.fileHistoryService.create({
        fileName: task.name,
        sourcePath: task.sourceFilePath,
        targetFilePath: task.targetFilePath,
        fileSize: task.fileSize,
        fileType: task.type,
        fileSuffix,
      })
    } catch (error) {
      logger.error.error('记录文件历史失败', {
        fileName: task.name,
        sourcePath: task.sourceFilePath,
        targetFilePath: task.targetFilePath,
        error: (error as Error).message,
      })
    }
  }

  /**
   * 批量生成 strm 文件
   */
  private async _generateStrmFiles(tasks: App.Generate.Task[]): Promise<void> {
    if (tasks.length === 0) return

    const configCache = await getConfigCache()
    const alistHost = configCache.get(ALIST_CONFIG.ALIST_REPLACE_HOST) ||  configCache.getRequired(ALIST_CONFIG.ALIST_HOST)
    const chunks = this._chunkArray(tasks, 500)
    const urlEncode = configCache.get(GENERATOR_CONFIG.URL_ENCODE)

    for (const chunk of chunks) {
      await Promise.all(
        chunk.map(async (task) => {
          const encodedPath = urlEncode && urlEncode === 'Y' ? task.sourceFilePath.split('/').map(item => encodeURIComponent(item)).join('/') : task.sourceFilePath
          const alistUrl = `${alistHost}/d${encodedPath}`
          const finalUrl = task.sign ? `${alistUrl}?sign=${task.sign}` : alistUrl

          await fs.writeFile(task.strmPath, finalUrl)
          
          // 记录文件历史
          await this._recordFileHistory(task)

          // 更新统计信息
          this.result.generatedFiles++

          logger.info.info('生成 strm 文件', {
            sourceFilePath: task.sourceFilePath,
            targetFilePath: task.targetFilePath,
            strmPath: task.strmPath,
          })
        }),
      )
    }
  }

  /**
   * 递归处理目录
   */
  private async _processDirectory(
    sourcePath: string,
    currentPath: string,
    targetBase: string,
    fileSuffix: string[],
    overwrite: boolean,
  ): Promise<void> {
    try {
      // 确保目标目录存在
      const relativePath = this._normalizePath(path.relative(sourcePath, currentPath))
      const targetPath = this._normalizePath(path.join(targetBase, relativePath))
      await fs.mkdir(targetPath, { recursive: true })

      // 获取当前目录文件列表
      const files = await alistService.listFiles(currentPath)
      
      // 处理媒体文件
      const mediaFiles = files.filter(
        file => !file.is_dir && fileSuffix.includes(path.extname(file.name).toLowerCase().slice(1)),
      )

      // 更新总文件数
      this.result.totalFiles += mediaFiles.length

      const tasks: App.Generate.Task[] = []
      for (const file of mediaFiles) {
        const task = await this._processFile(file, sourcePath, currentPath, targetBase, overwrite)
        if (task) {
          tasks.push(task)
        } else {
          // 更新跳过的文件数
          this.result.skippedFiles++
        }
      }

      // 生成 strm 文件
      await this._generateStrmFiles(tasks)

      // 递归处理子目录
      const dirs = files.filter(file => file.is_dir)
      for (const dir of dirs) {
        const nextPath = this._normalizePath(path.join(currentPath, dir.name))
        await this._processDirectory(sourcePath, nextPath, targetBase, fileSuffix, overwrite)
      }
    } catch (error) {
      logger.error.error('处理目录失败', {
        path: currentPath,
        error: (error as Error).message,
      })
      throw error
    }
  }

  /**
   * 生成 strm 文件
   */
  async generateStrm(
    source: string,
    target: string,
    fileSuffix: string[],
    overwrite: boolean = false,
  ): Promise<App.Generate.Result> {
    if (!source || !target || !fileSuffix?.length) {
      throw new Error('参数错误：source、target 和 fileSuffix 不能为空')
    }

    // 重置结果
    this.result = {
      success: true,
      message: '生成成功',
      totalFiles: 0,
      generatedFiles: 0,
      skippedFiles: 0,
    }

    // 统一路径格式
    source = this._normalizePath(source)
    target = this._normalizePath(target)

    try {
      // 确保目标目录存在
      await fs.mkdir(target, { recursive: true })
      logger.info.info('开始生成 strm 文件', { source, target, fileSuffix })

      // 开始处理目录
      await this._processDirectory(source, source, target, fileSuffix, overwrite)

      logger.info.info('strm 文件生成完成', { 
        source, 
        target,
        totalFiles: this.result.totalFiles,
        generatedFiles: this.result.generatedFiles,
        skippedFiles: this.result.skippedFiles,
      })
    } catch (error) {
      logger.error.error('生成 strm 文件失败', {
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      this.result.success = false
      this.result.message = (error as Error).message
    }

    return this.result
  }
}

export const generatorService = new GeneratorService()