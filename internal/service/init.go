package service

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// Init 初始化所有服务
func Init(ctx context.Context) {
	g.Log().Info(ctx, "Initializing services...")

	// 初始化Casdoor服务
	if err := Casdoor.Init(ctx); err != nil {
		g.Log().Fatal(ctx, "Failed to initialize Casdoor service:", err)
	}

	g.Log().Info(ctx, "All services initialized successfully")
}
