import { Sequelize } from 'sequelize-typescript'
import { dirname, join } from 'node:path'
import { existsSync, mkdirSync } from 'node:fs'
import config from '@/config.js'
import { logger } from '@/utils/logger.js'
import { Config } from '@/models/config.js'
import { Task } from '@/models/task.js'
import { TaskLog } from '@/models/task-log.js'
import { FileHistory } from '@/models/file-history.js'
import { setupAssociations } from '@/models/associations.js'
import { QueryTypes } from 'sequelize'


// 确保数据库目录存在
const dbDir = dirname(join(config.database.path, config.database.name))
if (!existsSync(dbDir)) {
  mkdirSync(dbDir, { recursive: true })
}

const sequelize = new Sequelize({
  dialect: 'sqlite',
  storage: join(config.database.path, config.database.name),
  models: [Config, Task, TaskLog, FileHistory],
  logging: (msg) => logger.debug.debug(msg),
})

/**
 * 删除数据库中的外键约束
 */
async function dropForeignKeys(): Promise<void> {
  try {
    // 获取所有表名
    const tables = await sequelize.query(
      `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`,
      { type: QueryTypes.SELECT }
    ) as { name: string }[]

    for (const { name } of tables) {
      // 获取表的外键信息
      const foreignKeys = await sequelize.query(
        `SELECT * FROM pragma_foreign_key_list('${name}')`,
        { type: QueryTypes.SELECT }
      ) as any[]

      if (foreignKeys.length > 0) {
        // 获取表的完整创建语句
        const [tableInfo] = await sequelize.query(
          `SELECT sql FROM sqlite_master WHERE type='table' AND name=?`,
          { 
            type: QueryTypes.SELECT,
            replacements: [name]
          }
        ) as { sql: string }[]

        if (!tableInfo?.sql) continue

        // 创建新的表结构（去掉两种形式的外键约束）
        let newSql = tableInfo.sql
          // 移除 FOREIGN KEY 语法的约束
          .replace(/,\s*FOREIGN KEY\s*\([^)]+\)\s*REFERENCES\s*[^)]+\)/g, '')
          // 移除列定义中的 REFERENCES 约束
          .replace(/REFERENCES\s+`?\w+`?\s*\([^)]+\)/g, '')

        if (newSql !== tableInfo.sql) {
          // 1. 重命名原表
          await sequelize.query(`ALTER TABLE "${name}" RENAME TO "${name}_old"`)
          // 2. 用新结构创建表
          await sequelize.query(newSql)
          const columns = await sequelize.query(
            `SELECT name FROM pragma_table_info('${name}_old')`,
            { type: QueryTypes.SELECT }
          ) as { name: string }[]
          
          const columnNames = columns.map(c => `"${c.name}"`).join(', ')
          // 4. 复制数据
          await sequelize.query(
            `INSERT INTO "${name}" (${columnNames}) SELECT ${columnNames} FROM "${name}_old"`
          )
          // 5. 删除旧表
          await sequelize.query(`DROP TABLE "${name}_old"`)

          logger.info.info(`表 ${name} 的外键约束已移除`)
        }
      }
    }
  } catch (error) {
    logger.error.error('移除外键约束时出错:', error)
    throw error
  }
}

export async function setupDatabase(): Promise<void> {
  try {
    await sequelize.authenticate()
    logger.info.info('数据库连接已成功建立')
    // 移除外键约束
    await dropForeignKeys()
    // 设置模型关联
    await setupAssociations()

    // 同步数据库结构
    await sequelize.sync({ alter: true })
    logger.info.info('数据库结构同步完成')


  }
  catch (error) {
    logger.error.error('无法连接到数据库:', error)
    // 确保在出错时也重新启用外键检查
    try {
      await sequelize.query('PRAGMA foreign_keys = ON')
    } catch (e) {
      logger.error.error('重新启用外键检查失败:', e)
    }
    throw error
  }
}

export { sequelize }
export * from '@/models/config.js'
export * from '@/models/task.js'
export * from '@/models/task-log.js'
export * from '@/models/file-history.js' 