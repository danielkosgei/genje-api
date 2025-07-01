use anyhow::Result;
use reqwest::{Client, ClientBuilder};
use std::time::Duration;

pub struct HttpClient {
    client: Client,
}

impl HttpClient {
    pub fn new() -> Result<Self> {
        let client = ClientBuilder::new()
            .user_agent("GenjeNewsAggregator/1.0 (+https://genje.ke)")
            .timeout(Duration::from_secs(30))
            .connect_timeout(Duration::from_secs(10))
            .tcp_keepalive(Duration::from_secs(60))
            .pool_max_idle_per_host(10)
            .build()?;

        Ok(Self { client })
    }

    pub fn client(&self) -> &Client {
        &self.client
    }

    pub async fn get_text(&self, url: &str) -> Result<String> {
        tracing::debug!("Fetching URL: {}", url);
        
        let response = self
            .client
            .get(url)
            .send()
            .await?
            .error_for_status()?;

        let text = response.text().await?;
        Ok(text)
    }

    pub async fn get_with_retry(&self, url: &str, max_retries: u32) -> Result<String> {
        let mut last_error = None;
        
        for attempt in 1..=max_retries {
            match self.get_text(url).await {
                Ok(content) => return Ok(content),
                Err(e) => {
                    tracing::warn!("Attempt {} failed for URL {}: {}", attempt, url, e);
                    last_error = Some(e);
                    
                    if attempt < max_retries {
                        // Exponential backoff
                        let delay = Duration::from_secs(2_u64.pow(attempt - 1));
                        tokio::time::sleep(delay).await;
                    }
                }
            }
        }

        Err(last_error.unwrap())
    }
} 