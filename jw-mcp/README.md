# csuft-jw-mcp

中南林业科技大学教务 **只读** MCP。给 Claude Code 用：打开登录页完成统一认证（含浏览器指纹/手机验证码），然后查询成绩、绩点/学分、等级考试和课表。

**没有任何教务写操作。** `logout` 只删本机 cookie。

## Claude Code 自行安装

在仓库里执行：

```bash
chmod +x jw-mcp/install.sh
./jw-mcp/install.sh
```

或手动：

```bash
cd jw-mcp && cargo build --release
claude mcp add --transport stdio csuft-jw -- "$(pwd)/target/release/csuft-jw-mcp"
```

安装后重启 Claude Code。对本机 Chrome 资料目录保存在 `~/.csuft-jw-mcp/chrome-profile`，同一台电脑再次登录会复用 FingerprintJS 的 `fpVisitorId`。

## 工具

| 工具 | 作用 |
|------|------|
| `csuft_jw_login` | 打开 Chrome，登录 CAS 并跟随跳转到教务。默认 `webvpn`，校园网可传 `mode=campus` |
| `csuft_jw_get_grades` | 完整成绩：课程、分数、学分、绩点、必修/选修、补考、分学期汇总 |
| `csuft_jw_get_gpa` | 平均绩点、平均分、基本分、必修学分 |
| `csuft_jw_get_level_exams` | 四六级等等级考试 |
| `csuft_jw_get_courses` | 指定学期+周课表 |
| `csuft_jw_get_student_info` | 姓名/学院/专业/班级 |
| `csuft_jw_status` | 本机会话状态 |
| `csuft_jw_logout` | 清除本机 cookie |

## 成绩查询覆盖

- 全部学期或 `term=2024-2025-1`
- 课程代码、名称、分数、学分、绩点、考试性质、课程属性
- 缓考不计入绩点
- 补考按教务规则折算
- 分学期 GPA 与学分汇总

## TLS 指纹

教务请求走 `wreq` + Chrome 131 / macOS 模拟，TLS（JA3/JA4）和 HTTP/2 设置按桌面 Chrome，而不是 rustls/reqwest 默认握手。可用 `csuft_jw_tls_probe` 对照 `tls.peet.ws`。

## 安全边界

- 只 GET/POST 查询页（成绩、课表、等级考试）
- 不评教、不改密、不改绑定、不提交表单写数据
- 会话只存在本机 `~/.csuft-jw-mcp/session.json`
