use crate::session::{Endpoints, Session};
use anyhow::{bail, Context, Result};
use serde::Deserialize;
use std::time::Duration;
use wreq::header::{HeaderMap, HeaderValue, COOKIE, CONTENT_TYPE, REFERER};
use wreq::{Client, Method, Url};
use wreq_util::{Emulation, EmulationOS, EmulationOption};

/// Chrome 131 on macOS: JA3/JA4 + HTTP/2 settings match desktop Chrome, not rustls.
pub fn chrome_emulation() -> EmulationOption {
    EmulationOption::builder()
        .emulation(Emulation::Chrome131)
        .emulation_os(EmulationOS::MacOS)
        .skip_http2(false)
        .skip_headers(false)
        .build()
}

pub fn build_http_client() -> Result<Client> {
    Ok(Client::builder()
        .emulation(chrome_emulation())
        .redirect(wreq::redirect::Policy::limited(8))
        .timeout(Duration::from_secs(30))
        .connect_timeout(Duration::from_secs(15))
        .no_proxy()
        .build()?)
}

fn headers_for(session: &Session, url: &str, method: &Method) -> Result<HeaderMap> {
    let parsed = Url::parse(url)?;
    let host = parsed.host_str().unwrap_or("");
    let mut headers = HeaderMap::new();
    let cookie = session.cookie_header_for(host);
    if !cookie.is_empty() {
        headers.insert(COOKIE, HeaderValue::from_str(&cookie)?);
    }
    if *method == Method::POST {
        headers.insert(CONTENT_TYPE, HeaderValue::from_static("application/x-www-form-urlencoded"));
        headers.insert(REFERER, HeaderValue::from_str(url)?);
    }
    Ok(headers)
}

pub async fn fetch_text(
    client: &Client,
    session: &Session,
    method: Method,
    url: &str,
    form: Option<&[(&str, &str)]>,
) -> Result<String> {
    let mut req = client.request(method.clone(), url);
    req = req.headers(headers_for(session, url, &method)?);
    if let Some(form) = form {
        req = req.form(&form.iter().copied().collect::<Vec<_>>());
    }
    let resp = req.send().await.with_context(|| format!("request {url}"))?;
    let status = resp.status();
    let final_url = resp.url().to_string();
    let text = resp.text().await?;
    if status.is_client_error() || status.is_server_error() {
        bail!("教务系统返回 HTTP {status} ({final_url})");
    }
    Ok(text)
}

#[derive(Debug, Deserialize)]
pub struct TlsProbe {
    #[serde(default, rename = "user_agent")]
    pub user_agent: String,
    #[serde(default, rename = "http_version")]
    pub http_version: String,
    #[serde(default)]
    pub tls: TlsSection,
    #[serde(default)]
    pub http2: serde_json::Value,
}

impl TlsProbe {
    pub fn ja3(&self) -> &str {
        &self.tls.ja3
    }
    pub fn ja3_hash(&self) -> &str {
        &self.tls.ja3_hash
    }
    pub fn ja4(&self) -> &str {
        &self.tls.ja4
    }
}

#[derive(Debug, Default, Deserialize)]
pub struct TlsSection {
    #[serde(default)]
    pub ja3: String,
    #[serde(default)]
    pub ja3_hash: String,
    #[serde(default)]
    pub ja4: String,
    #[serde(default)]
    pub ja4_r: String,
    #[serde(default)]
    pub tls_version_negotiated: String,
}

/// Hit a public TLS inspector with the same client used for 教务请求.
pub async fn probe_chrome_tls() -> Result<TlsProbe> {
    let client = build_http_client()?;
    let resp = client
        .get("https://tls.peet.ws/api/all")
        .send()
        .await
        .context("probe tls.peet.ws")?;
    let status = resp.status();
    let body = resp.text().await?;
    if !status.is_success() {
        bail!("tls probe HTTP {status}: {body}");
    }
    serde_json::from_str(&body).with_context(|| format!("parse tls probe: {body}"))
}

pub fn tls_probe_is_chrome(probe: &TlsProbe) -> Result<()> {
    let ua = probe.user_agent.to_ascii_lowercase();
    if !ua.contains("chrome/") {
        bail!("TLS probe UA is not Chrome: {}", probe.user_agent);
    }
    if ua.contains("curl/") || ua.contains("reqwest") {
        bail!("TLS probe UA still looks like a library: {}", probe.user_agent);
    }
    if probe.http_version != "h2" {
        bail!("expected HTTP/2, got {}", probe.http_version);
    }
    let ja4 = probe.ja4().to_ascii_lowercase();
    // Chrome 131 desktop JA4 is t13d…h2 (TLS1.3 + SNI + HTTP/2). rustls/reqwest is not t13d1516h2.
    if !ja4.starts_with("t13d") || !ja4.contains("h2") {
        bail!(
            "expected Chrome TLS1.3/HTTP2 JA4 (t13d…h2), got ja4={} ja3={}",
            probe.ja4(),
            probe.ja3()
        );
    }
    Ok(())
}

pub fn endpoints_for(session: &Session) -> Endpoints {
    Endpoints::from_mode(&session.mode)
}
