use crate::login::{login_in_browser, open_login_page_only};
use crate::query::{query_courses, query_credits, query_grades, query_level_exams, query_student_info};
use crate::session::{clear_session, load_session, save_session};
use anyhow::{Context, Result};
use serde_json::{json, Value};
use tokio::io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader};

const SERVER_NAME: &str = "csuft-jw-mcp";
const SERVER_VERSION: &str = env!("CARGO_PKG_VERSION");

pub async fn serve_stdio() -> Result<()> {
    let stdin = tokio::io::stdin();
    let mut reader = BufReader::new(stdin);
    let mut stdout = tokio::io::stdout();

    loop {
        let Some(message) = read_message(&mut reader).await? else {
            break;
        };
        if message.get("method").and_then(|m| m.as_str()) == Some("notifications/initialized") {
            continue;
        }
        if message.get("id").is_none() {
            continue;
        }
        let response = handle_message(message).await;
        write_message(&mut stdout, &response).await?;
    }
    Ok(())
}

pub async fn handle_message(msg: Value) -> Value {
    let id = msg.get("id").cloned().unwrap_or(Value::Null);
    let method = msg.get("method").and_then(|m| m.as_str()).unwrap_or("");
    let params = msg.get("params").cloned().unwrap_or(json!({}));
    match method {
        "initialize" => ok(
            id,
            json!({
                "protocolVersion": params.get("protocolVersion").cloned().unwrap_or(json!("2024-11-05")),
                "capabilities": { "tools": { "listChanged": false } },
                "serverInfo": { "name": SERVER_NAME, "version": SERVER_VERSION }
            }),
        ),
        "ping" => ok(id, json!({})),
        "tools/list" => ok(id, json!({ "tools": tool_list() })),
        "tools/call" => match call_tool(&params).await {
            Ok(result) => ok(id, result),
            Err(err) => ok(
                id,
                json!({
                    "content": [{ "type": "text", "text": format!("错误: {err}") }],
                    "isError": true
                }),
            ),
        },
        other => error(id, -32601, format!("Method not found: {other}")),
    }
}

fn tool_list() -> Value {
    json!([
        tool("csuft_jw_login", "打开本机 Chrome，进入中南林 CAS 登录页。请在弹出窗口完成指纹/手机验证码登录，直到跳进教务系统。只读，不改教务数据。", json!({
            "type": "object",
            "properties": {
                "mode": { "type": "string", "enum": ["webvpn", "campus"], "description": "默认 webvpn（校外）；campus=校园网" },
                "timeout_secs": { "type": "integer", "description": "等待登录完成的秒数，默认 180" }
            }
        })),
        tool("csuft_jw_open_login_page", "仅打开登录页（不捕获 cookie）。用于检查登录地址。", json!({
            "type": "object",
            "properties": { "mode": { "type": "string", "enum": ["campus", "webvpn"] } }
        })),
        tool("csuft_jw_status", "查看本地教务会话是否存在（不访问教务写接口）。", json!({ "type": "object", "properties": {} })),
        tool("csuft_jw_logout", "只删除本机保存的 cookie，不向教务系统发送任何写操作。", json!({ "type": "object", "properties": {} })),
        tool("csuft_jw_get_grades", "查询全部或指定学期成绩，含每门课分数、学分、绩点、性质，以及分学期汇总。", json!({
            "type": "object",
            "properties": {
                "term": { "type": "string", "description": "可选学期，如 2024-2025-1；留空表示全部" }
            }
        })),
        tool("csuft_jw_get_gpa", "查询学分与绩点：平均绩点、平均分、基本分、必修学分、已获得必修学分。", json!({ "type": "object", "properties": {} })),
        tool("csuft_jw_get_level_exams", "查询等级考试成绩（英语四六级、计算机等级等）。", json!({ "type": "object", "properties": {} })),
        tool("csuft_jw_get_courses", "查询指定学期、周次的课表。", json!({
            "type": "object",
            "properties": {
                "term": { "type": "string", "description": "学期，如 2025-2026-2" },
                "week": { "type": "integer", "description": "周次 1-20" }
            },
            "required": ["term", "week"]
        })),
        tool("csuft_jw_get_student_info", "从成绩页只读解析姓名、学院、专业、班级。", json!({ "type": "object", "properties": {} })),
        tool("csuft_jw_tls_probe", "用与教务请求相同的 Chrome TLS/HTTP2 指纹访问公开检测接口，返回 JA3/JA4。只读。", json!({ "type": "object", "properties": {} }))
    ])
}

fn tool(name: &str, description: &str, input_schema: Value) -> Value {
    json!({
        "name": name,
        "description": description,
        "inputSchema": input_schema
    })
}

