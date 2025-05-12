import { config as dotenvConfig } from 'dotenv'
import path from 'path'
import type { Config } from './types'

// 根据环境加载不同的配置文件
const envFile = process.env.NODE_ENV === 'production' ? '.env' : '.env.dev'
dotenvConfig({ path: path.resolve(process.cwd(), envFile) })

// 获取项目根目录
const projectRoot = path.resolve(__dirname, '../../../')

// 验证日志级别
const validateLogLevel = (level: string | undefined): 'info' | 'debug' | 'error' | 'warn' => {
  const validLevels = ['info', 'debug', 'error', 'warn'] as const
  const defaultLevel = process.env.NODE_ENV === 'production' ? 'info' : 'debug'
  return level && validLevels.includes(level as any) ? (level as any) : defaultLevel
}

const config: Config = {
  alist: {
    host: process.env.ALIST_HOST || 'http://192.168.92.120:5244',
    token: process.env.ALIST_TOKEN || '',
  },
  generator: {
    path: process.env.GENERATOR_PATH || '/',
    targetPath: process.env.GENERATOR_TARGET_PATH || '/Users/mccray/test',
    fileSuffix: (process.env.GENERATOR_FILE_SUFFIX || 'mp4,mkv,avi').split(','),
  },
  cron: {
    expression: process.env.CRON_EXPRESSION || '*/1 * * * *',
    enable: process.env.CRON_ENABLE === 'true',
  },
  server: {
    port: parseInt(process.env.PORT || '3000', 10),
  },
  logger: {
    // 日志根目录，使用项目根目录
    baseDir: process.env.LOG_BASE_DIR || path.join(projectRoot, 'data/logs'),
    // 应用名称，用于日志子目录
    appName: process.env.LOG_APP_NAME || 'alist-strm',
    // 日志级别
    level: validateLogLevel(process.env.LOG_LEVEL),
    // 日志保留天数
    maxDays: parseInt(process.env.LOG_MAX_DAYS || '30', 10),
    // 单个日志文件大小限制（MB）
    maxFileSize: parseInt(process.env.LOG_MAX_FILE_SIZE || '10', 10),
  },
}

export default config
