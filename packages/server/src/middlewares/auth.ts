import { Request, Response, NextFunction } from 'express'
import jwt from 'jsonwebtoken'
import { User } from '@/models/user.js'
import config from '@/config.js'

// 扩展 Express 的 Request 类型
declare global {
  namespace Express {
    interface Request {
      user?: App.Jwt.User
    }
  }
}

export const auth = async (req: Request, res: Response, next: NextFunction) => {
  try {
    const token = req.headers.authorization?.replace('Bearer ', '')
    
    if (!token) {
      return res.status(401).json({
        code: 401,
        message: '未授权访问',
      })
    }
    const jwtSecret = config.jwt.secret
    const user = jwt.verify(token, jwtSecret) as App.Jwt.User
    console.log('user', user)

    req.user = user
    next()
  } catch (error) {
    return res.status(401).json({
      code: 401,
      message: '无效的认证令牌',
    })
  }
}

// 可选的认证中间件，用于某些可以匿名访问的接口
export const optionalAuth = async (req: Request, res: Response, next: NextFunction) => {
  try {
    const token = req.headers.authorization?.replace('Bearer ', '')
    
    if (!token) {
      return next()
    }

    const jwtSecret = config.jwt.secret
    
    const decoded = jwt.verify(token, jwtSecret) as { id: number }
    const user = await User.findByPk(decoded.id)

    if (user && user.status === 'active') {
      req.user = user
    }
    next()
  } catch (error) {
    next()
  }
} 