import { Column, Table, DataType } from 'sequelize-typescript'
import { BaseModel } from './base.js'

@Table({
  tableName: 'file_histories',
  timestamps: true,
})
export class FileHistory extends BaseModel {
  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件名',
  })
  declare fileName: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '源路径',
  })
  declare sourcePath: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '目标文件路径',
  })
  declare targetFilePath: string

  @Column({
    type: DataType.BIGINT,
    allowNull: false,
    comment: '文件大小',
  })
  declare fileSize: number

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件类型',
  })
  declare fileType: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件后缀',
  })
  declare fileSuffix: string
} 