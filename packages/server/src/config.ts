import { config as dotenvConfig } from 'dotenv'
import { fileURLToPath } from 'node:url'
import { dirname, resolve, join } from 'node:path'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

// 根据环境加载不同的配置文件
const isDev = process.env.NODE_ENV === 'development'
const envFile = isDev ? '.env.development' : '.env'
dotenvConfig({ path: resolve(process.cwd(), envFile) })

// 获取项目根目录
const projectRoot = resolve(__dirname, '../../../')

// 验证日志级别
const validateLogLevel = (level: string | undefined): App.Config['logger']['level'] => {
  const validLevels: App.Config['logger']['level'][] = ['info', 'debug', 'error', 'warn']
  const defaultLevel: App.Config['logger']['level'] = isDev ? 'debug' : 'info'
  return level && validLevels.includes(level as App.Config['logger']['level']) ? (level as App.Config['logger']['level']) : defaultLevel
}

const config: App.Config = {
  server: {
    port: parseInt(process.env.PORT || '3210', 10),
  },
  logger: {
    // 日志根目录，使用项目根目录
    baseDir: process.env.LOG_BASE_DIR || join(projectRoot, isDev ? 'data/logs-dev' : 'data/logs'),
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
    path: process.env.DB_BASE_DIR || join(projectRoot, 'data/db'),
    // 数据库文件名
    name: process.env.DB_NAME || 'database.sqlite',
  },
  jwt: {
    // JWT 密钥
    secret: process.env.JWT_SECRET || '63fe1d02ac6da7fe325f3e7545f9b954dc76f25495f73f6d0c0dc82ad44d5fd3',
  },
  user: {
    // 初始用户名
    username: process.env.USER_NAME || 'admin',
    // 初始密码，如果未设置则生成随机密码
    password: process.env.USER_PASSWORD,
  },
}

export default config
