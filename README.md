# Genje - Kenyan News Aggregator

A modern, fast, and scalable news aggregator specifically designed for Kenyan news sources. Built with Rust for performance and reliability.

## Features

- **RSS Feed Parsing**: Automatically fetches and parses RSS feeds from major Kenyan news outlets
- **Web Scraping**: Intelligent web scraping for sources without RSS feeds
- **Multi-language Support**: Handles both English and Swahili content
- **Regional Classification**: Automatically categorizes news by Kenyan regions
- **Category Detection**: Smart categorization (politics, business, sports, health, technology, etc.)
- **RESTful API**: Clean JSON API for integration with web and mobile applications
- **PostgreSQL Database**: Robust data storage with full-text search capabilities
- **Docker Support**: Easy deployment with Docker and docker-compose

## Kenyan News Sources

Pre-configured with major Kenyan media outlets:
- Daily Nation
- The Standard
- Citizen Digital
- Capital FM
- And more...

## Quick Start

### Using Docker (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd genje
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start with docker-compose:
```bash
docker-compose up -d
```

### Manual Setup

1. Install dependencies:
   - Rust (latest stable)
   - PostgreSQL 13+

2. Set up the database:
```bash
createdb genje
```

3. Configure environment:
```bash
cp .env.example .env
# Edit .env with your database URL and other settings
```

4. Run migrations:
```bash
cargo run --bin main
```

5. Start the fetcher service:
```bash
cargo run --bin fetcher
```

6. Start the API server:
```bash
cargo run
```

## API Endpoints

### Articles
- `GET /api/articles` - Get articles with optional filters
  - Query parameters: `category`, `language`, `region`, `limit`, `offset`
- `GET /api/articles/recent` - Get recent articles
- `GET /api/articles/trending` - Get trending articles
- `GET /api/articles/search?q=query` - Search articles
- `POST /api/articles/:id/view` - Increment article view count

### Sources
- `GET /api/sources` - Get all active news sources

### Health
- `GET /health` - Service health check

## Configuration

The application can be configured via environment variables with the `GENJE_` prefix:

```bash
# Database
GENJE_DATABASE_URL=postgresql://user:pass@host:port/db
GENJE_DATABASE_MAX_CONNECTIONS=10

# Server
GENJE_SERVER_HOST=0.0.0.0
GENJE_SERVER_PORT=8080

# Fetcher
GENJE_FETCHER_FETCH_INTERVAL_MINUTES=30
GENJE_FETCHER_MAX_ARTICLES_PER_SOURCE=50
GENJE_FETCHER_CONCURRENT_FETCHES=5

# Logging
GENJE_LOGGING_LEVEL=info
```

## Architecture

### Components

1. **Main API Service** (`src/main.rs`)
   - REST API server using Axum
   - Article and source management
   - Search and filtering capabilities

2. **News Fetcher Service** (`src/bin/fetcher.rs`)
   - Periodic RSS feed parsing
   - Web scraping for non-RSS sources
   - Automatic content categorization

3. **Core Libraries**
   - `models/` - Data structures for articles and sources
   - `services/` - Business logic (RSS parsing, scraping, database)
   - `config/` - Configuration management
   - `utils/` - HTTP client and utilities

### Database Schema

- `news_sources` - News source configuration
- `articles` - Aggregated news articles with metadata
- Full-text search indexes for efficient querying

## Development

### Building

```bash
# Build all binaries
cargo build --release

# Build specific binary
cargo build --bin fetcher --release
```

### Testing

```bash
cargo test
```

### Database Migrations

Migrations are automatically applied on startup. Migration files are in the `migrations/` directory.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Roadmap

- [ ] Real-time notifications
- [ ] Sentiment analysis
- [ ] Multi-source deduplication
- [ ] Content recommendation engine
- [ ] Mobile app
- [ ] Analytics dashboard
- [ ] Support for more East African news sources 