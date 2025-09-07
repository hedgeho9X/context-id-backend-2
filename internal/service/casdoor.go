package service

import (
	"context"
	"context-id-backend/internal/dao"
	"context-id-backend/internal/model"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/joho/godotenv"
)

// CasdoorConfig Casdoor配置结构体
type CasdoorConfig struct {
	Endpoint         string // 内部通信地址（容器间）
	ExternalEndpoint string // 外部访问地址（浏览器）
	ClientId         string
	ClientSecret     string
	JwtSecret        string // JWT公钥，仅支持PEM格式公钥文件内容
	OrganizationName string
	ApplicationName  string
}

// AppConfig 应用配置结构体
type AppConfig struct {
	ExternalUrl        string // 应用外部访问地址
	CasdoorExternalUrl string // Casdoor外部访问地址
}

// StateStore State参数存储（用于CSRF防护）
type StateStore struct {
	states map[string]time.Time // state -> 过期时间
	mutex  sync.RWMutex
}

// CasdoorService Casdoor认证服务
type CasdoorService struct {
	config     *CasdoorConfig
	appConfig  *AppConfig
	stateStore *StateStore
}

var Casdoor = &CasdoorService{}

// Init 初始化Casdoor客户端 (参考tutorial的配置加载方式)
func (s *CasdoorService) Init(ctx context.Context) error {
	g.Log().Info(ctx, "正在初始化Casdoor服务...")

	// 1. 尝试加载环境变量文件
	envFiles := []string{".env"}
	loaded := false

	for _, envFile := range envFiles {
		if err := godotenv.Load(envFile); err == nil {
			g.Log().Info(ctx, "✅ 成功加载环境变量文件:", envFile)
			loaded = true
			break
		}
	}

	if !loaded {
		g.Log().Warning(ctx, "未找到环境变量文件，尝试从系统环境变量或配置文件加载")
	}

	// 2. 从环境变量或配置文件加载配置
	config, err := s.loadConfig(ctx)
	if err != nil {
		return fmt.Errorf("配置加载失败: %w", err)
	}

	// 3. 验证必需的配置
	if err := s.validateConfig(config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	s.config = config

	// 4. 初始化state存储
	s.stateStore = &StateStore{
		states: make(map[string]time.Time),
	}

	// 启动state清理goroutine
	go s.cleanupExpiredStates()

	// 5. 加载应用配置
	appConfig, err := s.loadAppConfig(ctx)
	if err != nil {
		return fmt.Errorf("应用配置加载失败: %w", err)
	}
	s.appConfig = appConfig

	// 5. 初始化Casdoor全局配置
	casdoorsdk.InitConfig(
		config.Endpoint,
		config.ClientId,
		config.ClientSecret,
		config.JwtSecret,
		config.OrganizationName,
		config.ApplicationName,
	)

	g.Log().Info(ctx, "✅ Casdoor服务初始化完成:")
	g.Log().Info(ctx, "   - Internal Endpoint:", config.Endpoint)
	g.Log().Info(ctx, "   - External Endpoint:", config.ExternalEndpoint)
	g.Log().Info(ctx, "   - ClientId:", config.ClientId)
	g.Log().Info(ctx, "   - Organization:", config.OrganizationName)
	g.Log().Info(ctx, "   - Application:", config.ApplicationName)

	// 记录JWT密钥信息
	if strings.Contains(config.JwtSecret, "-----BEGIN PUBLIC KEY-----") {
		g.Log().Info(ctx, "   - JWT密钥类型: RS256 (PEM公钥)")
	} else if len(config.JwtSecret) < 100 {
		g.Log().Info(ctx, "   - JWT密钥类型: HS256 (对称密钥)")
	} else {
		g.Log().Info(ctx, "   - JWT密钥: 已配置")
	}

	return nil
}

// safeSubstring 安全地截取字符串，避免越界
func (s *CasdoorService) safeSubstring(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen]
}

