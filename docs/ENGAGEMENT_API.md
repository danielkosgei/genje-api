# 🚀 Advanced Engagement & Trending API

The Genje News API now features a sophisticated engagement tracking system and advanced trending algorithm that considers 5 key factors to determine truly trending content.

## 🎯 Key Features

### **5-Factor Advanced Trending Algorithm**
1. **Engagement-Based Scoring** (35%) - Views, shares, comments, likes
2. **Velocity-Based Trending** (25%) - Rate of engagement growth
3. **Source Authority Weighting** (20%) - Credibility and reach scores
4. **Content Analysis** (10%) - NLP-based quality assessment
5. **Time Decay Function** (10%) - Recency with smart decay rates

### **Real-Time Engagement Tracking**
- Track views, shares, comments, and likes
- Automatic source authority calculation
- Performance-optimized with caching
- Privacy-conscious (IP-based, no user accounts required)

## 📊 API Endpoints

### **Track Engagement**
Record user interactions with articles:

```bash
# Track a view
curl -X POST "https://api.genje.co.ke/v1/articles/123/engagement" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "view",
    "user_ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "metadata": "{\"source\": \"mobile_app\"}"
  }'

# Track a share
curl -X POST "https://api.genje.co.ke/v1/articles/123/engagement" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "share",
    "metadata": "{\"platform\": \"twitter\"}"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Engagement tracked successfully",
    "article_id": 123,
    "event_type": "view"
  },
  "meta": {
    "timestamp": "2025-01-20T10:30:00Z"
  }
}
```

### **Get Article Engagement Metrics**
Retrieve engagement statistics for any article:

```bash
curl "https://api.genje.co.ke/v1/articles/123/engagement"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "article_id": 123,
    "views": 1250,
    "shares": 45,
    "comments": 12,
    "likes": 89,
    "last_updated": "2025-01-20T10:30:00Z"
  },
  "meta": {
    "timestamp": "2025-01-20T10:30:00Z"
  }
}
```

### **Advanced Trending Articles**
Get trending articles using the sophisticated 5-factor algorithm:

```bash
# Get trending articles for the last 24 hours
curl "https://api.genje.co.ke/v1/articles/trending/advanced?window=24h&limit=10"

# Get trending articles for the last hour (breaking news)
curl "https://api.genje.co.ke/v1/articles/trending/advanced?window=1h&limit=5"
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 123,
      "title": "Kenya's Economic Growth Outlook for 2025",
      "content": "...",
      "url": "https://example.com/article/123",
      "source": "Business Daily",
      "published_at": "2025-01-20T08:00:00Z",
      "category": "business",
      "engagement": {
        "views": 1250,
        "shares": 45,
        "comments": 12,
        "likes": 89
      },
      "trending_score": 0.87,
      "engagement_velocity": 0.65,
      "recency_score": 0.92,
      "authority_score": 0.78,
      "content_score": 0.71,
      "trending_reason": "Rapidly gaining engagement",
      "score_breakdown": {
        "engagement_weight": 0.35,
        "velocity_weight": 0.25,
        "authority_weight": 0.20,
        "content_weight": 0.10,
        "recency_weight": 0.10
      }
    }
  ],
  "meta": {
    "timestamp": "2025-01-20T10:30:00Z",
    "time_window": "24h",
    "algorithm": "5-factor advanced trending",
    "factors": [
      "engagement_score",
      "velocity_trending", 
      "source_authority",
      "content_analysis",
      "time_decay"
    ]
  }
}
```

### **Source Authority Metrics**
Check the authority, credibility, and reach scores for news sources:

```bash
curl "https://api.genje.co.ke/v1/sources/Business%20Daily/authority"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "source_name": "Business Daily",
    "authority_score": 0.78,
    "credibility_score": 0.82,
    "reach_score": 0.71,
    "total_articles": 1250,
    "avg_engagement": 145.5,
    "last_calculated": "2025-01-20T09:00:00Z"
  },
  "meta": {
    "timestamp": "2025-01-20T10:30:00Z"
  }
}
```

### **Top Engaged Articles**
Get the most engaged articles in a time window:

```bash
curl "https://api.genje.co.ke/v1/articles/top-engaged?window=24h&limit=10"
```

## 🔧 Integration Examples

