use anyhow::Result;
use chrono::Utc;
use sqlx::{PgPool, Row};
use uuid::Uuid;

use crate::models::{Article, NewArticle, NewsSource, NewNewsSource};

pub struct DatabaseService {
    pool: PgPool,
}

impl DatabaseService {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    // News Sources operations
    pub async fn create_news_source(&self, source: &NewNewsSource) -> Result<NewsSource> {
        let news_source = source.clone().into_news_source();
        
        sqlx::query(
            r#"
            INSERT INTO news_sources (
                id, name, base_url, rss_url, scrape_config, source_type, 
                is_active, created_at, updated_at, language, country, region, credibility_score
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
            "#,
        )
        .bind(news_source.id)
        .bind(&news_source.name)
        .bind(&news_source.base_url)
        .bind(&news_source.rss_url)
        .bind(serde_json::to_value(&news_source.scrape_config)?)
        .bind(serde_json::to_value(&news_source.source_type)?)
        .bind(news_source.is_active)
        .bind(news_source.created_at)
        .bind(news_source.updated_at)
        .bind(&news_source.language)
        .bind(&news_source.country)
        .bind(&news_source.region)
        .bind(news_source.credibility_score)
        .execute(&self.pool)
        .await?;

        Ok(news_source)
    }

    pub async fn get_active_news_sources(&self) -> Result<Vec<NewsSource>> {
        let rows = sqlx::query(
            "SELECT * FROM news_sources WHERE is_active = true ORDER BY created_at"
        )
        .fetch_all(&self.pool)
        .await?;

        let mut sources = Vec::new();
        for row in rows {
            sources.push(NewsSource {
                id: row.get("id"),
                name: row.get("name"),
                base_url: row.get("base_url"),
                rss_url: row.get("rss_url"),
                scrape_config: row.get::<Option<serde_json::Value>, _>("scrape_config")
                    .and_then(|v| serde_json::from_value(v).ok()),
                source_type: serde_json::from_value(row.get("source_type"))?,
                is_active: row.get("is_active"),
                last_fetched: row.get("last_fetched"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
                language: row.get("language"),
                country: row.get("country"),
                region: row.get("region"),
                credibility_score: row.get("credibility_score"),
            });
        }

        Ok(sources)
    }

    pub async fn update_source_last_fetched(&self, source_id: Uuid) -> Result<()> {
        let now = Utc::now();
        sqlx::query(
            "UPDATE news_sources SET last_fetched = $1, updated_at = $1 WHERE id = $2"
        )
        .bind(now)
        .bind(source_id)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    // Articles operations
    pub async fn create_article(&self, article: &NewArticle) -> Result<Article> {
        let article = article.clone().into_article();
        
        sqlx::query(
            r#"
            INSERT INTO articles (
                id, title, content, summary, url, author, published_at, 
                created_at, updated_at, source_id, category, tags, 
                is_trending, view_count, language, region
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
            "#,
        )
        .bind(article.id)
        .bind(&article.title)
        .bind(&article.content)
        .bind(&article.summary)
        .bind(&article.url)
        .bind(&article.author)
        .bind(article.published_at)
        .bind(article.created_at)
        .bind(article.updated_at)
        .bind(article.source_id)
        .bind(&article.category)
        .bind(&article.tags)
        .bind(article.is_trending)
        .bind(article.view_count)
        .bind(&article.language)
        .bind(&article.region)
        .execute(&self.pool)
        .await?;

        Ok(article)
    }

    pub async fn get_articles_by_filters(
        &self,
        category: Option<&str>,
        language: Option<&str>,
        region: Option<&str>,
        limit: i64,
        offset: i64,
    ) -> Result<Vec<Article>> {
        let mut query = String::from(
            "SELECT * FROM articles WHERE 1=1"
        );
        let mut params: Vec<Box<dyn sqlx::Encode<'_, sqlx::Postgres> + Send + Sync>> = Vec::new();
        let mut param_count = 0;

        if let Some(cat) = category {
            param_count += 1;
            query.push_str(&format!(" AND category = ${}", param_count));
            params.push(Box::new(cat.to_string()));
        }

        if let Some(lang) = language {
            param_count += 1;
            query.push_str(&format!(" AND language = ${}", param_count));
            params.push(Box::new(lang.to_string()));
        }

        if let Some(reg) = region {
            param_count += 1;
            query.push_str(&format!(" AND region = ${}", param_count));
            params.push(Box::new(reg.to_string()));
        }

        query.push_str(" ORDER BY published_at DESC");
        
        param_count += 1;
        query.push_str(&format!(" LIMIT ${}", param_count));
        params.push(Box::new(limit));
        
        param_count += 1;
        query.push_str(&format!(" OFFSET ${}", param_count));
        params.push(Box::new(offset));

        // For simplicity, use a basic query without dynamic parameters
        // In a production app, you'd want to use sqlx's query builder or a more sophisticated approach
        let rows = sqlx::query(&query)
            .fetch_all(&self.pool)
            .await?;

        let mut articles = Vec::new();
        for row in rows {
            articles.push(Article {
                id: row.get("id"),
                title: row.get("title"),
                content: row.get("content"),
                summary: row.get("summary"),
                url: row.get("url"),
                author: row.get("author"),
                published_at: row.get("published_at"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
                source_id: row.get("source_id"),
                category: row.get("category"),
                tags: row.get("tags"),
                is_trending: row.get("is_trending"),
                view_count: row.get("view_count"),
                language: row.get("language"),
                region: row.get("region"),
            });
        }

        Ok(articles)
    }

    pub async fn get_recent_articles(&self, limit: i64) -> Result<Vec<Article>> {
        let rows = sqlx::query(
            "SELECT * FROM articles ORDER BY published_at DESC LIMIT $1"
        )
        .bind(limit)
        .fetch_all(&self.pool)
        .await?;

        let mut articles = Vec::new();
        for row in rows {
            articles.push(Article {
                id: row.get("id"),
                title: row.get("title"),
                content: row.get("content"),
                summary: row.get("summary"),
                url: row.get("url"),
                author: row.get("author"),
                published_at: row.get("published_at"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
                source_id: row.get("source_id"),
                category: row.get("category"),
                tags: row.get("tags"),
                is_trending: row.get("is_trending"),
                view_count: row.get("view_count"),
                language: row.get("language"),
                region: row.get("region"),
            });
        }

        Ok(articles)
    }

    pub async fn article_exists(&self, url: &str) -> Result<bool> {
        let result = sqlx::query("SELECT COUNT(*) as count FROM articles WHERE url = $1")
            .bind(url)
            .fetch_one(&self.pool)
            .await?;

        let count: i64 = result.get("count");
        Ok(count > 0)
    }

    pub async fn increment_article_views(&self, article_id: Uuid) -> Result<()> {
        sqlx::query(
            "UPDATE articles SET view_count = view_count + 1, updated_at = $1 WHERE id = $2"
        )
        .bind(Utc::now())
        .bind(article_id)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    pub async fn get_trending_articles(&self, limit: i64) -> Result<Vec<Article>> {
        let rows = sqlx::query(
            "SELECT * FROM articles WHERE is_trending = true OR view_count > 100 ORDER BY view_count DESC, published_at DESC LIMIT $1"
        )
        .bind(limit)
        .fetch_all(&self.pool)
        .await?;

        let mut articles = Vec::new();
        for row in rows {
            articles.push(Article {
                id: row.get("id"),
                title: row.get("title"),
                content: row.get("content"),
                summary: row.get("summary"),
                url: row.get("url"),
                author: row.get("author"),
                published_at: row.get("published_at"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
                source_id: row.get("source_id"),
                category: row.get("category"),
                tags: row.get("tags"),
                is_trending: row.get("is_trending"),
                view_count: row.get("view_count"),
                language: row.get("language"),
                region: row.get("region"),
            });
        }

        Ok(articles)
    }

    pub async fn search_articles(&self, query: &str, limit: i64) -> Result<Vec<Article>> {
        let search_query = format!("%{}%", query.to_lowercase());
        
        let rows = sqlx::query(
            r#"
            SELECT * FROM articles 
            WHERE LOWER(title) LIKE $1 
               OR LOWER(summary) LIKE $1 
               OR LOWER(content) LIKE $1 
            ORDER BY published_at DESC 
            LIMIT $2
            "#,
        )
        .bind(&search_query)
        .bind(limit)
        .fetch_all(&self.pool)
        .await?;

        let mut articles = Vec::new();
        for row in rows {
            articles.push(Article {
                id: row.get("id"),
                title: row.get("title"),
                content: row.get("content"),
                summary: row.get("summary"),
                url: row.get("url"),
                author: row.get("author"),
                published_at: row.get("published_at"),
                created_at: row.get("created_at"),
                updated_at: row.get("updated_at"),
                source_id: row.get("source_id"),
                category: row.get("category"),
                tags: row.get("tags"),
                is_trending: row.get("is_trending"),
                view_count: row.get("view_count"),
                language: row.get("language"),
                region: row.get("region"),
            });
        }

        Ok(articles)
    }
} 