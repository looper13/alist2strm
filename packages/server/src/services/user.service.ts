import { User } from '@/models/user.js'
import { logger } from '@/utils/logger.js'
import bcrypt from 'bcrypt'
import jwt from 'jsonwebtoken'
import config from '@/config.js'

class UserService {
  private static instance: UserService

  private constructor() {
    // 私有构造函数
  }

  static getInstance(): UserService {
    if (!UserService.instance) {
      UserService.instance = new UserService()
    }
    return UserService.instance
  }

  /**
   * 创建用户
   */
  async create(data: {
    username: string
    password: string
    nickname?: string
    email?: string
  }): Promise<User> {
    try {
      // 检查用户名是否已存在
      const existingUser = await User.findOne({ where: { username: data.username } })
      if (existingUser) {
        throw new Error('用户名已存在')
      }

      // 加密密码
      const hashedPassword = await bcrypt.hash(data.password, 10)

      // 创建用户
      const user = await User.create({
        ...data,
        password: hashedPassword,
      })

      logger.info.info('用户创建成功', { username: data.username })
      return user
    } catch (error) {
      logger.error.error('创建用户失败:', error)
      throw error
    }
  }

  /**
   * 用户登录
   */
  async login(username: string, password: string): Promise<{ user: User; token: string }> {
    try {
      // 查找用户
      const user = await User.findOne({ where: { username } })
      if (!user) {
        throw new Error('用户名或密码错误')
      }

      // 验证密码
      const isPasswordValid = await bcrypt.compare(password, user.password)
      if (!isPasswordValid) {
        throw new Error('用户名或密码错误')
      }

      // 检查用户状态
      if (user.status === 'disabled') {
        throw new Error('账户已被禁用')
      }

      // 更新最后登录时间
      user.lastLoginAt = new Date()
      await user.save()

      const jwtUser: App.Jwt.User = {
        id: user.id,
        username: user.username,
        nickname: user.nickname,
      }

      // 生成 JWT token
      const token = jwt.sign(jwtUser, config.jwt.secret, { expiresIn: '7d' })

      logger.info.info('用户登录成功', { username })
      return { user, token }
    } catch (error) {
      logger.error.error('用户登录失败:', error)
      throw error
    }
  }

  /**
   * 更新用户信息
   */
  async update(userId: number, data: {
    nickname?: string
    email?: string
    password?: string
    oldPassword?: string
    status?: 'active' | 'disabled'
  }): Promise<User> {
    try {
      const user = await User.findByPk(userId)
      if (!user) {
        throw new Error('用户不存在')
      }

      // 如果要更新密码，先验证原密码
      if (data.password) {
        if (!data.oldPassword) {
          throw new Error('请输入原密码')
        }
        
        // 验证原密码是否正确
        const isPasswordValid = await bcrypt.compare(data.oldPassword, user.password)
        if (!isPasswordValid) {
          throw new Error('原密码错误')
        }

        // 更新为新密码
        user.password = await bcrypt.hash(data.password, 10)
        delete data.password // 防止下面的批量更新再次设置密码
        delete data.oldPassword
      }

      // 更新其他信息
      if (data.nickname !== undefined) user.nickname = data.nickname
      if (data.email !== undefined) user.email = data.email
      if (data.status !== undefined) user.status = data.status
      
      await user.save()

      logger.info.info('用户信息更新成功', { userId })
      return user
    } catch (error) {
      logger.error.error('更新用户信息失败:', error)
      throw error
    }
  }

  /**
   * 获取用户信息
   */
  async findById(id: number): Promise<User | null> {
    try {
      return await User.findByPk(id)
    } catch (error) {
      logger.error.error('获取用户信息失败:', error)
      throw error
    }
  }

  /**
   * 获取用户列表
   */
  async findAll(params: {
    page?: number
    pageSize?: number
    status?: 'active' | 'disabled'
  } = {}): Promise<{ rows: User[]; count: number }> {
    try {
      const { page = 1, pageSize = 10, status } = params
      const where = status ? { status } : {}

      const { rows, count } = await User.findAndCountAll({
        where,
        offset: (page - 1) * pageSize,
        limit: pageSize,
        order: [['createdAt', 'DESC']],
      })

      return { rows, count }
    } catch (error) {
      logger.error.error('获取用户列表失败:', error)
      throw error
    }
  }

  /**
   * 删除用户
   */
  async delete(id: number): Promise<void> {
    try {
      const user = await User.findByPk(id)
      if (!user) {
        throw new Error('用户不存在')
      }

      await user.destroy()
      logger.info.info('用户删除成功', { id })
    } catch (error) {
      logger.error.error('删除用户失败:', error)
      throw error
    }
  }

  /**
   * 检查并创建管理员用户
   */
  async initAdminUser(): Promise<void> {
    try {
      // 检查是否存在任何用户
      const userCount = await User.count()
      if (userCount === 0) {
        // 获取配置的用户名
        const username = config.user.username

        // 生成随机密码或使用配置的密码
        const generatePassword = () => {
          const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*'
          let password = ''
          for (let i = 0; i < 12; i++) {
            password += chars.charAt(Math.floor(Math.random() * chars.length))
          }
          return password
        }

        const adminPassword = config.user.password || generatePassword()
        
        // 创建管理员用户
        await this.create({
          username,
          password: adminPassword,
          nickname: '管理员',
        })

        // 使用醒目的方式打印密码
        logger.info.info('=========================================')
        logger.info.info('初始管理员账户已创建')
        logger.info.info(`用户名: ${username}`)
        logger.info.info(`密码: ${adminPassword}`)
        if (!config.user.password) {
          logger.info.info('请及时修改随机生成的密码！')
        }
        logger.info.info('=========================================')
      }
    } catch (error) {
      logger.error.error('初始化管理员账户失败:', error)
      throw error
    }
  }
}

export const userService = UserService.getInstance()
