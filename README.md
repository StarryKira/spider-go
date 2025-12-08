# Spider-Go 🕷️

一个基于 Golang 的教务系统数据爬虫和管理平台，提供成绩查询、课程表查询、考试安排查询等功能。

## ✨ 功能特性

### 📚 用户功能
- 用户注册/登录（JWT 认证）
- 绑定教务系统账号
- 查询全部成绩/按学期查询成绩
- 查询等级考试成绩（四六级等）
- 查询课程表（按周）
- 查询考试安排
- 📊 **成绩分析**：最近三个学期的成绩趋势分析
- 查看系统通知

### 👑 管理员功能
- 管理员登录（独立认证）
- 📝 通知管理（增删改查）
- 📈 日活统计（DAU）
- 🗓️ 学期配置管理
- 修改管理员密码

### 🔐 安全功能
- 邮箱验证码发送
- 密码加密存储（bcrypt）
- JWT 令牌认证
- 教务系统会话缓存（1小时）

### 📊 统计功能
- 日活统计（DAU）
- 自动记录用户活跃
- 数据保留 30 天
- 支持日期范围查询

## 🏗️ 项目架构

### 分层设计

```
┌─────────────────────────────────────────┐
│           API Layer (Gin)               │  路由层
├─────────────────────────────────────────┤
│         Controller Layer                │  控制器层
├─────────────────────────────────────────┤
│          Service Layer                  │  业务逻辑层
├──────────────┬──────────────────────────┤
│  Repository  │   Cache   │   Crawler    │  数据访问层
├──────────────┴──────────────────────────┤
│    MySQL    │   Redis   │   HTTP        │  基础设施层
└─────────────────────────────────────────┘
```

### 目录结构

```
spider-go/
├── config/                  # 配置文件
│   └── config.yaml
├── internal/
│   ├── api/                 # 路由配置
│   ├── app/                 # 应用初始化
│   │   ├── config.go        # 配置加载
│   │   ├── db.go            # 数据库初始化
│   │   ├── redis_client.go  # Redis 初始化
│   │   └── container.go     # 依赖注入容器
│   ├── cache/               # 缓存层
│   │   ├── session_cache.go # 会话缓存
│   │   ├── captcha_cache.go # 验证码缓存
│   │   ├── dau_cache.go     # 日活统计缓存
│   │   └── config_cache.go  # 系统配置缓存
│   ├── common/              # 公共模块
│   │   ├── errors.go        # 错误码定义
│   │   └── response.go      # 统一响应
│   ├── controller/          # 控制器层
│   ├── dto/                 # 数据传输对象
│   ├── middleware/          # 中间件
│   │   ├── jwtauth.go       # JWT 认证
│   │   └── admin_auth.go    # 管理员认证
│   ├── model/               # 数据模型
│   ├── repository/          # 数据仓储
│   ├── service/             # 业务服务
│   └── utils/               # 工具函数
└── main.go                  # 入口文件
```

## 🚀 快速开始

### 环境要求

- Go 1.25+
- MySQL 5.7+
- Redis 5.0+

### 安装步骤

1. **克隆项目**
```bash
git clone <your-repo-url>
cd spider-go
```

2. **安装依赖**
```bash
go mod download
```

3. **配置文件**

项目支持多环境配置，已内置开发和生产环境配置文件：

- `config/config.dev.yaml` - 开发环境配置（本地数据库）
- `config/config.prod.yaml` - 生产环境配置（远程数据库）

**开发环境配置示例** (`config.dev.yaml`)：
```yaml
app:
  port: 8080
  env: dev

database:
  host: 127.0.0.1
  port: 3306
  user: root
  pass: dev_password
  name: spider_go_dev

redis:
  session:
    host: 127.0.0.1:6379
    pass: ""
    db: 0
  captcha:
    host: 127.0.0.1:6379
    pass: ""
    db: 1

jwt:
  secret: "dev_secret_key_change_in_production"
  issuer: "spider-go-dev"
```

**生产环境配置示例** (`config.prod.yaml`)：
```yaml
app:
  port: 8080
  env: production

database:
  host: your_production_host
  port: 3306
  user: spider-go_prod
  pass: your_production_password
  name: spider-go_prod

redis:
  session:
    host: your_redis_host:6379
    pass: your_redis_password
    db: 0
  captcha:
    host: your_redis_host:6379
    pass: your_redis_password
    db: 1

jwt:
  secret: "CHANGE_THIS_TO_A_SECURE_RANDOM_STRING_IN_PRODUCTION"
  issuer: "spider-go"
```

4. **运行项目**

默认使用开发环境（`config.dev.yaml`）：
```bash
go run main.go
```

