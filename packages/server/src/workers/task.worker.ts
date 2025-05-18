const { parentPort, workerData } = require('worker_threads')
const { generatorService } = require('@/services/generator.service')

async function executeTask(
  taskId: number,
  sourcePath: string,
  targetPath: string,
  fileSuffix: string[],
  overwrite: boolean,
  batchSize: number
): Promise<void> {
  try {

    // 模拟任务进度
    for (let progress = 0; progress <= 100; progress += 10) {
      // 发送进度更新
      parentPort?.postMessage({
        type: 'progress',
        data: { progress },
      })
      await generatorService.generateStrm(sourcePath, targetPath, fileSuffix, overwrite, batchSize)
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
        taskId: taskId,
        error: error instanceof Error ? error.message : '未知错误',
      },
    })
  }
}

// 开始执行任务
// executeTask(workerData as TaskWorkerData) 