// generateState 生成随机state参数（用于CSRF防护）
func (s *CasdoorService) generateState(ctx context.Context) (string, error) {
	// 生成32字节随机数据
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(bytes)

	// 存储state，设置10分钟过期时间
	s.stateStore.mutex.Lock()
	s.stateStore.states[state] = time.Now().Add(10 * time.Minute)
	s.stateStore.mutex.Unlock()

	g.Log().Debug(ctx, "Generated state:", s.safeSubstring(state, 16)+"...")
	return state, nil
}

// validateState 验证state参数
func (s *CasdoorService) validateState(ctx context.Context, state string) error {
	if state == "" {
		return fmt.Errorf("state parameter is empty")
	}

	s.stateStore.mutex.Lock()
	defer s.stateStore.mutex.Unlock()

	expireTime, exists := s.stateStore.states[state]
	if !exists {
		g.Log().Warning(ctx, "Invalid state parameter:", s.safeSubstring(state, 16)+"...")
		return fmt.Errorf("invalid or expired state parameter")
	}

	// 检查是否过期
	if time.Now().After(expireTime) {
		delete(s.stateStore.states, state)
		g.Log().Warning(ctx, "Expired state parameter:", s.safeSubstring(state, 16)+"...")
		return fmt.Errorf("state parameter has expired")
	}

	// 使用后立即删除（一次性使用）
	delete(s.stateStore.states, state)
	g.Log().Debug(ctx, "State validation successful:", s.safeSubstring(state, 16)+"...")
	return nil
}

// cleanupExpiredStates 清理过期的state参数
func (s *CasdoorService) cleanupExpiredStates() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		s.stateStore.mutex.Lock()
		now := time.Now()
		for state, expireTime := range s.stateStore.states {
			if now.After(expireTime) {
				delete(s.stateStore.states, state)
			}
		}
		s.stateStore.mutex.Unlock()
	}
}

// loadConfig 加载配置 (参考tutorial实现)
func (s *CasdoorService) loadConfig(ctx context.Context) (*CasdoorConfig, error) {
	config := &CasdoorConfig{}

	// 优先从环境变量加载
	if endpoint := os.Getenv("CASDOOR_ENDPOINT"); endpoint != "" {
		config.Endpoint = endpoint
	}
	if externalEndpoint := os.Getenv("CASDOOR_EXTERNAL_ENDPOINT"); externalEndpoint != "" {
		config.ExternalEndpoint = externalEndpoint
	}
	if clientId := os.Getenv("CASDOOR_CLIENT_ID"); clientId != "" {
		config.ClientId = clientId
	}
	if clientSecret := os.Getenv("CASDOOR_CLIENT_SECRET"); clientSecret != "" {
		config.ClientSecret = clientSecret
	}
	if jwtSecret := os.Getenv("CASDOOR_JWT_SECRET"); jwtSecret != "" {
		g.Log().Info(ctx, "🔑 从环境变量加载JWT密钥:", jwtSecret)
		config.JwtSecret = s.loadJwtSecret(ctx, jwtSecret)
		g.Log().Info(ctx, "🔑 JWT密钥加载结果长度:", len(config.JwtSecret))
	}
	if orgName := os.Getenv("CASDOOR_ORGANIZATION_NAME"); orgName != "" {
		config.OrganizationName = orgName
	}
	if appName := os.Getenv("CASDOOR_APPLICATION_NAME"); appName != "" {
		config.ApplicationName = appName
	}

	// 如果环境变量没有设置，尝试从配置文件加载
	if config.Endpoint == "" || config.ClientId == "" {
		cfg := g.Cfg()

		// 尝试从配置文件加载，如果失败则使用默认值
		if config.Endpoint == "" {
			if endpoint, err := cfg.Get(ctx, "casdoor.endpoint"); err == nil && endpoint != nil {
				config.Endpoint = endpoint.String()
			} else {
				config.Endpoint = "http://localhost:8000"
			}
		}
		if config.ExternalEndpoint == "" {
			// 优先使用app.casdoorExternalUrl配置
			if casdoorExternalUrl, err := cfg.Get(ctx, "app.casdoorExternalUrl"); err == nil && casdoorExternalUrl != nil {
				config.ExternalEndpoint = casdoorExternalUrl.String()
			} else if externalEndpoint, err := cfg.Get(ctx, "casdoor.externalEndpoint"); err == nil && externalEndpoint != nil {
				config.ExternalEndpoint = externalEndpoint.String()
			} else {
				config.ExternalEndpoint = "http://localhost:8000"
			}
		}
		if config.ClientId == "" {
			if clientId, err := cfg.Get(ctx, "casdoor.clientId"); err == nil && clientId != nil {
				config.ClientId = clientId.String()
			}
		}
		if config.ClientSecret == "" {
			if clientSecret, err := cfg.Get(ctx, "casdoor.clientSecret"); err == nil && clientSecret != nil {
				config.ClientSecret = clientSecret.String()
			}
		}
		if config.JwtSecret == "" {
			if jwtSecret, err := cfg.Get(ctx, "casdoor.jwtSecret"); err == nil && jwtSecret != nil {
				config.JwtSecret = s.loadJwtSecret(ctx, jwtSecret.String())
			} else {
				// 尝试默认的公钥文件路径
				defaultPaths := []string{
					"./certs/token_jwt_public_key.pem",
				}
				loaded := false
				for _, path := range defaultPaths {
					if _, err := os.Stat(path); err == nil {
						config.JwtSecret = s.loadJwtSecret(ctx, path)
						g.Log().Info(ctx, "✅ 使用默认公钥文件:", path)
						loaded = true
						break
					}
				}
				if !loaded {
					g.Log().Warning(ctx, "⚠️ 未找到默认公钥文件，请确保已配置JWT密钥")
					config.JwtSecret = "" // 设置为空，让后续验证处理
				}
			}
		}
		if config.OrganizationName == "" {
			if orgName, err := cfg.Get(ctx, "casdoor.organizationName"); err == nil && orgName != nil {
				config.OrganizationName = orgName.String()
			} else {
				config.OrganizationName = "hello"
			}
		}
		if config.ApplicationName == "" {
			if appName, err := cfg.Get(ctx, "casdoor.applicationName"); err == nil && appName != nil {
				config.ApplicationName = appName.String()
			} else {
				config.ApplicationName = "context-ID-DEV"
			}
		}
	}

	return config, nil
}

