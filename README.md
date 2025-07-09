# Genje API

[![Go CI/CD](https://github.com/danielkosgei/genje-api/actions/workflows/go.yml/badge.svg)](https://github.com/danielkosgei/genje-api/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/danielkosgei/genje-api/branch/main/graph/badge.svg)](https://codecov.io/gh/danielkosgei/genje-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/danielkosgei/genje-api)](https://goreportcard.com/report/github.com/danielkosgei/genje-api)
[![Docker Pulls](https://img.shields.io/docker/pulls/danielkosgei/genje-api)](https://hub.docker.com/r/danielkosgei/genje-api)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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

## CI/CD Pipeline

This project uses GitHub Actions for continuous integration and deployment with a comprehensive workflow that ensures code quality, security, and reliable deployments.

### Workflow Overview

The CI/CD pipeline (`go.yml`) runs on:
- **Push to `main`** - Full pipeline including production deployment
- **Push to `staging`** - Full pipeline including staging deployment  
- **Pull Requests** to `main` or `staging` - Testing and validation only

### Pipeline Jobs

#### 1. **Test** 🧪
- **Matrix Strategy**: Tests on Go versions 1.23 and 1.24.4
- **Coverage**: Generates coverage reports and uploads to Codecov
- **Caching**: Go modules cached for faster builds
- **Dependencies**: Verifies and downloads all dependencies

```bash
# Local equivalent
make test
make test-coverage
```

#### 2. **Lint** 🔍
- **Code Quality**: Runs `golangci-lint` with 5-minute timeout
- **Standards**: Enforces Go best practices and style guidelines
- **Fast Feedback**: Fails fast on code quality issues

```bash
# Local equivalent  
make lint
```

#### 3. **Build** 🔨
- **Dependency**: Requires test and lint to pass
- **Artifacts**: Builds binary and uploads as GitHub artifact
- **Retention**: Artifacts kept for 30 days

```bash
# Local equivalent
make build
```

#### 4. **Security** 🔒
- **Scanner**: Uses Gosec to detect security vulnerabilities
- **SARIF**: Uploads results to GitHub Security tab
- **Parallel**: Runs independently for faster feedback

#### 5. **Docker** 🐳
- **Trigger**: Only on `main` branch pushes
- **Multi-arch**: Builds for `linux/amd64` and `linux/arm64`
- **Registry**: Pushes to Docker Hub with multiple tags
- **Caching**: Uses GitHub Actions cache for faster builds

#### 6. **Deploy** 🚀
- **Staging**: Deploys on `staging` branch pushes
- **Production**: Deploys on `main` branch pushes  
- **Environments**: Uses GitHub environment protection rules
- **Manual Approval**: Can require manual approval for production

#### 7. **Release** 📦
- **Trigger**: Only on version tags (e.g., `v1.0.0`)
- **GoReleaser**: Creates GitHub releases with binaries
- **Cross-platform**: Builds for multiple OS/architecture combinations

### Branch Strategy

```
main           # Production branch - triggers prod deployment
  ↑
staging        # Staging branch - triggers staging deployment  
  ↑
feature/*      # Feature branches - create PRs to staging
```

### Required Secrets

Set these in your GitHub repository settings → Secrets and variables → Actions:

```bash
# Docker Hub (for image publishing)
DOCKER_USERNAME=your-dockerhub-username
DOCKER_PASSWORD=your-dockerhub-token

# Codecov (for coverage reporting)  
CODECOV_TOKEN=your-codecov-token

# Deployment secrets (add as needed)
# KUBECONFIG, AWS_ACCESS_KEY_ID, etc.
```

### Environment Setup

1. **Enable GitHub Actions**
   ```bash
   # Workflows are automatically enabled when you push .github/workflows/go.yml
   ```

2. **Configure Environments** (Optional)
   - Go to Settings → Environments
   - Create `staging` and `production` environments
   - Add protection rules (required reviewers, wait timers)

3. **Set up Codecov**
   - Sign up at [codecov.io](https://codecov.io)
   - Link your GitHub repository
   - Add `CODECOV_TOKEN` secret

### Local Testing

Before pushing, ensure your code passes all checks:

```bash
# Full local validation
make test && make lint && make build

# Match CI environment
go test -v -coverprofile=coverage.out -covermode=atomic ./...
```

### Deployment Customization

The deployment jobs currently contain placeholder commands. Customize them for your infrastructure:

**Staging Deployment Example:**
```yaml
- name: Deploy to staging
  run: |
    kubectl apply -f k8s/staging/ 
    kubectl rollout status deployment/genje-api -n staging
```

**Production Deployment Example:**  
```yaml
- name: Deploy to production
  run: |
    aws ecs update-service --cluster prod --service genje-api
    aws ecs wait services-stable --cluster prod --services genje-api
```

### Workflow Triggers & Behavior

| **Event** | **Branch** | **Jobs Run** | **Deployment** |
|-----------|------------|--------------|----------------|
| Push | `main` | All jobs | ✅ Production |
| Push | `staging` | All jobs | ✅ Staging |
| Pull Request | `main`/`staging` | Test, Lint, Build, Security | ❌ None |
| Tag | `v*.*.*` | All jobs + Release | ❌ None |

### Monitoring & Observability

- **🔍 Code Coverage**: View coverage reports on [Codecov](https://codecov.io/gh/danielkosgei/genje-api)
- **🔒 Security Scan**: Check security issues in GitHub Security tab
- **📊 Go Report**: Code quality metrics on [Go Report Card](https://goreportcard.com/report/github.com/danielkosgei/genje-api)
- **🐳 Docker Hub**: Track image pulls and versions
- **📋 GitHub Actions**: Monitor workflow runs and build history

### Release Process

1. **Development**: Work on feature branches, create PRs to `staging`
2. **Staging**: Merge to `staging` branch for testing
3. **Production**: Merge `staging` to `main` for production deployment
4. **Release**: Tag `main` branch to create official releases

```bash
# Create a release
git checkout main
git pull origin main
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

### Troubleshooting CI/CD

**❌ Linting Failures**
```bash
# Fix locally before pushing
make lint
make format
```

**❌ Test Failures**  
```bash
# Run the exact same tests as CI
go test -v -coverprofile=coverage.out -covermode=atomic ./...
```

**❌ Docker Build Failures**
```bash
# Test Docker build locally
make docker-build
docker run --rm genje-api:latest /bin/sh -c "echo 'Container works'"
```

**❌ Missing Secrets**
- Check repository Settings → Secrets and variables → Actions  
- Ensure `DOCKER_USERNAME`, `DOCKER_PASSWORD`, `CODECOV_TOKEN` are set

**❌ Deployment Failures**
- Check GitHub Actions logs for specific error messages
- Verify environment variables in deployment jobs
- Test deployment commands locally first

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

