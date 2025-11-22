# API 文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: JWT Bearer Token
- **响应格式**: JSON

## 统一响应格式

### 成功响应
```json
{
  "code": 0,
  "message": "成功",
  "data": { ... }
}
```

### 错误响应
```json
{
  "code": 10002,
  "message": "参数错误"
}
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 10001 | 内部服务错误 |
| 10002 | 参数错误 |
| 10003 | 未授权 |
| 10004 | 禁止访问 |
| 10005 | 资源不存在 |
| 20001 | 用户不存在 |
| 20002 | 用户已存在 |
| 20003 | 密码错误 |
| 20004 | 令牌无效 |
| 30001 | 未绑定教务系统 |
| 30002 | 教务系统登录失败 |
| 30003 | 教务系统参数错误 |
| 30004 | 教务系统请求失败 |
| 30005 | 教务系统数据解析失败 |
| 40001 | 缓存错误 |

---

# 📱 用户接口

## 1. 用户注册

### 请求

```http
POST /api/register
Content-Type: application/json

{
  "name": "张三",
  "email": "zhangsan@example.com",
  "password": "your_password"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "注册成功"
  }
}
```

## 2. 用户登录

### 请求

```http
POST /api/login
Content-Type: application/json

{
  "email": "zhangsan@example.com",
  "password": "your_password"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**说明**：
- 登录成功后会自动记录到日活统计
- Token 有效期 7 天
- 同时会设置 `access_token` Cookie

## 3. 发送验证码

### 请求

```http
POST /api/captcha/send
Content-Type: application/json

{
  "email": "zhangsan@example.com"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "验证码已发送，请查收邮件"
  }
}
```

**说明**：
- 验证码 6 位数字
- 有效期 5 分钟
- 存储在 Redis DB 1

## 4. 重置密码

### 请求

```http
POST /api/reset
Content-Type: application/json

{
  "email": "zhangsan@example.com",
  "code": "123456",
  "password": "new_password"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "密码重置成功"
  }
}
```

---

# 🔐 用户认证接口

**所有以下接口需要在 Header 中携带 JWT Token**

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 5. 获取用户信息

### 请求

```http
GET /api/user/info
Authorization: Bearer {token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "uid": 1,
    "email": "zhangsan@example.com",
    "name": "张三",
    "sid": "202012345678",
    "created_at": "2024-01-01T00:00:00Z",
    "avatar": ""
  }
}
```

**说明**：密码字段已自动隐藏

## 6. 绑定教务系统账号

### 请求

```http
POST /api/user/bind
Authorization: Bearer {token}
Content-Type: application/json

{
  "sid": "202012345678",
  "spwd": "jwc_password"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "绑定成功"
  }
}
```

**说明**：
- 绑定后会清除旧的会话缓存
- 下次查询成绩时自动登录教务系统

## 7. 获取成绩（RESTful 规范）

### 请求

**获取所有成绩**：
```http
GET /api/user/grades
Authorization: Bearer {token}
```

**按学期获取成绩**：
```http
GET /api/user/grades?term=2024-2025-1
Authorization: Bearer {token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "grades": [
      {
        "serialNo": "1",
        "Year": "2024-2025-1",
        "Code": "CS101",
        "subject": "数据结构",
        "score": "92",
        "credit": 4.0,
        "gpa": 4.2,
        "Status": 0,
        "property": "必修"
      }
    ],
    "gpa": {
      "averageGPA": 3.85,      // 平均绩点
      "averageScore": 88.5,    // 平均分
      "basicScore": 87.2       // 学业基本分
    }
  }
}
```

**说明**：
- 不传 `term` 参数：查询所有成绩
- 传 `term` 参数：查询指定学期成绩
- 参数通过 query params 传递（符合 RESTful 规范）

## 8. 获取等级考试成绩

### 请求

```http
GET /api/user/grades/level
Authorization: Bearer {token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "no": "1",
      "CourseName": "大学英语四级",
      "LevelGrade": "560",
      "Time": "2024-06"
    }
  ]
}
```

## 9. 获取成绩分析（最近三个学期）

### 请求

```http
GET /api/user/grades/analysis
Authorization: Bearer {token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "current_term": "2024-2025-2",
    "terms_data": [
      {
        "term": "2024-2025-2",
        "gpa": {
          "averageGPA": 3.92,      // 平均绩点
          "averageScore": 89.2,    // 平均分
          "basicScore": 88.5       // 学业基本分
        }
      },
      {
        "term": "2024-2025-1",
        "gpa": {
          "averageGPA": 3.85,
          "averageScore": 88.0,
          "basicScore": 87.0
        }
      },
      {
        "term": "2023-2024-2",
        "gpa": {
          "averageGPA": 3.75,
          "averageScore": 86.5,
          "basicScore": 85.8
        }
      }
    ],
    "overall_gpa": {
      "averageGPA": 3.85,
      "averageScore": 88.5,
      "basicScore": 87.2
    },
    "trend_analysis": {
      "gpa_trend": "上升",
      "score_trend": "上升",
      "best_term": "2024-2025-2",
      "best_term_gpa": 3.92,
      "worst_term": "2023-2024-2",
      "worst_term_gpa": 3.75
    }
  }
}
```

**说明**：
- ⚠️ **不返回具体科目成绩**，只返回统计数据
- 自动从 Redis 读取当前学期配置
- 自动计算前两个学期
- 提供趋势分析和学期对比
- 保护学生隐私

## 10. 获取课程表（RESTful 规范）

### 请求

```http
GET /api/user/courses?week=5&term=2024-2025-1
Authorization: Bearer {token}
```

**参数说明**：
- `week`: 周次（1-20）
- `term`: 学期（格式：2024-2025-1）

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "weekno": 5,
    "starttime": "",
    "endtime": "",
    "days": [
      {
        "weekday": 1,
        "courses": [
          {
            "name": "数据结构",
            "teacher": "张老师",
            "classroom": "A1-203",
            "weekday": 1,
            "start_period": 1,
            "end_period": 2
          }
        ]
      }
    ]
  }
}
```

## 11. 获取考试安排（RESTful 规范）

### 请求

```http
GET /api/user/exams?term=2024-2025-1
Authorization: Bearer {token}
```

**参数说明**：
- `term`: 学期（格式：2024-2025-1）

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "serial_no": "1",
      "class_no": "CS101",
      "class_name": "数据结构",
      "time": "2024-01-15 09:00-11:00",
      "place": "A1-101",
      "execution": "正考"
    }
  ]
}
```

