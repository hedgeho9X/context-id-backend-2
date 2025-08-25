package main

import (
	"context-id-backend/internal/controller"
	"context-id-backend/internal/middleware"
	"context-id-backend/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.GetInitCtx()

	// 初始化服务
	service.Init(ctx)

	s := g.Server()

	// 全局中间件
	s.Use(
		ghttp.MiddlewareCORS,
		middleware.Auth,
	)

	// 路由注册
	s.Group("/api/v1", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.Auth)
		controller.Register(group)
	})

	// 静态文件服务
	s.AddStaticPath("/", "public")

	// 健康检查
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status":  "ok",
			"message": "Context-ID Backend is running",
		})
	})

	s.Run()
}
