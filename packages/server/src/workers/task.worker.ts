import { parentPort, workerData } from 'worker_threads'

interface TaskWorkerData {
  taskId: number
  sourcePath: string
  targetPath: string
  // 其他任务配置...
}

async function executeTask(data: TaskWorkerData): Promise<void> {
  try {
    const { taskId } = data

    // 模拟任务进度
    for (let progress = 0; progress <= 100; progress += 10) {
      // 发送进度更新
      parentPort?.postMessage({
        type: 'progress',
        data: { progress },
      })

      // 模拟任务执行
      await new Promise(resolve => setTimeout(resolve, 1000))
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
        taskId: data.taskId,
        error: error instanceof Error ? error.message : '未知错误',
      },
    })
  }
}

// 开始执行任务
executeTask(workerData as TaskWorkerData) 