使用环境变量指定环境：
```bash
# Windows (PowerShell)
$env:GO_ENV="production"; go run main.go

# Linux/Mac
export GO_ENV=production
go run main.go
```

使用命令行参数指定环境（优先级最高）：
```bash
go run main.go -env=production
# 或
go run main.go -env=dev
```

编译后运行：
```bash
go build -o spider-go.exe
./spider-go.exe -env=production
```

5. **访问服务**
```
服务地址: http://localhost:8080
```

## 🔑 默认账号

### 管理员账号（首次启动自动创建）
- **邮箱**: `admin@spider-go.com`
- **密码**: `123456`
- **权限**: 全部管理权限

⚠️ **首次登录后请立即修改密码！**

## 🎯 核心功能说明

### 1. 依赖注入容器

项目使用依赖注入容器统一管理所有依赖，避免全局变量：

```go
container, err := app.NewContainer("./config")
// 自动初始化：
// - 配置加载
// - 数据库连接
// - Redis 连接
// - 所有 Service 和 Controller
// - 默认管理员
```

### 2. 缓存策略

#### Redis DB 0（会话 + 日活 + 配置）
- 用户登录会话（1小时过期）
- 日活统计数据（30天过期）
- 系统配置（永久）

#### Redis DB 1（验证码）
- 邮箱验证码（5分钟过期）

### 3. 日活统计

- ✅ 用户登录自动记录
- ✅ 使用 JWT 访问任意接口自动记录
- ✅ 同一用户一天只计数一次（Redis Set 自动去重）
- ✅ 本地缓存优化，减少 Redis 压力
- ✅ 数据保留 30 天

### 4. 成绩分析

- 📊 显示最近三个学期的成绩
- 📈 GPA 趋势分析（上升/下降/稳定）
- 🏆 最好/最差学期统计
- 📉 各学期对比

## 📖 API 文档

详见 [API_DOCUMENTATION.md](./API_DOCUMENTATION.md)

## 🛠️ 技术栈

- **Web 框架**: Gin
- **数据库 ORM**: GORM
- **缓存**: Redis (go-redis/v9)
- **JWT**: golang-jwt/jwt
- **HTML 解析**: goquery
- **邮件发送**: gomail.v2
- **配置管理**: Viper
- **密码加密**: bcrypt

## 📁 数据库表

### users（用户表）
| 字段 | 类型 | 说明 |
|------|------|------|
| uid | int | 用户ID（主键） |
| email | string | 邮箱（唯一） |
| name | string | 用户名 |
| password | string | 密码（加密） |
| sid | string | 教务系统学号 |
| spwd | string | 教务系统密码 |
| created_at | datetime | 创建时间 |
| avatar | string | 头像链接 |

### administrators（管理员表）
| 字段 | 类型 | 说明 |
|------|------|------|
| uid | int | 管理员ID（主键） |
| email | string | 邮箱（唯一） |
| name | string | 管理员名 |
| password | string | 密码（加密） |
| created_at | datetime | 创建时间 |
| avatar | string | 头像链接 |

### notices（通知表）
| 字段 | 类型 | 说明 |
|------|------|------|
| nid | int | 通知ID（主键） |
| content | text | 通知内容 |
| notice_type | string | 通知类型 |
| is_show | bool | 是否显示 |
| create_time | datetime | 创建时间 |
| update_time | datetime | 更新时间 |
| is_top | bool | 是否置顶 |
| is_html | bool | 是否HTML格式 |

## 🔧 开发建议

### 环境配置管理

项目通过环境区分开发和生产配置：

**环境指定优先级**：
1. 命令行参数 `-env=production` （最高优先级）
2. 环境变量 `GO_ENV=production`
3. 默认 `dev` 环境

**生产部署示例**：
```bash
# 方式1: 使用命令行参数（推荐）
./spider-go -env=production

# 方式2: 使用环境变量
export GO_ENV=production
./spider-go

# Docker 部署
docker run -e GO_ENV=production your-image
```

**配置文件选择规则**：
- `GO_ENV=dev` → 加载 `config/config.dev.yaml`
- `GO_ENV=production` → 加载 `config/config.prod.yaml`
- 未设置或其他值 → 默认加载 `config/config.dev.yaml`

### 日志

建议添加结构化日志（如 zap 或 logrus）

### 监控

建议添加：
- Prometheus 指标
- 链路追踪（Jaeger）
- 性能监控

## 📝 待办事项

- [ ] 添加单元测试
- [ ] 添加集成测试
- [ ] 添加 API 限流
- [ ] 添加数据库连接池配置
- [ ] 添加优雅关闭机制
- [ ] 添加健康检查接口


## 👥 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题，请通过 Issue 联系。


## TODO
根据日活刷新缓存


