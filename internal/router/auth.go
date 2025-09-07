package router

import (
	"context-id-backend/internal/controller"
	"context-id-backend/internal/middleware"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(group *ghttp.RouterGroup) {
	// 认证相关路由 - 不需要认证的公开API
	group.GET("/auth/login-url", controller.Auth.GetLoginURL)   // 获取Casdoor登录URL
	group.GET("/auth/signup-url", controller.Auth.GetSignupURL) // 获取Casdoor注册URL

	// OAuth 2.0 Token交换API (前后端分离架构)
	// 注意: 这是后端API，不是OAuth回调URL
	// Casdoor的redirect_uri应该指向前端页面，而不是这个API
	group.POST("/auth/callback", controller.Auth.Login) // 前端用code+state交换token

	// 用户信息相关API
	group.GET("/user", controller.Auth.GetCurrentUser) // 通过token获取用户信息

	// 认证相关路由 - 需要认证的受保护API
	group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
		authGroup.Middleware(middleware.Auth)
		authGroup.GET("/profile-url", controller.Auth.GetMyProfileURL) // 获取用户资料页面URL
	})
}
