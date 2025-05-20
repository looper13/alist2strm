import { Column, Table, DataType } from 'sequelize-typescript'
import { BaseModel } from './base.js'

export type TaskStatus = 'pending' | 'running' | 'completed' | 'failed' | 'stopped'

@Table({
  tableName: 'tasks',
  timestamps: true,
})
export class Task extends BaseModel {
  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '任务名称',
  })
  declare name: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '源路径',
  })
  declare sourcePath: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '目标路径',
  })
  declare targetPath: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件后缀',
  })
  declare fileSuffix: string

  @Column({
    type: DataType.BOOLEAN,
    allowNull: false,
    defaultValue: false,
    comment: '是否覆盖',
  })
  declare overwrite: boolean

  @Column({
    type: DataType.BOOLEAN,
    allowNull: false,
    defaultValue: true,
    comment: '是否启用',
  })
  declare enabled: boolean

  @Column({
    type: DataType.STRING,
    allowNull: true,
    comment: 'Cron 表达式',
  })
  declare cron: string

  @Column({
    type: DataType.BOOLEAN,
    allowNull: false,
    defaultValue: false,
    comment: '是否正在运行',
  })
  declare running: boolean

  @Column({
    type: DataType.DATE,
    allowNull: true,
    comment: '最后运行时间',
  })
  declare lastRunAt: Date | null
} 