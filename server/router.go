package main

import (
	"github.com/MccRay-s/alist2strm/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 中间件
	// r.Use(middleware.RequestID())    // 请求ID中间件
	r.Use(middleware.AccessLogger()) // 访问日志中间件
	r.Use(gin.Recovery())            // 错误恢复中间件

	// API 路由组
	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "alist2strm server is running",
			})
		})

		// 用户相关路由
		// user := api.Group("/user")
		// {
		// 	user.POST("/login", func(c *gin.Context) {
		// 		// TODO: 实现登录逻辑
		// 		c.JSON(200, gin.H{"message": "login endpoint"})
		// 	})
		// }

		// // 任务相关路由
		// task := api.Group("/task")
		// {
		// 	task.GET("/", func(c *gin.Context) {
		// 		// TODO: 实现获取任务列表
		// 		c.JSON(200, gin.H{"message": "task list endpoint"})
		// 	})
		// }

		// // 配置相关路由
		// config := api.Group("/config")
		// {
		// 	config.GET("/", func(c *gin.Context) {
		// 		// TODO: 实现获取配置
		// 		c.JSON(200, gin.H{"message": "config endpoint"})
		// 	})
		// }
	}

	return r
}
