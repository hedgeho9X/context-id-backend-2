# Context-ID Backend API 详细文档

## 目录
- [概述](#概述)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [认证机制](#认证机制)
- [API 端点](#api-端点)
- [数据模型](#数据模型)
- [错误处理](#错误处理)
- [配置说明](#配置说明)
- [部署指南](#部署指南)
- [测试指南](#测试指南)

## 概述

Context-ID Backend 是一个基于 GoFrame v2 框架开发的后端 API 服务，集成了 Casdoor 认证系统，提供完整的用户认证、授权和用户管理功能。

### 主要特性

- 🔐 **Casdoor 集成**: 完整的 OAuth2 认证流程
- 🚀 **高性能**: 基于 GoFrame v2 框架
- 🗄️ **数据库支持**: PostgreSQL 数据库集成
- 🔒 **JWT 认证**: 安全的 Token 验证机制
- 📝 **RESTful API**: 标准的 REST API 设计
- 🎯 **中间件支持**: 认证、CORS 等中间件
- 📊 **健康检查**: 服务状态监控
- 🌐 **静态文件服务**: 支持前端页面托管

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.23.0+ | 编程语言 |
| GoFrame | v2.9.2 | Web 框架 |
| Casdoor | v0.42.0 | 认证系统 |
| PostgreSQL | - | 主数据库 |
| JWT | v4.5.0 | Token 认证 |
| Docker | - | 容器化部署 |

## 快速开始

### 1. 环境准备

```bash
# 克隆项目
git clone <repository-url>
cd Context-ID-backend

# 安装依赖
go mod tidy
```

### 2. 配置设置

创建配置文件：
```bash
cp config.example.env .env
```

编辑 `.env` 文件：
```env
CASDOOR_ENDPOINT=http://localhost:8000
CASDOOR_CLIENT_ID=your_client_id
CASDOOR_CLIENT_SECRET=your_client_secret
CASDOOR_JWT_SECRET=your_jwt_secret
CASDOOR_ORGANIZATION_NAME=your_org_name
CASDOOR_APPLICATION_NAME=your_app_name
```

### 3. 启动服务

```bash
# 开发模式
go run main.go

# 或使用脚本
./scripts/start.sh
```

服务将在 `http://localhost:8080` 启动

## 认证机制

### OAuth2 流程

1. **获取登录 URL**: 客户端调用 `/api/v1/auth/login-url` 获取 Casdoor 登录页面
2. **用户授权**: 用户在 Casdoor 页面完成登录
3. **回调处理**: Casdoor 重定向到 `/api/v1/auth/callback` 并携带授权码
4. **Token 交换**: 系统使用授权码换取访问令牌
5. **用户信息**: 使用访问令牌获取用户信息

### Token 使用

所有需要认证的 API 请求都需要在 Header 中携带 JWT Token：

```http
Authorization: Bearer <your_jwt_token>
```

## API 端点

### 基础信息

| 属性 | 值 |
|------|-----|
| Base URL | `http://localhost:8080` |
| API Version | `v1` |
| API Prefix | `/api/v1` |
| Content-Type | `application/json` |

### 1. 系统端点

#### 1.1 服务器信息

```http
GET /
```

**描述**: 获取服务器基本信息和可用端点

**响应示例**:
```json
{
  "status": "ok",
  "message": "Context-ID Backend API Server",
  "version": "1.0.0",
  "api": {
    "v1": "/api/v1"
  },
  "test_pages": {
    "static": {
      "login": "/login",
      "callback": "/callback",
      "dashboard": "/dashboard",
      "error": "/error"
    },
    "template": {
      "login": "/template/login",
      "callback": "/template/callback",
      "dashboard": "/template/dashboard",
      "error": "/template/error"
    }
  },
  "static_files": {
    "static": "/static",
    "templates": "/templates"
  }
}
```

#### 1.2 健康检查

```http
GET /health
```

**描述**: 检查服务健康状态

**响应示例**:
```json
{
  "status": "ok",
  "message": "Context-ID Backend is running",
  "apis": {
    "v1_health": "/api/v1/health"
  }
}
```

#### 1.3 API v1 信息

```http
GET /api/v1/
```

**描述**: 获取 API v1 版本信息和端点列表

**响应示例**:
```json
{
  "status": "ok",
  "message": "Context-ID API v1",
  "version": "1.0.0",
  "endpoints": {
    "auth": {
      "login_url": "/api/v1/auth/login-url",
      "signup_url": "/api/v1/auth/signup-url",
      "callback": "/api/v1/auth/callback",
      "user_info": "/api/v1/user"
    },
    "protected": {
      "my_profile": "/api/v1/auth/my-profile-url"
    }
  }
}
```

#### 1.4 API 健康检查

```http
GET /api/v1/health
```

**描述**: API v1 版本健康检查

**响应示例**:
```json
{
  "status": "ok",
  "message": "API v1 is running"
}
```

### 2. 认证端点

#### 2.1 获取登录 URL

```http
GET /api/v1/auth/login-url
```

**描述**: 获取 Casdoor 登录页面 URL

**查询参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|---------|------|
| redirect_uri | string | 否 | `http://localhost:8080/api/v1/auth/callback` | 登录成功后的回调地址 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "login_url": "http://localhost:8000/login/oauth/authorize?client_id=ca09f4429d05c4226155&redirect_uri=http://localhost:8080/api/v1/auth/callback&response_type=code&scope=read&state=casdoor"
  }
}
```

#### 2.2 获取注册 URL

```http
GET /api/v1/auth/signup-url
```

**描述**: 获取 Casdoor 用户注册页面 URL

**查询参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|---------|------|
| redirect_uri | string | 否 | `http://localhost:8080/api/v1/auth/callback` | 注册成功后的回调地址 |
| enable_password | boolean | 否 | true | 是否启用密码注册模式 |

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "signup_url": "http://localhost:8000/signup/oauth/authorize?client_id=ca09f4429d05c4226155&redirect_uri=http://localhost:8080/api/v1/auth/callback&response_type=code&scope=read&state=casdoor"
  }
}
```

#### 2.3 OAuth 回调处理

```http
GET /api/v1/auth/callback
```

**描述**: 处理 Casdoor OAuth2 回调，完成用户登录

**查询参数**:
| 参数 | 类型 | 必需 | 描述 |
|------|------|------|------|
| code | string | 是 | OAuth2 授权码 |
| state | string | 是 | OAuth2 状态参数 |

**响应示例**:
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ...",
    "user": {
      "username": "admin",
      "displayName": "管理员",
      "email": "admin@example.com",
      "phone": "13800000000",
      "avatar": "http://localhost:8000/img/admin.png"
    }
  }
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "缺少授权码"
}
```

#### 2.4 获取用户信息

```http
GET /api/v1/user
```

**描述**: 通过 JWT Token 获取当前登录用户信息

**请求头**:
```http
Authorization: Bearer <jwt_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user": {
      "username": "admin",
      "displayName": "管理员",
      "email": "admin@example.com",
      "phone": "13800000000",
      "avatar": "http://localhost:8000/img/admin.png"
    }
  }
}
```

**错误响应**:
```json
{
  "code": 401,
  "message": "缺少Authorization头"
}
```

#### 2.5 获取用户资料页面 URL (需要认证)

```http
GET /api/v1/auth/my-profile-url
```

**描述**: 获取当前用户在 Casdoor 中的个人资料页面 URL

**请求头**:
```http
Authorization: Bearer <jwt_token>
```

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "my_profile_url": "http://localhost:8000/account?access_token=eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ..."
  }
}
```

**错误响应**:
```json
{
  "code": 400,
  "message": "缺少访问令牌"
}
```

### 3. 静态页面端点

#### 3.1 测试页面 (Static)

| 路径 | 描述 | 文件路径 |
|------|------|----------|
| `/login` | 登录页面 | `static/index.html` |
| `/callback` | 回调页面 | `static/callback.html` |
| `/dashboard` | 用户仪表板 | `static/dashboard.html` |
| `/error` | 错误页面 | `static/error.html` |

#### 3.2 测试页面 (Template)

| 路径 | 描述 | 文件路径 |
|------|------|----------|
| `/template/login` | 登录页面模板 | `templates/index.html` |
| `/template/callback` | 回调页面模板 | `templates/callback.html` |
| `/template/dashboard` | 仪表板模板 | `templates/dashboard.html` |
| `/template/error` | 错误页面模板 | `templates/error.html` |

#### 3.3 静态文件服务

| 路径 | 描述 |
|------|------|
| `/static/*` | 静态文件目录 |
| `/templates/*` | 模板文件目录 |

## 数据模型

### 1. 用户模型 (User)

```go
type User struct {
    Id          uint64      `json:"id" db:"id"`                    // 用户ID
    Username    string      `json:"username" db:"username"`        // 用户名
    Email       string      `json:"email" db:"email"`              // 邮箱
    DisplayName string      `json:"displayName" db:"display_name"` // 显示名称
    Avatar      string      `json:"avatar" db:"avatar"`            // 头像URL
    Phone       string      `json:"phone" db:"phone"`              // 手机号
    Status      int         `json:"status" db:"status"`            // 状态 (1:活跃, 0:禁用)
    CreatedAt   *gtime.Time `json:"createdAt" db:"created_at"`     // 创建时间
    UpdatedAt   *gtime.Time `json:"updatedAt" db:"updated_at"`     // 更新时间
}
```

### 2. 用户登录请求 (UserLoginReq)

```go
type UserLoginReq struct {
    Code  string `json:"code" v:"required#授权码不能为空"`   // OAuth2授权码
    State string `json:"state" v:"required#状态码不能为空"` // OAuth2状态参数
}
```

### 3. 用户登录响应 (UserLoginRes)

```go
type UserLoginRes struct {
    Token string `json:"token"` // JWT访问令牌
    User  *User  `json:"user"`  // 用户信息
}
```

### 4. 用户信息响应 (UserInfoRes)

```go
type UserInfoRes struct {
    User *User `json:"user"` // 用户信息
}
```

### 5. Casdoor 用户信息 (UserInfo)

```go
type UserInfo struct {
    Username    string `json:"username"`    // 用户名
    DisplayName string `json:"displayName"` // 显示名称
    Email       string `json:"email"`       // 邮箱
    Phone       string `json:"phone"`       // 手机号
    Avatar      string `json:"avatar"`      // 头像URL
}
```

## 错误处理

### 标准错误响应格式

```json
{
  "code": <error_code>,
  "message": "<error_message>",
  "data": null
}
```

### 常见错误码

| 状态码 | 错误码 | 描述 | 解决方案 |
|--------|--------|------|----------|
| 400 | 400 | 请求参数错误 | 检查请求参数格式和必需字段 |
| 401 | 401 | 认证失败 | 检查 Authorization 头和 Token 有效性 |
| 403 | 403 | 权限不足 | 确认用户权限或联系管理员 |
| 404 | 404 | 资源不存在 | 检查请求路径是否正确 |
| 500 | 500 | 服务器内部错误 | 查看服务器日志或联系技术支持 |

### 具体错误示例

#### 1. 缺少授权码
```json
{
  "code": 400,
  "message": "缺少授权码"
}
```

#### 2. Token 无效
```json
{
  "code": 401,
  "message": "token无效"
}
```

#### 3. 缺少认证信息
```json
{
  "code": 401,
  "message": "缺少Authorization头"
}
```

#### 4. 登录失败
```json
{
  "code": 500,
  "message": "登录失败: OAuth token exchange failed"
}
```

## 配置说明

### 1. 环境变量配置

支持通过以下环境变量配置系统：

```bash
# Casdoor 配置
CASDOOR_ENDPOINT=http://localhost:8000          # Casdoor 服务地址
CASDOOR_CLIENT_ID=your_client_id                # 应用 Client ID
CASDOOR_CLIENT_SECRET=your_client_secret        # 应用 Client Secret
CASDOOR_JWT_SECRET=your_jwt_secret              # JWT 签名密钥
CASDOOR_ORGANIZATION_NAME=your_org_name         # 组织名称
CASDOOR_APPLICATION_NAME=your_app_name          # 应用名称
```

### 2. 配置文件 (conf/config.yaml)

```yaml
# 服务器配置
server:
  address: ":8080"
  accessLogEnabled: true
  errorLogEnabled: true
  pprofEnabled: true

# 数据库配置
database:
  default:
    link: "pgsql:postgres:123456@tcp(localhost:5432)/contextid?sslmode=disable"
    debug: true

# Casdoor配置
casdoor:
  endpoint: "http://localhost:8000"
  clientId: "ca09f4429d05c4226155"
  clientSecret: "71a26d426be2d0312fcd2e10a072ebaa1ce51ed0"
  jwtSecret: "jwt-secret-key"
  organizationName: "hello"
  applicationName: "context-ID"

# Redis配置（可选）
redis:
  default:
    address: "redis:6379"
    db: 0

# 日志配置
logger:
  level: "all"
  stdout: true
```

### 3. 配置优先级

系统按以下优先级加载配置：

1. **环境变量文件**: `.env` → `config.env` → `config.example.env`
2. **系统环境变量**: 直接从系统环境变量读取
3. **配置文件**: `conf/config.yaml`
4. **默认值**: 内置默认配置

## 部署指南

### 1. Docker 部署

#### 构建镜像
```bash
docker build -t context-id-backend .
```

#### 运行容器
```bash
docker run -d \
  --name context-id-backend \
  -p 8080:8080 \
  -e CASDOOR_ENDPOINT=http://casdoor:8000 \
  -e CASDOOR_CLIENT_ID=your_client_id \
  -e CASDOOR_CLIENT_SECRET=your_client_secret \
  -e CASDOOR_JWT_SECRET=your_jwt_secret \
  context-id-backend
```

### 2. Docker Compose 部署

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f context-id-backend
```

### 3. 生产环境部署

#### 系统要求
- Go 1.23.0+
- PostgreSQL 12+
- Redis 6+ (可选)
- Casdoor 服务

#### 部署步骤

1. **编译应用**
```bash
go build -o context-id-backend main.go
```

2. **配置环境**
```bash
# 创建配置文件
cp config.example.env .env
# 编辑配置文件
vim .env
```

3. **启动服务**
```bash
# 使用 systemd (推荐)
sudo systemctl start context-id-backend
sudo systemctl enable context-id-backend

# 或使用脚本
./scripts/start.sh
```

4. **配置反向代理** (Nginx 示例)
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 测试指南

### 1. 手动测试

#### 获取登录 URL
```bash
curl -X GET "http://localhost:8080/api/v1/auth/login-url"
```

#### 获取用户信息
```bash
curl -X GET "http://localhost:8080/api/v1/user" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 2. 浏览器测试

1. 访问 `http://localhost:8080/login` 进入登录页面
2. 点击登录按钮跳转到 Casdoor
3. 完成登录后自动跳转到回调页面
4. 访问 `http://localhost:8080/dashboard` 测试 API 调用

### 3. API 测试工具

推荐使用以下工具进行 API 测试：
- **Postman**: 图形化 API 测试工具
- **curl**: 命令行测试工具
- **HTTPie**: 更友好的命令行工具

#### Postman 集合示例

```json
{
  "info": {
    "name": "Context-ID Backend API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Get Login URL",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/api/v1/auth/login-url",
          "host": ["{{base_url}}"],
          "path": ["api", "v1", "auth", "login-url"]
        }
      }
    },
    {
      "name": "Get User Info",
      "request": {
        "method": "GET",
        "header": [
          {
            "key": "Authorization",
            "value": "Bearer {{jwt_token}}",
            "type": "text"
          }
        ],
        "url": {
          "raw": "{{base_url}}/api/v1/user",
          "host": ["{{base_url}}"],
          "path": ["api", "v1", "user"]
        }
      }
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    },
    {
      "key": "jwt_token",
      "value": ""
    }
  ]
}
```

### 4. 集成测试

#### 完整登录流程测试

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "1. 获取登录 URL"
LOGIN_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/auth/login-url")
echo $LOGIN_RESPONSE

LOGIN_URL=$(echo $LOGIN_RESPONSE | jq -r '.data.login_url')
echo "登录 URL: $LOGIN_URL"

echo "2. 请在浏览器中访问上述 URL 完成登录"
echo "3. 从回调页面获取 access_token"
read -p "请输入获取到的 access_token: " ACCESS_TOKEN

echo "4. 使用 token 获取用户信息"
USER_RESPONSE=$(curl -s -X GET "$BASE_URL/api/v1/user" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo $USER_RESPONSE
```

## 常见问题

### 1. Casdoor 连接失败

**问题**: 无法连接到 Casdoor 服务

**解决方案**:
- 检查 `CASDOOR_ENDPOINT` 配置是否正确
- 确认 Casdoor 服务是否正常运行
- 检查网络连接和防火墙设置

### 2. JWT Token 验证失败

**问题**: Token 验证返回 401 错误

**解决方案**:
- 检查 `CASDOOR_JWT_SECRET` 配置是否与 Casdoor 一致
- 确认 Token 格式正确 (Bearer + 空格 + token)
- 检查 Token 是否已过期

### 3. 数据库连接失败

**问题**: 数据库操作失败

**解决方案**:
- 检查数据库连接字符串配置
- 确认数据库服务正常运行
- 检查数据库用户权限

### 4. 回调 URL 不匹配

**问题**: OAuth 回调失败

**解决方案**:
- 在 Casdoor 中配置正确的回调 URL
- 确认回调 URL 可以正常访问
- 检查域名和端口配置

## 更新日志

### v1.0.0 (2024-01-XX)
- ✅ 初始版本发布
- ✅ 集成 Casdoor 认证系统
- ✅ 实现 OAuth2 登录流程
- ✅ 添加 JWT Token 验证
- ✅ 支持用户信息获取
- ✅ 提供静态页面服务
- ✅ 添加健康检查端点
- ✅ 支持 Docker 部署

## 技术支持

如有问题或建议，请通过以下方式联系：

- 📧 **邮箱**: support@context-id.com
- 🐛 **Issue**: GitHub Issues
- 📖 **文档**: 项目 Wiki
- 💬 **讨论**: GitHub Discussions

---

**最后更新**: 2024-01-XX  
**文档版本**: v1.0.0  
**API 版本**: v1