async fn call_tool(params: &Value) -> Result<Value> {
    let name = params.get("name").and_then(|v| v.as_str()).unwrap_or("");
    let args = params.get("arguments").cloned().unwrap_or(json!({}));
    match name {
        "csuft_jw_login" => {
            let mode = args.get("mode").and_then(|v| v.as_str()).unwrap_or("webvpn");
            let timeout = args.get("timeout_secs").and_then(|v| v.as_u64()).unwrap_or(180);
            let outcome = login_in_browser(mode, timeout).await?;
            save_session(&outcome.session)?;
            Ok(text(format!(
                "登录成功。模式={}，落地页={}，cookie={} 个。现在可以查询成绩。",
                outcome.session.mode,
                outcome.landed_url,
                outcome.session.cookies.len()
            )))
        }
        "csuft_jw_open_login_page" => {
            let mode = args.get("mode").and_then(|v| v.as_str()).unwrap_or("webvpn");
            let url = open_login_page_only(mode)?;
            Ok(text(format!("已尝试打开登录页: {url}")))
        }
        "csuft_jw_status" => {
            let status = match load_session()? {
                Some(s) => json!({
                    "logged_in": !s.is_empty(),
                    "mode": s.mode,
                    "cookie_count": s.cookies.len(),
                    "updated_at": s.updated_at,
                    "write_access": false
                }),
                None => json!({ "logged_in": false, "write_access": false }),
            };
            Ok(text(serde_json::to_string_pretty(&status)?))
        }
        "csuft_jw_logout" => {
            let removed = clear_session()?;
            Ok(text(if removed {
                "已清除本机会话".into()
            } else {
                "本机没有保存的会话".into()
            }))
        }
        "csuft_jw_get_grades" => {
            let session = require_session()?;
            let term = args.get("term").and_then(|v| v.as_str());
            let payload = query_grades(&session, term).await?;
            Ok(text(serde_json::to_string_pretty(&payload)?))
        }
        "csuft_jw_get_gpa" => {
            let session = require_session()?;
            let credits = query_credits(&session).await?;
            Ok(text(serde_json::to_string_pretty(&credits)?))
        }
        "csuft_jw_get_level_exams" => {
            let session = require_session()?;
            let items = query_level_exams(&session).await?;
            Ok(text(serde_json::to_string_pretty(&items)?))
        }
        "csuft_jw_get_courses" => {
            let session = require_session()?;
            let term = args
                .get("term")
                .and_then(|v| v.as_str())
                .ok_or_else(|| anyhow::anyhow!("缺少 term"))?;
            let week = args
                .get("week")
                .and_then(|v| v.as_i64())
                .ok_or_else(|| anyhow::anyhow!("缺少 week"))? as i32;
            let schedule = query_courses(&session, term, week).await?;
            Ok(text(serde_json::to_string_pretty(&schedule)?))
        }
        "csuft_jw_get_student_info" => {
            let session = require_session()?;
            let info = query_student_info(&session).await?;
            Ok(text(serde_json::to_string_pretty(&info)?))
        }
        "csuft_jw_tls_probe" => {
            let probe = crate::client::probe_chrome_tls().await?;
            crate::client::tls_probe_is_chrome(&probe)?;
            Ok(text(serde_json::to_string_pretty(&serde_json::json!({
                "ok": true,
                "user_agent": probe.user_agent,
                "ja3_hash": probe.ja3_hash(),
                "ja3": probe.ja3(),
                "ja4": probe.ja4(),
                "http_version": probe.http_version,
                "tls_version": probe.tls.tls_version_negotiated
            }))?))
        }
        other => anyhow::bail!("未知工具: {other}"),
    }
}

fn require_session() -> Result<crate::session::Session> {
    load_session()?.filter(|s| !s.is_empty()).ok_or_else(|| {
        anyhow::anyhow!("尚未登录。请先调用 csuft_jw_login，在弹出的 Chrome 窗口完成统一认证。")
    })
}

fn text(s: String) -> Value {
    json!({ "content": [{ "type": "text", "text": s }] })
}

fn ok(id: Value, result: Value) -> Value {
    json!({ "jsonrpc": "2.0", "id": id, "result": result })
}

fn error(id: Value, code: i64, message: String) -> Value {
    json!({ "jsonrpc": "2.0", "id": id, "error": { "code": code, "message": message } })
}

async fn read_message<R: AsyncReadExt + Unpin>(reader: &mut BufReader<R>) -> Result<Option<Value>> {
    let mut content_length: Option<usize> = None;
    loop {
        let mut line = String::new();
        let n = reader.read_line(&mut line).await?;
        if n == 0 {
            return Ok(None);
        }
        let trimmed = line.trim_end();
        if trimmed.is_empty() {
            break;
        }
        if let Some(rest) = trimmed.strip_prefix("Content-Length:") {
            content_length = Some(rest.trim().parse()?);
        }
    }
    let Some(len) = content_length else {
        return Ok(None);
    };
    let mut buf = vec![0u8; len];
    reader.read_exact(&mut buf).await?;
    Ok(Some(serde_json::from_slice(&buf)?))
}

async fn write_message<W: AsyncWriteExt + Unpin>(writer: &mut W, value: &Value) -> Result<()> {
    let body = serde_json::to_vec(value)?;
    let header = format!("Content-Length: {}\r\n\r\n", body.len());
    writer.write_all(header.as_bytes()).await?;
    writer.write_all(&body).await?;
    writer.flush().await?;
    Ok(())
}

/// Encode a JSON-RPC body with MCP stdio headers. Used by tests of the real framer.
pub fn encode_message(value: &Value) -> Vec<u8> {
    let body = serde_json::to_vec(value).expect("json");
    let mut out = format!("Content-Length: {}\r\n\r\n", body.len()).into_bytes();
    out.extend_from_slice(&body);
    out
}

pub fn decode_message(bytes: &[u8]) -> Result<Value> {
    let text = std::str::from_utf8(bytes)?;
    let (headers, body) = text.split_once("\r\n\r\n").context("missing header separator")?;
    let mut len = None;
    for line in headers.lines() {
        if let Some(rest) = line.strip_prefix("Content-Length:") {
            len = Some(rest.trim().parse::<usize>()?);
        }
    }
    let len = len.context("missing Content-Length")?;
    let body_bytes = body.as_bytes();
    anyhow::ensure!(body_bytes.len() >= len, "truncated body");
    Ok(serde_json::from_slice(&body_bytes[..len])?)
}
