package main

import (
	"context-id-backend/internal/router"
	"context-id-backend/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"

	// 导入PostgreSQL数据库驱动
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
)

func main() {
	// 获取初始化上下文
	// gctx.GetInitCtx() 用于获取一个初始化的上下文对象
	// 这个上下文对象可以用于整个应用程序的生命周期管理
	ctx := gctx.GetInitCtx()

	// 设置配置文件路径
	g.Cfg().GetAdapter().(*gcfg.AdapterFile).SetPath("conf")

	// 初始化服务
	service.Init(ctx)

	s := g.Server()

	// 设置服务器地址为8080端口
	s.SetAddr(":8080")

	// 全局中间件
	s.Use(
		ghttp.MiddlewareCORS,
	)

	// 初始化所有路由
	router.InitRoutes(s)

	s.Run()
}
