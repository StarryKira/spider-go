w# Spider-Go API 文档

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **开发环境端口**: `8080`
- **Content-Type**: `application/json`
- **字符编码**: `UTF-8`

## 认证说明

### JWT 认证
需要认证的接口需要在 HTTP 请求头中携带 JWT Token：

```
Authorization: Bearer <token>
```

### 用户 Token
用户登录后获得，用于访问 `/api/user/*` 下的认证接口。

### 管理员 Token
管理员登录后获得，用于访问 `/api/admin/*` 下的认证接口。

---

## 1. 用户模块

### 1.1 用户注册

**接口地址**: `POST /api/user/register`

**请求参数**:
```json
{
  "email": "student@example.com",
  "password": "password123",
  "captcha": "123456"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码（至少6位） |
| captcha | string | 是 | 邮箱验证码 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uid": 1,
    "email": "student@example.com",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 1.2 用户登录

**接口地址**: `POST /api/user/login`

**请求参数**:
```json
{
  "email": "student@example.com",
  "password": "password123"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uid": 1,
    "email": "student@example.com",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

---

### 1.3 重置密码

**接口地址**: `POST /api/user/reset-password`

**请求参数**:
```json
{
  "email": "student@example.com",
  "new_password": "newpassword123",
  "captcha": "123456"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| new_password | string | 是 | 新密码（至少6位） |
| captcha | string | 是 | 邮箱验证码 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 1.4 获取用户信息

**接口地址**: `GET /api/user/info`

**认证**: 需要用户 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uid": 1,
    "email": "student@example.com",
    "sid": "202012345678",
    "is_bind": true
  }
}
```

---

### 1.5 绑定教务系统

**接口地址**: `POST /api/user/bind`

**认证**: 需要用户 Token

**请求参数**:
```json
{
  "sid": "202012345678",
  "spwd": "jwcpassword"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sid | string | 是 | 教务系统学号 |
| spwd | string | 是 | 教务系统密码 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**注意**:
- 绑定时会验证教务系统账号密码是否正确
- 连续3次失败会被锁定30分钟

---

### 1.6 检查绑定状态

**接口地址**: `GET /api/user/is-bind`

**认证**: 需要用户 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "is_bind": true
  }
}
```

---

## 2. 验证码模块

### 2.1 发送邮箱验证码

**接口地址**: `POST /api/captcha/send`

**请求参数**:
```json
{
  "email": "student@example.com"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 接收验证码的邮箱 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**注意**:
- 验证码有效期为 5 分钟
- 同一邮箱 1 分钟内只能发送一次

---

## 3. 成绩模块

### 3.1 获取成绩

**接口地址**: `GET /api/user/grades`

**认证**: 需要用户 Token

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 否 | 学期（格式：2024-2025-1），不传则返回所有学期 |

**请求示例**:
```
GET /api/user/grades?term=2024-2025-1
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "course_name": "高等数学",
      "course_code": "MATH101",
      "credit": 4.0,
      "score": "85",
      "grade_point": 3.5,
      "term": "2024-2025-1",
      "exam_type": "期末考试",
      "course_nature": "必修"
    }
  ]
}
```

---

### 3.2 获取等级考试成绩

**接口地址**: `GET /api/user/grades/level`

**认证**: 需要用户 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "exam_name": "大学英语四级",
      "score": "520",
      "exam_date": "2024-06-15",
      "result": "通过"
    }
  ]
}
```

---

## 4. 课程模块

### 4.1 获取课程表

**接口地址**: `GET /api/user/courses`

**认证**: 需要用户 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "course_name": "高等数学",
      "teacher": "张三",
      "week": "1-16周",
      "weekday": 1,
      "section": "1-2节",
      "location": "教学楼A101",
      "week_nums": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]
    }
  ]
}
```

**字段说明**:
- `weekday`: 星期几（1-7，周一到周日）
- `section`: 第几节课
- `week_nums`: 上课周次数组

---

## 5. 考试模块

### 5.1 获取考试安排

**接口地址**: `GET /api/user/exams`

**认证**: 需要用户 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "course_name": "高等数学",
      "exam_time": "2025-01-10 09:00-11:00",
      "location": "教学楼A101",
      "seat_number": "15",
      "exam_type": "期末考试"
    }
  ]
}
```

---

## 6. 管理员模块

### 6.1 管理员登录

**接口地址**: `POST /api/admin/login`

