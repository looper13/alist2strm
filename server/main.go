package main

import (
	"log"

	"github.com/MccRay-s/alist2strm/config"
	"github.com/MccRay-s/alist2strm/database"
	"github.com/MccRay-s/alist2strm/service"
	"github.com/MccRay-s/alist2strm/utils"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	// log.Printf("服务器配置加载完成，端口: %s", cfg.Server.Port)

	// 初始化日志系统
	if err := utils.InitLogger(cfg); err != nil {
		log.Fatalf("日志系统初始化失败: %v", err)
	}
	// 初始化数据库
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	utils.Info("数据库初始化完成")

	// 初始化默认用户（如果没有用户的话）
	if err := service.User.InitializeDefaultUser(); err != nil {
		utils.Error("初始化默认用户失败", "error", err.Error())
		log.Fatalf("初始化默认用户失败: %v", err)
	}

	// 初始化服务
	// 获取 logger 实例
	logger := utils.InfoLogger.Desugar()
	// 初始化 AList 服务
	alistService := service.InitializeAListService(logger)
	if alistService == nil {
		utils.Warn("AList 服务初始化失败，部分功能可能不可用")
	} else {
		utils.Info("AList 服务初始化完成")
	}

	// 初始化 STRM 生成服务
	strmService := service.GetStrmGeneratorService()
	strmService.Initialize(logger)

	// 初始化任务队列
	service.GetTaskQueue()
	utils.Info("任务队列初始化完成")

	// 初始化任务调度器
	taskScheduler := service.GetTaskScheduler()

	// 启动任务队列执行器
	service.StartTaskQueue()
	utils.Info("任务队列执行器已启动")

	// 启动任务调度器
	taskCount := taskScheduler.Start()
	if taskCount > 0 {
		utils.Info("任务调度器启动完成", "定时任务数量", taskCount)
	} else {
		utils.Info("任务调度器启动完成，没有启用的定时任务")
	}

	utils.Info("服务初始化完成")

	// 设置路由
	r := SetupRoutes()

	// 启动服务器
	utils.Info("服务器启动，监听端口: " + cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		utils.Error("服务器启动失败", "error", err)
		log.Fatalf("服务器启动失败: %v", err)
	}

	// 程序退出时同步日志
	defer utils.Sync()
}
