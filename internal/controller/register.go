package controller

import "github.com/gogf/gf/v2/net/ghttp"

// Register 注册所有控制器路由
func Register(group *ghttp.RouterGroup) {
	// 认证相关路由
	group.Group("/auth", func(authGroup *ghttp.RouterGroup) {
		// 这些路由不需要认证
		authGroup.GET("/url", Auth.GetLoginURL)
		authGroup.POST("/callback", Auth.Callback)
		authGroup.POST("/login", Auth.Login)

		// 这些路由需要认证
		authGroup.GET("/user", Auth.GetUserInfo)
		authGroup.POST("/logout", Auth.Logout)
	})

	// 其他业务路由可以在这里添加
	// group.Group("/memory", func(memoryGroup *ghttp.RouterGroup) {
	//     // 记忆系统相关路由
	// })
}
