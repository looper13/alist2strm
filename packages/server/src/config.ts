import { config as dotenvConfig } from 'dotenv'
import path from 'path'
import type { Config } from './types'

// 根据环境加载不同的配置文件
const isDev = process.env.NODE_ENV === 'development'
const envFile = isDev ? '.env.dev' : '.env'
dotenvConfig({ path: path.resolve(process.cwd(), envFile) })

// 获取项目根目录
const projectRoot = path.resolve(__dirname, '../../../')

// 验证日志级别
const validateLogLevel = (level: string | undefined): Config['logger']['level'] => {
  const validLevels: Config['logger']['level'][] = ['info', 'debug', 'error', 'warn']
  const defaultLevel: Config['logger']['level'] = isDev ? 'debug' : 'info'
  return level && validLevels.includes(level as Config['logger']['level']) ? (level as Config['logger']['level']) : defaultLevel
}

const config: Config = {
  alist: {
    host: process.env.ALIST_HOST || 'http://localhost:5244',
    token: process.env.ALIST_TOKEN || '',
  },
  server: {
    port: parseInt(process.env.PORT || '3000', 10),
  },
  logger: {
    // 日志根目录，使用项目根目录
    baseDir: process.env.LOG_BASE_DIR || path.join(projectRoot, isDev ? 'data/logs-dev' : 'data/logs'),
    // 应用名称，用于日志子目录
    appName: process.env.LOG_APP_NAME || 'alist-strm',
    // 日志级别
    level: validateLogLevel(process.env.LOG_LEVEL),
    // 日志保留天数
    maxDays: parseInt(process.env.LOG_MAX_DAYS || (isDev ? '7' : '30'), 10),
    // 单个日志文件大小限制（MB）
    maxFileSize: parseInt(process.env.LOG_MAX_FILE_SIZE || '10', 10),
  },
  database: {
    // 数据库文件存储路径
    path: process.env.DB_PATH || path.join(projectRoot, 'data/db'),
    // 数据库文件名
    name: process.env.DB_NAME || 'database.sqlite',
  },
}

export default config
