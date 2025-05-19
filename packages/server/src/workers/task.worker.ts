import { parentPort, workerData } from 'node:worker_threads'
import { generatorService } from '@/services/generator.service.js'

interface TaskWorkerData {
  taskId: number
  sourcePath: string
  targetPath: string
  fileSuffix: string
  overwrite: boolean
  batchSize: number
}

async function executeTask(taskData: TaskWorkerData) {
  try {
    const { taskId, sourcePath, targetPath, fileSuffix, overwrite, batchSize } = taskData
    // 将逗号分隔的字符串转换为数组
    const suffixes = fileSuffix.split(',').map(s => s.trim())

    // 模拟任务进度
    for (let progress = 0; progress <= 100; progress += 10) {
      // 发送进度更新
      parentPort?.postMessage({
        type: 'progress',
        data: progress,
      })
      await generatorService.generateStrm(sourcePath, targetPath, suffixes, overwrite, batchSize)
    }

    // 发送完成消息
    parentPort?.postMessage({
      type: 'completed',
      data: { taskId },
    })
  }
  catch (error) {
    // 发送错误消息
    parentPort?.postMessage({
      type: 'error',
      data: {
        taskId: taskData.taskId,
        error: error instanceof Error ? error.message : '未知错误',
      },
    })
  }
}

// 开始执行任务
if (parentPort) {
  executeTask(workerData as TaskWorkerData)
} 