---

# 👑 管理员接口

## 12. 管理员登录

### 请求

```http
POST /api/admin/login
Content-Type: application/json

{
  "email": "admin@spider-go.com",
  "password": "123456"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**说明**：
- 管理员 Token 有效期 24 小时
- Token 包含 `is_admin: true` 标识

## 13. 获取管理员信息

### 请求

```http
GET /api/admin/info
Authorization: Bearer {admin_token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "uid": 1,
    "email": "admin@spider-go.com",
    "name": "Haruka",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

## 14. 修改管理员密码

### 请求

```http
POST /api/admin/reset
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "email": "admin@spider-go.com",
  "password": "new_password"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "重置成功"
  }
}
```

---

# 📝 通知管理接口

## 15. 创建通知（管理员）

### 请求

```http
POST /api/admin/notices
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "content": "系统将于今晚 22:00 进行维护",
  "notice_type": "系统通知",
  "is_show": true,
  "is_top": true,
  "is_html": false
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "创建成功"
  }
}
```

## 16. 更新通知（管理员）

### 请求

```http
PUT /api/admin/notices/1
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "content": "维护已完成",
  "notice_type": "系统通知",
  "is_show": true,
  "is_top": false,
  "is_html": false
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "更新成功"
  }
}
```

## 17. 删除通知（管理员）

### 请求

```http
DELETE /api/admin/notices/1
Authorization: Bearer {admin_token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "删除成功"
  }
}
```

## 18. 获取所有通知（管理员）

### 请求

```http
GET /api/admin/notices
Authorization: Bearer {admin_token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "nid": 1,
      "content": "系统维护通知",
      "notice_type": "系统通知",
      "is_show": true,
      "create_time": "2024-01-01T00:00:00Z",
      "update_time": "2024-01-01T00:00:00Z",
      "is_top": true,
      "is_html": false
    }
  ]
}
```

## 19. 获取可见通知（公开）

### 请求

```http
GET /api/notices
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": [
    {
      "nid": 1,
      "content": "系统维护通知",
      "notice_type": "系统通知",
      "is_show": true,
      "create_time": "2024-01-01T00:00:00Z",
      "update_time": "2024-01-01T00:00:00Z",
      "is_top": true,
      "is_html": false
    }
  ]
}
```

**说明**：
- 只返回 `is_show = true` 的通知
- 按置顶和创建时间排序

---

# 📊 统计接口（管理员）

## 20. 获取日活统计

### 请求

```http
GET /api/admin/statistics/dau
Authorization: Bearer {admin_token}

# 获取今天的日活
GET /api/admin/statistics/dau

# 获取指定日期的日活
GET /api/admin/statistics/dau?date=2024-01-20
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "date": "2024-01-20",
    "dau": 1523
  }
}
```

## 21. 获取日活范围统计

### 请求

```http
GET /api/admin/statistics/dau/range?start_date=2024-01-01&end_date=2024-01-07
Authorization: Bearer {admin_token}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "2024-01-01": 1420,
    "2024-01-02": 1535,
    "2024-01-03": 1601,
    "2024-01-04": 1589,
    "2024-01-05": 1623,
    "2024-01-06": 1456,
    "2024-01-07": 1502
  }
}
```

**说明**：
- 最多查询 90 天范围
- 数据保留 30 天
- 自动过期清理

---

# ⚙️ 系统配置接口

## 22. 获取当前学期（公开）

### 请求

```http
GET /api/config/term
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "current_term": "2024-2025-1"
  }
}
```

## 23. 设置当前学期（管理员）

### 请求

```http
POST /api/admin/config/term
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "term": "2024-2025-2"
}
```

### 响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message": "设置成功",
    "current_term": "2024-2025-2"
  }
}
```

**学期格式说明**：
- 格式：`YYYY-YYYY-[1|2]`
- `2024-2025-1`：2024-2025学年第一学期
- `2024-2025-2`：2024-2025学年第二学期
- 两个年份必须连续（endYear = startYear + 1）

---

# 📋 完整接口列表

## 公开接口（无需认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/login` | 用户登录 |
| POST | `/api/register` | 用户注册 |
| POST | `/api/reset` | 重置密码 |
| POST | `/api/captcha/send` | 发送验证码 |
| GET | `/api/notices` | 获取可见通知 |
| GET | `/api/config/term` | 获取当前学期 |
| POST | `/api/admin/login` | 管理员登录 |

