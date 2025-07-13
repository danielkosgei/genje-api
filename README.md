# Genje API

[![Go CI/CD](https://github.com/danielkosgei/genje-api/actions/workflows/go.yml/badge.svg)](https://github.com/danielkosgei/genje-api/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/danielkosgei/genje-api/branch/main/graph/badge.svg)](https://codecov.io/gh/danielkosgei/genje-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielkosgei/genje-api)](https://goreportcard.com/report/github.com/danielkosgei/genje-api)
[![GHCR](https://img.shields.io/badge/GHCR-danielkosgei%2Fgenje--api-blue)](https://github.com/danielkosgei/genje-api/pkgs/container/genje-api)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Genje API** is a comprehensive news aggregation service that automatically collects and organizes articles from multiple Kenyan news sources via RSS feeds. It provides a powerful RESTful API for accessing, filtering, and managing news content with advanced features like search, categorization, and automatic summarization.


## **API Documentation**

### **Base URL**
```
http://api.genje.co.ke
```

### **Example Usage**

#### **Get Recent Articles**
```bash
GET /v1/articles?limit=10&page=1

# Response
{
  "articles": [
    {
      "id": 1,
      "title": "Kenya's Economic Growth Outlook for 2024",
      "content": "The Central Bank of Kenya projects...",
      "summary": "Economic experts predict steady growth...",
      "url": "https://standardmedia.co.ke/business/article/2024/01/15/kenya-economic-growth",
      "author": "Jane Doe",
      "source": "Standard Business",
      "published_at": "2024-01-15T10:30:00Z",
      "created_at": "2024-01-15T10:35:00Z",
      "category": "business",
      "image_url": "https://standardmedia.co.ke/images/business/economic-growth.jpg"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1250
  }
}
```

#### **Search Articles**
```bash
GET /v1/articles/search?q=election&category=politics&limit=5

# Response
{
  "success": true,
  "data": [
    {
      "id": 142,
      "title": "2024 Election Preparations Underway",
      "content": "The Independent Electoral and Boundaries Commission...",
      "url": "https://nation.africa/kenya/news/politics/election-preparations-2024",
      "source": "Daily Nation",
      "category": "politics",
      "published_at": "2024-01-14T08:00:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 5,
      "total": 23
    },
    "generated_at": "2024-01-15T12:00:00Z",
    "query": "election",
    "filters": {
      "category": "politics"
    }
  }
}
```

#### **Get Articles by Category**
```bash
GET /v1/articles/by-category/sports?limit=3

# Response
{
  "success": true,
  "data": [
    {
      "id": 89,
      "title": "Harambee Stars Prepares for AFCON Qualifiers",
      "content": "The national football team has intensified training...",
      "source": "Standard Sports",
      "category": "sports",
      "published_at": "2024-01-15T06:00:00Z"
    }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 3,
      "total": 156
    },
    "generated_at": "2024-01-15T12:00:00Z"
  }
}
```

#### **Get Article Summary**
```bash
POST /v1/articles/123/summarize

# Response
{
  "summary": "The Central Bank of Kenya has announced new monetary policies. The policies aim to stabilize the currency and control inflation. Implementation will begin next quarter."
}
```

#### **Get Trending Articles**
```bash
GET /v1/articles/trending?window=24h&limit=5

# Response
{
  "success": true,
  "data": [
    {
      "id": 95,
      "title": "Breaking: New Infrastructure Project Launched",
      "score": 8.5,
      "trending_reason": "recent_engagement",
      "published_at": "2024-01-15T09:00:00Z"
    }
  ],
  "meta": {
    "generated_at": "2024-01-15T12:00:00Z",
    "algorithm": "recent_engagement",
    "time_window": "24h"
  }
}
```

#### **Get Statistics**
```bash
GET /v1/stats

# Response
{
  "success": true,
  "data": {
    "total_articles": 12450,
    "total_sources": 15,
    "categories": 8,
    "last_updated": "2024-01-15T11:30:00Z"
  },
  "meta": {
    "generated_at": "2024-01-15T12:00:00Z"
  }
}
```

## **Error Handling**

### **HTTP Status Codes**

| Status Code | Description | Example |
|-------------|-------------|---------|
| `200` | Success | Request completed successfully |
| `400` | Bad Request | Invalid query parameters |
| `404` | Not Found | Article or resource not found |
| `500` | Internal Server Error | Database connection failed |

### **Error Response Format**

```json
{
  "error": "Invalid query parameters",
  "code": 400,
  "details": "Page parameter must be a positive integer"
}
```

### **Common Error Scenarios**

#### **Invalid Article ID**
```bash
GET /v1/articles/invalid-id

# Response (400 Bad Request)
{
  "error": "Invalid article ID",
  "code": 400
}
```

#### **Article Not Found**
```bash
GET /v1/articles/99999

# Response (404 Not Found)
{
  "error": "Article not found",
  "code": 404
}
```

#### **Invalid Search Query**
```bash
GET /v1/articles/search

# Response (400 Bad Request)
{
  "error": "Query parameter 'q' is required",
  "code": 400
}
```

## **Authentication**

Currently, the Genje API does not require authentication and is designed for public access. All endpoints are accessible without API keys or tokens.


## **API Endpoints Reference**

### **Health & System**
- `GET /health` - API health check
- `GET /v1/status` - Detailed system status
- `GET /` - API information and available endpoints

### **Articles**
- `GET /v1/articles` - Get articles with pagination and filters
- `GET /v1/articles/{id}` - Get specific article
- `POST /v1/articles/{id}/summarize` - Generate article summary
- `GET /v1/articles/search` - Full-text search
- `GET /v1/articles/trending` - Get trending articles
- `GET /v1/articles/recent` - Get recent articles
- `GET /v1/articles/feed` - Get cursor-based article feed
- `GET /v1/articles/by-source/{sourceId}` - Get articles by source
- `GET /v1/articles/by-category/{category}` - Get articles by category

### **Sources**
- `GET /v1/sources` - Get all active sources
- `GET /v1/sources/{id}` - Get specific source
- `POST /v1/sources` - Create new source
- `PUT /v1/sources/{id}` - Update existing source
- `DELETE /v1/sources/{id}` - Delete source
- `POST /v1/sources/{id}/refresh` - Refresh specific source

### **Categories & Statistics**
- `GET /v1/categories` - Get all available categories
- `GET /v1/stats` - Get global statistics
- `GET /v1/stats/sources` - Get per-source statistics
- `GET /v1/stats/categories` - Get per-category statistics
- `GET /v1/stats/timeline` - Get timeline statistics
- `GET /v1/trends` - Get trending topics

### **System Operations**
- `POST /v1/refresh` - Trigger manual news aggregation
- `GET /v1/openapi.json` - OpenAPI specification
- `GET /v1/schema` - API schema information

## **Architecture**

The Genje API follows a clean architecture pattern:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │───▶│    Handlers     │───▶│    Services     │
│   (cURL, etc.)  │    │  (Controllers)  │    │ (Business Logic)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                               ┌─────────────────┐
                                               │   Repositories  │
                                               │  (Data Access)  │
                                               └─────────────────┘
                                                       │
                                                       ▼
                                               ┌─────────────────┐
                                               │   SQLite DB     │
                                               │   (Storage)     │
                                               └─────────────────┘
```

## Start Development Here

### **Use Docker (Recommended)**

```bash
# Pull and run the latest image
docker run -d \
  --name genje-api \
  -p 8080:8080 \
  -e DATABASE_URL="/app/data/genje.db" \
  -e AGGREGATION_INTERVAL="30m" \
  -v genje-data:/app/data \
  ghcr.io/danielkosgei/genje-api:main
```

```
# Check if it's running
curl http://localhost:8080/health
```

### **Local Development**

```bash
# Clone the repository
git clone https://github.com/danielkosgei/genje-api.git
cd genje-api

# Setup development environment
make dev-setup

# Configure environment (edit as needed)
cp .env.example .env

# Run the application
make run

# The API will be available at http://localhost:8080
```

### **Development Commands**

```bash
# Build the application
make build

# Run in development mode
make run

# Run tests
make test

# Run tests with coverage
make test-coverage

# Lint the code
make lint

# Format code
make format

# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Clean build artifacts
make clean

# Setup development environment
make dev-setup
```

## **Configuration**

### **Key Configuration Options**

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `./genje.db?...` | SQLite database connection string |
| `AGGREGATION_INTERVAL` | `30m` | How often to fetch news (5m, 30m, 1h, 2h) |
| `REQUEST_TIMEOUT` | `30s` | HTTP timeout for RSS requests |
| `USER_AGENT` | `Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)` | User agent for RSS requests |
| `MAX_CONTENT_SIZE` | `10000` | Maximum article content length |
| `MAX_SUMMARY_SIZE` | `300` | Maximum summary length |



## **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Contact**: For questions or support, please open an issue on GitHub.

**🌟 Star this repo** if you find it helpful!
