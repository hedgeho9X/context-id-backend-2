package router

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterStaticRoutes 注册静态文件和测试页面路由
func RegisterStaticRoutes(s *ghttp.Server) {
	// 静态文件目录
	s.AddStaticPath("/static", "static")
	s.AddStaticPath("/templates", "templates")

	// Static测试页面路由
	s.BindHandler("/login", func(r *ghttp.Request) {
		r.Response.ServeFile("static/index.html")
	})

	s.BindHandler("/callback", func(r *ghttp.Request) {
		r.Response.ServeFile("static/callback.html")
	})

	s.BindHandler("/dashboard", func(r *ghttp.Request) {
		r.Response.ServeFile("static/dashboard.html")
	})

	s.BindHandler("/error", func(r *ghttp.Request) {
		r.Response.ServeFile("static/error.html")
	})

	// Template测试页面路由
	s.BindHandler("/template/login", func(r *ghttp.Request) {
		r.Response.ServeFile("templates/index.html")
	})

	s.BindHandler("/template/callback", func(r *ghttp.Request) {
		r.Response.ServeFile("templates/callback.html")
	})

	s.BindHandler("/template/dashboard", func(r *ghttp.Request) {
		r.Response.ServeFile("templates/dashboard.html")
	})

	s.BindHandler("/template/error", func(r *ghttp.Request) {
		r.Response.ServeFile("templates/error.html")
	})
}
