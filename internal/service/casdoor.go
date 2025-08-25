package service

import (
	"context"
	"context-id-backend/internal/dao"
	"context-id-backend/internal/model"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type CasdoorService struct {
	client *casdoorsdk.Client
}

var Casdoor = &CasdoorService{}

// Init 初始化Casdoor客户端
func (s *CasdoorService) Init(ctx context.Context) error {
	cfg := g.Cfg()

	endpoint := cfg.MustGet(ctx, "casdoor.endpoint").String()
	clientId := cfg.MustGet(ctx, "casdoor.clientId").String()
	clientSecret := cfg.MustGet(ctx, "casdoor.clientSecret").String()
	jwtSecret := cfg.MustGet(ctx, "casdoor.jwtSecret").String()
	organizationName := cfg.MustGet(ctx, "casdoor.organizationName").String()
	applicationName := cfg.MustGet(ctx, "casdoor.applicationName").String()

	s.client = casdoorsdk.NewClient(endpoint, clientId, clientSecret, jwtSecret, organizationName, applicationName)

	g.Log().Info(ctx, "Casdoor client initialized successfully")
	return nil
}

// GetAuthURL 获取Casdoor登录URL
func (s *CasdoorService) GetAuthURL(ctx context.Context, redirectURI, state string) string {
	originalURL := s.client.GetSigninUrl(redirectURI)
	g.Log().Info(ctx, "Original Casdoor URL:", originalURL)

	// 替换URL中的主机名为localhost（供前端访问）
	baseURL := originalURL
	baseURL = strings.Replace(baseURL, "http://casdoor:8000", "http://localhost:8000", -1)
	baseURL = strings.Replace(baseURL, "http://contextid-casdoor:8000", "http://localhost:8000", -1)
	baseURL = strings.Replace(baseURL, "casdoor:8000", "localhost:8000", -1)

	// 如果提供了自定义state，替换URL中的state参数
	if state != "" && state != "random_state" {
		// 使用正则表达式替换state参数
		if strings.Contains(baseURL, "state=") {
			re := regexp.MustCompile(`state=[^&]*`)
			baseURL = re.ReplaceAllString(baseURL, "state="+state)
		}
	}

	g.Log().Info(ctx, "Final auth URL:", baseURL)
	return baseURL
}

// GetToken 通过授权码获取token
func (s *CasdoorService) GetToken(ctx context.Context, code, state string) (string, error) {
	token, err := s.client.GetOAuthToken(code, state)
	if err != nil {
		g.Log().Error(ctx, "Failed to get OAuth token:", err)
		return "", err
	}
	return token.AccessToken, nil
}

// ParseJwtToken 解析JWT token获取用户信息
func (s *CasdoorService) ParseJwtToken(ctx context.Context, token string) (*casdoorsdk.Claims, error) {
	claims, err := s.client.ParseJwtToken(token)
	if err != nil {
		g.Log().Error(ctx, "Failed to parse JWT token:", err)
		return nil, err
	}
	return claims, nil
}

// GetUserInfo 获取用户信息
func (s *CasdoorService) GetUserInfo(ctx context.Context, username string) (*casdoorsdk.User, error) {
	user, err := s.client.GetUser(username)
	if err != nil {
		g.Log().Error(ctx, "Failed to get user info:", err)
		return nil, err
	}
	return user, nil
}

// SyncUser 同步Casdoor用户到本地数据库
func (s *CasdoorService) SyncUser(ctx context.Context, casdoorUser *casdoorsdk.User) (*model.User, error) {
	// 检查用户是否已存在
	existingUser, err := dao.User.GetByUsername(ctx, casdoorUser.Name)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return nil, err
	}

	user := &model.User{
		Username:    casdoorUser.Name,
		Email:       casdoorUser.Email,
		DisplayName: casdoorUser.DisplayName,
		Avatar:      casdoorUser.Avatar,
		Phone:       casdoorUser.Phone,
		Status:      1,
	}

	if existingUser != nil {
		// 更新现有用户
		user.Id = existingUser.Id
		user.CreatedAt = existingUser.CreatedAt
		user.UpdatedAt = gtime.Now()
		err = dao.User.Update(ctx, user)
	} else {
		// 创建新用户
		user.CreatedAt = gtime.Now()
		user.UpdatedAt = gtime.Now()
		err = dao.User.Create(ctx, user)
	}

	if err != nil {
		g.Log().Error(ctx, "Failed to sync user:", err)
		return nil, err
	}

	g.Log().Info(ctx, "User synced successfully:", user.Username)
	return user, nil
}

// Login 用户登录处理
func (s *CasdoorService) Login(ctx context.Context, code, state string) (*model.UserLoginRes, error) {
	// 获取访问令牌
	token, err := s.GetToken(ctx, code, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	// 解析JWT token
	claims, err := s.ParseJwtToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 获取用户信息
	casdoorUser, err := s.GetUserInfo(ctx, claims.User.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 同步用户到本地数据库
	user, err := s.SyncUser(ctx, casdoorUser)
	if err != nil {
		return nil, fmt.Errorf("failed to sync user: %w", err)
	}

	// 生成本地JWT token（可选，也可以直接使用Casdoor的token）
	localToken, err := s.generateLocalToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate local token: %w", err)
	}

	return &model.UserLoginRes{
		Token: localToken,
		User:  user,
	}, nil
}

// generateLocalToken 生成本地JWT token
func (s *CasdoorService) generateLocalToken(ctx context.Context, user *model.User) (string, error) {
	// 这里可以使用GoFrame的JWT功能或者其他JWT库
	// 为了简化，这里返回一个简单的token格式
	tokenData := map[string]interface{}{
		"user_id":  user.Id,
		"username": user.Username,
		"email":    user.Email,
		"exp":      gtime.Now().Add(24 * 7 * time.Hour).Unix(), // 7天过期
	}

	tokenBytes, err := json.Marshal(tokenData)
	if err != nil {
		return "", err
	}

	// 在实际项目中，你应该使用JWT库来生成签名的token
	// 这里为了演示简化处理
	return string(tokenBytes), nil
}

// VerifyToken 验证token
func (s *CasdoorService) VerifyToken(ctx context.Context, token string) (*model.User, error) {
	// 首先尝试解析本地token
	var tokenData map[string]interface{}
	err := json.Unmarshal([]byte(token), &tokenData)
	if err == nil {
		// 本地token
		if userId, ok := tokenData["user_id"].(float64); ok {
			return dao.User.GetById(ctx, uint64(userId))
		}
	}

	// 尝试解析Casdoor token
	claims, err := s.ParseJwtToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// 根据用户名获取本地用户信息
	user, err := dao.User.GetByUsername(ctx, claims.User.Name)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
