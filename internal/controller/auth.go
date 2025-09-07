package controller

import (
	"context-id-backend/internal/model"
	"context-id-backend/internal/service"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type AuthController struct{}

var Auth = &AuthController{}

// GetLoginURL 获取Casdoor登录URL（带安全state参数）
func (c *AuthController) GetLoginURL(r *ghttp.Request) {
	ctx := r.Context()

	// 要求前端必须传入redirect_uri参数
	redirectURI := r.Get("redirect_uri").String()
	if redirectURI == "" {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"code":    400,
			"message": "redirect_uri参数不能为空",
		})
		return
	}

	loginURL, _, err := service.Casdoor.GetLoginURL(ctx, redirectURI)
	if err != nil {
		g.Log().Error(ctx, "Failed to generate login URL:", err)
		r.Response.Status = 500
		r.Response.WriteJson(g.Map{
			"code":    500,
			"message": "生成登录URL失败: " + err.Error(),
		})
		return
	}

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

	// 要求前端必须传入redirect_uri参数
	redirectURI := r.Get("redirect_uri").String()
	if redirectURI == "" {
		r.Response.Status = 400
		r.Response.WriteJson(g.Map{
			"code":    400,
			"message": "redirect_uri参数不能为空",
		})
		return
	}

	// 注册应该使用OAuth2流程，以支持注册后自动登录和重定向
	// enablePassword = false: 完整OAuth2注册流程，注册后可以重定向
	enablePassword := false
	signupURL, _, err := service.Casdoor.GetSignupURL(ctx, enablePassword, redirectURI)
	if err != nil {
		g.Log().Error(ctx, "Failed to generate signup URL:", err)
		r.Response.Status = 500
		r.Response.WriteJson(g.Map{
			"code":    500,
			"message": "生成注册URL失败: " + err.Error(),
		})
		return
	}

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

// 注意: 移除了GET /auth/callback方法
// 在前后端分离架构中，Casdoor的redirect_uri应该指向前端页面
// 前端页面获取code+state后，通过POST /auth/callback API交换token

// Login OAuth2 Token交换API（前后端分离架构）
// 前端通过POST请求发送code+state，后端返回access_token
// 这是标准的OAuth2授权码流程的token交换步骤
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
