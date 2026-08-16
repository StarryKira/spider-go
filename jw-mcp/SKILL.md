---
name: csuft-jw-mcp
description: Install and use the read-only CSUFT academic MCP (grades, GPA, schedule, level exams).
---

# 安装中南林教务只读 MCP

1. 确认本机有 Rust（`cargo --version`）和 Chrome。
2. 在仓库执行：

```bash
chmod +x jw-mcp/install.sh
./jw-mcp/install.sh
```

3. 重启 Claude Code。
4. 先调用 `csuft_jw_login`（默认 WebVPN）。校园网再传 `mode=campus`。在弹出的 Chrome 里完成登录，直到进入教务。
5. 再用 `csuft_jw_get_grades` / `csuft_jw_get_gpa` / `csuft_jw_get_level_exams` / `csuft_jw_get_courses`。

本 MCP 只读，不会向教务系统写入。
