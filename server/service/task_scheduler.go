package service

import (
	"sync"
	"time"

	"github.com/MccRay-s/alist2strm/model/task"
	"github.com/MccRay-s/alist2strm/repository"
	"github.com/MccRay-s/alist2strm/utils"
	"github.com/robfig/cron/v3"
)

// TaskScheduler 任务调度器
type TaskScheduler struct {
	cron      *cron.Cron
	entryIDs  map[uint]cron.EntryID
	taskMutex sync.RWMutex
}

// 全局单例
var (
	scheduler     *TaskScheduler
	schedulerOnce sync.Once
)

// GetTaskScheduler 获取任务调度器实例
func GetTaskScheduler() *TaskScheduler {
	schedulerOnce.Do(func() {
		scheduler = &TaskScheduler{
			// 使用标准的 5 字段 Cron 格式 (分、时、日、月、周)，而不是 6 字段格式 (秒、分、时、日、月、周)
			cron:      cron.New(),
			entryIDs:  make(map[uint]cron.EntryID),
			taskMutex: sync.RWMutex{},
		}
	})
	return scheduler
}

// Start 启动调度器并返回加载的任务数量
func (s *TaskScheduler) Start() int {
	utils.Info("正在启动任务调度器...")

	// 启动cron调度器
	s.cron.Start()
	utils.Info("Cron 调度器已启动")

	// 加载所有启用的任务
	if err := s.LoadAllEnabledTasks(); err != nil {
		utils.Error("加载定时任务失败", "error", err.Error())
		return 0
	}

	// 返回加载后的任务数
	s.taskMutex.RLock()
	newTaskCount := len(s.entryIDs)
	taskIDs := make([]uint, 0, newTaskCount)
	for id := range s.entryIDs {
		taskIDs = append(taskIDs, id)
	}
	s.taskMutex.RUnlock()

	utils.Info("定时任务调度器启动成功", "task_count", newTaskCount, "task_ids", taskIDs)

	// 输出所有任务的下一次执行时间
	for _, id := range taskIDs {
		nextTime := s.GetNextRunTime(id)
		if nextTime != nil {
			utils.Info("任务下次执行时间", "task_id", id, "next_run", nextTime.Format("2006-01-02 15:04:05"))
		} else {
			utils.Warn("无法获取任务下次执行时间", "task_id", id)
		}
	}

	return newTaskCount
}

// Stop 停止调度器
func (s *TaskScheduler) Stop() {
	s.cron.Stop()
	utils.Info("定时任务调度器已停止")
}

// LoadAllEnabledTasks 加载所有启用的任务
func (s *TaskScheduler) LoadAllEnabledTasks() error {
	// 获取所有已启用且有cron表达式的任务
	tasks, err := repository.Task.GetAllEnabled()
	if err != nil {
		return err
	}

	// 逐个添加到调度器
	for _, t := range tasks {
		if err := s.AddTask(&t); err != nil {
			utils.Error("添加任务到调度器失败", "task_id", t.ID, "error", err.Error())
		}
	}

	return nil
}

