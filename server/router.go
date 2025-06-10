package main

import (
	"github.com/MccRay-s/alist2strm/controller"
	"github.com/MccRay-s/alist2strm/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置路由
func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.RequestID())    // 请求ID中间件
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

		// 公开路由（不需要认证）
		public := api.Group("/user")
		{
			public.POST("/login", controller.User.Login)       // 用户登录
			public.POST("/register", controller.User.Register) // 用户注册
		}

		// 需要认证的路由
		auth := api.Group("")
		auth.Use(middleware.JWTAuth()) // 应用JWT认证中间件
		{
			// 用户相关路由
			user := auth.Group("/user")
			{
				user.GET("/me", controller.User.Me)            // 获取当前用户信息
				user.GET("/:id", controller.User.GetUserInfo)  // 获取指定用户信息
				user.PUT("/:id", controller.User.UpdateUser)   // 更新用户信息
				user.GET("/list", controller.User.GetUserList) // 获取用户列表
			}

			// 配置相关路由
			config := auth.Group("/config")
			{
				config.POST("/", controller.Config.Create)                   // 创建配置
				config.GET("/:id", controller.Config.GetConfigInfo)          // 获取指定配置信息
				config.GET("/code/:code", controller.Config.GetConfigByCode) // 根据代码获取配置
				config.PUT("/:id", controller.Config.UpdateConfig)           // 更新配置信息
				config.DELETE("/:id", controller.Config.DeleteConfig)        // 删除配置
				config.GET("/list", controller.Config.GetConfigList)         // 获取配置列表
			}
		}

	}

	return r
}