// loadAppConfig 加载应用配置
func (s *CasdoorService) loadAppConfig(ctx context.Context) (*AppConfig, error) {
	config := &AppConfig{}
	cfg := g.Cfg()

	// 从配置文件加载
	if externalUrl, err := cfg.Get(ctx, "app.externalUrl"); err == nil && externalUrl != nil {
		config.ExternalUrl = externalUrl.String()
	} else {
		config.ExternalUrl = "http://localhost:8080"
	}

	if casdoorExternalUrl, err := cfg.Get(ctx, "app.casdoorExternalUrl"); err == nil && casdoorExternalUrl != nil {
		config.CasdoorExternalUrl = casdoorExternalUrl.String()
	} else {
		config.CasdoorExternalUrl = "http://localhost:8000"
	}

	// 环境变量覆盖配置文件
	if externalUrl := os.Getenv("APP_EXTERNAL_URL"); externalUrl != "" {
		config.ExternalUrl = externalUrl
	}
	if casdoorExternalUrl := os.Getenv("APP_CASDOOR_EXTERNAL_URL"); casdoorExternalUrl != "" {
		config.CasdoorExternalUrl = casdoorExternalUrl
	}

	g.Log().Info(ctx, "✅ 应用配置加载完成:")
	g.Log().Info(ctx, "   - External URL:", config.ExternalUrl)
	g.Log().Info(ctx, "   - Casdoor External URL:", config.CasdoorExternalUrl)

	return config, nil
}

// loadJwtSecret 加载JWT密钥 (参考tutorial的成功方法)
func (s *CasdoorService) loadJwtSecret(ctx context.Context, jwtSecret string) string {
	// 如果是文件路径，读取文件内容
	if strings.HasPrefix(jwtSecret, "/") || strings.HasPrefix(jwtSecret, "./") || strings.HasSuffix(jwtSecret, ".pem") {
		if content, err := os.ReadFile(jwtSecret); err == nil {
			g.Log().Info(ctx, "✅ 成功从文件加载JWT密钥:", jwtSecret)
			// 简单处理，移除多余的空白字符但不做严格验证
			return strings.TrimSpace(string(content))
		} else {
			g.Log().Warning(ctx, "❌ 无法读取JWT密钥文件:", jwtSecret, "错误:", err)
			return jwtSecret // 回退到原始值
		}
	} else {
		// 直接处理密钥内容（参考tutorial方法）
		return strings.ReplaceAll(jwtSecret, "\\n", "\n")
	}
}

