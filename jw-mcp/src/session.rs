use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Cookie {
    pub name: String,
    pub value: String,
    #[serde(default)]
    pub domain: String,
    #[serde(default = "default_path")]
    pub path: String,
}

fn default_path() -> String {
    "/".to_string()
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Session {
    pub mode: String,
    pub cookies: Vec<Cookie>,
    pub updated_at: DateTime<Utc>,
}

impl Session {
    pub fn new(mode: &str, cookies: Vec<Cookie>) -> Self {
        Self {
            mode: mode.to_string(),
            cookies,
            updated_at: Utc::now(),
        }
    }

    pub fn is_empty(&self) -> bool {
        self.cookies.is_empty()
    }

    pub fn cookie_header_for(&self, host: &str) -> String {
        self.cookies
            .iter()
            .filter(|c| cookie_matches_host(&c.domain, host))
            .map(|c| format!("{}={}", c.name, c.value))
            .collect::<Vec<_>>()
            .join("; ")
    }
}

fn cookie_matches_host(domain: &str, host: &str) -> bool {
    if domain.is_empty() {
        return true;
    }
    let domain = domain.trim_start_matches('.');
    host == domain || host.ends_with(&format!(".{domain}"))
}

pub fn data_dir() -> Result<PathBuf> {
    if let Ok(custom) = std::env::var("CSUFT_JW_MCP_DIR") {
        if !custom.is_empty() {
            let dir = PathBuf::from(custom);
            fs::create_dir_all(&dir)?;
            return Ok(dir);
        }
    }
    let dir = dirs::home_dir()
        .context("cannot resolve home directory")?
        .join(".csuft-jw-mcp");
    fs::create_dir_all(&dir)?;
    Ok(dir)
}

pub fn session_path() -> Result<PathBuf> {
    Ok(data_dir()?.join("session.json"))
}

pub fn load_session() -> Result<Option<Session>> {
    let path = session_path()?;
    if !path.exists() {
        return Ok(None);
    }
    let raw = fs::read_to_string(&path)?;
    let session: Session = serde_json::from_str(&raw)?;
    Ok(Some(session))
}

pub fn save_session(session: &Session) -> Result<()> {
    let path = session_path()?;
    fs::write(path, serde_json::to_string_pretty(session)?)?;
    Ok(())
}

pub fn clear_session() -> Result<bool> {
    let path = session_path()?;
    if path.exists() {
        fs::remove_file(path)?;
        return Ok(true);
    }
    Ok(false)
}

#[derive(Debug, Clone)]
pub struct Endpoints {
    pub mode: String,
    pub login_url: String,
    pub grade_url: String,
    pub level_url: String,
    pub course_url: String,
    pub main_url: String,
}

impl Endpoints {
    pub fn from_mode(mode: &str) -> Self {
        match mode {
            "campus" => Self {
                mode: "campus".into(),
                login_url: "https://cas.csuft.edu.cn/cas/login?service=http%3A%2F%2Fjwgl.csuft.edu.cn%2Fjsxsd%2Fframework%2FxsMain.jsp".into(),
                grade_url: "http://jwgl.csuft.edu.cn/jsxsd/kscj/cjcx_list".into(),
                level_url: "http://jwgl.csuft.edu.cn/jsxsd/kscj/djkscj_list".into(),
                course_url: "http://jwgl.csuft.edu.cn/jsxsd/xskb/xskb_list.do".into(),
                main_url: "http://jwgl.csuft.edu.cn/jsxsd/framework/xsMain.jsp".into(),
            },
            _ => Self {
                mode: "webvpn".into(),
                login_url: "https://https-cas-csuft-edu-cn-443.webvpn.csuft.edu.cn/cas/login?service=https%3A%2F%2Fhttp-jwgl-csuft-edu-cn-80.webvpn.csuft.edu.cn%2Fjsxsd%2Fframework%2FxsMain.jsp".into(),
                grade_url: "https://http-jwgl-csuft-edu-cn-80.webvpn.csuft.edu.cn/jsxsd/kscj/cjcx_list".into(),
                level_url: "https://http-jwgl-csuft-edu-cn-80.webvpn.csuft.edu.cn/jsxsd/kscj/djkscj_list".into(),
                course_url: "https://http-jwgl-csuft-edu-cn-80.webvpn.csuft.edu.cn/jsxsd/xskb/xskb_list.do".into(),
                main_url: "https://http-jwgl-csuft-edu-cn-80.webvpn.csuft.edu.cn/jsxsd/framework/xsMain.jsp".into(),
            },
        }
    }
}
