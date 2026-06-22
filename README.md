# Spider-Go Docker Web 版

一个面向中南林业科技大学教务系统的 Go 服务，提供成绩、课表、考试安排等查询能力。本仓库在原项目基础上补充了可直接使用的 Web 前端、Docker Compose 部署，以及新版 CAS/WebVPN 登录适配。

> **原项目与作者**
>
> 本项目基于 [StarryKira/spider-go](https://github.com/StarryKira/spider-go) 修改，原作者为 [StarryKira](https://github.com/StarryKira)。后端主体、接口设计和核心教务查询能力来自原项目。
>
> 本仓库由 [printf-wt](https://github.com/printf-wt) 在原项目基础上维护和改进，并非学校官方项目。

## 本仓库的改进

- 新增学生端和管理员端 Web 页面
- 新增 Dockerfile、Docker Compose、MySQL 和 Redis 一键部署
- 适配当前 CAS 登录表单与 MFA 状态参数
- 完成 WebVPN 授权回调和会话 Cookie 获取
- 改进教务账号绑定错误提示，保留后端真实错误原因
- 提供脱敏的生产配置示例，避免误提交邮箱授权码等凭据

## 功能

- 邮箱验证码注册、登录
- 绑定本人教务账号
- 查询成绩、课表和考试安排
- 管理员登录、学期配置、通知管理和基础统计
- 校园网直连与校外 WebVPN 两种教务访问模式

## 环境要求

- Docker Desktop，或 Docker Engine 24+
- Docker Compose v2（使用 `docker compose` 命令）
- 能正常访问学校 CAS、WebVPN 和教务系统

仅使用 Docker 部署时，不需要在宿主机安装 Go、MySQL、Redis 或 Node.js。

## 快速部署

### 1. 克隆仓库

```bash
git clone https://github.com/printf-wt/spider-go-fixed.git
cd spider-go-fixed
```

### 2. 创建本地配置

Linux/macOS：

```bash
cp deploy/config.production.example.yaml deploy/config.production.yaml
```

Windows PowerShell：

```powershell
Copy-Item deploy/config.production.example.yaml deploy/config.production.yaml
```

`deploy/config.production.yaml` 已被 `.gitignore` 忽略，不会被正常的 Git 提交上传。

### 3. 修改必要配置

打开 `deploy/config.production.yaml`，至少检查以下内容：

```yaml
jwt:
  secret: "替换为足够长的随机字符串"

jwc:
  # 校外部署通常使用 webvpn；校园网环境可使用 campus
  mode: "webvpn"

email:
  smtp_host: "smtp.qq.com"
  smtp_port: 465
  username: "你的QQ邮箱@qq.com"
  password: "QQ邮箱SMTP授权码，不是QQ密码"
```

数据库和 Redis 的默认值与 `compose.yaml` 一致，可直接用于本地部署。若修改数据库或 Redis 密码，必须同时修改 `compose.yaml` 和 `deploy/config.production.yaml` 中对应的值。

### 4. 启动服务

```bash
docker compose up -d --build
```

检查状态：

```bash
docker compose ps
```

四个服务 `frontend`、`app`、`mysql`、`redis` 均显示 `healthy` 后即可访问。

### 5. 打开页面

- Web 前端：<http://localhost:3000>
- 后端 API：<http://localhost:8080>
- API 说明：[API.md](./API.md)

## 首次使用

### 管理员

首次启动且数据库中没有管理员时，后端会创建默认管理员：

- 邮箱：`admin@spider-go.com`
- 密码：`123456`

进入前端后选择“管理员登录”。**登录成功后请立即在管理员账户页面修改默认密码。** 已存在管理员时，重新启动不会覆盖管理员账号。

### 学生用户

1. 在前端选择学生端并注册 Spider-Go 账号。
2. 填写邮箱并获取验证码；这一步依赖 SMTP 配置。
3. 登录后输入本人的学号和统一身份认证密码完成绑定。
4. 绑定成功后即可查询成绩、课表和考试安排。

教务密码必须完整输入，末尾的英文句号等标点也属于密码。教务系统或 WebVPN 改版后，登录流程可能需要再次适配。

## 常用运维命令

```bash
# 查看状态
docker compose ps

# 查看后端日志
docker compose logs -f app

# 停止服务但保留数据
docker compose down

# 拉取更新并重新构建
git pull
docker compose up -d --build
```

MySQL 和 Redis 数据保存在 Docker 命名卷中。`docker compose down -v` 会删除账号、绑定信息和缓存数据，请勿在没有备份时执行。

## 局域网或公网访问

默认端口只绑定到本机 `127.0.0.1`。需要从其他设备访问时，可将 `compose.yaml` 中的：

```yaml
- "127.0.0.1:3000:80"
```

改为：

```yaml
- "3000:80"
```

公网部署建议使用 Nginx/Caddy 配置 HTTPS 和域名，不要直接暴露 MySQL、Redis 或后端管理接口。同时修改 CORS、默认管理员密码、JWT 密钥及所有数据库密码。

## 常见问题

### 启动时提示找不到配置文件

确认已从示例文件创建 `deploy/config.production.yaml`。

### 邮箱验证码发送失败

检查 SMTP 主机、端口、邮箱和授权码。QQ 邮箱应使用单独生成的 SMTP 授权码，而不是 QQ 登录密码。

### 教务账号提示用户名或密码错误

- 先在学校官方统一身份认证页面验证同一组凭据。
- 检查密码大小写、空格和末尾标点。
- 校外部署确认 `jwc.mode` 为 `webvpn`。
- 使用 `docker compose logs -f app` 查看后端返回的具体原因。

### 页面无法从其他设备打开

默认仅允许本机访问。按“局域网或公网访问”一节修改端口绑定，并检查宿主机防火墙。

## 安全与使用说明

- 不要提交 `deploy/config.production.yaml`、`.env`、邮箱授权码、教务密码或访问令牌。
- 只绑定和查询本人有权访问的教务账号，不要用于批量抓取、越权访问或商业用途。
- 本项目依赖学校系统页面结构，学校系统升级可能导致部分功能暂时失效。
- 仓库未发现原项目提供的独立许可证文件，本仓库不额外授予超出原项目的使用许可；使用、修改或再分发前请确认原作者的许可要求。

## 致谢

感谢 [StarryKira](https://github.com/StarryKira) 创建并开源原始 [spider-go](https://github.com/StarryKira/spider-go) 项目。本仓库的改进建立在原项目工作的基础上。
