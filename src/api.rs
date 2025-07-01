use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use uuid::Uuid;

use crate::DatabaseService;

pub fn create_router(db_service: Arc<DatabaseService>) -> Router {
    Router::new()
        .route("/", get(root))
        .route("/health", get(health))
        .route("/api/articles", get(get_articles))
        .route("/api/articles/recent", get(get_recent_articles))
        .route("/api/articles/trending", get(get_trending_articles))
        .route("/api/articles/search", get(search_articles))
        .route("/api/articles/:id/view", post(increment_article_views))
        .route("/api/sources", get(get_sources))
        .with_state(db_service)
}

async fn root() -> &'static str {
    "Genje - Kenyan News Aggregator"
}

async fn health() -> Json<serde_json::Value> {
    Json(serde_json::json!({
        "status": "healthy",
        "service": "genje-api",
        "timestamp": chrono::Utc::now().to_rfc3339()
    }))
}

#[derive(Deserialize)]
struct ArticleQuery {
    category: Option<String>,
    language: Option<String>,
    region: Option<String>,
    limit: Option<i64>,
    offset: Option<i64>,
}

async fn get_articles(
    State(db): State<Arc<DatabaseService>>,
    Query(params): Query<ArticleQuery>,
) -> Result<Json<ArticlesResponse>, StatusCode> {
    let limit = params.limit.unwrap_or(20).min(100);
    let offset = params.offset.unwrap_or(0);

    match db
        .get_articles_by_filters(
            params.category.as_deref(),
            params.language.as_deref(),
            params.region.as_deref(),
            limit,
            offset,
        )
        .await
    {
        Ok(articles) => Ok(Json(ArticlesResponse { articles })),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}

#[derive(Deserialize)]
struct LimitQuery {
    limit: Option<i64>,
}

async fn get_recent_articles(
    State(db): State<Arc<DatabaseService>>,
    Query(params): Query<LimitQuery>,
) -> Result<Json<ArticlesResponse>, StatusCode> {
    let limit = params.limit.unwrap_or(20).min(100);

    match db.get_recent_articles(limit).await {
        Ok(articles) => Ok(Json(ArticlesResponse { articles })),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}

async fn get_trending_articles(
    State(db): State<Arc<DatabaseService>>,
    Query(params): Query<LimitQuery>,
) -> Result<Json<ArticlesResponse>, StatusCode> {
    let limit = params.limit.unwrap_or(20).min(100);

    match db.get_trending_articles(limit).await {
        Ok(articles) => Ok(Json(ArticlesResponse { articles })),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}

#[derive(Deserialize)]
struct SearchQuery {
    q: String,
    limit: Option<i64>,
}

async fn search_articles(
    State(db): State<Arc<DatabaseService>>,
    Query(params): Query<SearchQuery>,
) -> Result<Json<ArticlesResponse>, StatusCode> {
    let limit = params.limit.unwrap_or(20).min(100);

    match db.search_articles(&params.q, limit).await {
        Ok(articles) => Ok(Json(ArticlesResponse { articles })),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}

async fn increment_article_views(
    State(db): State<Arc<DatabaseService>>,
    Path(id): Path<Uuid>,
) -> Result<Json<serde_json::Value>, StatusCode> {
    match db.increment_article_views(id).await {
        Ok(_) => Ok(Json(serde_json::json!({"success": true}))),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}

async fn get_sources(
    State(db): State<Arc<DatabaseService>>,
) -> Result<Json<SourcesResponse>, StatusCode> {
    match db.get_active_news_sources().await {
        Ok(sources) => Ok(Json(SourcesResponse { sources })),
        Err(_) => Err(StatusCode::INTERNAL_SERVER_ERROR),
    }
}

#[derive(Serialize)]
struct ArticlesResponse {
    articles: Vec<genje::Article>,
}

#[derive(Serialize)]
struct SourcesResponse {
    sources: Vec<genje::NewsSource>,
} 