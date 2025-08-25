# Context-ID Backend

基于GoFrame框架的Context-ID记忆系统后端，集成Casdoor认证服务。

## 功能特性

- 🚀 基于GoFrame v2框架
- 🔐 集成Casdoor统一认证
- 🐘 使用PostgreSQL数据库
- 🐳 Docker容器化部署
- 📊 Redis缓存支持
- 🔒 JWT令牌认证
- 🌐 RESTful API设计

## 技术栈

- **后端框架**: GoFrame v2
- **认证服务**: Casdoor
- **数据库**: PostgreSQL 15
- **缓存**: Redis 7
- **容器化**: Docker & Docker Compose

## 快速开始

### 环境要求

- Docker >= 20.10
- Docker Compose >= 2.2
- Go >= 1.21 (仅开发环境需要)

### 部署步骤

1. **克隆项目**
   ```bash
   git clone <your-repo-url>
   cd Context-ID-backend
   ```

2. **启动所有服务**
   ```bash
   docker-compose up -d
   ```

3. **检查服务状态**
   ```bash
   docker-compose ps
   ```

4. **查看日志**
   ```bash
   # 查看所有服务日志
   docker-compose logs -f
   
   # 查看特定服务日志
   docker-compose logs -f backend
   docker-compose logs -f casdoor
   ```

### 服务端口

| 服务 | 端口 | 描述 |
|------|------|------|
| GoFrame后端 | 8080 | API服务 |
| Casdoor | 8000 | 认证服务 |
| PostgreSQL | 5432 | 主数据库 |
| Casdoor DB | 5433 | Casdoor数据库 |
| Redis | 6379 | 缓存服务 |

### 访问地址

- **API文档**: http://localhost:8080/health
- **Casdoor管理后台**: http://localhost:8000
  - 默认账号: `built-in/admin`
  - 默认密码: `123`

## API接口

### 认证相关

1. **获取登录URL**
   ```http
   GET /api/v1/auth/url?redirect_uri=http://localhost:3000/callback&state=random_state
   ```

2. **用户登录**
   ```http
   POST /api/v1/auth/login
   Content-Type: application/json
   
   {
     "code": "authorization_code",
     "state": "random_state"
   }
   ```

3. **获取用户信息**
   ```http
   GET /api/v1/auth/user
   Authorization: Bearer <token>
   ```

4. **用户登出**
   ```http
   POST /api/v1/auth/logout
   Authorization: Bearer <token>
   ```

## 配置说明

### GoFrame配置 (conf/config.yaml)

```yaml
server:
  address: ":8080"

database:
  default:
    link: "pgsql:user=postgres password=123456 host=postgres port=5432 dbname=contextid sslmode=disable"

casdoor:
  endpoint: "http://casdoor:8000"
  clientId: "context-id-app"
  clientSecret: "context-id-secret"
  jwtSecret: "jwt-secret-key"
  organizationName: "built-in"
  applicationName: "context-id-app"
```

### Casdoor配置 (casdoor/app.conf)

主要配置项：
- `dataSourceName`: PostgreSQL连接字符串
- `authState`: 认证状态标识
- `initDataFile`: 初始数据文件路径

## 开发指南

### 本地开发

1. **安装依赖**
   ```bash
   go mod tidy
   ```

2. **启动数据库服务**
   ```bash
   docker-compose up -d postgres casdoor-db redis casdoor
   ```

3. **运行后端服务**
   ```bash
   go run main.go
   ```

### 项目结构

```
Context-ID-backend/
├── main.go                 # 应用入口
├── conf/                   # 配置文件
├── internal/              # 内部代码
│   ├── controller/        # 控制器
│   ├── service/          # 业务逻辑
│   ├── middleware/       # 中间件
│   ├── model/           # 数据模型
│   └── dao/             # 数据访问
├── sql/                  # 数据库脚本
├── casdoor/             # Casdoor配置
├── docker-compose.yml   # Docker编排
└── Dockerfile          # Docker构建文件
```

## 常见问题

### 1. 服务启动失败

检查端口是否被占用：
```bash
netstat -tlnp | grep :8080
netstat -tlnp | grep :8000
```

### 2. 数据库连接失败

确保PostgreSQL服务正常运行：
```bash
docker-compose logs postgres
```

### 3. Casdoor认证失败

1. 检查Casdoor服务状态
2. 确认应用配置正确
3. 检查重定向URI配置

### 4. 重置数据

```bash
# 停止所有服务
docker-compose down

# 删除数据卷（注意：这会删除所有数据）
docker-compose down -v

# 重新启动
docker-compose up -d
```

## 生产部署建议

1. **安全配置**
   - 修改默认密码
   - 使用HTTPS
   - 配置防火墙
   - 定期备份数据

2. **性能优化**
   - 配置数据库连接池
   - 启用Redis缓存
   - 配置负载均衡

3. **监控告警**
   - 配置健康检查
   - 设置日志收集
   - 监控资源使用

## 许可证

MIT License