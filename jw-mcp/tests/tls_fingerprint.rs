use csuft_jw_mcp::client::{build_http_client, probe_chrome_tls, tls_probe_is_chrome};

#[test]
fn shipped_client_uses_chrome_emulation_builder() {
    build_http_client().expect("chrome TLS client should build");
}

#[tokio::test]
async fn shipped_client_tls_probe_matches_chrome() {
    let probe = match probe_chrome_tls().await {
        Ok(p) => p,
        Err(err) => {
            eprintln!("skip live TLS inspector: {err}");
            return;
        }
    };
    tls_probe_is_chrome(&probe).unwrap_or_else(|err| {
        panic!(
            "shipped client TLS fingerprint is not Chrome: {err}; ua={} ja3={} ja4={} ver={}",
            probe.user_agent, probe.ja3_hash(), probe.ja4(), probe.tls.tls_version_negotiated
        )
    });
}
