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
	userService := service.GetUserService()
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
		api.POST("/register", handler.RegisterHandler)
		api.POST("/login", handler.LoginHandler)

		// 需要认证的路由
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			// 用户相关路由
			auth.GET("/users/info", handler.GetUserInfoHandler)
			auth.PUT("/users/:id", handler.UpdateUserInfoHandler)

			// 配置相关路由
			auth.POST("/configs", handler.CreateConfig)
			auth.PUT("/configs/:id", handler.UpdateConfig)
			auth.DELETE("/configs/:id", handler.DeleteConfig)
			auth.GET("/configs/:id", handler.GetConfig)
			auth.GET("/configs/code/:code", handler.GetConfigByCode)
			auth.GET("/configs/all", handler.ListConfigs)

			// 任务相关路由
			auth.POST("/tasks", handler.CreateTask)
			auth.PUT("/tasks/:id", handler.UpdateTask)
			auth.DELETE("/tasks/:id", handler.DeleteTask)
			auth.GET("/tasks/:id", handler.GetTask)
			auth.GET("/tasks/all", handler.ListTasks)
			auth.PUT("/tasks/:id/status", handler.SetTaskStatus)
			auth.PUT("/tasks/:id/reset", handler.ResetTaskStatus)
			auth.POST("/tasks/:id/execute", handler.ExecuteTaskHandler)
		}
	}

	// 启动服务器
	utils.Info("服务器准备启动", zap.String("port", config.GlobalConfig.Server.Port))
	if err := r.Run(config.GlobalConfig.Server.Port); err != nil {
		utils.Fatal("服务器启动失败", zap.Error(err))
	}
}
