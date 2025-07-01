use anyhow::Result;
use dotenv::dotenv;
use genje::{DatabaseService, Settings};
use sqlx::PgPool;
use std::sync::Arc;
use tokio::net::TcpListener;
use tracing::{info, warn};
use tracing_subscriber;

mod api;

use api::create_router;

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

    info!("Starting Genje News Aggregator");
    info!("Configuration loaded: {:?}", settings);

    // Initialize database connection
    info!("Connecting to database: {}", settings.database_url());
    let pool = PgPool::connect(settings.database_url()).await?;

    // Run database migrations
    info!("Running database migrations...");
    sqlx::migrate!("./migrations").run(&pool).await?;

    // Initialize database service
    let db_service = Arc::new(DatabaseService::new(pool));

    // Initialize news sources if none exist
    initialize_news_sources(&db_service).await?;

    // Create the API router
    let app = create_router(db_service);

    // Start the server
    let listener = TcpListener::bind(&settings.server_address()).await?;
    info!("Server starting on {}", settings.server_address());

    axum::serve(listener, app).await?;

    Ok(())
}

async fn initialize_news_sources(db_service: &DatabaseService) -> Result<()> {
    use genje::get_kenyan_news_sources;

    // Check if we already have sources
    let existing_sources = db_service.get_active_news_sources().await?;
    if !existing_sources.is_empty() {
        info!("Found {} existing news sources", existing_sources.len());
        return Ok(());
    }

    info!("Initializing default Kenyan news sources...");
    let default_sources = get_kenyan_news_sources();

    for source in default_sources {
        match db_service.create_news_source(&source).await {
            Ok(_) => info!("Added news source: {}", source.name),
            Err(e) => warn!("Failed to add news source {}: {}", source.name, e),
        }
    }

    info!("News sources initialization completed");
    Ok(())
}
