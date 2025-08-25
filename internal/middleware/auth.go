package middleware

import (
	"context-id-backend/internal/service"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// Auth 认证中间件
func Auth(r *ghttp.Request) {
	ctx := r.Context()

	// 跳过认证的路径
	skipPaths := []string{
		"/health",
		"/api/v1/auth/login",
		"/api/v1/auth/callback",
		"/api/v1/auth/url",
		"/favicon.ico",
	}
	
	// 跳过静态文件
	staticExtensions := []string{".html", ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".svg"}
	for _, ext := range staticExtensions {
		if strings.HasSuffix(r.URL.Path, ext) {
			r.Middleware.Next()
			return
		}
	}
	
	// 如果是根路径，也跳过认证（用于访问首页）
	if r.URL.Path == "/" {
		r.Middleware.Next()
		return
	}

	for _, path := range skipPaths {
		if strings.HasPrefix(r.URL.Path, path) {
			r.Middleware.Next()
			return
		}
	}

	// 获取Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"code":    401,
			"message": "未提供认证信息",
		})
		return
	}

	// 检查Bearer token格式
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"code":    401,
			"message": "认证格式错误",
		})
		return
	}

	token := parts[1]

	// 验证token
	user, err := service.Casdoor.VerifyToken(ctx, token)
	if err != nil {
		g.Log().Error(ctx, "Token verification failed:", err)
		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"code":    401,
			"message": "认证失败",
		})
		return
	}

	// 将用户信息存储到上下文中
	r.SetCtxVar("user", user)
	r.Middleware.Next()
}
