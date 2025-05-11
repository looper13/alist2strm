import { config as dotenvConfig } from 'dotenv'
import type { Config } from './types'

dotenvConfig()

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
}

export default config