**请求参数**:
```json
{
  "email": "admin@spider-go.com",
  "password": "123456"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 管理员邮箱 |
| password | string | 是 | 管理员密码 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uid": 1,
    "email": "admin@spider-go.com",
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

**默认管理员账号**:
- 邮箱: `admin@spider-go.com`
- 密码: `123456`（生产环境请立即修改）

---

### 6.2 获取管理员信息

**接口地址**: `GET /api/admin/info`

**认证**: 需要管理员 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "uid": 1,
    "email": "admin@spider-go.com",
    "name": "系统管理员"
  }
}
```

---

### 6.3 修改管理员密码

**接口地址**: `POST /api/admin/reset`

**认证**: 需要管理员 Token

**请求参数**:
```json
{
  "old_password": "123456",
  "new_password": "newpassword123"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 旧密码 |
| new_password | string | 是 | 新密码（至少6位） |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 6.4 群发邮件

**接口地址**: `POST /api/admin/broadcast-email`

**认证**: 需要管理员 Token

**请求参数**:
```json
{
  "subject": "系统维护通知",
  "body": "系统将于今晚22:00进行维护，预计维护时间1小时。"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| subject | string | 是 | 邮件主题 |
| body | string | 是 | 邮件内容 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 150,
    "success": 148,
    "failed": 2
  }
}
```

**注意**:
- 会向所有注册用户发送邮件
- 发送过程为异步，不会阻塞响应

---

## 7. 通知模块

### 7.1 获取可见通知（公开）

**接口地址**: `GET /api/notices`

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "title": "系统维护通知",
      "content": "系统将于今晚22:00进行维护...",
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

### 7.2 获取通知详情（公开）

**接口地址**: `GET /api/notices/:id`

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 通知ID |

**请求示例**:
```
GET /api/notices/1
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "title": "系统维护通知",
    "content": "系统将于今晚22:00进行维护，预计维护时间1小时。请各位用户提前做好准备。",
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

---

### 7.3 获取所有通知（管理员）

**接口地址**: `GET /api/admin/notices`

**认证**: 需要管理员 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "title": "系统维护通知",
      "content": "系统将于今晚22:00进行维护...",
      "is_visible": true,
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

### 7.4 创建通知（管理员）

**接口地址**: `POST /api/admin/notices`

**认证**: 需要管理员 Token

