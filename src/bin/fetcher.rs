use anyhow::Result;
use dotenv::dotenv;
use genje::{DatabaseService, RssParser, Settings, SourceType, WebScraper};
use sqlx::PgPool;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::{interval, sleep};
use tracing::{error, info, warn};
use tracing_subscriber;

#[tokio::main]
async fn main() -> Result<()> {
    // Load environment variables
    dotenv().ok();

    // Load configuration
    let settings = Settings::new()?;

    // Initialize logging
    let subscriber = tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::new(&settings.logging.level))
        .finish();
    tracing::subscriber::set_global_default(subscriber)?;

    info!("Starting Genje News Fetcher");

    // Initialize database connection
    let pool = PgPool::connect(settings.database_url()).await?;
    let db_service = Arc::new(DatabaseService::new(pool));

    // Initialize parsers
    let rss_parser = Arc::new(RssParser::new());
    let web_scraper = Arc::new(WebScraper::new());

    // Create fetcher
    let fetcher = NewsFetcher::new(db_service, rss_parser, web_scraper, settings.fetcher);

    info!("News fetcher initialized, starting periodic fetching...");

    // Run the fetcher
    fetcher.run().await
}

struct NewsFetcher {
    db_service: Arc<DatabaseService>,
    rss_parser: Arc<RssParser>,
    web_scraper: Arc<WebScraper>,
    config: genje::FetcherSettings,
}

impl NewsFetcher {
    fn new(
        db_service: Arc<DatabaseService>,
        rss_parser: Arc<RssParser>,
        web_scraper: Arc<WebScraper>,
        config: genje::FetcherSettings,
    ) -> Self {
        Self {
            db_service,
            rss_parser,
            web_scraper,
            config,
        }
    }

    async fn run(&self) -> Result<()> {
        let mut interval = interval(Duration::from_secs(self.config.fetch_interval_minutes * 60));

        // Run initial fetch
        self.fetch_all_sources().await;

        loop {
            interval.tick().await;
            self.fetch_all_sources().await;
        }
    }

    async fn fetch_all_sources(&self) {
        info!("Starting news fetch cycle");

        match self.db_service.get_active_news_sources().await {
            Ok(sources) => {
                info!("Found {} active news sources", sources.len());

                // Process all sources sequentially to avoid Send/Sync issues with scraper
                for source in sources {
                    Self::fetch_from_source(
                        &source,
                        &self.db_service,
                        &self.rss_parser,
                        &self.web_scraper,
                        self.config.max_articles_per_source,
                    )
                    .await;
                }
            }
            Err(e) => error!("Failed to get news sources: {}", e),
        }

        info!("News fetch cycle completed");
    }

    async fn fetch_from_source(
        source: &genje::NewsSource,
        db_service: &DatabaseService,
        rss_parser: &RssParser,
        web_scraper: &WebScraper,
        max_articles: usize,
    ) {
        info!("Fetching from source: {}", source.name);

        let articles = match source.source_type {
            SourceType::RSS => {
                match rss_parser.fetch_and_parse(source).await {
                    Ok(articles) => articles,
                    Err(e) => {
                        error!("Failed to fetch RSS from {}: {}", source.name, e);
                        return;
                    }
                }
            }
            SourceType::WebScraping => {
                match web_scraper.scrape_articles(source).await {
                    Ok(articles) => articles,
                    Err(e) => {
                        error!("Failed to scrape from {}: {}", source.name, e);
                        return;
                    }
                }
            }
            SourceType::API => {
                warn!("API source type not yet implemented for {}", source.name);
                return;
            }
        };

        info!("Found {} articles from {}", articles.len(), source.name);

        // Limit the number of articles and save them
        let mut saved_count = 0;
        for article in articles.into_iter().take(max_articles) {
            // Check if article already exists
            match db_service.article_exists(&article.url).await {
                Ok(true) => {
                    // Article already exists, skip
                    continue;
                }
                Ok(false) => {
                    // New article, save it
                    match db_service.create_article(&article).await {
                        Ok(_) => {
                            saved_count += 1;
                        }
                        Err(e) => {
                            error!("Failed to save article from {}: {}", source.name, e);
                        }
                    }
                }
                Err(e) => {
                    error!("Failed to check if article exists: {}", e);
                    continue;
                }
            }
        }

        info!("Saved {} new articles from {}", saved_count, source.name);

        // Update last fetched timestamp
        if let Err(e) = db_service.update_source_last_fetched(source.id).await {
            error!("Failed to update last_fetched for {}: {}", source.name, e);
        }

        // Add a small delay between sources to be respectful
        sleep(Duration::from_secs(2)).await;
    }
}