### **JavaScript/Node.js**
```javascript
class GenjeEngagementAPI {
  constructor(baseURL = 'https://api.genje.co.ke') {
    this.baseURL = baseURL;
  }

  async trackEngagement(articleId, eventType, metadata = {}) {
    const response = await fetch(`${this.baseURL}/v1/articles/${articleId}/engagement`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        event_type: eventType,
        metadata: JSON.stringify(metadata)
      })
    });
    return response.json();
  }

  async getAdvancedTrending(window = '24h', limit = 20) {
    const response = await fetch(
      `${this.baseURL}/v1/articles/trending/advanced?window=${window}&limit=${limit}`
    );
    return response.json();
  }

  async getEngagementMetrics(articleId) {
    const response = await fetch(`${this.baseURL}/v1/articles/${articleId}/engagement`);
    return response.json();
  }
}

// Usage
const api = new GenjeEngagementAPI();

// Track a view when user reads an article
await api.trackEngagement(123, 'view', { source: 'web_app' });

// Track a share when user shares on social media
await api.trackEngagement(123, 'share', { platform: 'twitter' });

// Get trending articles
const trending = await api.getAdvancedTrending('6h', 10);
console.log('Trending articles:', trending.data);
```

### **Python**
```python
import requests
import json

class GenjeEngagementAPI:
    def __init__(self, base_url='https://api.genje.co.ke'):
        self.base_url = base_url

    def track_engagement(self, article_id, event_type, metadata=None):
        url = f"{self.base_url}/v1/articles/{article_id}/engagement"
        payload = {
            'event_type': event_type,
            'metadata': json.dumps(metadata or {})
        }
        response = requests.post(url, json=payload)
        return response.json()

    def get_advanced_trending(self, window='24h', limit=20):
        url = f"{self.base_url}/v1/articles/trending/advanced"
        params = {'window': window, 'limit': limit}
        response = requests.get(url, params=params)
        return response.json()

    def get_engagement_metrics(self, article_id):
        url = f"{self.base_url}/v1/articles/{article_id}/engagement"
        response = requests.get(url)
        return response.json()

# Usage
api = GenjeEngagementAPI()

# Track engagement
api.track_engagement(123, 'view', {'source': 'mobile_app'})
api.track_engagement(123, 'like', {'user_segment': 'premium'})

# Get trending articles with detailed scoring
trending = api.get_advanced_trending('1h', 5)
for article in trending['data']:
    print(f"Article: {article['title']}")
    print(f"Trending Score: {article['trending_score']:.2f}")
    print(f"Reason: {article['trending_reason']}")
    print("---")
```

## 📈 Algorithm Details

### **Engagement Scoring Formula**
```
engagement_score = (views * 0.1) + (shares * 0.4) + (comments * 0.3) + (likes * 0.2)
normalized_score = min(engagement_score / 10000, 1.0)
```

### **Velocity Calculation**
```
velocity = (current_period_engagement - previous_period_engagement) / previous_period_engagement
normalized_velocity = clamp(velocity, -1.0, 1.0)
```

### **Source Authority Components**
- **Authority Score**: Based on article volume and engagement (30% volume, 70% engagement)
- **Credibility Score**: Comments and likes ratio to shares
- **Reach Score**: Average views per article

### **Content Analysis Factors**
- Title optimization (length, numbers, questions, breaking news indicators)
- Content quality (length, structure, formatting)
- Keyword trending analysis

### **Time Decay Function**
```
recency_score = e^(-age_in_seconds / half_life_seconds)
```
Different half-life values for different time windows:
- 1h window: 30-minute half-life
- 24h window: 8-hour half-life
- 7d window: 48-hour half-life

## 🚀 Performance Features

- **Caching**: Trending results cached for 15 minutes
- **Background Processing**: Source authority updates run asynchronously
- **Optimized Queries**: Indexed database queries for fast response times
- **Fallback**: Graceful degradation to simple trending if advanced algorithm fails

## 🔒 Privacy & Security

- **No User Accounts**: IP-based tracking without personal data
- **Configurable Retention**: Engagement events can be purged after specified periods
- **Rate Limiting**: Built-in protection against abuse
- **Validation**: All inputs validated and sanitized

This advanced engagement system transforms your news API into a sophisticated content discovery platform that truly understands what's trending and why!