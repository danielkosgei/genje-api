use config::{Config, ConfigError, Environment, File};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Settings {
    pub database: DatabaseSettings,
    pub server: ServerSettings,
    pub fetcher: FetcherSettings,
    pub logging: LoggingSettings,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DatabaseSettings {
    pub url: String,
    pub max_connections: u32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ServerSettings {
    pub host: String,
    pub port: u16,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct FetcherSettings {
    pub fetch_interval_minutes: u64,
    pub max_articles_per_source: usize,
    pub concurrent_fetches: usize,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct LoggingSettings {
    pub level: String,
}

impl Settings {
    pub fn new() -> Result<Self, ConfigError> {
        let mut config = Config::builder()
            // Default values
            .set_default("database.url", "postgresql://postgres:password@localhost:5432/genje")?
            .set_default("database.max_connections", 10)?
            .set_default("server.host", "0.0.0.0")?
            .set_default("server.port", 8080)?
            .set_default("fetcher.fetch_interval_minutes", 30)?
            .set_default("fetcher.max_articles_per_source", 50)?
            .set_default("fetcher.concurrent_fetches", 5)?
            .set_default("logging.level", "info")?;

        // Add in configuration file if it exists
        if std::path::Path::new("config.toml").exists() {
            config = config.add_source(File::with_name("config"));
        }

        // Add in environment variables (with a prefix of GENJE)
        config = config.add_source(Environment::with_prefix("GENJE").separator("_"));

        config.build()?.try_deserialize()
    }

    pub fn database_url(&self) -> &str {
        &self.database.url
    }

    pub fn server_address(&self) -> String {
        format!("{}:{}", self.server.host, self.server.port)
    }
} 