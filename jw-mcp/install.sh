#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT"

if ! command -v cargo >/dev/null 2>&1; then
  echo "需要先安装 Rust: https://rustup.rs" >&2
  exit 1
fi

echo "正在编译 csuft-jw-mcp ..."
cargo build --release
BIN="$ROOT/target/release/csuft-jw-mcp"
chmod +x "$BIN"

MCP_JSON="$ROOT/.mcp.json"
cat > "$MCP_JSON" <<EOF
{
  "mcpServers": {
    "csuft-jw": {
      "command": "$BIN",
      "args": []
    }
  }
}
EOF

if command -v claude >/dev/null 2>&1; then
  claude mcp remove csuft-jw >/dev/null 2>&1 || true
  claude mcp add --transport stdio csuft-jw -- "$BIN" || {
    echo "claude mcp add 失败，已写入 $MCP_JSON，可把该文件复制到项目根或 ~/.claude.json"
  }
else
  echo "未检测到 claude CLI。已生成 $MCP_JSON"
  echo "Claude Code 可执行: claude mcp add --transport stdio csuft-jw -- $BIN"
fi

echo
echo "安装完成。"
echo "二进制: $BIN"
echo "这是只读 MCP：登录只打开浏览器并保存 cookie，查询成绩/课表/绩点/等级考试，不向教务系统写入。"
echo "首次使用请让 Claude 调用 csuft_jw_login（默认 WebVPN；校园网传 mode=campus）。"
