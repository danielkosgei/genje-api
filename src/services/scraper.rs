use anyhow::{anyhow, Result};
use chrono::{DateTime, Utc};
use reqwest::Client;
use scraper::{Html, Selector};

use crate::models::{NewArticle, NewsSource, ScrapingConfig};

pub struct WebScraper {
    client: Client,
}

impl WebScraper {
    pub fn new() -> Self {
        Self {
            client: Client::builder()
                .user_agent("GenjeNewsAggregator/1.0")
                .build()
                .expect("Failed to create HTTP client"),
        }
    }

    pub async fn scrape_articles(&self, source: &NewsSource) -> Result<Vec<NewArticle>> {
        let scrape_config = source
            .scrape_config
            .as_ref()
            .ok_or_else(|| anyhow!("No scraping configuration for source: {}", source.name))?;

        tracing::info!("Scraping articles from: {}", source.base_url);

        let response = self.client.get(&source.base_url).send().await?;
        let html_content = response.text().await?;
        let document = Html::parse_document(&html_content);

        let article_selector = Selector::parse(&scrape_config.article_selector)
            .map_err(|_| anyhow!("Invalid article selector: {}", scrape_config.article_selector))?;

        let mut articles = Vec::new();

        for article_element in document.select(&article_selector) {
            match self.extract_article(&article_element, source, scrape_config).await {
                Ok(article) => articles.push(article),
                Err(e) => {
                    tracing::warn!("Failed to extract article: {}", e);
                    continue;
                }
            }
        }

        tracing::info!(
            "Successfully scraped {} articles from {}",
            articles.len(),
            source.name
        );

        Ok(articles)
    }

    async fn extract_article(
        &self,
        element: &scraper::ElementRef<'_>,
        source: &NewsSource,
        config: &ScrapingConfig,
    ) -> Result<NewArticle> {
        // Extract title
        let title_selector = Selector::parse(&config.title_selector)
            .map_err(|_| anyhow!("Invalid title selector"))?;
        
        let title = element
            .select(&title_selector)
            .next()
            .and_then(|el| el.text().next())
            .ok_or_else(|| anyhow!("Could not extract title"))?
            .trim()
            .to_string();

        // Extract URL - try to find a link within the article element
        let url = self.extract_article_url(element, &source.base_url)?;

        // If we have a full article URL, fetch the full content
        let (content, summary) = if url.starts_with("http") {
            self.fetch_full_article(&url, config).await.unwrap_or_else(|_| {
                (None, self.extract_summary_from_element(element, config))
            })
        } else {
            (None, self.extract_summary_from_element(element, config))
        };

        // Extract author if configured
        let author = if let Some(author_selector_str) = &config.author_selector {
            if let Ok(author_selector) = Selector::parse(author_selector_str) {
                element
                    .select(&author_selector)
                    .next()
                    .and_then(|el| el.text().next())
                    .map(|s| s.trim().to_string())
            } else {
                None
            }
        } else {
            None
        };

        // Extract publication date
        let published_at = if let (Some(date_selector_str), Some(date_format)) = 
            (&config.date_selector, &config.date_format) {
            let date_selector = Selector::parse(date_selector_str).ok();
            if let Some(selector) = date_selector {
                element
                    .select(&selector)
                    .next()
                    .and_then(|el| el.text().next())
                    .and_then(|date_str| self.parse_date(date_str.trim(), date_format).ok())
                    .unwrap_or_else(Utc::now)
            } else {
                Utc::now()
            }
        } else {
            Utc::now()
        };

        // Basic categorization and language detection
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
            source_id: source.id,
            category,
            tags: Vec::new(), // Could be extracted from meta tags or content
            language,
            region,
        })
    }

    fn extract_article_url(&self, element: &scraper::ElementRef<'_>, base_url: &str) -> Result<String> {
        // Try to find a link in the article element
        let link_selectors = ["a", "a[href]", ".title a", ".headline a"];
        
        for selector_str in &link_selectors {
            if let Ok(selector) = Selector::parse(selector_str) {
                if let Some(link_element) = element.select(&selector).next() {
                    if let Some(href) = link_element.value().attr("href") {
                        return Ok(self.resolve_url(href, base_url));
                    }
                }
            }
        }

        Err(anyhow!("Could not find article URL"))
    }

    fn resolve_url(&self, href: &str, base_url: &str) -> String {
        if href.starts_with("http") {
            href.to_string()
        } else if href.starts_with('/') {
            format!("{}{}", base_url.trim_end_matches('/'), href)
        } else {
            format!("{}/{}", base_url.trim_end_matches('/'), href)
        }
    }

    async fn fetch_full_article(
        &self,
        url: &str,
        config: &ScrapingConfig,
    ) -> Result<(Option<String>, Option<String>)> {
        let response = self.client.get(url).send().await?;
        let html_content = response.text().await?;
        let document = Html::parse_document(&html_content);

        // Extract full content
        let content = if let Ok(content_selector) = Selector::parse(&config.content_selector) {
            document
                .select(&content_selector)
                .next()
                .map(|el| {
                    el.text()
                        .collect::<Vec<_>>()
                        .join(" ")
                        .trim()
                        .to_string()
                })
        } else {
            None
        };

        // Create summary from content if available
        let summary = content.as_ref().map(|c| {
            let words: Vec<&str> = c.split_whitespace().take(50).collect();
            words.join(" ")
        });

        Ok((content, summary))
    }

    fn extract_summary_from_element(
        &self,
        element: &scraper::ElementRef<'_>,
        _config: &ScrapingConfig,
    ) -> Option<String> {
        // Try common summary selectors
        let summary_selectors = [".summary", ".excerpt", ".description", "p"];
        
        for selector_str in &summary_selectors {
            if let Ok(selector) = Selector::parse(selector_str) {
                if let Some(summary_element) = element.select(&selector).next() {
                    let summary = summary_element
                        .text()
                        .collect::<Vec<_>>()
                        .join(" ")
                        .trim()
                        .to_string();
                    
                    if !summary.is_empty() {
                        return Some(summary);
                    }
                }
            }
        }

        None
    }

    fn parse_date(&self, date_str: &str, format: &str) -> Result<DateTime<Utc>> {
        let dt = DateTime::parse_from_str(date_str, format)
            .map_err(|_| anyhow!("Failed to parse date: {}", date_str))?;
        Ok(dt.with_timezone(&Utc))
    }

    // Reuse language detection and categorization methods from RSS parser
    fn detect_language(&self, title: &str, summary: Option<&str>) -> Option<String> {
        let content = format!("{} {}", title, summary.unwrap_or(""));
        let content_lower = content.to_lowercase();

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