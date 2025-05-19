import { Column, Table, DataType } from 'sequelize-typescript'
import { BaseModel } from './base.js'

@Table({
  tableName: 'configs',
  timestamps: true,
})
export class Config extends BaseModel {
  @Column({
    type: DataType.STRING,
    allowNull: false,
    unique: true,
    comment: '配置名称',
  })
  name!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    unique: true,
    comment: '配置代码',
  })
  code!: string

  @Column({
    type: DataType.TEXT,
    allowNull: false,
    comment: '配置值',
  })
  value!: string
} 