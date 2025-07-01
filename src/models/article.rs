use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Article {
    pub id: Uuid,
    pub title: String,
    pub content: Option<String>,
    pub summary: Option<String>,
    pub url: String,
    pub author: Option<String>,
    pub published_at: DateTime<Utc>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub source_id: Uuid,
    pub category: Option<String>,
    pub tags: Vec<String>,
    pub is_trending: bool,
    pub view_count: i64,
    pub language: String, // "en", "sw" for Swahili, etc.
    pub region: Option<String>, // Kenyan regions: "nairobi", "mombasa", "kisumu", etc.
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewArticle {
    pub title: String,
    pub content: Option<String>,
    pub summary: Option<String>,
    pub url: String,
    pub author: Option<String>,
    pub published_at: DateTime<Utc>,
    pub source_id: Uuid,
    pub category: Option<String>,
    pub tags: Vec<String>,
    pub language: String,
    pub region: Option<String>,
}

impl NewArticle {
    pub fn into_article(self) -> Article {
        let now = Utc::now();
        Article {
            id: Uuid::new_v4(),
            title: self.title,
            content: self.content,
            summary: self.summary,
            url: self.url,
            author: self.author,
            published_at: self.published_at,
            created_at: now,
            updated_at: now,
            source_id: self.source_id,
            category: self.category,
            tags: self.tags,
            is_trending: false,
            view_count: 0,
            language: self.language,
            region: self.region,
        }
    }
} 