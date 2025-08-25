package controller

import (
	"context-id-backend/internal/model"
	"context-id-backend/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type AuthController struct{}

var Auth = &AuthController{}

// GetLoginURL 获取Casdoor登录URL
func (c *AuthController) GetLoginURL(r *ghttp.Request) {
	ctx := r.Context()

	redirectURI := r.Get("redirect_uri", "http://localhost:8080/api/v1/auth/callback").String()
	state := r.Get("state", "random_state").String()

	loginURL := service.Casdoor.GetAuthURL(ctx, redirectURI, state)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"login_url": loginURL,
		},
	})
}

// Callback Casdoor回调处理
func (c *AuthController) Callback(r *ghttp.Request) {
	ctx := r.Context()

	var req *model.UserLoginReq
	if err := r.Parse(&req); err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 处理登录
	result, err := service.Casdoor.Login(ctx, req.Code, req.State)
	if err != nil {
		g.Log().Error(ctx, "Login failed:", err)
		r.Response.Status = 500
		r.Response.WriteJson(g.Map{
			"code":    500,
			"message": "登录失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "登录成功",
		"data":    result,
	})
}

// Login 用户登录（处理前端传来的code和state）
func (c *AuthController) Login(r *ghttp.Request) {
	ctx := r.Context()

	var req *model.UserLoginReq
	if err := r.Parse(&req); err != nil {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 处理登录
	result, err := service.Casdoor.Login(ctx, req.Code, req.State)
	if err != nil {
		g.Log().Error(ctx, "Login failed:", err)
		r.Response.Status = 500
		r.Response.WriteJson(g.Map{
			"code":    500,
			"message": "登录失败: " + err.Error(),
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "登录成功",
		"data":    result,
	})
}

// GetUserInfo 获取当前用户信息
func (c *AuthController) GetUserInfo(r *ghttp.Request) {
	user := r.GetCtxVar("user").Interface().(*model.User)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"data": &model.UserInfoRes{
			User: user,
		},
	})
}

// Logout 用户登出
func (c *AuthController) Logout(r *ghttp.Request) {
	// 在实际项目中，你可能需要将token加入黑名单
	// 这里简单返回成功
	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "登出成功",
	})
}
