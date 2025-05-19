import { alistService } from './alist.service.js'
import { logger } from '@/utils/logger.js'
import fs from 'node:fs/promises'
import path from 'node:path'
import { configCache } from './config-cache.service.js'
import type { GenerateResult, GenerateTask } from '@/types/index.js'


class GeneratorService {

  private async _fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath)
      return true
    } catch {
      return false
    }
  }

  private _chunkArray<T>(array: T[], size: number): T[][] {
    const chunks: T[][] = []
    for (let i = 0; i < array.length; i += size) {
      chunks.push(array.slice(i, i + size))
    }
    return chunks
  }

  /**
   * 递归生成 strm 文件
   * @param sourcePath 源文件路径
   * @param targetPath 目标文件路径
   * @param fileSuffix 文件后缀
   * @param overwrite 是否覆盖
   */
  private async _doGenFilesStrm(
    sourcePath: string, 
    targetPath: string, 
    fileSuffix: string[],
    overwrite: boolean = false,
    batchSize: number = 500
  ){
    const resolveFiles = async (currentPath: string, targetBase: string) => {
      try{
        const tasks: GenerateTask[] = []
        const files = await alistService.listFiles(currentPath)
        const dirs = files.filter((f) => f.is_dir) || []
        const mediaFiles = files.filter(
          (f) => !f.is_dir && fileSuffix.includes(path.extname(f.name).toLowerCase().slice(1)),
        ) || []

        for(const file of mediaFiles){
          const sourceFilePath = path.join(currentPath, file.name)
          const relativePath = path.relative(sourcePath, currentPath)
          const targetFilePath = path.join(targetBase, relativePath, file.name)
          const strmPath = targetFilePath + '.strm'

          if(overwrite || !this._fileExists(strmPath)){
            tasks.push({
              sourceFilePath,
              targetFilePath,
              strmPath,
              name: file.name,
              sign: file.sign,
            })
          }
        }

        if(tasks.length){
          const alistHost = configCache.getRequired('ALIST_HOST')
          const chunks = this._chunkArray(tasks, batchSize)
           // 处理每个并发块
          for (const chunk of chunks) {
            await Promise.all(
              chunk.map(async (task) => {
                const enCodedPath = task.sourceFilePath
                  .split('/')
                  .map((item) => encodeURIComponent(item))
                  .join('/')

                  const alistUrl = `${alistHost}/d${enCodedPath}`
                  // 如果存储启用了签名并且文件有签名，则添加签名参数
                  const finalUrl = task.sign ? `${alistUrl}?sign=${task.sign}` : alistUrl
                  await fs.writeFile(task.strmPath, finalUrl)
                  logger.info.info('生成 strm 文件', {
                    sourceFilePath: task.sourceFilePath,
                    targetFilePath: task.targetFilePath,
                    strmPath: task.strmPath,
                  })
              })
            )
          }
        }

        for(const dir of dirs){
          const nextSourcePath = path.join(currentPath, dir.name)
          const nextTargetPath = path.join(
            targetBase,
            path.relative(sourcePath, currentPath),
            dir.name,
          )
          await fs.mkdir(nextTargetPath, { recursive: true })
          await resolveFiles(nextSourcePath, targetBase)
        }
      }
      catch(error){
        logger.error.error('获取文件列表失败', {
          error: (error as Error).message,
        })
      }

      const files = await alistService.listFiles(currentPath)
      const targetPath = path.join(targetBase, path.basename(currentPath))
      await fs.mkdir(targetPath, { recursive: true })
      return files.map(file => ({
        ...file,
        path: path.join(targetPath, file.name),
      }))
    }
    await resolveFiles(sourcePath, targetPath)
  }



  /**
   * 生成 strm 文件
   * @param source 源文件路径
   * @param target 目标文件路径
   * @param fileSuffix 文件后缀
   * @param overwrite 是否覆盖
   * @param batchSize 批量大小
   */
  async generateStrm(
    source: string, 
    target: string, 
    fileSuffix: string[], 
    overwrite: boolean = false, 
    batchSize: number = 500
  ) : Promise<GenerateResult> {
    logger.info.info('生成 strm 文件', {
      source,
      target,
      fileSuffix,
      overwrite,
      batchSize,
    })
    if (!source || !target || !fileSuffix || fileSuffix.length === 0) {
      throw new Error('参数错误')
    }

    // 检查目标目录是否存在
    await fs.mkdir(target, { recursive: true })
    logger.info.info('目标目录存在检测完成')

    // 初始化结果统计
    const result: GenerateResult = {
      success: true,
      message: '生成成功',
      totalFiles: 0,
      generatedFiles: 0,
      skippedFiles: 0,
    }

    try{
      const files = await this._doGenFilesStrm(source, target, fileSuffix, overwrite, batchSize)
    }
    catch(error){
      logger.error.error('生成过程中发生错误', {
        error: (error as Error).message,
        stack: (error as Error).stack,
      })
      result.success = false
      result.message = (error as Error).message
      return result
    }

    const files = await alistService.listFiles('/')
    console.error('files', files)
    return result
  }
}

export const generatorService = new GeneratorService()