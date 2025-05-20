import { Column, Table, DataType, ForeignKey } from 'sequelize-typescript'
import { BaseModel } from './base.js'
import { Task } from './task.js'

@Table({
  tableName: 'task_logs',
  timestamps: true,
})
export class TaskLog extends BaseModel {
  @ForeignKey(() => Task)
  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    comment: '任务ID',
  })
  declare taskId: number

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '状态',
  })
  declare status: string

  @Column({
    type: DataType.TEXT,
    allowNull: true,
    comment: '消息',
  })
  declare message: string | null

  @Column({
    type: DataType.DATE,
    allowNull: false,
    comment: '开始时间',
  })
  declare startTime: Date

  @Column({
    type: DataType.DATE,
    allowNull: true,
    comment: '结束时间',
  })
  declare endTime: Date | null

  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    defaultValue: 0,
    comment: '总文件数',
  })
  declare totalFile: number

  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    defaultValue: 0,
    comment: '已生成文件数',
  })
  declare generatedFile: number

  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    defaultValue: 0,
    comment: '跳过文件数',
  })
  declare skipFile: number
} 