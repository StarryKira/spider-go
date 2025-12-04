# API 接口文档

## 目录
- [1. 概述](#1-概述)
- [2. 通用说明](#2-通用说明)
- [3. 公开接口](#3-公开接口)
  - [3.1 用户登录](#31-用户登录)
  - [3.2 用户注册](#32-用户注册)
  - [3.3 重置密码](#33-重置密码)
  - [3.4 发送邮箱验证码](#34-发送邮箱验证码)
  - [3.5 获取可见通知](#35-获取可见通知)
  - [3.6 获取当前学期](#36-获取当前学期)
- [4. 用户接口（需要认证）](#4-用户接口需要认证)
  - [4.1 绑定教务系统账号](#41-绑定教务系统账号)
  - [4.2 获取用户信息](#42-获取用户信息)
  - [4.3 检查是否绑定](#43-检查是否绑定)
  - [4.4 获取成绩](#44-获取成绩)
  - [4.5 获取等级考试成绩](#45-获取等级考试成绩)
  - [4.6 获取成绩分析](#46-获取成绩分析)
  - [4.7 获取课程表](#47-获取课程表)
  - [4.8 获取考试安排](#48-获取考试安排)
- [5. 管理员接口](#5-管理员接口)
  - [5.1 管理员登录](#51-管理员登录)
  - [5.2 获取管理员信息](#52-获取管理员信息)
  - [5.3 修改管理员密码](#53-修改管理员密码)
  - [5.4 创建通知](#54-创建通知)
  - [5.5 更新通知](#55-更新通知)
  - [5.6 删除通知](#56-删除通知)
  - [5.7 获取所有通知](#57-获取所有通知)
  - [5.8 获取日活统计](#58-获取日活统计)
  - [5.9 获取日活范围统计](#59-获取日活范围统计)
  - [5.10 设置当前学期](#510-设置当前学期)

---

## 1. 概述

本文档描述了教务系统 API 的所有接口规范，用于前后端对接。

**Base URL**: `http://your-domain/api`

---

## 2. 通用说明

### 2.1 认证方式

用户接口和管理员接口需要在请求头中携带 JWT Token：

```
Authorization: Bearer <token>
```

或通过 Cookie 携带：
```
Cookie: access_token=<token>
```

### 2.2 通用响应格式

#### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 具体数据
  }
}
```

#### 错误响应
```json
{
  "code": 错误码,
  "message": "错误描述"
}
```

### 2.3 常见错误码

| 错误码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 3. 公开接口

### 3.1 用户登录

**接口**: `POST /api/login`

**描述**: 用户使用邮箱和密码登录

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 用户邮箱 |
| password | string | 是 | 用户密码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| token | string | JWT 访问令牌，有效期 7 天 |

**错误响应示例**:
```json
{
  "code": 401,
  "message": "用户不存在或密码错误"
}
```

**备注**:
- 登录成功后，Token 会同时通过响应体返回和设置到 Cookie 中
- Cookie 有效期为 7 天

---

### 3.2 用户注册

**接口**: `POST /api/register`

**描述**: 使用邮箱验证码注册新用户

**请求体**:
```json
{
  "name": "张三",
  "email": "user@example.com",
  "captcha": "123456",
  "password": "password123"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 用户姓名 |
| email | string | 是 | 用户邮箱 |
| captcha | string | 是 | 邮箱验证码（6位数字） |
| password | string | 是 | 用户密码（建议至少8位） |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "注册成功"
  }
}
```

**错误响应示例**:
```json
{
  "code": 400,
  "message": "邮箱已被注册"
}
```

或

```json
{
  "code": 400,
  "message": "验证码错误或已过期"
}
```

**备注**:
- 注册前需要先调用发送验证码接口
- 验证码有效期为 5 分钟

---

### 3.3 重置密码

**接口**: `POST /api/reset`

**描述**: 使用邮箱验证码重置密码

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "newPassword123",
  "captcha": "123456"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 用户邮箱 |
| password | string | 是 | 新密码 |
| captcha | string | 是 | 邮箱验证码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "重置成功"
  }
}
```

**错误响应示例**:
```json
{
  "code": 404,
  "message": "用户不存在"
}
```

---

### 3.4 发送邮箱验证码

**接口**: `POST /api/captcha/send`

**描述**: 向指定邮箱发送验证码

**请求体**:
```json
{
  "email": "user@example.com"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 接收验证码的邮箱地址 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "验证码已发送"
  }
}
```

**错误响应示例**:
```json
{
  "code": 500,
  "message": "发送验证码失败"
}
```

**备注**:
- 验证码为6位数字
- 有效期5分钟
- 同一邮箱60秒内只能发送一次

---

### 3.5 获取可见通知

**接口**: `GET /api/notices`

**描述**: 获取所有可见的系统通知

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "nid": 1,
      "title": "系统维护通知",
      "content": "系统将于本周六进行维护...",
      "created_at": "2025-12-01T10:00:00Z",
      "is_visible": true
    }
  ]
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| nid | int | 通知ID |
| title | string | 通知标题 |
| content | string | 通知内容 |
| created_at | string | 创建时间 (ISO 8601格式) |
| is_visible | bool | 是否可见 |

---

### 3.6 获取当前学期

**接口**: `GET /api/config/term`

**描述**: 获取系统当前学期配置

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "term": "2024-2025-1"
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| term | string | 当前学期，格式：学年-学期（1为第一学期，2为第二学期） |

---

## 4. 用户接口（需要认证）

以下接口需要在请求头中携带有效的 JWT Token。

### 4.1 绑定教务系统账号

**接口**: `POST /api/user/bind`

**描述**: 绑定学校教务系统账号（学号和密码）

**认证**: 需要用户登录

**请求体**:
```json
{
  "sid": "2021001001",
  "spwd": "jwc_password"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sid | string | 是 | 学号 |
| spwd | string | 是 | 教务系统密码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "绑定成功"
  }
}
```

**错误响应示例**:
```json
{
  "code": 401,
  "message": "未授权"
}
```

或

```json
{
  "code": 400,
  "message": "学号和密码不能为空"
}
```

**备注**:
- 绑定前需要先登录获取 Token
- 绑定后会清除旧的教务系统会话缓存
- 可以重复调用来更新绑定信息

---

### 4.2 获取用户信息

**接口**: `GET /api/user/info`

**描述**: 获取当前登录用户的详细信息

**认证**: 需要用户登录

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "uid": 1,
    "email": "user@example.com",
    "name": "张三",
    "sid": "2021001001",
    "created_at": "2025-01-01T00:00:00Z",
    "avatar": "https://example.com/avatar.jpg"
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| uid | int | 用户ID |
| email | string | 用户邮箱 |
| name | string | 用户姓名 |
| sid | string | 学号（如已绑定） |
| created_at | string | 账号创建时间 (ISO 8601格式) |
| avatar | string | 用户头像URL |

**错误响应示例**:
```json
{
  "code": 401,
  "message": "未授权"
}
```

**备注**:
- 响应中不会包含密码等敏感信息
- 如果未绑定教务系统，sid 字段为空字符串

---

### 4.3 检查是否绑定

**接口**: `GET /api/user/isbind`

**描述**: 检查当前用户是否已绑定教务系统账号

**认证**: 需要用户登录

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "is_bind": true
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| is_bind | bool | 是否已绑定教务系统账号。true: 已绑定，false: 未绑定 |

**错误响应示例**:
```json
{
  "code": 401,
  "message": "未授权"
}
```

**备注**:
- 该接口用于前端判断是否需要引导用户进行教务系统账号绑定
- 绑定状态取决于用户的 sid 和 spwd 字段是否都不为空
- 可在应用启动或进入需要教务系统数据的页面时调用

**使用场景示例**:
```javascript
// 前端示例代码
async function checkBindStatus() {
  const response = await fetch('/api/user/isbind', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  const data = await response.json();

  if (!data.data.is_bind) {
    // 引导用户去绑定页面
    router.push('/bind');
  } else {
    // 已绑定，继续正常流程
    loadUserData();
  }
}
```

---

### 4.4 获取成绩

**接口**: `GET /api/user/grades`

**描述**: 获取用户的成绩信息

**认证**: 需要用户登录并绑定教务系统

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 否 | 学期，格式：2024-2025-1。不传则返回所有学期 |

**请求示例**:
```
GET /api/user/grades?term=2024-2025-1
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "course_name": "高等数学",
      "course_code": "MATH101",
      "credit": 4.0,
      "grade": "95",
      "gpa": 4.0,
      "term": "2024-2025-1"
    }
  ]
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| course_name | string | 课程名称 |
| course_code | string | 课程代码 |
| credit | float | 学分 |
| grade | string | 成绩 |
| gpa | float | 绩点 |
| term | string | 学期 |

---

### 4.5 获取等级考试成绩

**接口**: `GET /api/user/grades/level`

**描述**: 获取用户的等级考试成绩（如英语四六级、计算机等级考试等）

**认证**: 需要用户登录并绑定教务系统

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "exam_name": "大学英语四级",
      "exam_date": "2024-12-15",
      "score": "560",
      "result": "通过"
    }
  ]
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| exam_name | string | 考试名称 |
| exam_date | string | 考试日期 |
| score | string | 成绩/分数 |
| result | string | 考试结果 |

---

### 4.6 获取成绩分析

**接口**: `GET /api/user/grades/analysis`

**描述**: 获取用户近期学期的成绩分析数据

**认证**: 需要用户登录并绑定教务系统

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "average_gpa": 3.75,
    "total_credits": 120,
    "terms": [
      {
        "term": "2024-2025-1",
        "gpa": 3.8,
        "credits": 20
      }
    ]
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| average_gpa | float | 平均绩点 |
| total_credits | int | 总学分 |
| terms | array | 各学期详细数据 |

---

### 4.7 获取课程表

**接口**: `GET /api/user/courses`

**描述**: 获取用户的课程表信息

**认证**: 需要用户登录并绑定教务系统

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 否 | 周次，不传则返回当前周 |
| term | string | 否 | 学期，不传则使用当前学期 |

**请求示例**:
```
GET /api/user/courses?week=1&term=2024-2025-1
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "course_name": "高等数学",
      "teacher": "张教授",
      "location": "教学楼A101",
      "day_of_week": 1,
      "start_section": 1,
      "end_section": 2,
      "weeks": "1-16周"
    }
  ]
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| course_name | string | 课程名称 |
| teacher | string | 授课教师 |
| location | string | 上课地点 |
| day_of_week | int | 星期几（1-7，1为周一） |
| start_section | int | 开始节次 |
| end_section | int | 结束节次 |
| weeks | string | 上课周次 |

---

### 4.8 获取考试安排

**接口**: `GET /api/user/exams`

**描述**: 获取用户的考试安排信息

**认证**: 需要用户登录并绑定教务系统

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 否 | 学期，不传则使用当前学期 |

**请求示例**:
```
GET /api/user/exams?term=2024-2025-1
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "course_name": "高等数学",
      "exam_time": "2025-01-15 09:00",
      "location": "教学楼A101",
      "seat_number": "25",
      "duration": "120分钟"
    }
  ]
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| course_name | string | 考试课程名称 |
| exam_time | string | 考试时间 |
| location | string | 考试地点 |
| seat_number | string | 座位号 |
| duration | string | 考试时长 |

---

## 5. 管理员接口

### 5.1 管理员登录

**接口**: `POST /api/admin/login`

**描述**: 管理员使用用户名和密码登录

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 管理员用户名 |
| password | string | 是 | 管理员密码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| token | string | 管理员 JWT 访问令牌 |

---

### 5.2 获取管理员信息

**接口**: `GET /api/admin/info`

**描述**: 获取当前登录管理员的信息

**认证**: 需要管理员登录

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

---

### 5.3 修改管理员密码

**接口**: `POST /api/admin/reset`

**描述**: 修改当前登录管理员的密码

**认证**: 需要管理员登录

**请求体**:
```json
{
  "old_password": "old123",
  "new_password": "new123"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 原密码 |
| new_password | string | 是 | 新密码 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "密码修改成功"
  }
}
```

---

### 5.4 创建通知

**接口**: `POST /api/admin/notices`

**描述**: 创建新的系统通知

**认证**: 需要管理员登录

**请求体**:
```json
{
  "title": "系统维护通知",
  "content": "系统将于本周六进行维护...",
  "is_visible": true
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 通知标题 |
| content | string | 是 | 通知内容 |
| is_visible | bool | 是 | 是否可见 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "nid": 1,
    "message": "创建成功"
  }
}
```

---

### 5.5 更新通知

**接口**: `PUT /api/admin/notices/:nid`

**描述**: 更新指定的系统通知

**认证**: 需要管理员登录

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nid | int | 是 | 通知ID |

**请求体**:
```json
{
  "title": "更新后的标题",
  "content": "更新后的内容",
  "is_visible": false
}
```

**请求示例**:
```
PUT /api/admin/notices/1
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "更新成功"
  }
}
```

---

### 5.6 删除通知

**接口**: `DELETE /api/admin/notices/:nid`

**描述**: 删除指定的系统通知

**认证**: 需要管理员登录

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nid | int | 是 | 通知ID |

**请求示例**:
```
DELETE /api/admin/notices/1
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "删除成功"
  }
}
```

---

### 5.7 获取所有通知

**接口**: `GET /api/admin/notices`

**描述**: 获取所有系统通知（包括不可见的）

**认证**: 需要管理员登录

**请求参数**: 无

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "nid": 1,
      "title": "系统维护通知",
      "content": "系统将于本周六进行维护...",
      "created_at": "2025-12-01T10:00:00Z",
      "is_visible": true
    }
  ]
}
```

---

### 5.8 获取日活统计

**接口**: `GET /api/admin/statistics/dau`

**描述**: 获取指定日期的日活跃用户统计

**认证**: 需要管理员登录

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| date | string | 否 | 日期，格式：YYYY-MM-DD。不传则返回今天 |

**请求示例**:
```
GET /api/admin/statistics/dau?date=2025-12-01
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "date": "2025-12-01",
    "active_users": 150
  }
}
```

**响应参数说明**:
| 参数 | 类型 | 说明 |
|------|------|------|
| date | string | 日期 |
| active_users | int | 活跃用户数 |

---

### 5.9 获取日活范围统计

**接口**: `GET /api/admin/statistics/dau/range`

**描述**: 获取指定时间范围的日活跃用户统计

**认证**: 需要管理员登录

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start_date | string | 是 | 开始日期，格式：YYYY-MM-DD |
| end_date | string | 是 | 结束日期，格式：YYYY-MM-DD |

**请求示例**:
```
GET /api/admin/statistics/dau/range?start_date=2025-12-01&end_date=2025-12-07
```

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "date": "2025-12-01",
      "active_users": 150
    },
    {
      "date": "2025-12-02",
      "active_users": 165
    }
  ]
}
```

---

### 5.10 设置当前学期

**接口**: `POST /api/admin/config/term`

**描述**: 设置系统当前学期

**认证**: 需要管理员登录

**请求体**:
```json
{
  "term": "2024-2025-2"
}
```

**请求参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| term | string | 是 | 学期，格式：学年-学期 |

**成功响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "message": "设置成功"
  }
}
```

---

## 附录

### A. 学期格式说明

学期格式为：`学年-学期`
- 学年格式：`YYYY-YYYY`（如 2024-2025）
- 学期：`1` 表示第一学期（秋季），`2` 表示第二学期（春季）
- 完整示例：`2024-2025-1`

### B. 时间格式说明

所有时间字段均采用 ISO 8601 格式：
- 日期时间：`YYYY-MM-DDTHH:mm:ssZ`（如 2025-12-01T10:00:00Z）
- 日期：`YYYY-MM-DD`（如 2025-12-01）

### C. 认证流程说明

1. 用户/管理员登录获取 Token
2. 在后续请求的 Header 中携带 Token：`Authorization: Bearer <token>`
3. 或通过 Cookie 自动携带：`access_token=<token>`
4. Token 有效期为 7 天，过期后需重新登录

### D. CORS 配置说明

如需跨域访问，请确保服务器端已正确配置 CORS 策略。前端请求时需要携带凭证：

```javascript
fetch(url, {
  credentials: 'include',  // 携带 Cookie
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
```

---

**文档版本**: v1.0
**最后更新**: 2025-12-04
**维护者**: 后端开发团队