**请求参数**:
```json
{
  "title": "系统维护通知",
  "content": "系统将于今晚22:00进行维护...",
  "is_visible": true
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 通知标题 |
| content | string | 是 | 通知内容 |
| is_visible | bool | 是 | 是否可见 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "title": "系统维护通知",
    "content": "系统将于今晚22:00进行维护...",
    "is_visible": true,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

---

### 7.5 更新通知（管理员）

**接口地址**: `PUT /api/admin/notices/:id`

**认证**: 需要管理员 Token

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 通知ID |

**请求参数**:
```json
{
  "title": "系统维护通知（更新）",
  "content": "系统维护时间调整为明晚22:00...",
  "is_visible": true
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 7.6 删除通知（管理员）

**接口地址**: `DELETE /api/admin/notices/:id`

**认证**: 需要管理员 Token

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 通知ID |

**请求示例**:
```
DELETE /api/admin/notices/1
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## 8. 配置管理模块

### 8.1 获取当前学期（公开）

**接口地址**: `GET /api/config/term`

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "term": "2024-2025-1"
  }
}
```

---

### 8.2 获取学期日期（公开）

**接口地址**: `GET /api/config/semester-dates`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 是 | 学期（格式：2024-2025-1） |

**请求示例**:
```
GET /api/config/semester-dates?term=2024-2025-1
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "term": "2024-2025-1",
    "start_date": "2024-09-01",
    "end_date": "2025-01-15"
  }
}
```

---

### 8.3 设置当前学期（管理员）

**接口地址**: `POST /api/admin/config/term`

**认证**: 需要管理员 Token

**请求参数**:
```json
{
  "term": "2024-2025-1"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 是 | 学期（格式：YYYY-YYYY-[1\|2]） |

**学期格式说明**:
- 格式：`学年开始年份-学年结束年份-学期`
- 示例：`2024-2025-1` 表示 2024-2025 学年第一学期
- 学期只能是 1 或 2

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### 8.4 设置学期日期（管理员）

**接口地址**: `POST /api/admin/config/semester-dates`

**认证**: 需要管理员 Token

**请求参数**:
```json
{
  "term": "2024-2025-1",
  "start_date": "2024-09-01",
  "end_date": "2025-01-15"
}
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 是 | 学期（格式：2024-2025-1） |
| start_date | string | 是 | 开学日期（格式：YYYY-MM-DD） |
| end_date | string | 是 | 放假日期（格式：YYYY-MM-DD） |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**注意**:
- 开学日期必须早于放假日期
- 日期格式必须为 YYYY-MM-DD

---

## 9. 统计模块

### 9.1 获取今日DAU（管理员）

**接口地址**: `GET /api/admin/statistics/dau`

**认证**: 需要管理员 Token

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "date": "2024-01-15",
    "count": 156
  }
}
```

**说明**:
- DAU (Daily Active Users): 日活跃用户数
- 每个用户每天只计数一次

---

### 9.2 获取DAU范围（管理员）

**接口地址**: `GET /api/admin/statistics/dau/range`

**认证**: 需要管理员 Token

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 是 | 开始日期（格式：YYYY-MM-DD） |
| end_date | string | 是 | 结束日期（格式：YYYY-MM-DD） |

**请求示例**:
```
GET /api/admin/statistics/dau/range?start_date=2024-01-01&end_date=2024-01-31
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "data": [
      {
        "date": "2024-01-01",
        "count": 120
      },
      {
        "date": "2024-01-02",
        "count": 135
      }
    ]
  }
}
```

**注意**:
- 日期范围最多 31 天
- 开始日期不能晚于结束日期

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 40000 | 请求参数错误 |
| 40001 | 用户不存在 |
| 40002 | 密码错误 |
| 40003 | 验证码错误或已过期 |
| 40004 | Token无效 |
| 40005 | Token已过期 |
| 40006 | 权限不足 |
| 40007 | 用户已存在 |
| 40008 | 账号已被锁定 |
| 50000 | 服务器内部错误 |
| 50001 | 数据库错误 |
| 50002 | 缓存错误 |
| 50003 | 教务系统登录失败 |
| 50004 | 教务系统解析失败 |

---

## 错误响应格式

所有错误响应统一格式：

```json
{
  "code": 40002,
  "message": "密码错误",
  "data": null
}
```

---

## 数据缓存说明

### 用户数据缓存
- 成绩、课程表、考试安排数据会缓存 1 小时
- 缓存失效后会自动从教务系统重新获取
- 用户可以通过重新绑定来强制刷新数据

### 会话缓存
- 教务系统登录会话缓存 1 小时
- 会话失效后会自动重新登录
- 连续登录失败 3 次会锁定 30 分钟

---

## 限流说明

### 验证码发送限流
- 同一邮箱 1 分钟内只能发送 1 次验证码
- 同一 IP 1 分钟内最多发送 5 次验证码

### 登录限流
- 教务系统登录失败 3 次后锁定 30 分钟
- 用户登录失败 5 次后锁定 15 分钟

---

## 开发环境配置

### 本地开发
```bash
# 运行应用（开发环境）
./spider-go.exe -env=dev

# 或使用 go run
go run main.go -env=dev
```

### 生产环境
```bash
# 运行应用（生产环境）
./spider-go.exe -env=production
```

### 环境变量
也可以通过环境变量指定：
```bash
export GO_ENV=dev
./spider-go.exe
```

---

## 附录

### A. 教务系统模式

系统支持两种教务系统访问模式：

1. **campus 模式**（校园网内）
   - 直接访问教务系统 URL
   - 速度更快
   - 仅限校园网内使用

2. **webvpn 模式**（外网访问）
   - 通过 WebVPN 访问教务系统
   - 支持外网访问
   - 配置在 `config.yaml` 中

### B. 数据库表结构

#### users 表
- `uid`: 用户ID（主键）
- `email`: 邮箱（唯一）
- `password`: 密码（加密存储）
- `sid`: 学号
- `spwd`: 教务系统密码（加密存储）
- `is_bind`: 是否已绑定
- `locked_until`: 锁定截止时间

#### administrators 表
- `uid`: 管理员ID（主键）
- `email`: 邮箱（唯一）
- `password`: 密码（加密存储）
- `name`: 姓名

#### notices 表
- `id`: 通知ID（主键）
- `title`: 标题
- `content`: 内容
- `is_visible`: 是否可见
- `created_at`: 创建时间
- `updated_at`: 更新时间

---

## 联系方式

如有问题，请提交 Issue 到 GitHub 仓库。
