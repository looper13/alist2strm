package main

import (
	"log"

	"github.com/MccRay-s/alist2strm/config"
	"github.com/MccRay-s/alist2strm/database"
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
	utils.Info("日志系统初始化完成")

	// 初始化数据库
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	utils.Info("数据库初始化完成")

	// 设置路由
	r := SetupRoutes()

	utils.Info("测试日志记录功能", "user_id", 123, "username", "admin", "ip", "192.168.1.100")
	// 测试日志记录功能
	utils.Debug("调试信息", "key", "value")
	utils.Warn("警告信息", "key", "value")
	utils.Error("错误信息", "key", "value")
	utils.Info("服务器已成功启动，等待请求...")

	// 启动服务器
	utils.Info("服务器启动，监听端口: " + cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		utils.Error("服务器启动失败", "error", err)
		log.Fatalf("服务器启动失败: %v", err)
	}

	// 程序退出时同步日志
	defer utils.Sync()
}