// AddTask 添加任务到调度器
func (s *TaskScheduler) AddTask(t *task.Task) error {
	// 如果任务没有启用或cron表达式为空，则不添加
	if !t.Enabled || t.Cron == "" {
		return nil
	}

	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()

	// 如果任务已存在，先移除
	if _, exists := s.entryIDs[t.ID]; exists {
		s.RemoveTaskNoLock(t.ID)
	}

	// 创建任务执行函数
	taskFunc := func() {
		now := time.Now()
		utils.Info("调度器触发任务执行", "task_id", t.ID, "name", t.Name, "time", now.Format("2006-01-02 15:04:05"))

		// 使用一个单独的goroutine执行数据库操作，避免阻塞调度器
		go func() {
			// 检查任务当前状态
			currentTask, err := repository.Task.GetByID(t.ID)
			if err != nil {
				utils.Error("获取任务信息失败", "task_id", t.ID, "error", err.Error())
				return
			}

			// 如果任务已禁用或正在运行，则跳过本次执行
			if !currentTask.Enabled {
				utils.Info("任务已禁用，跳过本次执行", "task_id", t.ID, "name", currentTask.Name)
				return
			}

			if currentTask.Running {
				utils.Info("任务正在运行中，跳过本次执行", "task_id", t.ID, "name", currentTask.Name)
				return
			}

			// 如果任务已在队列中，也跳过
			taskQueue := GetTaskQueue()
			if taskQueue.IsTaskInQueue(t.ID) {
				utils.Info("任务已在队列中，跳过本次调度", "task_id", t.ID, "name", currentTask.Name)
				return
			}

			// 将任务添加到队列
			utils.Info("定时任务触发，添加到执行队列", "task_id", t.ID, "name", currentTask.Name, "queue_length", taskQueue.GetQueueLength())
			taskQueue.AddTask(t.ID)

			// 再次检查队列状态
			utils.Info("添加到队列后的状态", "task_id", t.ID, "in_queue", taskQueue.IsTaskInQueue(t.ID), "executor_running", taskQueue.IsExecutorRunning(), "queue_length", taskQueue.GetQueueLength())
		}()
	}

	// 添加到cron调度器
	utils.Info("正在添加任务到调度器", "task_id", t.ID, "cron", t.Cron, "task_name", t.Name)
	entryID, err := s.cron.AddFunc(t.Cron, taskFunc)
	if err != nil {
		utils.Error("添加任务到调度器失败", "task_id", t.ID, "cron", t.Cron, "error", err.Error())
		return err
	}

	// 记录entryID
	s.entryIDs[t.ID] = entryID

	// 获取并记录下次执行时间，帮助诊断
	entry := s.cron.Entry(entryID)
	utils.Info("任务已成功添加到调度器", "task_id", t.ID, "cron", t.Cron, "next_run", entry.Next.Format("2006-01-02 15:04:05"))
	return nil
}

// RemoveTask 从调度器移除任务
func (s *TaskScheduler) RemoveTask(taskID uint) {
	s.taskMutex.Lock()
	defer s.taskMutex.Unlock()
	s.RemoveTaskNoLock(taskID)
}

// RemoveTaskNoLock 从调度器移除任务(无锁版本)
func (s *TaskScheduler) RemoveTaskNoLock(taskID uint) {
	if entryID, exists := s.entryIDs[taskID]; exists {
		s.cron.Remove(entryID)
		delete(s.entryIDs, taskID)
		utils.Info("任务已从调度器移除", "task_id", taskID)
	}
}

// UpdateTask 更新任务调度
func (s *TaskScheduler) UpdateTask(t *task.Task) error {
	// 首先检查任务是否需要添加到调度器
	needsScheduling := t.Enabled && t.Cron != ""

	// 创建任务的本地副本，避免后续引用可能带来的问题
	taskCopy := *t

	s.taskMutex.Lock()
	// 先移除任务
	s.RemoveTaskNoLock(t.ID)
	s.taskMutex.Unlock()

	// 如果需要调度，则添加任务（AddTask会自行获取锁）
	if needsScheduling {
		// 使用go routine异步添加任务，避免长时间阻塞HTTP请求
		go func() {
			if err := s.AddTask(&taskCopy); err != nil {
				utils.Error("异步添加任务到调度器失败", "task_id", taskCopy.ID, "error", err.Error())
			}
		}()
	}

	return nil
}

// GetNextRunTime 获取下次执行时间
func (s *TaskScheduler) GetNextRunTime(taskID uint) *time.Time {
	s.taskMutex.RLock()
	defer s.taskMutex.RUnlock()

	if entryID, exists := s.entryIDs[taskID]; exists {
		entry := s.cron.Entry(entryID)
		if entry.ID != 0 { // 有效的entry
			nextTime := entry.Next
			return &nextTime
		}
	}

	return nil
}
