package main

import (
	"alist2strm/config"
	"alist2strm/internal/model"
	"alist2strm/internal/router"
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

	// 设置路由
	router.SetupRoutes(r)

	// 启动服务器
	utils.Info("服务器准备启动", zap.String("port", config.GlobalConfig.Server.Port))
	if err := r.Run(config.GlobalConfig.Server.Port); err != nil {
		utils.Fatal("服务器启动失败", zap.Error(err))
	}
}
