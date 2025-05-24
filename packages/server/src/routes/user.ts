import { Router } from 'express'
import type { Request, Response, NextFunction, Router as RouterType } from 'express'
import { HttpError } from '@/middlewares/error.js'
import { auth } from '@/middlewares/auth.js'
import { userService } from '@/services/user.service.js'
import { success, pageResult } from '@/utils/response.js'
import { logger } from '@/utils/logger.js'

const router: RouterType = Router()

// 用户注册
router.post('/register', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { username, password, nickname, email } = req.body

    // 验证必填字段
    if (!username || !password) {
      throw new HttpError('用户名和密码不能为空', 400)
    }

    const user = await userService.create({
      username,
      password,
      nickname,
      email,
    })

    success(res, user, '注册成功')
  } catch (error) {
    if (error instanceof Error && error.message === '用户名已存在') {
      next(new HttpError('用户名已存在', 400))
    } else {
      logger.error.error('用户注册失败:', error)
      next(new HttpError('用户注册失败', 500))
    }
  }
})

// 用户登录
router.post('/login', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { username, password } = req.body

    // 验证必填字段
    if (!username || !password) {
      throw new HttpError('用户名和密码不能为空', 400)
    }

    const result = await userService.login(username, password)
    success(res, result, '登录成功')
  } catch (error) {
    if (error instanceof Error) {
      switch (error.message) {
        case '用户名或密码错误':
          next(new HttpError('用户名或密码错误', 401))
          break
        case '账户已被禁用':
          next(new HttpError('账户已被禁用', 403))
          break
        default:
          logger.error.error('用户登录失败:', error)
          next(new HttpError('登录失败', 500))
      }
    } else {
      next(new HttpError('登录失败', 500))
    }
  }
})

// 获取当前用户信息
router.get('/me', auth, async (req: Request, res: Response) => {
  const jwtUser = req.user as App.Jwt.User
  const user = await userService.findById(jwtUser.id)
  success(res, user, '获取成功')
})

// 更新用户信息
router.put('/me', auth, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { nickname, email, password } = req.body
    const user = await userService.update(req.user!.id, {
      nickname,
      email,
      password,
    })

    success(res, user, '更新成功')
  } catch (error) {
    logger.error.error('更新用户信息失败:', error)
    next(new HttpError('更新失败', 500))
  }
})

// 获取用户列表
router.get('/', auth, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { page = 1, pageSize = 10, status } = req.query
    
    const result = await userService.findAll({
      page: Number(page),
      pageSize: Number(pageSize),
      status: status as 'active' | 'disabled' | undefined,
    })

    pageResult(res, {
      list: result.rows,
      total: result.count,
      page: Number(page),
      pageSize: Number(pageSize)
    })
  } catch (error) {
    logger.error.error('获取用户列表失败:', error)
    next(new HttpError('获取用户列表失败', 500))
  }
})

// 获取指定用户信息
router.get('/:id', auth, async (req: Request, res: Response, next: NextFunction) => {
  try {
    const user = await userService.findById(Number(req.params.id))
    if (!user) {
      throw new HttpError('用户不存在', 404)
    }
    success(res, user, '获取成功')
  } catch (error) {
    if (error instanceof HttpError) {
      next(error)
    } else {
      logger.error.error('获取用户信息失败:', error)
      next(new HttpError('获取用户信息失败', 500))
    }
  }
})

// 删除用户
router.delete('/:id', auth, async (req: Request, res: Response, next: NextFunction) => {
  try {
    await userService.delete(Number(req.params.id))
    success(res, null, '删除成功')
  } catch (error) {
    if (error instanceof Error && error.message === '用户不存在') {
      next(new HttpError('用户不存在', 404))
    } else {
      logger.error.error('删除用户失败:', error)
      next(new HttpError('删除用户失败', 500))
    }
  }
})

export default router 