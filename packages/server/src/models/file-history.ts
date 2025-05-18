import { Column, Table, DataType } from 'sequelize-typescript'
import { BaseModel } from './base'

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
  fileName!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '源路径',
  })
  sourcePath!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '目标文件路径',
  })
  targetFilePath!: string

  @Column({
    type: DataType.BIGINT,
    allowNull: false,
    comment: '文件大小',
  })
  fileSize!: number

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件类型',
  })
  fileType!: string

  @Column({
    type: DataType.STRING,
    allowNull: false,
    comment: '文件后缀',
  })
  fileSuffix!: string
} 