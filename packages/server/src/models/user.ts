import { Table, Column, DataType } from 'sequelize-typescript'
import { BaseModel } from './base.js'

@Table({
  tableName: 'users',
  timestamps: true,
})
export class User extends BaseModel {
  @Column({
    type: DataType.STRING,
    allowNull: false,
    unique: true,
    comment: '用户名',
  })
  declare username: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '密码',
  })
  declare password: string

  @Column({
    type: DataType.STRING,
    allowNull: true,
    comment: '昵称',
  })
  declare nickname: string

  @Column({
    type: DataType.STRING,
    allowNull: true,
    comment: '邮箱',
  })
  declare email?: string

  @Column({
    type: DataType.ENUM('active', 'disabled'),
    allowNull: false,
    defaultValue: 'active',
    comment: '状态',
  })
  declare status: 'active' | 'disabled'

  @Column({
    type: DataType.DATE,
    allowNull: true,
    comment: '最后登录时间',
  })
  declare lastLoginAt?: Date

  // 不返回密码字段
  toJSON() {
    const values = { ...this.get() }
    delete values.password
    return values
  }
} 