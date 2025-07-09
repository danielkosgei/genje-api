# Configuration Guide

This document explains all the configuration options available for the Genje News API.

## Quick Setup

1. Copy the configuration template:
   ```bash
   cp .example.env .env
   ```

2. Edit `.env` with your specific values:
   ```bash
   nano .env
   ```

3. The application will automatically load the `.env` file on startup.

## Required Configuration

### PORT
- **Description**: HTTP server port
- **Default**: `8080`
- **Example**: `PORT=8080`

### DATABASE_URL
- **Description**: SQLite database connection string
- **Default**: `./genje.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON`
- **Development**: `./genje-dev.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON`
- **Production**: `/data/genje.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=2000&_foreign_keys=ON`

## News Aggregation Settings

### AGGREGATION_INTERVAL
- **Description**: How often to fetch news from RSS feeds
- **Default**: `30m`
- **Format**: Duration string (`5m`, `30m`, `1h`, `2h`)
- **Development**: `5m` (for faster testing)
- **Production**: `30m` or `1h`

### REQUEST_TIMEOUT
- **Description**: HTTP timeout for RSS feed requests
- **Default**: `30s`
- **Format**: Duration string (`30s`, `1m`, `2m`)

### USER_AGENT
- **Description**: User-Agent header sent to news sources
- **Default**: `Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)`
- **Note**: Some sites block requests without proper user agents

## Advanced Configuration

These settings have sensible defaults and typically don't need to be changed:

### Content Limits
```env
MAX_CONTENT_SIZE=10000    # Max article content length
MAX_SUMMARY_SIZE=300      # Max summary length
```

### Database Connection Pool
```env
DB_MAX_OPEN_CONNS=25      # Max open connections
DB_MAX_IDLE_CONNS=25      # Max idle connections  
DB_CONN_MAX_LIFETIME=5m   # Connection lifetime
```

### Logging
```env
LOG_LEVEL=INFO            # DEBUG, INFO, WARN, ERROR
LOG_JSON=false            # Enable JSON logging
```

## Environment-Specific Examples

### Development Environment
```env
PORT=8080
DATABASE_URL=./genje-dev.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON
AGGREGATION_INTERVAL=5m
REQUEST_TIMEOUT=30s
USER_AGENT=Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)
LOG_LEVEL=DEBUG
```

### Production Environment
```env
PORT=8080
DATABASE_URL=/data/genje.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=2000&_foreign_keys=ON
AGGREGATION_INTERVAL=30m
REQUEST_TIMEOUT=30s
USER_AGENT=Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)
LOG_LEVEL=INFO
LOG_JSON=true
```

### Docker Environment
```env
PORT=8080
DATABASE_URL=/app/data/genje.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON
AGGREGATION_INTERVAL=30m
REQUEST_TIMEOUT=30s
USER_AGENT=Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)
```

## SQLite Parameters Explained

The `DATABASE_URL` includes several SQLite-specific parameters:

- **`_journal_mode=WAL`**: Write-Ahead Logging mode for better concurrency
- **`_synchronous=NORMAL`**: Balance between safety and performance  
- **`_cache_size=1000`**: Cache 1000 pages in memory (approximately 4MB)
- **`_foreign_keys=ON`**: Enable foreign key constraint checking

## Validation

The application validates configuration on startup and will fail with clear error messages if:

- Required environment variables are missing
- Duration strings are malformed
- Database connection fails
- Invalid numeric values are provided

## Configuration Loading Order

1. Default values (hardcoded in `internal/config/config.go`)
2. Environment variables from the system
3. Variables from `.env` file (if present)

Environment variables and `.env` file values override defaults.

## Security Notes

- Never commit `.env` files to version control
- Use strong database paths in production
- Consider restricting CORS origins in production
- Enable structured logging for production monitoring
- Use absolute paths for production database files
- Ensure proper file permissions on the database file and directory 