import { DataTypes, Model, Sequelize } from 'sequelize'
import type { TaskAttributes, TaskCreationAttributes } from '../types'
import { sequelize } from './index'

class Task extends Model<TaskAttributes, TaskCreationAttributes> {
  declare id: number
  declare name: string
  declare sourcePath: string
  declare targetPath: string
  declare fileSuffix: string
  declare overwrite: boolean
  declare enabled: boolean
  declare cronExpression: string | null
  declare lastRunAt: Date | null
  declare createdAt: Date
  declare updatedAt: Date

  static associate(models: any) {
    Task.hasMany(models.TaskLog, { foreignKey: 'taskId' })
  }

  static initModel(sequelize: Sequelize): typeof Task {
    Task.init(
      {
        id: {
          type: DataTypes.INTEGER,
          autoIncrement: true,
          primaryKey: true,
        },
        name: {
          type: DataTypes.STRING,
          allowNull: false,
        },
        sourcePath: {
          type: DataTypes.STRING,
          allowNull: false,
        },
        targetPath: {
          type: DataTypes.STRING,
          allowNull: false,
        },
        fileSuffix: {
          type: DataTypes.STRING,
          allowNull: false,
          defaultValue: 'mp4,mkv,avi',
        },
        overwrite: {
          type: DataTypes.BOOLEAN,
          allowNull: false,
          defaultValue: false,
        },
        enabled: {
          type: DataTypes.BOOLEAN,
          allowNull: false,
          defaultValue: true,
        },
        cronExpression: {
          type: DataTypes.STRING,
          allowNull: true,
        },
        lastRunAt: {
          type: DataTypes.DATE,
          allowNull: true,
        },
        createdAt: {
          type: DataTypes.DATE,
          allowNull: false,
        },
        updatedAt: {
          type: DataTypes.DATE,
          allowNull: false,
        },
      },
      {
        sequelize,
        modelName: 'Task',
      },
    )
    return Task
  }
}

export default Task
