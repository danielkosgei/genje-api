# Genje API

News aggregation service that collects articles from multiple Kenyan news sources via RSS feeds and provides a RESTful API for accessing and managing news content.

## API Endpoints

```
GET /health
GET /api/v1/articles?page=1&limit=20&category=news&source=Standard&search=politics
GET /api/v1/articles/{id}
POST /api/v1/articles/{id}/summarize
GET /api/v1/sources
GET /api/v1/categories
POST /api/v1/refresh
```

### Development

1. **Clone the repository**
   ```bash
   git clone https://github.com/danielkosgei/genje-api.git
   cd genje-api
   ```

2. **Install dependencies**
   ```bash
   make dev-setup
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the application**
   ```bash
   make run
   ```

### Using Docker (Production)

1. **Build the image**
   ```bash
   make docker-build
   ```

2. **Run the container**
   ```bash
   make docker-run
   ```

## Configuration

The application uses environment variables for configuration. See `.env.example` for all available options and `docs/CONFIGURATION.md` for detailed documentation.

### Quick Setup
```bash
cp .env.example .env
nano .env
```

### Key Variables
- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - SQLite database connection string
- `AGGREGATION_INTERVAL` - How often to fetch news (default: 30m)
- `REQUEST_TIMEOUT` - HTTP request timeout (default: 30s)
- `USER_AGENT` - User agent for RSS requests

For complete configuration options, see [Configuration Guide](docs/CONFIGURATION.md).

## Development

### Available Commands

```bash
make build          # Build the application
make run            # Run in development mode
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Lint the code
make format         # Format the code
make clean          # Clean build artifacts
make tidy           # Tidy dependencies
```

### Testing

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Test specific package
go test ./internal/services/...
```

### Code Quality

```bash
# Format code
make format

# Lint code
make lint

# Check for issues
go vet ./...
```

## Deployment

### Docker Production Setup

```bash
# Build production image
docker build -t genje-api:latest .

# Run with environment file
docker run -d \
  --name genje-api \
  -p 8080:8080 \
  --env-file .env \
  --restart unless-stopped \
  genje-api:latest
```

### Systemd Service (Linux)

```ini
[Unit]
Description=Genje News API
After=network.target

[Service]
Type=simple
User=genje
WorkingDirectory=/opt/genje
ExecStart=/opt/genje/bin/genje-api
Restart=always
RestartSec=5
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```

