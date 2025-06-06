package main

import (
	"alist2strm/config"
	"alist2strm/internal/handler"
	"alist2strm/internal/middleware"
	"alist2strm/internal/model"
	"alist2strm/internal/service"
	"alist2strm/internal/utils"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	utils.InitLogger()
	defer utils.Logger.Sync()
	utils.Info("开始启动服务...")

	// 初始化配置
	config.Init()
	utils.Info("配置初始化完成")

	// 确保数据库目录存在
	dbDir := filepath.Dir(config.GlobalConfig.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		utils.Fatal("创建数据库目录失败",
			zap.String("path", dbDir),
			zap.Error(err))
	}
	utils.Info("数据库目录检查完成", zap.String("path", dbDir))

	// 初始化数据库
	if err := model.InitDB(); err != nil {
		utils.Fatal("初始化数据库失败", zap.Error(err))
	}
	utils.Info("数据库初始化完成")

	// 创建默认用户
	userService := service.NewUserService(model.DB, utils.Logger)
	if err := userService.CreateDefaultUser(
		config.GlobalConfig.USER.UserName,
		config.GlobalConfig.USER.UserPassword,
	); err != nil {
		utils.Fatal("创建默认用户失败", zap.Error(err))
	} else {
		utils.Info("默认用户检查完成")
	}

	// 创建 Gin 实例
	r := gin.Default()

	// 注册路由
	api := r.Group("/api")
	{
		// 公开路由
		api.POST("/register", handler.Register)
		api.POST("/login", handler.Login)

		// 需要认证的路由
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			// 用户相关路由
			auth.GET("/user/info", handler.GetUserInfo)

			// 配置相关路由
			auth.POST("/config", handler.CreateConfig)
			auth.PUT("/config/:id", handler.UpdateConfig)
			auth.DELETE("/config/:id", handler.DeleteConfig)
			auth.GET("/config/:id", handler.GetConfig)
			auth.GET("/config/code/:code", handler.GetConfigByCode)
			auth.GET("/configs", handler.ListConfigs)

			// 任务相关路由
			auth.POST("/task", handler.CreateTask)
			auth.PUT("/task/:id", handler.UpdateTask)
			auth.DELETE("/task/:id", handler.DeleteTask)
			auth.GET("/task/:id", handler.GetTask)
			auth.GET("/tasks", handler.ListTasks)
			auth.PUT("/task/:id/status", handler.SetTaskStatus)
			auth.PUT("/task/:id/reset", handler.ResetTaskStatus)
		}
	}

	// 启动服务器
	utils.Info("服务器准备启动", zap.String("port", config.GlobalConfig.Server.Port))
	if err := r.Run(config.GlobalConfig.Server.Port); err != nil {
		utils.Fatal("服务器启动失败", zap.Error(err))
	}
}
