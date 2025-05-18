import { Column, Table, DataType, ForeignKey, BelongsTo } from 'sequelize-typescript'
import { BaseModel } from './base'
import { Task } from './task'

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
  taskId!: number

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '状态',
  })
  status!: string

  @Column({
    type: DataType.TEXT,
    allowNull: true,
    comment: '消息',
  })
  message!: string | null

  @Column({
    type: DataType.DATE,
    allowNull: false,
    comment: '开始时间',
  })
  startTime!: Date

  @Column({
    type: DataType.DATE,
    allowNull: true,
    comment: '结束时间',
  })
  endTime!: Date | null

  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    defaultValue: 0,
    comment: '总文件数',
  })
  totalFile!: number

  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    defaultValue: 0,
    comment: '已生成文件数',
  })
  generatedFile!: number

  @Column({
    type: DataType.INTEGER,
    allowNull: false,
    defaultValue: 0,
    comment: '跳过文件数',
  })
  skipFile!: number

  @BelongsTo(() => Task)
  task!: Task
} 