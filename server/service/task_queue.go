package service

import (
	"sync"
	"time"

	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/utils"
)

// TaskQueue 任务队列
type TaskQueue struct {
	queue    []uint        // 任务ID队列
	running  bool          // 执行器是否正在运行
	mutex    sync.Mutex    // 互斥锁
	cond     *sync.Cond    // 条件变量，用于任务通知
	shutdown chan struct{} // 关闭信号
}

// 全局队列单例
var (
	taskQueue     *TaskQueue
	taskQueueOnce sync.Once
)

// GetTaskQueue 获取任务队列实例
func GetTaskQueue() *TaskQueue {
	taskQueueOnce.Do(func() {
		// 创建一个互斥锁
		taskQueue = &TaskQueue{
			queue:    make([]uint, 0),
			running:  false,
			mutex:    sync.Mutex{},
			shutdown: make(chan struct{}),
		}
		// 使用已创建的互斥锁初始化条件变量
		taskQueue.cond = sync.NewCond(&taskQueue.mutex)
	})
	return taskQueue
}

// StartTaskQueue 启动任务队列执行器
func StartTaskQueue() {
	tq := GetTaskQueue()

	// 确保执行器状态正确
	tq.mutex.Lock()
	tq.running = false
	tq.mutex.Unlock()

	// 启动任务执行器
	go tq.executor()

	utils.Info("任务队列执行器已启动")

	// 启动一个后台检查，确保执行器正在运行
	go func() {
		time.Sleep(3 * time.Second)
		utils.Info("任务队列执行器状态检查",
			"running", tq.IsExecutorRunning(),
			"queue_length", tq.GetQueueLength())
	}()
}

// AddTask 添加任务到队列
func (tq *TaskQueue) AddTask(taskID uint) {
	utils.Info("正在尝试添加任务到队列", "task_id", taskID)

	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	// 检查任务是否已在队列中
	for _, id := range tq.queue {
		if id == taskID {
			utils.Info("任务已在队列中，跳过添加", "task_id", taskID)
			return
		}
	}

	// 检查执行器状态
	executorRunning := tq.running
	queueLen := len(tq.queue)

	// 添加到队列
	tq.queue = append(tq.queue, taskID)
	utils.Info("任务已添加到队列", "task_id", taskID, "queue_length", len(tq.queue), "executor_running", executorRunning)

	// 通知执行器有新任务
	utils.Info("发送信号通知执行器处理新任务", "task_id", taskID, "queue_before", queueLen, "queue_after", len(tq.queue))
	tq.cond.Signal()
}

// executor 任务执行器
func (tq *TaskQueue) executor() {
	utils.Info("任务执行器已启动，等待任务...")

	for {
		var taskID uint
		var hasTask bool

		// 获取任务
		tq.mutex.Lock()
		queueLen := len(tq.queue)

		utils.Info("执行器检查队列", "queue_length", queueLen, "running_state", tq.running)

		// 等待队列中有任务
		if queueLen == 0 {
			utils.Info("队列为空，执行器进入等待状态")

			// 使用条件变量等待任务或关闭信号
			waitStart := time.Now()

			// 创建一个通道，用于监听关闭信号
			shutdownCh := tq.shutdown

			// 在条件变量等待期间已经持有锁，Wait会释放锁并在返回前重新获取锁
			tq.cond.Wait()

			waitDuration := time.Since(waitStart)
			utils.Info("执行器收到信号，退出等待状态", "wait_duration", waitDuration.String(), "queue_length", len(tq.queue))

			// 检查是否收到关闭信号 (需要在重新获取锁之后检查，避免竞争)
			select {
			case <-shutdownCh:
				// 收到关闭信号，释放锁并退出
				utils.Info("执行器收到关闭信号，退出执行")
				tq.mutex.Unlock()
				return
			default:
				// 没有关闭信号，继续处理
			}
		}

		// 取出队列头部的任务
		if len(tq.queue) > 0 {
			taskID = tq.queue[0]
			tq.queue = tq.queue[1:]
			tq.running = true
			hasTask = true
		}
		tq.mutex.Unlock()

		// 如果有任务，则执行
		if hasTask {
			// 创建一个单独的goroutine来执行任务，避免阻塞队列处理
			go func(id uint) {
				// 执行任务（不在锁内执行，避免阻塞其他操作）
				utils.Info("开始执行任务", "task_id", id)

				// 记录开始时间
				startTime := time.Now()

				// 执行任务
				_, err := Task.ExecuteStrmGeneration(id)

				// 计算持续时间（秒）
				endTime := time.Now()
				durationSeconds := int64(endTime.Sub(startTime).Seconds())

				// 获取最新的任务日志，更新持续时间
				taskLogs, total, logErr := repository.TaskLog.GetLatestByTaskID(id, 1)
				if logErr == nil && total > 0 && len(taskLogs) > 0 {
					latestLog := taskLogs[0]

					// 更新持续时间
					updateData := map[string]interface{}{
						"duration": durationSeconds,
					}

					if updateErr := repository.TaskLog.UpdatePartial(latestLog.ID, updateData); updateErr != nil {
						utils.Error("更新任务日志持续时间失败", "task_id", id, "log_id", latestLog.ID, "error", updateErr.Error())
					} else {
						utils.Info("更新任务日志持续时间", "task_id", id, "duration", durationSeconds)
					}
				}

				// 任务已经在ExecuteStrmGeneration方法中设置和重置了运行状态

				// 记录执行结果
				if err != nil {
					utils.Error("任务执行失败", "task_id", id, "error", err.Error(), "duration", durationSeconds)
				} else {
					utils.Info("任务执行成功", "task_id", id, "duration", durationSeconds)
				}

				// 更新状态
				tq.mutex.Lock()
				tq.running = false
				tq.mutex.Unlock()
			}(taskID)

			// 短暂休息，避免CPU占用过高和立即处理下一个任务
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// IsTaskInQueue 检查任务是否在队列中
func (tq *TaskQueue) IsTaskInQueue(taskID uint) bool {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	for _, id := range tq.queue {
		if id == taskID {
			return true
		}
	}
	return false
}

// IsExecutorRunning 检查执行器是否正在运行
func (tq *TaskQueue) IsExecutorRunning() bool {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.running
}

// GetQueueLength 获取队列长度
func (tq *TaskQueue) GetQueueLength() int {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return len(tq.queue)
}

// RemoveTaskFromQueue 从队列中移除指定任务
func (tq *TaskQueue) RemoveTaskFromQueue(taskID uint) bool {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	for i, id := range tq.queue {
		if id == taskID {
			// 移除任务
			tq.queue = append(tq.queue[:i], tq.queue[i+1:]...)
			return true
		}
	}
	return false
}

// Shutdown 关闭任务队列
func (tq *TaskQueue) Shutdown() {
	close(tq.shutdown)
	utils.Info("任务队列关闭")
}
