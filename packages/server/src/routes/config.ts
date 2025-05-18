import { Router } from 'express'
import { ConfigService } from '@/services/config.service'
import { logger } from '@/utils/logger'

const router = Router()
const configService = new ConfigService()

// 创建配置
router.post('/', async (req, res) => {
  try {
    const config = await configService.create(req.body)
    res.json({
      code: 0,
      data: config,
      message: '创建成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 创建配置:', error)
    res.status(500).json({
      code: 500,
      message: '创建失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 更新配置
router.put('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const config = await configService.update(id, req.body)
    if (!config) {
      res.status(404).json({
        code: 404,
        message: '配置不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: config,
      message: '更新成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 更新配置:', error)
    res.status(500).json({
      code: 500,
      message: '更新失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 删除配置
router.delete('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const success = await configService.delete(id)
    if (!success) {
      res.status(404).json({
        code: 404,
        message: '配置不存在',
      })
      return
    }
    res.json({
      code: 0,
      message: '删除成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 删除配置:', error)
    res.status(500).json({
      code: 500,
      message: '删除失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 分页查询配置
router.get('/', async (req, res) => {
  try {
    const { page, pageSize, keyword } = req.query
    const result = await configService.findByPage({
      page: page ? parseInt(page as string, 10) : undefined,
      pageSize: pageSize ? parseInt(pageSize as string, 10) : undefined,
      keyword: keyword as string,
    })
    res.json({
      code: 0,
      data: result,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 分页查询配置:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 查询所有配置
router.get('/all', async (req, res) => {
  try {
    const configs = await configService.findAll()
    res.json({
      code: 0,
      data: configs,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询所有配置:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

// 根据ID查询配置
router.get('/:id', async (req, res) => {
  try {
    const id = parseInt(req.params.id, 10)
    const config = await configService.findById(id)
    if (!config) {
      res.status(404).json({
        code: 404,
        message: '配置不存在',
      })
      return
    }
    res.json({
      code: 0,
      data: config,
      message: '查询成功',
    })
  }
  catch (error) {
    logger.error.error('路由处理异常 - 查询配置:', error)
    res.status(500).json({
      code: 500,
      message: '查询失败',
      error: error instanceof Error ? error.message : '未知错误',
    })
  }
})

export default router 