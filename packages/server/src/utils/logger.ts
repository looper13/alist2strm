import log4js from 'log4js'
import path from 'path'

// 配置日志
log4js.configure({
  appenders: {
    console: { type: 'console' },
    file: {
      type: 'dateFile',
      filename: path.join(process.cwd(), 'logs', 'app.log'),
      pattern: 'yyyy-MM-dd',
      maxLogSize: 1024 * 1024 * 10,
      compress: true,
      keepFileExt: true,
      numBackups: 7,
    },
    error: {
      type: 'dateFile',
      filename: path.join(process.cwd(), 'logs', 'error.log'),
      pattern: 'yyyy-MM-dd',
      maxLogSize: 1024 * 1024 * 10,
      compress: true,
      keepFileExt: true,
      numBackups: 7,
    },
  },
  categories: {
    default: { appenders: ['console', 'file'], level: 'info' },
    error: { appenders: ['console', 'error'], level: 'error' },
    alist: { appenders: ['console', 'file'], level: 'debug' },
  },
})

// 创建不同的日志记录器
export const logger = log4js.getLogger('default')
export const errorLogger = log4js.getLogger('error')
export const alistLogger = log4js.getLogger('alist')

// 确保在应用退出时正确关闭日志
process.on('exit', () => {
  log4js.shutdown()
})
