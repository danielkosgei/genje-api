package docs

import "genje-api/internal/models"

// GenerateOpenAPISpec generates a comprehensive OpenAPI 3.0 specification
func GenerateOpenAPISpec() models.OpenAPISpec {
	return models.OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: models.OpenAPIInfo{
			Title:       "Genje News API",
			Description: "Comprehensive Kenyan news aggregation service providing access to articles from multiple sources with advanced filtering, search, and analytics capabilities.",
			Version:     "1.0.0",
			Contact: models.OpenAPIContact{
				Name:  "Genje API Team",
				Email: "api@genje.co.ke",
				URL:   "https://api.genje.co.ke",
			},
		},
		Paths: map[string]interface{}{
			"/v1/articles": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List articles",
					"description": "Retrieve a paginated list of articles with optional filtering",
					"tags":        []string{"Articles"},
					"parameters": []map[string]interface{}{
						{
							"name":        "page",
							"in":          "query",
							"description": "Page number (default: 1)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "default": 1},
						},
						{
							"name":        "limit",
							"in":          "query",
							"description": "Items per page (default: 20, max: 100)",
							"required":    false,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 100, "default": 20},
						},
						{
							"name":        "category",
							"in":          "query",
							"description": "Filter by category",
							"required":    false,
							"schema": map[string]interface{}{
								"type": "string",
								"enum": []string{"news", "sports", "business", "politics", "technology", "entertainment", "health", "world", "opinion", "general"},
							},
						},
						{
							"name":        "source",
							"in":          "query",
							"description": "Filter by source name",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string"},
						},
						{
							"name":        "search",
							"in":          "query",
							"description": "Search in title and content",
							"required":    false,
							"schema":      map[string]interface{}{"type": "string", "minLength": 2},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ArticlesResponse",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Bad request",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
			"/v1/articles/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get article by ID",
					"description": "Retrieve a specific article by its ID",
					"tags":        []string{"Articles"},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"description": "Article ID",
							"required":    true,
							"schema":      map[string]interface{}{"type": "integer", "minimum": 1},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ArticleResponse",
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "Article not found",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/ErrorResponse",
									},
								},
							},
						},
					},
				},
			},
		},
		Components: map[string]interface{}{
			"schemas": map[string]interface{}{
				"Article": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":           map[string]interface{}{"type": "integer", "example": 123},
						"title":        map[string]interface{}{"type": "string", "example": "Kenya's Economic Growth Outlook for 2024"},
						"content":      map[string]interface{}{"type": "string", "example": "The Central Bank of Kenya projects..."},
						"summary":      map[string]interface{}{"type": "string", "example": "Economic experts predict steady growth..."},
						"url":          map[string]interface{}{"type": "string", "format": "uri", "example": "https://standardmedia.co.ke/business/article/2024/01/15/kenya-economic-growth"},
						"author":       map[string]interface{}{"type": "string", "example": "Jane Doe"},
						"source":       map[string]interface{}{"type": "string", "example": "Standard Business"},
						"published_at": map[string]interface{}{"type": "string", "format": "date-time", "example": "2024-01-15T10:30:00Z"},
						"created_at":   map[string]interface{}{"type": "string", "format": "date-time", "example": "2024-01-15T10:35:00Z"},
						"category":     map[string]interface{}{"type": "string", "example": "business"},
						"image_url":    map[string]interface{}{"type": "string", "format": "uri", "example": "https://standardmedia.co.ke/images/business/economic-growth.jpg"},
					},
					"required": []string{"id", "title", "url", "source", "published_at", "created_at", "category"},
				},
				"ArticlesResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"$ref": "#/components/schemas/Article"}},
						"meta": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp":  map[string]interface{}{"type": "string", "format": "date-time"},
								"pagination": map[string]interface{}{"$ref": "#/components/schemas/PaginationMeta"},
							},
						},
					},
				},
				"ArticleResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": true},
						"data":    map[string]interface{}{"$ref": "#/components/schemas/Article"},
						"meta": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
							},
						},
					},
				},
				"PaginationMeta": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"page":        map[string]interface{}{"type": "integer", "example": 1},
						"limit":       map[string]interface{}{"type": "integer", "example": 20},
						"total":       map[string]interface{}{"type": "integer", "example": 1250},
						"total_pages": map[string]interface{}{"type": "integer", "example": 63},
						"has_next":    map[string]interface{}{"type": "boolean", "example": true},
						"has_prev":    map[string]interface{}{"type": "boolean", "example": false},
					},
				},
				"ErrorResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"success": map[string]interface{}{"type": "boolean", "example": false},
						"error": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"code":    map[string]interface{}{"type": "string", "example": "VALIDATION_ERROR"},
								"message": map[string]interface{}{"type": "string", "example": "Invalid query parameters"},
								"details": map[string]interface{}{"type": "string", "example": "Page parameter must be a positive integer"},
							},
						},
						"meta": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"timestamp": map[string]interface{}{"type": "string", "format": "date-time"},
							},
						},
					},
				},
			},
		},
	}
}