## 用户接口（需要用户 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/user/bind` | 绑定教务系统 |
| GET | `/api/user/info` | 获取用户信息 |
| GET | `/api/user/grades` | 获取成绩（query: term 可选） |
| GET | `/api/user/grades/level` | 获取等级考试成绩 |
| GET | `/api/user/grades/analysis` | 获取成绩分析（仅统计） |
| GET | `/api/user/courses` | 获取课程表（query: week, term） |
| GET | `/api/user/exams` | 获取考试安排（query: term） |

## 管理员接口（需要管理员 JWT）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/info` | 获取管理员信息 |
| POST | `/api/admin/reset` | 修改管理员密码 |
| POST | `/api/admin/notices` | 创建通知 |
| PUT | `/api/admin/notices/:nid` | 更新通知 |
| DELETE | `/api/admin/notices/:nid` | 删除通知 |
| GET | `/api/admin/notices` | 获取所有通知 |
| GET | `/api/admin/statistics/dau` | 获取日活统计 |
| GET | `/api/admin/statistics/dau/range` | 获取日活范围统计 |
| POST | `/api/admin/config/term` | 设置当前学期 |

---

# 🔄 典型业务流程

## 用户注册登录流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant S as 服务器
    participant R as Redis
    
    U->>S: 1. POST /api/register
    S->>S: 2. 密码加密存储
    S-->>U: 3. 注册成功
    
    U->>S: 4. POST /api/login
    S->>R: 5. 记录日活（DAU）
    S->>S: 6. 生成 JWT Token
    S-->>U: 7. 返回 Token
    
    U->>S: 8. GET /api/user/info (带 Token)
    S->>R: 9. 自动记录日活
    S-->>U: 10. 返回用户信息
```

## 查询成绩流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant S as 服务器
    participant R as Redis
    participant J as 教务系统
    
    U->>S: 1. POST /api/user/bind (绑定账号)
    S-->>U: 2. 绑定成功
    
    U->>S: 3. GET /api/user/grades/all
    S->>R: 4. 检查会话缓存
    alt 缓存不存在
        S->>J: 5. 登录教务系统
        J-->>S: 6. 返回 Cookies
        S->>R: 7. 缓存 Cookies (1小时)
    end
    S->>J: 8. 带 Cookies 查询成绩
    J-->>S: 9. 返回 HTML
    S->>S: 10. 解析 HTML
    S-->>U: 11. 返回成绩数据
```

