package router

import (
	"alist2strm/internal/controller"
	"alist2strm/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine) {
	// 创建 API 组
	api := r.Group("/api")

	// 公开路由
	setupPublicRoutes(api)

	// 需要认证的路由
	setupAuthenticatedRoutes(api)
}

// setupPublicRoutes 设置公开路由
func setupPublicRoutes(api *gin.RouterGroup) {
	// 用户控制器
	userController := controller.NewUserController()

	// 用户注册和登录
	api.POST("/register", userController.Register)
	api.POST("/login", userController.Login)
}

// setupAuthenticatedRoutes 设置需要认证的路由
func setupAuthenticatedRoutes(api *gin.RouterGroup) {
	// 添加认证中间件
	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())

	// 初始化所有控制器
	userController := controller.NewUserController()
	configController := controller.NewConfigController()
	taskController := controller.NewTaskController()
	fileHistoryController := controller.NewFileHistoryController()
	notificationController := controller.NewNotificationController()
	validationController := controller.NewValidationController()

	// 注册各模块路由
	userController.RegisterRoutes(auth)
	configController.RegisterRoutes(auth)
	taskController.RegisterRoutes(auth)
	fileHistoryController.RegisterRoutes(auth)
	notificationController.RegisterRoutes(auth)
	validationController.RegisterRoutes(auth)
}
