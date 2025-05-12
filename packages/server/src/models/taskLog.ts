import { DataTypes, Model, Sequelize } from 'sequelize'
import type { TaskLogAttributes, TaskLogCreationAttributes } from '../types'
import { sequelize } from './index'
import Task from './task'

class TaskLog extends Model<TaskLogAttributes, TaskLogCreationAttributes> {
  declare id: number
  declare taskId: number
  declare status: 'pending' | 'success' | 'error'
  declare startTime: Date
  declare endTime: Date | null
  declare totalFiles: number | null
  declare generatedFiles: number | null
  declare skippedFiles: number | null
  declare error: string | null
  declare createdAt: Date
  declare updatedAt: Date

  static associate(models: any) {
    TaskLog.belongsTo(models.Task, { foreignKey: 'taskId' })
  }

  static initModel(sequelize: Sequelize): typeof TaskLog {
    TaskLog.init(
      {
        id: {
          type: DataTypes.INTEGER,
          autoIncrement: true,
          primaryKey: true,
        },
        taskId: {
          type: DataTypes.INTEGER,
          allowNull: false,
          references: {
            model: Task,
            key: 'id',
          },
        },
        status: {
          type: DataTypes.ENUM('pending', 'success', 'error'),
          allowNull: false,
          defaultValue: 'pending',
        },
        startTime: {
          type: DataTypes.DATE,
          allowNull: false,
        },
        endTime: {
          type: DataTypes.DATE,
          allowNull: true,
        },
        totalFiles: {
          type: DataTypes.INTEGER,
          allowNull: true,
        },
        generatedFiles: {
          type: DataTypes.INTEGER,
          allowNull: true,
        },
        skippedFiles: {
          type: DataTypes.INTEGER,
          allowNull: true,
        },
        error: {
          type: DataTypes.TEXT,
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
        modelName: 'TaskLog',
      },
    )
    return TaskLog
  }
}

export default TaskLog
