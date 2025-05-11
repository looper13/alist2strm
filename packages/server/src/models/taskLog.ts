import { DataTypes, Model, Sequelize } from 'sequelize'
import type { TaskLogAttributes, TaskLogCreationAttributes } from '../types'
import { sequelize } from './index'
import Task from './task'

class TaskLog extends Model<TaskLogAttributes, TaskLogCreationAttributes> {
  public id!: number
  public taskId!: number
  public status!: 'pending' | 'success' | 'error'
  public startTime!: Date
  public endTime!: Date | null
  public totalFiles!: number | null
  public generatedFiles!: number | null
  public skippedFiles!: number | null
  public error!: string | null
  public createdAt!: Date
  public updatedAt!: Date

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