---

# 🧪 测试示例

## 使用 curl 测试

### 1. 用户注册
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "password": "123456"
  }'
```

### 2. 用户登录
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "zhangsan@example.com",
    "password": "123456"
  }'
```

### 3. 获取成绩（需要 Token）
```bash
# 获取所有成绩
curl -X GET http://localhost:8080/api/user/grades \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 按学期获取成绩（RESTful 规范 - 使用 query params）
curl -X GET "http://localhost:8080/api/user/grades?term=2024-2025-1" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 4. 获取课程表（需要 Token）
```bash
# RESTful 规范 - 使用 query params
curl -X GET "http://localhost:8080/api/user/courses?week=5&term=2024-2025-1" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 5. 获取考试安排（需要 Token）
```bash
# RESTful 规范 - 使用 query params
curl -X GET "http://localhost:8080/api/user/exams?term=2024-2025-1" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 6. 管理员登录
```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@spider-go.com",
    "password": "123456"
  }'
```

### 7. 设置当前学期（管理员）
```bash
curl -X POST http://localhost:8080/api/admin/config/term \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "term": "2024-2025-1"
  }'
```

### 8. 查看日活统计（管理员）
```bash
curl -X GET http://localhost:8080/api/admin/statistics/dau \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE"
```

---

# 🔒 安全说明

## 认证机制

1. **用户认证**: JWT Token（7天有效期）
2. **管理员认证**: JWT Token（24小时有效期）+ `is_admin` 标识
3. **会话缓存**: Redis（1小时自动过期）

## 密码安全

- 使用 bcrypt 加密存储
- Cost: 默认（10）
- 教务系统密码存储在数据库（需加密传输）

## 验证码安全

- 6位数字验证码
- 5分钟有效期
- 使用后立即删除
- 使用 Lua 脚本保证原子性

## CORS 配置

生产环境建议配置 CORS 中间件：

```go
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
    AllowOrigins: []string{"https://your-frontend.com"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Authorization", "Content-Type"},
}))
```

---

# 📈 性能优化

## 缓存策略

### 会话缓存（1小时）
- 减少教务系统登录次数
- 避免频繁加密登录

### 日活本地缓存
- 内存缓存已记录用户
- 减少 Redis 写入
- 每天 0 点自动清空

### 异步记录
- 日活统计使用 goroutine 异步记录
- 不影响接口响应速度

## 数据库优化

- 使用索引：`email`（唯一）
- 连接池配置（可在 `db.go` 中调整）

---

# 🐛 常见问题

## Q1: 首次启动报错 "当前学期未设置"

**解决**：使用管理员账号登录后设置当前学期

```bash
# 1. 管理员登录
POST /api/admin/login

# 2. 设置当前学期
POST /api/admin/config/term
{
  "term": "2024-2025-1"
}
```

## Q2: 查询成绩失败 "未绑定教务系统"

**解决**：先绑定教务系统账号

```bash
POST /api/user/bind
{
  "sid": "你的学号",
  "spwd": "教务系统密码"
}
```

## Q3: 发送验证码失败

**解决**：检查 SMTP 配置
- 确认邮箱配置正确
- QQ 邮箱需要使用 SMTP 授权码，不是邮箱密码
- 端口 465 需要 SSL，587 需要 TLS

## Q4: Redis 连接失败

**解决**：
- 检查 Redis 是否启动
- 检查配置文件中的地址和密码
- 检查防火墙设置

---

# 📚 开发文档

## 添加新接口

1. **定义 DTO**（`internal/dto/`）
2. **实现 Service**（`internal/service/`）
3. **实现 Controller**（`internal/controller/`）
4. **注册路由**（`internal/api/routes.go`）
5. **更新 Container**（`internal/app/container.go`）

## 添加新的缓存

1. **定义接口**（`internal/cache/xxx_cache.go`）
2. **在 Container 中注册**
3. **在 Service 中使用**

## 错误处理

使用统一的错误处理：

```go
return common.NewAppError(common.CodeInvalidParams, "自定义错误信息")
```

---

# 📞 技术支持

如有问题，请通过以下方式联系：

- 提交 Issue
- 发送邮件

**项目维护者**: Haruka ❤️