// validateConfig 验证配置 (参考tutorial实现)
func (s *CasdoorService) validateConfig(config *CasdoorConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("casdoor endpoint 不能为空")
	}
	if config.ClientId == "" {
		return fmt.Errorf("casdoor client ID 不能为空")
	}
	if config.ClientSecret == "" {
		return fmt.Errorf("casdoor client secret 不能为空")
	}
	if config.JwtSecret == "" {
		return fmt.Errorf("casdoor JWT secret 不能为空")
	}
	if config.OrganizationName == "" {
		return fmt.Errorf("casdoor organization name 不能为空")
	}
	if config.ApplicationName == "" {
		return fmt.Errorf("casdoor application name 不能为空")
	}
	return nil
}

// getExternalEndpoint 获取外部访问的endpoint（用于生成URL）
func (s *CasdoorService) getExternalEndpoint() string {
	// 优先使用应用配置中的Casdoor外部URL
	if s.appConfig != nil && s.appConfig.CasdoorExternalUrl != "" {
		return s.appConfig.CasdoorExternalUrl
	}
	if s.config.ExternalEndpoint != "" {
		return s.config.ExternalEndpoint
	}
	return s.config.Endpoint
}

// GetLoginURL 获取Casdoor登录URL（带安全state参数）
func (s *CasdoorService) GetLoginURL(ctx context.Context, redirectURI string) (string, string, error) {
	if redirectURI == "" {
		return "", "", fmt.Errorf("redirectURI不能为空")
	}

	// 生成安全的state参数
	state, err := s.generateState(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// 使用SDK生成URL，然后替换endpoint
	loginURL := casdoorsdk.GetSigninUrl(redirectURI)

	// 替换内部endpoint为外部endpoint
	externalEndpoint := s.getExternalEndpoint()
	if s.config.Endpoint != externalEndpoint {
		loginURL = strings.Replace(loginURL, s.config.Endpoint, externalEndpoint, 1)
	}

	// 替换URL中的state参数为我们生成的安全state
	// Casdoor SDK默认使用应用名称作为state
	loginURL = strings.Replace(loginURL, "state="+s.config.ApplicationName, "state="+state, 1)

	g.Log().Info(ctx, "Generated secure login URL with state:", s.safeSubstring(state, 16)+"...")

	return loginURL, state, nil
}

// GetSignupURL 获取Casdoor注册URL
func (s *CasdoorService) GetSignupURL(ctx context.Context, enablePassword bool, redirectURI string) (string, string, error) {
	if redirectURI == "" {
		return "", "", fmt.Errorf("redirectURI不能为空")
	}

	var signupURL string
	var state string

	if enablePassword {
		// 简化注册页面 (仅密码注册)，不支持重定向
		signupURL = casdoorsdk.GetSignupUrl(enablePassword, redirectURI)
		state = "" // 简化模式不需要state
	} else {
		// 完整OAuth2注册流程，支持注册后重定向
		// 生成安全的state参数
		var err error
		state, err = s.generateState(ctx)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate state: %w", err)
		}

		// 使用SDK生成注册URL
		signupURL = casdoorsdk.GetSignupUrl(enablePassword, redirectURI)

		// 替换URL中的state参数为我们生成的安全state
		// 注册URL实际上是基于登录URL生成的，所以也需要替换state
		signupURL = strings.Replace(signupURL, "state="+s.config.ApplicationName, "state="+state, 1)
	}

	// 替换内部endpoint为外部endpoint
	externalEndpoint := s.getExternalEndpoint()
	if s.config.Endpoint != externalEndpoint {
		signupURL = strings.Replace(signupURL, s.config.Endpoint, externalEndpoint, 1)
	}

	g.Log().Info(ctx, "Generated signup URL (enablePassword=%t, state=%s):", enablePassword, s.safeSubstring(state, 16)+"...")

	return signupURL, state, nil
}

