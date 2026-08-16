use anyhow::Result;

#[tokio::main]
async fn main() -> Result<()> {
    csuft_jw_mcp::mcp::serve_stdio().await
}
