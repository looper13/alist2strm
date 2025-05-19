import { Router } from 'express'
import { FileHistoryService } from '@/services/file-history.service.js'
import { logger } from '@/utils/logger.js'

const router = Router()
const fileHistoryService = new FileHistoryService()

// 创建文件历史
router.post('/', async (req, res) => {
  try {
    const fileHistory = await fileHistoryService.create(req.body)
    res.json({
      code: 0,
      data: fileHistory,
      message: '创建成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 创建文件历史:', error)
    res.status(500).json({
      code: 500,
      message: '创建失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 分页查询文件历史
router.get('/', async (req, res) => {
  try {
    const { page, pageSize, keyword, fileType, fileSuffix, startTime, endTime } = req.query
    const result = await fileHistoryService.findByPage({
      page: page ? parseInt(page as string, 10) : undefined,
      pageSize: pageSize ? parseInt(pageSize as string, 10) : undefined,
      keyword: keyword as string,
      fileType: fileType as string,
      fileSuffix: fileSuffix as string,
      startTime: startTime ? new Date(startTime as string) : undefined,
      endTime: endTime ? new Date(endTime as string) : undefined,
    })
    res.json({
      code: 0,
      data: result,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 分页查询文件历史:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 根据ID查询文件历史
router.get('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const fileHistory = await fileHistoryService.findById(id)
    if (!fileHistory) {
      res.status(404).json({
        code: 404,
        message: '文件历史不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: fileHistory,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询文件历史:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 检查文件是否存在
router.get('/check', async (req, res) => {
  try {
    const { sourcePath, fileName } = req.query
    if (!sourcePath || !fileName) {
      res.status(400).json({
        code: 400,
        message: '参数错误',
      })
      return
    }
    const exists = await fileHistoryService.checkFileExists(sourcePath as string, fileName as string)
    res.json({
      code: 0,
      data: { exists },
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 检查文件是否存在:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

export default router 