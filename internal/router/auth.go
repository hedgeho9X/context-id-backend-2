package router

import (
	"context-id-backend/internal/controller"
	"context-id-backend/internal/middleware"

	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterAuthRoutes 注册认证相关路由
func RegisterAuthRoutes(group *ghttp.RouterGroup) {
	// 认证相关路由 - 不需要认证的
	group.GET("/auth/login-url", controller.Auth.GetLoginURL)
	group.GET("/auth/signup-url", controller.Auth.GetSignupURL)
	group.GET("/auth/callback", controller.Auth.Callback)

	// 获取用户信息URL（无需认证）
	group.GET("/user", controller.Auth.GetCurrentUser)

	// 认证相关路由 - 需要认证的
	group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
		authGroup.Middleware(middleware.Auth)
		authGroup.GET("/profile-url", controller.Auth.GetMyProfileURL)
	})
}
