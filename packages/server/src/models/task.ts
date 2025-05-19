import { Column, Table, DataType, HasMany } from 'sequelize-typescript'
import { BaseModel } from './base.js'
import { TaskLog } from './task-log.js'

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
  name!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '源路径',
  })
  sourcePath!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '目标路径',
  })
  targetPath!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件后缀',
  })
  fileSuffix!: string

  @Column({
    type: DataType.BOOLEAN,
    allowNull: false,
    defaultValue: false,
    comment: '是否覆盖',
  })
  overwrite!: boolean

  @Column({
    type: DataType.BOOLEAN,
    allowNull: false,
    defaultValue: true,
    comment: '是否启用',
  })
  enabled!: boolean

  @Column({
    type: DataType.STRING,
    allowNull: true,
    comment: 'Cron 表达式',
  })
  cron!: string

  @Column({
    type: DataType.BOOLEAN,
    allowNull: false,
    defaultValue: false,
    comment: '是否正在运行',
  })
  running!: boolean

  @Column({
    type: DataType.DATE,
    allowNull: true,
    comment: '最后运行时间',
  })
  lastRunAt!: Date | null

  @HasMany(() => TaskLog)
  logs!: TaskLog[]
} 