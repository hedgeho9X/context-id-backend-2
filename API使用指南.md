# Context-ID Backend API 使用指南

## 概述

本项目已成功整合了 `casdoor-tutorial/main.go` 中的成功方法，将其转换为完整的 API 服务。现在 Context-ID-backend 提供了一套完整的 Casdoor 认证 API。

## 主要改进

### 1. 服务层改进 (`internal/service/casdoor.go`)
- ✅ 整合了 tutorial 中的成功方法
- ✅ 添加了 `HandleCallback` 方法，直接使用 `casdoorsdk.GetOAuthToken` 和 `casdoorsdk.ParseJwtToken`
- ✅ 添加了 `ValidateToken` 方法，用于验证 JWT token
- ✅ 添加了 `GetLoginURL` 方法，简化登录URL获取
- ✅ 新增了 `UserInfo` 结构体，与 tutorial 保持一致

### 2. 控制器层改进 (`internal/controller/auth.go`)
- ✅ 更新了 `GetLoginURL` 方法，使用新的服务方法
- ✅ 更新了 `Callback` 方法，使用 tutorial 中的成功实现
- ✅ 更新了 `Login` 方法，直接使用 `HandleCallback`
- ✅ 新增了 `GetCurrentUser` 方法，支持 token 验证

### 3. 路由改进 (`internal/controller/register.go`)
- ✅ 新增了 `/api/v1/login-url` 路由
- ✅ 新增了 `/api/v1/user` 路由（支持 token 验证）

### 4. 主服务器改进 (`main.go`)
- ✅ 添加了静态文件服务
- ✅ 改进了根路径处理，支持浏览器访问
- ✅ 新增了 HTML 页面路由：`/login`, `/callback`, `/dashboard`

### 5. HTML模板更新
- ✅ 更新了 `index.html` 中的 API 调用路径
- ✅ 更新了 `dashboard.html` 中的 API 调用和响应处理
- ✅ 复制模板文件到 `static/` 目录

## API 端点

### 认证相关 API

| 方法 | 端点 | 描述 | 认证要求 |
|------|------|------|----------|
| GET | `/api/v1/login-url` | 获取 Casdoor 登录 URL | 无 |
| GET | `/api/v1/auth/callback` | Casdoor 回调处理（GET） | 无 |
| POST | `/api/v1/auth/callback` | Casdoor 回调处理（POST） | 无 |
| POST | `/api/v1/auth/login` | 用户登录 | 无 |
| GET | `/api/v1/user` | 获取当前用户信息（通过 token） | Bearer Token |
| GET | `/api/v1/auth/user` | 获取用户信息（中间件认证） | Bearer Token |
| POST | `/api/v1/auth/logout` | 用户登出 | Bearer Token |

### 页面路由

| 路径 | 描述 |
|------|------|
| `/` | 主页（API 信息或登录页面） |
| `/login` | 登录页面 |
| `/callback` | 回调页面 |
| `/dashboard` | 用户仪表板 |
| `/health` | 健康检查 |

## 使用示例

### 1. 获取登录 URL

```bash
curl -X GET http://localhost:8080/api/v1/login-url
```

响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "login_url": "http://localhost:8000/login/oauth/authorize?client_id=ca09f4429d05c4226155&redirect_uri=http://localhost:8080/api/v1/auth/callback&response_type=code&scope=read&state=casdoor"
  }
}
```

### 2. 登录处理

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"code": "授权码", "state": "状态值"}'
```

响应：
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "JWT_TOKEN_HERE",
    "user": {
      "username": "user123",
      "displayName": "用户显示名",
      "email": "user@example.com",
      "phone": "13800000000",
      "avatar": "http://example.com/avatar.jpg"
    }
  }
}
```

### 3. 获取用户信息

```bash
curl -X GET http://localhost:8080/api/v1/user \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user": {
      "username": "user123",
      "displayName": "用户显示名",
      "email": "user@example.com",
      "phone": "13800000000",
      "avatar": "http://example.com/avatar.jpg"
    }
  }
}
```

## 启动服务

### 1. 配置环境变量

参考 `casdoor-tutorial/main.go` 的实现方式，服务现在支持使用 `.env` 文件配置：

```bash
# 复制配置模板
cp config.example.env .env

# 编辑 .env 文件，填入您的 Casdoor 配置
# CASDOOR_ENDPOINT=http://localhost:8000
# CASDOOR_CLIENT_ID=your_client_id
# CASDOOR_CLIENT_SECRET=your_client_secret
# CASDOOR_JWT_SECRET=your_jwt_secret
# CASDOOR_ORGANIZATION_NAME=your_org_name
# CASDOOR_APPLICATION_NAME=your_app_name
```

### 2. 安装依赖并启动

```bash
cd /Users/jerry/Documents/Code/project/Context-ID/Context-ID-backend

# 安装依赖
go mod tidy

# 启动服务
go run main.go
```

### 3. 配置加载优先级

服务按以下优先级加载配置：
1. **环境变量文件**：`.env` → `config.env` → `config.example.env`
2. **系统环境变量**：直接从系统环境变量读取
3. **配置文件**：`conf/config.yaml`（作为备用）
4. **默认值**：内置默认配置

4. 访问 `http://localhost:8080` 查看 API 文档
5. 访问 `http://localhost:8080/login` 开始登录流程

## 测试流程

1. **浏览器测试**：
   - 访问 `http://localhost:8080/login`
   - 点击登录按钮，跳转到 Casdoor
   - 完成登录后自动跳转回 `/callback` 页面
   - 访问 `/dashboard` 测试 API 调用

2. **API 测试**：
   - 使用 curl 或 Postman 测试各个 API 端点
   - 测试 token 验证功能

## 注意事项

- 确保 Casdoor 服务正常运行
- 检查防火墙和端口配置
- 验证 JWT secret 配置正确
- 确保回调 URL 在 Casdoor 中正确配置

## 技术栈

- **后端框架**：GoFrame v2
- **认证系统**：Casdoor
- **数据库**：PostgreSQL
- **前端**：原生 HTML/CSS/JavaScript
