use anyhow::{anyhow, Result};
use chrono::{DateTime, Utc};
use reqwest::Client;
use rss::Channel;
use uuid::Uuid;

use crate::models::{NewArticle, NewsSource};

pub struct RssParser {
    client: Client,
}

impl RssParser {
    pub fn new() -> Self {
        Self {
            client: Client::new(),
        }
    }

    pub async fn fetch_and_parse(&self, source: &NewsSource) -> Result<Vec<NewArticle>> {
        let rss_url = source
            .rss_url
            .as_ref()
            .ok_or_else(|| anyhow!("No RSS URL configured for source: {}", source.name))?;

        tracing::info!("Fetching RSS feed from: {}", rss_url);

        let response = self
            .client
            .get(rss_url)
            .header("User-Agent", "GenjeNewsAggregator/1.0")
            .send()
            .await?;

        let rss_content = response.text().await?;
        let channel = Channel::read_from(rss_content.as_bytes())?;

        let mut articles = Vec::new();

        for item in channel.items() {
            match self.parse_rss_item(item, source.id).await {
                Ok(article) => articles.push(article),
                Err(e) => {
                    tracing::warn!("Failed to parse RSS item: {}", e);
                    continue;
                }
            }
        }

        tracing::info!(
            "Successfully parsed {} articles from {}",
            articles.len(),
            source.name
        );

        Ok(articles)
    }

    async fn parse_rss_item(&self, item: &rss::Item, source_id: Uuid) -> Result<NewArticle> {
        let title = item
            .title()
            .ok_or_else(|| anyhow!("RSS item missing title"))?
            .to_string();

        let url = item
            .link()
            .ok_or_else(|| anyhow!("RSS item missing link"))?
            .to_string();

        let published_at = if let Some(pub_date) = item.pub_date() {
            self.parse_date(pub_date)?
        } else {
            Utc::now()
        };

        let content = item.content().map(|c| c.to_string());
        let summary = item.description().map(|d| {
            // Clean up HTML tags from description
            self.clean_html_content(d)
        });

        let author = item.author().map(|a| a.to_string());

        // Extract tags from categories
        let tags: Vec<String> = item
            .categories()
            .iter()
            .map(|cat| cat.name().to_string())
            .collect();

        // Determine language and region based on content analysis
        let language = self.detect_language(&title, summary.as_deref()).unwrap_or_else(|| "en".to_string());
        let region = self.detect_kenyan_region(&title, summary.as_deref());
        let category = self.categorize_article(&title, summary.as_deref());

        Ok(NewArticle {
            title,
            content,
            summary,
            url,
            author,
            published_at,
            source_id,
            category,
            tags,
            language,
            region,
        })
    }

    fn parse_date(&self, date_str: &str) -> Result<DateTime<Utc>> {
        // Try multiple date formats commonly used in RSS feeds
        let formats = [
            "%a, %d %b %Y %H:%M:%S %z",
            "%a, %d %b %Y %H:%M:%S GMT",
            "%Y-%m-%dT%H:%M:%S%z",
            "%Y-%m-%d %H:%M:%S",
        ];

        for format in &formats {
            if let Ok(dt) = DateTime::parse_from_str(date_str, format) {
                return Ok(dt.with_timezone(&Utc));
            }
        }

        // If all parsing fails, try chrono's built-in parsing
        date_str
            .parse::<DateTime<Utc>>()
            .map_err(|_| anyhow!("Unable to parse date: {}", date_str))
    }

    fn clean_html_content(&self, html: &str) -> String {
        // Basic HTML tag removal - in a production app, consider using a proper HTML parser
        html.replace(&['<', '>'][..], "")
            .chars()
            .filter(|c| !c.is_control())
            .collect::<String>()
            .trim()
            .to_string()
    }

    fn detect_language(&self, title: &str, summary: Option<&str>) -> Option<String> {
        let content = format!("{} {}", title, summary.unwrap_or(""));
        let content_lower = content.to_lowercase();

        // Simple language detection for Swahili vs English
        let swahili_keywords = [
            "habari", "serikali", "rais", "wabunge", "mkuu", "mkoa", "wilaya",
            "shule", "hospitali", "polisi", "uchumi", "biashara", "kisiasa",
        ];

        let swahili_count = swahili_keywords
            .iter()
            .filter(|&word| content_lower.contains(word))
            .count();

        if swahili_count > 2 {
            Some("sw".to_string())
        } else {
            Some("en".to_string())
        }
    }

    fn detect_kenyan_region(&self, title: &str, summary: Option<&str>) -> Option<String> {
        let content = format!("{} {}", title, summary.unwrap_or("")).to_lowercase();

        let regions = [
            ("nairobi", "nairobi"),
            ("mombasa", "mombasa"),
            ("kisumu", "kisumu"),
            ("nakuru", "nakuru"),
            ("eldoret", "eldoret"),
            ("thika", "kiambu"),
            ("malindi", "kilifi"),
            ("garissa", "garissa"),
            ("kakamega", "kakamega"),
            ("kitale", "trans-nzoia"),
        ];

        for (keyword, region) in &regions {
            if content.contains(keyword) {
                return Some(region.to_string());
            }
        }

        None
    }

    fn categorize_article(&self, title: &str, summary: Option<&str>) -> Option<String> {
        let content = format!("{} {}", title, summary.unwrap_or("")).to_lowercase();

        let categories = [
            (vec!["politics", "political", "parliament", "election", "government", "president"], "politics"),
            (vec!["business", "economy", "economic", "trade", "market", "finance"], "business"),
            (vec!["sports", "football", "rugby", "athletics", "olympics"], "sports"),
            (vec!["health", "medical", "hospital", "doctor", "disease", "covid"], "health"),
            (vec!["technology", "tech", "digital", "innovation", "startup"], "technology"),
            (vec!["education", "school", "university", "student", "teacher"], "education"),
            (vec!["entertainment", "music", "film", "celebrity", "arts"], "entertainment"),
        ];

        for (keywords, category) in &categories {
            if keywords.iter().any(|&keyword| content.contains(keyword)) {
                return Some(category.to_string());
            }
        }

        None
    }
} 