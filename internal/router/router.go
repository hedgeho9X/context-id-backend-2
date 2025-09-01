package router

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// InitRoutes 初始化所有路由
func InitRoutes(s *ghttp.Server) {
	// 注册静态文件和测试页面路由
	RegisterStaticRoutes(s)

	// 注册API v1路由组
	s.Group("/api/v1", func(v1Group *ghttp.RouterGroup) {
		// API基础路由
		RegisterAPIRoutes(v1Group)

		// 认证相关路由
		RegisterAuthRoutes(v1Group)
	})

	// 根路径处理 - 返回服务器信息
	s.BindHandler("/", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"message": "Context-ID Backend API Server",
			"version": "1.0.0",
			"api": g.Map{
				"v1": "/api/v1",
			},
			"test_pages": g.Map{
				"static": g.Map{
					"login":     "/login",
					"callback":  "/callback",
					"dashboard": "/dashboard",
					"error":     "/error",
				},
				"template": g.Map{
					"login":     "/template/login",
					"callback":  "/template/callback",
					"dashboard": "/template/dashboard",
					"error":     "/template/error",
				},
			},
			"static_files": g.Map{
				"static":    "/static",
				"templates": "/templates",
			},
		})
	})

	// 全局健康检查
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"message": "Context-ID Backend is running",
			"apis": g.Map{
				"v1_health": "/api/v1/health",
			},
		})
	})
}