// GetMyProfileURL 获取当前用户资料页面URL
func (s *CasdoorService) GetMyProfileURL(ctx context.Context, accessToken string) string {
	myProfileURL := casdoorsdk.GetMyProfileUrl(accessToken)

	// 替换内部endpoint为外部endpoint
	externalEndpoint := s.getExternalEndpoint()
	if s.config.Endpoint != externalEndpoint {
		myProfileURL = strings.Replace(myProfileURL, s.config.Endpoint, externalEndpoint, 1)
	}

	g.Log().Info(ctx, "Generated my profile URL:", myProfileURL)

	return myProfileURL
}

// GetToken 通过授权码获取token (使用tutorial中的成功方法)
func (s *CasdoorService) GetToken(ctx context.Context, code, state string) (string, error) {
	token, err := casdoorsdk.GetOAuthToken(code, state)
	if err != nil {
		g.Log().Error(ctx, "Failed to get OAuth token:", err)
		return "", err
	}
	return token.AccessToken, nil
}

// ParseJwtToken 解析JWT token获取用户信息 (使用tutorial中的成功方法)
func (s *CasdoorService) ParseJwtToken(ctx context.Context, token string) (*casdoorsdk.Claims, error) {
	claims, err := casdoorsdk.ParseJwtToken(token)
	if err != nil {
		g.Log().Error(ctx, "Failed to parse JWT token:", err)
		return nil, err
	}
	return claims, nil
}

// GetUserInfo 获取用户信息 (使用tutorial中的成功方法)
func (s *CasdoorService) GetUserInfo(ctx context.Context, username string) (*casdoorsdk.User, error) {
	user, err := casdoorsdk.GetUser(username)
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
	// 直接解析Casdoor JWT token
	claims, err := s.ParseJwtToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// 从JWT claims中构建用户信息，不需要查询本地数据库
	user := &model.User{
		Username:    claims.User.Name,
		Email:       claims.User.Email,
		DisplayName: claims.User.DisplayName,
		Avatar:      claims.User.Avatar,
		Phone:       claims.User.Phone,
		Status:      1, // 默认状态为活跃
	}

	return user, nil
}

// UserInfo 用户信息结构体 (从tutorial中复制)
type UserInfo struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Avatar      string `json:"avatar"`
}

// HandleCallback 处理OAuth回调 (使用tutorial中的成功方法，添加安全验证)
func (s *CasdoorService) HandleCallback(ctx context.Context, code, state string) (*UserInfo, string, error) {
	// 1. 验证state参数（CSRF防护）
	if err := s.validateState(ctx, state); err != nil {
		g.Log().Error(ctx, "State validation failed:", err)
		return nil, "", fmt.Errorf("CSRF protection: %w", err)
	}

	// 2. 获取OAuth token
	token, err := casdoorsdk.GetOAuthToken(code, state)
	if err != nil {
		g.Log().Error(ctx, "Failed to get OAuth token:", err)
		return nil, "", err
	}

	// 解析JWT token获取用户信息
	claims, err := casdoorsdk.ParseJwtToken(token.AccessToken)
	if err != nil {
		g.Log().Error(ctx, "Failed to parse JWT token:", err)
		return nil, "", err
	}

	// 转换为我们的用户信息格式
	userInfo := &UserInfo{
		Username:    claims.User.Name,
		DisplayName: claims.User.DisplayName,
		Email:       claims.User.Email,
		Phone:       claims.User.Phone,
		Avatar:      claims.User.Avatar,
	}

	return userInfo, token.AccessToken, nil
}

// ValidateToken 验证token (使用tutorial中的成功方法)
func (s *CasdoorService) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	claims, err := casdoorsdk.ParseJwtToken(token)
	if err != nil {
		g.Log().Error(ctx, "Failed to validate token:", err)
		return nil, err
	}

	userInfo := &UserInfo{
		Username:    claims.User.Name,
		DisplayName: claims.User.DisplayName,
		Email:       claims.User.Email,
		Phone:       claims.User.Phone,
		Avatar:      claims.User.Avatar,
	}

	return userInfo, nil
}
