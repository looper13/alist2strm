import log4js from 'log4js'
import path from 'path'
import fs from 'fs'
import config from '../config'

// 定义日志目录
const logDirs = {
  info: '',
  error: '',
  debug: '',
  warn: '',
  access: '',
}

// 基础日志配置
const baseLogConfig = {
  type: 'dateFile',
  pattern: 'yyyy-MM-dd.log',
  alwaysIncludePattern: true,
  maxLogSize: config.logger.maxFileSize * 1024 * 1024, // 转换为字节
  numBackups: config.logger.maxDays,
  compress: true,
  keepFileExt: true,
  encoding: 'utf-8',
  mode: 0o666,
}

// 创建不同级别的日志记录器
export const logger = {
  info: log4js.getLogger('info'),
  error: log4js.getLogger('error'),
  debug: log4js.getLogger('debug'),
  warn: log4js.getLogger('warn'),
  access: log4js.getLogger('access'),
}

// 导出日志中间件
export const httpLogger = log4js.connectLogger(logger.access, {
  level: 'auto',
  format: (req, res, format) =>
    format(
      `访问日志: :remote-addr - ":method :url HTTP/:http-version" 状态码::status 内容长度::content-length 来源:":referrer" 用户代理:":user-agent"`,
    ),
})

// 初始化日志系统
export function setupLogger(): void {
  // 确保基础日志目录存在
  const baseLogDir = path.resolve(config.logger.baseDir)
  if (!fs.existsSync(baseLogDir)) {
    fs.mkdirSync(baseLogDir, { recursive: true })
  }

  // 初始化日志目录
  Object.keys(logDirs).forEach((key) => {
    logDirs[key as keyof typeof logDirs] = path.join(baseLogDir, config.logger.appName, key)
    const dir = logDirs[key as keyof typeof logDirs]
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true })
    }
  })

  // 配置日志
  log4js.configure({
    appenders: {
      console: {
        type: 'console',
        layout: {
          type: 'pattern',
          pattern: '[%d{yyyy-MM-dd hh:mm:ss.SSS}] [%p] [%c] - %m',
        },
      },
      info: {
        ...baseLogConfig,
        filename: path.join(logDirs.info, 'app'),
        layout: {
          type: 'pattern',
          pattern: '[%d{yyyy-MM-dd hh:mm:ss.SSS}] [%p] [%c] %f:%l - %m',
        },
      },
      error: {
        ...baseLogConfig,
        filename: path.join(logDirs.error, 'error'),
        layout: {
          type: 'pattern',
          pattern: '[%d{yyyy-MM-dd hh:mm:ss.SSS}] [%p] [%c] %f:%l - %m%n%s%n',
        },
      },
      debug: {
        ...baseLogConfig,
        filename: path.join(logDirs.debug, 'debug'),
      },
      warn: {
        ...baseLogConfig,
        filename: path.join(logDirs.warn, 'warn'),
      },
      access: {
        ...baseLogConfig,
        filename: path.join(logDirs.access, 'access'),
        layout: {
          type: 'pattern',
          pattern: '[%d{yyyy-MM-dd hh:mm:ss.SSS}] [%p] - %m',
        },
      },
    },
    categories: {
      default: {
        appenders: ['console', 'info'],
        level: config.logger.level,
        enableCallStack: true,
      },
      info: {
        appenders: ['console', 'info'],
        level: 'info',
        enableCallStack: true,
      },
      error: {
        appenders: ['console', 'error'],
        level: 'error',
        enableCallStack: true,
      },
      debug: {
        appenders: ['console', 'debug'],
        level: 'debug',
        enableCallStack: true,
      },
      warn: {
        appenders: ['console', 'warn'],
        level: 'warn',
        enableCallStack: true,
      },
      access: {
        appenders: ['console', 'access'],
        level: 'info',
      },
    },
    pm2: true,
    disableClustering: false,
  })

  // 添加启动日志
  logger.info.info('日志系统初始化完成', {
    baseDir: baseLogDir,
    level: config.logger.level,
    dirs: Object.keys(logDirs),
  })

  // 添加未捕获异常的处理
  process.on('uncaughtException', (err) => {
    logger.error.error('未捕获的异常:', err)
  })

  process.on('unhandledRejection', (reason, promise) => {
    logger.error.error('未处理的 Promise 拒绝:', promise, '原因:', reason)
  })

  // 确保在应用退出时正确关闭日志
  process.on('exit', () => {
    log4js.shutdown()
  })
}
