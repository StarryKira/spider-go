use crate::session::{data_dir, Cookie, Endpoints, Session};
use anyhow::{bail, Context, Result};
use futures_util::{SinkExt, StreamExt};
use serde_json::{json, Value};
use std::net::TcpListener;
use std::path::PathBuf;
use std::process::{Child, Command, Stdio};
use std::time::Duration;
use tokio::time::{sleep, Instant};
use tokio_tungstenite::connect_async;

pub struct LoginOutcome {
    pub session: Session,
    pub landed_url: String,
}

pub async fn login_in_browser(mode: &str, timeout_secs: u64) -> Result<LoginOutcome> {
    let endpoints = Endpoints::from_mode(mode);
    let chrome = find_chrome().context("未找到 Chrome / Chromium，无法打开登录页")?;
    let port = pick_free_port()?;
    let profile = data_dir()?.join("chrome-profile");
    std::fs::create_dir_all(&profile)?;

    let mut child = launch_chrome(&chrome, port, &profile, &endpoints.login_url)?;
    let result = wait_and_capture(port, timeout_secs, &endpoints.mode).await;
    let _ = child.kill();
    result
}

async fn wait_and_capture(port: u16, timeout_secs: u64, mode: &str) -> Result<LoginOutcome> {
    let deadline = Instant::now() + Duration::from_secs(timeout_secs.max(30));
    let http = reqwest::Client::builder()
        .timeout(Duration::from_secs(3))
        .build()?;

    let mut last_url = String::new();
    while Instant::now() < deadline {
        if let Ok(pages) = list_pages(&http, port).await {
            for page in pages {
                let url = page.get("url").and_then(|v| v.as_str()).unwrap_or("");
                last_url = url.to_string();
                if is_jwgl_url(url) {
                    if let Some(ws) = page.get("webSocketDebuggerUrl").and_then(|v| v.as_str()) {
                        let cookies = cdp_cookies(ws).await?;
                        if cookies.iter().any(|c| !c.value.is_empty()) {
                            return Ok(LoginOutcome {
                                session: Session::new(mode, cookies),
                                landed_url: url.to_string(),
                            });
                        }
                    }
                }
            }
        }
        sleep(Duration::from_millis(800)).await;
    }
    bail!("等待登录超时。最后停留在: {last_url}。请在打开的 Chrome 窗口完成统一认证（含手机验证码），直到跳进教务系统。");
}

fn is_jwgl_url(url: &str) -> bool {
    let u = url.to_ascii_lowercase();
    (u.contains("jwgl") || u.contains("jsxsd")) && !u.contains("/cas/login")
}

async fn list_pages(http: &reqwest::Client, port: u16) -> Result<Vec<Value>> {
    let body = http
        .get(format!("http://127.0.0.1:{port}/json"))
        .send()
        .await?
        .error_for_status()?
        .json::<Value>()
        .await?;
    Ok(body.as_array().cloned().unwrap_or_default())
}

async fn cdp_cookies(ws_url: &str) -> Result<Vec<Cookie>> {
    let (mut ws, _) = connect_async(ws_url).await.context("连接 Chrome DevTools 失败")?;
    let req = json!({"id": 1, "method": "Network.getAllCookies"});
    ws.send(tokio_tungstenite::tungstenite::Message::Text(req.to_string().into()))
        .await?;

    let deadline = Instant::now() + Duration::from_secs(8);
    while Instant::now() < deadline {
        let Some(msg) = ws.next().await else { break };
        let msg = msg?;
        let text = match msg {
            tokio_tungstenite::tungstenite::Message::Text(t) => t.to_string(),
            _ => continue,
        };
        let v: Value = serde_json::from_str(&text)?;
        if v.get("id") == Some(&json!(1)) {
            let cookies = v["result"]["cookies"]
                .as_array()
                .cloned()
                .unwrap_or_default()
                .into_iter()
                .filter_map(|c| {
                    Some(Cookie {
                        name: c.get("name")?.as_str()?.to_string(),
                        value: c.get("value")?.as_str()?.to_string(),
                        domain: c.get("domain").and_then(|x| x.as_str()).unwrap_or("").to_string(),
                        path: c.get("path").and_then(|x| x.as_str()).unwrap_or("/").to_string(),
                    })
                })
                .collect();
            return Ok(cookies);
        }
    }
    bail!("Chrome 未返回 cookie")
}

fn launch_chrome(chrome: &PathBuf, port: u16, profile: &PathBuf, login_url: &str) -> Result<Child> {
    let child = Command::new(chrome)
        .arg(format!("--remote-debugging-port={port}"))
        .arg(format!("--user-data-dir={}", profile.display()))
        .arg("--no-first-run")
        .arg("--no-default-browser-check")
        .arg("--disable-sync")
        .arg("--new-window")
        .arg(login_url)
        .stdin(Stdio::null())
        .stdout(Stdio::null())
        .stderr(Stdio::null())
        .spawn()
        .context("启动 Chrome 失败")?;
    Ok(child)
}

fn find_chrome() -> Result<PathBuf> {
    let candidates = [
        "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
        "/Applications/Chromium.app/Contents/MacOS/Chromium",
        "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
        "/usr/bin/google-chrome",
        "/usr/bin/chromium",
        "/usr/bin/chromium-browser",
        "google-chrome",
        "chromium",
        "chrome",
    ];
    for c in candidates {
        let path = PathBuf::from(c);
        if path.exists() {
            return Ok(path);
        }
        if Command::new(c).arg("--version").output().is_ok() {
            return Ok(PathBuf::from(c));
        }
    }
    bail!("未找到 Chrome")
}

fn pick_free_port() -> Result<u16> {
    let listener = TcpListener::bind("127.0.0.1:0")?;
    Ok(listener.local_addr()?.port())
}

pub fn open_login_page_only(mode: &str) -> Result<String> {
    let endpoints = Endpoints::from_mode(mode);
    #[cfg(target_os = "macos")]
    {
        Command::new("open").arg(&endpoints.login_url).status().ok();
    }
    #[cfg(target_os = "linux")]
    {
        Command::new("xdg-open").arg(&endpoints.login_url).status().ok();
    }
    #[cfg(target_os = "windows")]
    {
        Command::new("cmd").args(["/C", "start", "", &endpoints.login_url]).status().ok();
    }
    Ok(endpoints.login_url)
}
