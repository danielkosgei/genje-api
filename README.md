# Genje News API

[![Go CI/CD](https://github.com/danielkosgei/genje-api/actions/workflows/go.yml/badge.svg)](https://github.com/danielkosgei/genje-api/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/danielkosgei/genje-api/branch/main/graph/badge.svg)](https://codecov.io/gh/danielkosgei/genje-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielkosgei/genje-api)](https://goreportcard.com/report/github.com/danielkosgei/genje-api)
[![GHCR](https://img.shields.io/badge/GHCR-danielkosgei%2Fgenje--api-blue)](https://github.com/danielkosgei/genje-api/pkgs/container/genje-api)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **RESTful News API**

**Genje API** is a news aggregation service that automatically collects, processes, and organizes articles from news sources. It provides developers with a powerful, RESTful API featuring advanced NLP-powered summarization, intelligent search, real-time aggregation, and comprehensive filtering capabilities.

## **Quick Start for Developers**

### **Base URL**
```
https://api.genje.co.ke  # Production
http://localhost:8080    # Local Development
```

Get intelligent, context-aware summaries powered by TF-IDF analysis:

```bash
# Get or generate article summary
GET /v1/articles/123/summary

# Response
{
  "success": true,
  "data": {
    "summary": "President William Ruto has pledged to expand youth employment in both the Climate Worx and affordable housing programmes, aiming to double current figures within the next three months."
  },
  "meta": {
    "timestamp": "2025-07-20T13:57:23Z"
  }
}
```

#### **Get Articles with Advanced Filtering**
```bash
GET /v1/articles?page=1&limit=20&category=politics&search=election

# New Standardized Response Format
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Kenya's Economic Growth Outlook for 2025",
      "content": "The Central Bank of Kenya projects...",
      "summary": "Economic experts predict steady growth driven by infrastructure investments and digital transformation initiatives.",
      "url": "https://standardmedia.co.ke/business/article/2025/01/15/kenya-economic-growth",
      "author": "Jane Doe",
      "source": "Standard Business",
      "published_at": "2025-01-15T10:30:00Z",
      "created_at": "2025-01-15T10:35:00Z",
      "category": "business",
      "image_url": "https://standardmedia.co.ke/images/business/economic-growth.jpg"
    }
  ],
  "meta": {
    "timestamp": "2025-07-20T12:00:00Z",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1250,
      "total_pages": 63,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

#### **Intelligent Search with Relevance Ranking**
```bash
GET /v1/articles/search?q=election&category=politics&sort_by=relevance

# Advanced search with multiple filters
{
  "success": true,
  "data": [
    {
      "id": 142,
      "title": "2025 Election Preparations Underway",
      "summary": "The Independent Electoral and Boundaries Commission has announced comprehensive preparations for the upcoming elections, including voter registration drives and security measures.",
      "url": "https://nation.africa/kenya/news/politics/election-preparations-2025",
      "source": "Daily Nation",
      "category": "politics",
      "published_at": "2025-01-14T08:00:00Z"
    }
  ],
  "meta": {
    "timestamp": "2025-07-20T12:00:00Z",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 23,
      "total_pages": 2,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

#### **RESTful Resource Access**
```bash
# Get articles from specific source
GET /v1/sources/123/articles

# Get articles by category (RESTful way)
GET /v1/categories/sports/articles

# Both return the same standardized format
{
  "success": true,
  "data": [...],
  "meta": {
    "timestamp": "2025-07-20T12:00:00Z",
    "pagination": {...}
  }
}
```

#### **Rich Analytics & Statistics**
```bash
GET /v1/stats

# Comprehensive statistics
{
  "success": true,
  "data": {
    "total_articles": 12450,
    "total_sources": 15,
    "categories": 8,
    "last_updated": "2025-07-20T11:30:00Z"
  },
  "meta": {
    "timestamp": "2025-07-20T12:00:00Z"
  }
}

# Timeline statistics
GET /v1/stats/timeline?days=30

# Source-specific statistics  
GET /v1/stats/sources
```

#### **Trending & Recent Content**
```bash
# Get trending articles with advanced scoring
GET /v1/articles/trending?window=24h&limit=10

# Get recent articles from last N hours
GET /v1/articles/recent?hours=6&limit=15

# Cursor-based feed for infinite scroll
GET /v1/articles/feed?cursor=abc123&limit=20
```

## **Error Handling & Rate Limiting**

### **Standardized Error Response Format**

All errors follow a consistent structure for easy parsing:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": "Page parameter must be a positive integer"
  },
  "meta": {
    "timestamp": "2025-07-20T12:00:00Z",
    "request_id": "abc123def456"
  }
}
```

### **HTTP Status Codes & Error Types**

| Status | Error Code | Description | Example |
|--------|------------|-------------|---------|
| `200` | - | Success | Request completed successfully |
| `201` | - | Created | Resource created successfully |
| `204` | - | No Content | Resource deleted successfully |
| `400` | `VALIDATION_ERROR` | Bad Request | Invalid parameters or request body |
| `404` | `NOT_FOUND` | Not Found | Article, source, or resource not found |
| `429` | `RATE_LIMIT_EXCEEDED` | Too Many Requests | Rate limit exceeded (100 req/min) |
| `500` | `INTERNAL_ERROR` | Server Error | Database or system error |

### **Rate Limiting**

- **Limit**: 100 requests per minute per IP address
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Window`, `Retry-After`
- **Response**: 429 status with retry information

### **Common Error Examples**

```bash
# Invalid article ID
GET /v1/articles/invalid-id
# → 400 Bad Request with VALIDATION_ERROR

# Article not found  
GET /v1/articles/99999
# → 404 Not Found with NOT_FOUND

# Missing required search query
GET /v1/articles/search
# → 400 Bad Request: "Query parameter 'q' is required"

# Rate limit exceeded
# → 429 Too Many Requests with Retry-After header
```

## **Authentication**

Currently, the Genje API does not require authentication and is designed for public access. All endpoints are accessible without API keys or tokens.


## **Complete API Reference**

### **Health & System**
```bash
GET /health                    # API health check
GET /v1/status                 # Detailed system status  
GET /                          # API information and available endpoints
GET /v1/openapi.json          # OpenAPI 3.0 specification
GET /v1/schema                # API schema information
```

### **Articles Resource**
```bash
# Core article operations
GET    /v1/articles                    # List articles with pagination & filters
GET    /v1/articles/{id}              # Get specific article by ID
GET    /v1/articles/{id}/summary      # Get/generate NLP-powered summary
POST   /v1/articles/{id}/summary      # Generate new summary (same endpoint)

# Advanced article queries  
GET    /v1/articles/search            # Full-text search with relevance ranking
GET    /v1/articles/feed              # Cursor-based infinite scroll feed
GET    /v1/articles/trending          # Trending articles with scoring algorithm
GET    /v1/articles/recent            # Recent articles from last N hours
```

### **Sources Resource**
```bash
# Source management (RESTful CRUD)
GET    /v1/sources                    # List all active sources
POST   /v1/sources                    # Create new news source
GET    /v1/sources/{id}               # Get specific source details
PUT    /v1/sources/{id}               # Update source (full replacement)
PATCH  /v1/sources/{id}               # Update source (partial update)
DELETE /v1/sources/{id}               # Delete source (returns 204 No Content)

# Source operations
POST   /v1/sources/{id}/refresh       # Trigger refresh for specific source
GET    /v1/sources/{id}/articles      # Get articles from specific source
```

### **Categories Resource**
```bash
GET    /v1/categories                 # List all available categories
GET    /v1/categories/{name}/articles # Get articles from specific category
```

### **Statistics & Analytics**
```bash
GET    /v1/stats                      # Global statistics overview
GET    /v1/stats/sources              # Per-source article statistics  
GET    /v1/stats/categories           # Per-category article statistics
GET    /v1/stats/timeline             # Timeline statistics (last N days)
GET    /v1/trends                     # Trending topics and keywords
```

### **System Operations**
```bash
POST   /v1/refresh                    # Trigger manual news aggregation
```

### **🔧 Query Parameters**

#### **Articles Filtering**
```bash
GET /v1/articles?page=1&limit=20&category=politics&source=Daily%20Nation&search=election&sort_by=date&sort_order=desc
```

| Parameter | Type | Description | Default | Max |
|-----------|------|-------------|---------|-----|
| `page` | integer | Page number | 1 | - |
| `limit` | integer | Items per page | 20 | 100 |
| `category` | string | Filter by category | - | - |
| `source` | string | Filter by source name | - | - |
| `search` | string | Search in title/content | - | - |
| `sort_by` | string | Sort field (`date`, `title`, `source`) | `date` | - |
| `sort_order` | string | Sort direction (`asc`, `desc`) | `desc` | - |

#### **Search Parameters**
```bash
GET /v1/articles/search?q=election&category=politics&from=2025-01-01&to=2025-12-31&sort_by=relevance
```

| Parameter | Type | Description | Required |
|-----------|------|-------------|----------|
| `q` | string | Search query | Yes |
| `category` | string | Filter by category | No |
| `source` | string | Filter by source | No |
| `from` | date | Start date (YYYY-MM-DD) | No |
| `to` | date | End date (YYYY-MM-DD) | No |
| `sort_by` | string | `relevance`, `date`, `source` | No |

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

## **Advanced NLP Summarization**

Our intelligent summarization engine uses sophisticated NLP techniques to generate meaningful summaries:

### **Key Features**
- **TF-IDF Analysis**: Mathematical foundation for semantic importance
- **Multi-Criteria Scoring**: 8 different metrics including position, title relevance, entity recognition
- **Language Awareness**: Supports both English and Kiswahili content
- **Content Diversity**: Clustering algorithms prevent redundant information
- **Context Understanding**: Uses title and discourse markers for better relevance

### **Scoring Algorithm**
```
Final Score = (TF-IDF × 0.25) + (Position × 0.20) + (Title Similarity × 0.15) + 
              (Length × 0.10) + (Entities × 0.10) + (Numerical × 0.08) + 
              (Centrality × 0.07) + (Discourse × 0.05)
```

### **Performance**
- **Processing Time**: 80-500ms per article
- **Quality**: High correlation with article main topics
- **Cache**: Summaries stored for instant retrieval

## **SDK & Integration Examples**

### **JavaScript/Node.js**
```javascript
// Install: npm install axios
const axios = require('axios');

const genjeAPI = axios.create({
  baseURL: 'https://api.genje.co.ke',
  timeout: 10000,
});

// Get articles with NLP summaries
async function getArticlesWithSummaries() {
  try {
    const { data } = await genjeAPI.get('/v1/articles', {
      params: { limit: 10, category: 'politics' }
    });
    
    // Get summaries for each article
    for (const article of data.data) {
      const summary = await genjeAPI.get(`/v1/articles/${article.id}/summary`);
      article.ai_summary = summary.data.data.summary;
    }
    
    return data;
  } catch (error) {
    console.error('API Error:', error.response?.data || error.message);
  }
}
```

### **Python**
```python
import requests
from typing import List, Dict, Optional

class GenjeAPI:
    def __init__(self, base_url: str = "https://api.genje.co.ke"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.timeout = 10
    
    def get_articles(self, category: Optional[str] = None, 
                    limit: int = 20, page: int = 1) -> Dict:
        """Get articles with optional filtering"""
        params = {"limit": limit, "page": page}
        if category:
            params["category"] = category
            
        response = self.session.get(f"{self.base_url}/v1/articles", params=params)
        response.raise_for_status()
        return response.json()
    
    def get_summary(self, article_id: int) -> str:
        """Get AI-powered article summary"""
        response = self.session.get(f"{self.base_url}/v1/articles/{article_id}/summary")
        response.raise_for_status()
        return response.json()["data"]["summary"]
    
    def search_articles(self, query: str, **filters) -> Dict:
        """Search articles with NLP ranking"""
        params = {"q": query, **filters}
        response = self.session.get(f"{self.base_url}/v1/articles/search", params=params)
        response.raise_for_status()
        return response.json()

# Usage
api = GenjeAPI()
articles = api.get_articles(category="business", limit=5)
summary = api.get_summary(article_id=123)
```

### **cURL Examples**
```bash
# Get trending articles with summaries
curl -X GET "https://api.genje.co.ke/v1/articles/trending?limit=5" \
  -H "Accept: application/json"

# Search with multiple filters
curl -X GET "https://api.genje.co.ke/v1/articles/search" \
  -G \
  -d "q=election" \
  -d "category=politics" \
  -d "sort_by=relevance" \
  -H "Accept: application/json"

# Get article summary
curl -X GET "https://api.genje.co.ke/v1/articles/123/summary" \
  -H "Accept: application/json"
```

## **Configuration**

### **Environment Variables**

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | `./genje.db?...` | SQLite database connection string |
| `AGGREGATION_INTERVAL` | `30m` | News fetch interval (5m, 30m, 1h, 2h) |
| `REQUEST_TIMEOUT` | `30s` | HTTP timeout for RSS requests |
| `USER_AGENT` | `Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)` | User agent for RSS requests |
| `MAX_CONTENT_SIZE` | `10000` | Maximum article content length |
| `MAX_SUMMARY_SIZE` | `300` | Maximum summary length |

### **Production Deployment**
```bash
# Docker Compose
version: '3.8'
services:
  genje-api:
    image: ghcr.io/danielkosgei/genje-api:main
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=/app/data/genje.db
      - AGGREGATION_INTERVAL=30m
      - PORT=8080
    volumes:
      - genje-data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  genje-data:
```

## **Architecture & Performance**

### **System Architecture**
- **Language**: Go 1.24+ with modern concurrency patterns
- **Database**: SQLite with WAL mode for high concurrency
- **Caching**: In-memory caching for summaries and frequent queries
- **Rate Limiting**: Token bucket algorithm (100 req/min per IP)
- **Monitoring**: Structured logging with request tracing

### **Performance Metrics**
- **Response Time**: < 100ms for cached content, < 500ms for NLP processing
- **Throughput**: 1000+ requests/second on standard hardware
- **Availability**: 99.9% uptime with health checks and auto-recovery
- **Storage**: Efficient SQLite with optimized indexes

### **News Sources**
Currently aggregating from 15+ major Kenyan news outlets:
- The Standard (multiple categories)
- Daily Nation
- Capital FM News
- Nairobi Wire
- Business Daily
- Diaspora Messenger
- And more...

## **Contributing**

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### **Development Setup**
```bash
git clone https://github.com/danielkosgei/genje-api.git
cd genje-api
make dev-setup
make run
```

### **Running Tests**
```bash
make test              # Run all tests
make test-coverage     # Run with coverage report
make lint              # Run linter
```

## **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## **Support & Contact**

- **Issues**: [GitHub Issues](https://github.com/danielkosgei/genje-api/issues)
- **Discussions**: [GitHub Discussions](https://github.com/danielkosgei/genje-api/discussions)
- **Email**: support@genje.co.ke

---

**🌟 Star this repo** if you find it helpful!
