package router

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterAPIRoutes 注册API相关路由
func RegisterAPIRoutes(group *ghttp.RouterGroup) {
	// 根路径处理 - 返回API信息
	group.GET("/", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"message": "Context-ID API v1",
			"version": "1.0.0",
			"endpoints": g.Map{
				"auth": g.Map{
					"login_url":  "/api/v1/auth/login-url",
					"signup_url": "/api/v1/auth/signup-url",
					"callback":   "/api/v1/auth/callback",
					"user_info":  "/api/v1/user",
				},
				"protected": g.Map{
					"my_profile": "/api/v1/auth/my-profile-url",
				},
			},
		})
	})

	// 健康检查
	group.GET("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"message": "API v1 is running",
		})
	})
}
