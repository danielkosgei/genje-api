use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewsSource {
    pub id: Uuid,
    pub name: String,
    pub base_url: String,
    pub rss_url: Option<String>,
    pub scrape_config: Option<ScrapingConfig>,
    pub source_type: SourceType,
    pub is_active: bool,
    pub last_fetched: Option<DateTime<Utc>>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub language: String,
    pub country: String,
    pub region: Option<String>,
    pub credibility_score: Option<f32>, // 0-100 credibility rating
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SourceType {
    RSS,
    WebScraping,
    API,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScrapingConfig {
    pub article_selector: String,
    pub title_selector: String,
    pub content_selector: String,
    pub author_selector: Option<String>,
    pub date_selector: Option<String>,
    pub date_format: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewNewsSource {
    pub name: String,
    pub base_url: String,
    pub rss_url: Option<String>,
    pub scrape_config: Option<ScrapingConfig>,
    pub source_type: SourceType,
    pub language: String,
    pub country: String,
    pub region: Option<String>,
    pub credibility_score: Option<f32>,
}

impl NewNewsSource {
    pub fn into_news_source(self) -> NewsSource {
        let now = Utc::now();
        NewsSource {
            id: Uuid::new_v4(),
            name: self.name,
            base_url: self.base_url,
            rss_url: self.rss_url,
            scrape_config: self.scrape_config,
            source_type: self.source_type,
            is_active: true,
            last_fetched: None,
            created_at: now,
            updated_at: now,
            language: self.language,
            country: self.country,
            region: self.region,
            credibility_score: self.credibility_score,
        }
    }
}

// Common Kenyan news sources
pub fn get_kenyan_news_sources() -> Vec<NewNewsSource> {
    vec![
        NewNewsSource {
            name: "Daily Nation".to_string(),
            base_url: "https://nation.africa".to_string(),
            rss_url: Some("https://nation.africa/kenya/rss".to_string()),
            scrape_config: None,
            source_type: SourceType::RSS,
            language: "en".to_string(),
            country: "kenya".to_string(),
            region: None,
            credibility_score: Some(85.0),
        },
        NewNewsSource {
            name: "The Standard".to_string(),
            base_url: "https://www.standardmedia.co.ke".to_string(),
            rss_url: Some("https://www.standardmedia.co.ke/rss/headlines.php".to_string()),
            scrape_config: None,
            source_type: SourceType::RSS,
            language: "en".to_string(),
            country: "kenya".to_string(),
            region: None,
            credibility_score: Some(80.0),
        },
        NewNewsSource {
            name: "Citizen Digital".to_string(),
            base_url: "https://citizentv.co.ke".to_string(),
            rss_url: Some("https://citizentv.co.ke/feed/".to_string()),
            scrape_config: None,
            source_type: SourceType::RSS,
            language: "en".to_string(),
            country: "kenya".to_string(),
            region: None,
            credibility_score: Some(82.0),
        },
        NewNewsSource {
            name: "Capital FM".to_string(),
            base_url: "https://www.capitalfm.co.ke".to_string(),
            rss_url: Some("https://www.capitalfm.co.ke/news/feed/".to_string()),
            scrape_config: None,
            source_type: SourceType::RSS,
            language: "en".to_string(),
            country: "kenya".to_string(),
            region: None,
            credibility_score: Some(78.0),
        },
    ]
} 