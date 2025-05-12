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
}

class GeneratorService {
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
  }: GenerateOptions): Promise<GenerateResult> {
    logger.info.info('开始生成 STRM 文件', { sourcePath, targetPath, fileSuffix, overwrite })

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

    // 递归生成 STRM 文件
    await this._generateStrmRecursive(
      sourcePath,
      targetPath,
      fileSuffix,
      overwrite,
      storage,
      result,
    )

    logger.info.info('STRM 文件生成完成', { result })
    return result
  }

  /**
   * 递归生成 STRM 文件
   *
   * @param sourcePath 源文件路径
   * @param targetPath 目标文件路径
   * @param fileSuffix 需要生成 STRM 文件的文件后缀名数组
   * @param overwrite 是否覆盖已存在的 STRM 文件
   * @param storage 存储配置
   * @param result 生成结果对象
   */
  private async _generateStrmRecursive(
    sourcePath: string,
    targetPath: string,
    fileSuffix: string[],
    overwrite: boolean,
    storage: AlistStorage,
    result: GenerateResult,
  ) {
    logger.debug.debug('正在处理目录', { sourcePath, targetPath })
    const files = await alistService.listFiles(sourcePath)

    for (const file of files) {
      const sourceFilePath = path.join(sourcePath, file.name)
      const targetFilePath = path.join(
        file.is_dir ? targetPath : path.dirname(targetPath),
        file.name,
      )

      if (file.is_dir) {
        // 递归处理子目录
        await fs.mkdir(targetFilePath, { recursive: true })
        await this._generateStrmRecursive(
          sourceFilePath,
          targetFilePath,
          fileSuffix,
          overwrite,
          storage,
          result,
        )
      } else {
        result.totalFiles++
        const ext = path.extname(file.name).toLowerCase()
        if (!fileSuffix.includes(ext.slice(1))) {
          result.skippedFiles++
          continue
        }

        const strmPath = targetFilePath + '.strm'
        if (!overwrite && (await fs.stat(strmPath).catch(() => null))) {
          result.skippedFiles++
          continue
        }

        try {
          // 获取文件详情以获取 sign
          const fileInfo = await alistService.getFileInfo(sourceFilePath)
          // 生成 AList 直接访问地址
          const alistUrl = `${config.alist.host}/d${sourceFilePath}`

          // 如果存储需要签名，添加从文件详情获取的 sign
          const finalUrl =
            storage.enable_sign && fileInfo.sign ? `${alistUrl}?sign=${fileInfo.sign}` : alistUrl

          // 生成 STRM 文件内容
          await fs.writeFile(strmPath, finalUrl)
          result.generatedFiles++
          logger.debug.debug('已生成 STRM 文件', {
            source: sourceFilePath,
            target: strmPath,
            hasSign: Boolean(storage.enable_sign && fileInfo.sign),
          })
        } catch (error) {
          logger.error.error('生成 STRM 文件失败', {
            source: sourceFilePath,
            error: (error as Error).message,
            stack: (error as Error).stack,
          })
        }
      }
    }
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
