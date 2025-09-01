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

	loginURL := service.Casdoor.GetLoginURL(ctx, redirectURI)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"login_url": loginURL,
		},
	})
}

// GetSignupURL 获取Casdoor注册URL
func (c *AuthController) GetSignupURL(r *ghttp.Request) {
	ctx := r.Context()

	redirectURI := r.Get("redirect_uri", "http://localhost:8080/api/v1/auth/callback").String()
	// enablePassword := r.Get("enable_password", "true").Bool()
	//默认要求使用密码
	enablePassword := true
	signupURL := service.Casdoor.GetSignupURL(ctx, enablePassword, redirectURI)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"signup_url": signupURL,
		},
	})
}

// GetMyProfileURL 获取当前用户资料页面URL (需要token)
func (c *AuthController) GetMyProfileURL(r *ghttp.Request) {
	ctx := r.Context()

	accessToken := r.GetHeader("Authorization")

	// 移除 "Bearer " 前缀
	if len(accessToken) > 7 && accessToken[:7] == "Bearer " {
		accessToken = accessToken[7:]
	}

	if accessToken == "" {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"code":    400,
			"message": "缺少访问令牌",
		})
		return
	}

	myProfileURL := service.Casdoor.GetMyProfileURL(ctx, accessToken)

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"my_profile_url": myProfileURL,
		},
	})
}

// Callback Casdoor回调处理 (标准OAuth2 GET请求)
func (c *AuthController) Callback(r *ghttp.Request) {
	ctx := r.Context()

	// 只处理GET请求（标准OAuth2回调）
	code := r.Get("code").String()
	state := r.Get("state").String()

	if code == "" {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"code":    400,
			"message": "缺少授权码",
		})
		return
	}

	// 使用code交换token
	userInfo, token, err := service.Casdoor.HandleCallback(ctx, code, state)
	if err != nil {
		g.Log().Error(ctx, "Login failed:", err)
		r.Response.Status = 500
		r.Response.WriteJson(g.Map{
			"code":    500,
			"message": "登录失败: " + err.Error(),
		})
		return
	}

	// 返回JSON格式的用户信息和token
	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "登录成功",
		"data": g.Map{
			"access_token": token,
			"user":         userInfo,
		},
	})
}

// Login 用户登录（处理前端传来的code和state）(使用tutorial中的成功方法)
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

	// 使用tutorial中的HandleCallback方法
	userInfo, token, err := service.Casdoor.HandleCallback(ctx, req.Code, req.State)
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
		"data": g.Map{
			"token": token,
			"user":  userInfo,
		},
	})
}

// GetCurrentUser 获取当前用户信息 (类似tutorial中的/api/user，使用token验证)
func (c *AuthController) GetCurrentUser(r *ghttp.Request) {
	ctx := r.Context()

	// 从Header获取token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"code":    401,
			"message": "缺少Authorization头",
		})
		return
	}

	// 提取token（去掉"Bearer "前缀）
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	// 验证token并获取用户信息
	userInfo, err := service.Casdoor.ValidateToken(ctx, token)
	if err != nil {
		r.Response.Status = 401
		r.Response.WriteJson(g.Map{
			"code":    401,
			"message": "token无效",
		})
		return
	}

	r.Response.WriteJson(g.Map{
		"code":    200,
		"message": "success",
		"data": g.Map{
			"user": userInfo,
		},
	})
}
