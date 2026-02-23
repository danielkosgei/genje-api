<p align="center">
  <h1 align="center">Jalada</h1>
  <p align="center">
    A dossier on Kenya's 2027 elections.
    <br />
    <em>Neutral. Comprehensive. Sourced.</em>
  </p>
</p>

<p align="center">
  <a href="https://github.com/danielkosgei/genje-api/actions/workflows/go.yml"><img src="https://github.com/danielkosgei/genje-api/actions/workflows/go.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/danielkosgei/genje-api/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-AGPL--3.0-blue.svg" alt="License: AGPL-3.0"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.22-00ADD8.svg" alt="Go Version">
  <img src="https://img.shields.io/badge/database-PostgreSQL%2016-336791.svg" alt="PostgreSQL">
</p>

---

**Jalada** (Swahili for *"record"* or *"file/dossier"*) is a free, open-source REST API that provides structured, sourced data on Kenyan politicians, political parties, elections, and real-time political news, built for the 2027 general election cycle.

It tracks **456 elected officials** across all seats (President, Governors, Senators, MPs, Women Representatives), **28 political parties**, **47 counties**, **290 constituencies**, and aggregates live news from Kenyan media outlets every 15 minutes.

The project is designed to power election dashboards, civic tech apps, media tools, and academic research with verifiable, neutral data.

## Features

- **Politician Dossiers** - Bio, education, career history, party affiliations, status (active/deceased)
- **Political Parties & Coalitions** - Full membership rosters, ideology, leadership
- **Election Tracking** - 2022 results, 2027 timeline milestones, candidacies
- **Live News Aggregation** - Auto-scraped from Kenyan RSS feeds every 15 minutes, tagged as election-related, linked to politician profiles
- **Trending & Analytics** - Politician mention rankings, sentiment, promise tracking, integrity flags
- **Geography** - All 47 counties, 290 constituencies, wards, polling stations
- **Self-Documenting** - Hit `/` or `/v1/` for the full endpoint and schema reference as JSON

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/danielkosgei/genje-api.git
cd genje-api
docker compose up -d
```

The API will be available at `http://localhost:8080`. The database is automatically migrated and seeded on first boot.

### Local Development

```bash
# Prerequisites: Go 1.22+, PostgreSQL 16+

cp .env.example .env        # edit DATABASE_URL to point to your Postgres
make run                     # builds and runs the server
```

### Verify

```bash
# Health check
curl http://localhost:8080/health

# Full API schema
curl http://localhost:8080/

# Search politicians
curl "http://localhost:8080/v1/politicians?q=ruto"

# Politician dossier
curl http://localhost:8080/v1/politicians/william-ruto
```

## API Reference

Hit the root URL (`/` or `/v1/`) for the complete self-documenting JSON schema with all 40 endpoints, parameters, and response types.

### Endpoints Overview

| Group | Endpoint | Description |
|-------|----------|-------------|
| **Politicians** | `GET /v1/politicians` | List and search all 456 politicians |
| | `GET /v1/politicians/{slug}` | Full dossier (bio, education, career, party, integrity) |
| | `GET /v1/politicians/{slug}/news` | News articles mentioning this politician |
| | `GET /v1/politicians/{slug}/court-cases` | Court cases and legal proceedings |
| | `GET /v1/politicians/{slug}/promises` | Campaign promises and fulfilment status |
| | `GET /v1/politicians/{slug}/achievements` | Notable achievements |
| | `GET /v1/politicians/{slug}/controversies` | Controversies and scandals |
| | `GET /v1/politicians/{slug}/assets` | Declared assets (EACC filings) |
| | `GET /v1/politicians/{slug}/voting-record` | Parliamentary voting record |
| | `GET /v1/politicians/{slug}/attendance` | Parliamentary attendance |
| | `GET /v1/politicians/{slug}/affiliations` | Political affiliations graph |
| | `GET /v1/politicians/{slug}/sentiment` | Public sentiment analysis |
| | `GET /v1/politicians/{slug}/events` | Associated events and rallies |
| **Parties** | `GET /v1/parties` | All 28 political parties |
| | `GET /v1/parties/{slug}` | Party detail with member roster |
| **Coalitions** | `GET /v1/coalitions` | Political coalitions |
| | `GET /v1/coalitions/{slug}` | Coalition detail with member parties |
| **Elections** | `GET /v1/elections` | All elections (2022, 2027) |
| | `GET /v1/elections/{id}/timeline` | Election milestones |
| | `GET /v1/elections/{id}/candidates` | Registered candidates |
| | `GET /v1/elections/{id}/results` | Results by constituency |
| **Geography** | `GET /v1/counties` | All 47 counties |
| | `GET /v1/counties/{code}/constituencies` | Constituencies in a county |
| | `GET /v1/constituencies/{code}` | Constituency detail |
| **News** | `GET /v1/news` | Aggregated news (auto-updated) |
| | `GET /v1/news/{id}` | Article detail |
| | `GET /v1/sources` | Official data sources |
| **Analytics** | `GET /v1/analytics/trending` | Trending politicians by mentions |
| | `GET /v1/analytics/sentiment` | Aggregate sentiment |
| | `GET /v1/analytics/promises` | Promise fulfilment stats |
| | `GET /v1/analytics/integrity` | Integrity flags summary |
| | `GET /v1/analytics/attendance` | Attendance rankings |
| **Timeline** | `GET /v1/timeline` | 2027 election timeline |
| | `GET /v1/events` | Political events and rallies |

### Pagination

List endpoints support pagination via `limit` and `offset` query parameters:

```json
{
  "data": [...],
  "total": 456,
  "limit": 20,
  "offset": 0,
  "has_more": true
}
```

### Errors

All errors follow a consistent format:

```json
{
  "error": "Not Found",
  "code": 404,
  "message": "politician not found"
}
```

## Data Sources

| Source | Data |
|--------|------|
| [IEBC](https://www.iebc.or.ke) | Candidate lists, election results, constituencies, polling stations |
| [Kenya Gazette](http://kenyalaw.org/kenya_gazette/) | Official appointments, party registrations |
| [EACC](https://eacc.go.ke) | Asset declarations, integrity investigations |
| [Hansard](http://www.parliament.go.ke) | Voting records, parliamentary attendance |
| [Wikipedia, 13th Parliament](https://en.wikipedia.org/wiki/13th_Parliament_of_Kenya) | MP/Senator/Women Rep verification |
| Kenyan Media RSS | Capital FM, The Standard, Nation, Citizen, KTN, The Star, Tuko |

## Project Structure

```
jalada/
├── cmd/server/              # Application entrypoint
│   └── main.go
├── internal/
│   ├── config/              # Environment configuration
│   ├── database/            # PostgreSQL pool, migrations
│   │   └── migrations/      # SQL migration files
│   ├── handlers/            # HTTP handlers and router
│   ├── middleware/           # CORS, logging, rate limiting, request ID
│   ├── models/              # Domain types (17 model files)
│   ├── repository/          # Database queries (8 repo files)
│   ├── scraper/             # RSS fetcher, politician mention linker, scheduler
│   ├── seeder/              # Seed data loader
│   │   └── data/            # Embedded JSON seed files
│   └── services/            # Business logic layer
├── .github/workflows/       # CI/CD (test, lint, Docker build)
├── Dockerfile               # Multi-stage production build
├── docker-compose.yml       # Local dev stack (API + PostgreSQL)
├── Makefile                 # Build, test, migrate, lint targets
└── LICENSE                  # AGPL-3.0
```

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22+ |
| Database | PostgreSQL 16 |
| Router | [chi](https://github.com/go-chi/chi) |
| DB Driver | [pgx](https://github.com/jackc/pgx) with connection pooling |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) (embedded SQL) |
| RSS Parsing | [gofeed](https://github.com/mmcdole/gofeed) |
| Logging | [zerolog](https://github.com/rs/zerolog) (structured JSON) |
| Rate Limiting | `golang.org/x/time/rate` (token bucket) |
| Containerization | Docker, Docker Compose |
| CI/CD | GitHub Actions |

## Development

```bash
make build          # compile binary
make run            # build + run
make test           # run tests with race detector
make lint           # golangci-lint
make format         # gofmt + goimports
make tidy           # go mod tidy

make migrate-up     # apply migrations
make migrate-down   # rollback last migration
make migrate-create # create new migration (NAME=xxx)

make docker-build   # build Docker image
make compose-up     # start docker compose stack
make compose-down   # stop docker compose stack
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | | PostgreSQL connection string |
| `ENV` | `development` | Environment (`development`, `production`) |
| `AGGREGATION_INTERVAL` | `15m` | News scraping interval |
| `REQUEST_TIMEOUT` | `30s` | HTTP client timeout for scrapers |
| `USER_AGENT` | `Jalada/1.0` | User-Agent for outbound requests |
| `LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `LOG_JSON` | `false` | JSON-formatted log output |

## Contributing

Contributions are welcome. Jalada is built for the public interest. If you have access to official Kenyan data sources or can help improve accuracy, please open an issue or PR.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/add-mca-data`)
3. Commit your changes
4. Push and open a Pull Request

Please ensure your code passes `make lint` and `make test` before submitting.

### Areas That Need Help

- [ ] MCA (Member of County Assembly) data for all ~1,450 wards
- [ ] Historical election results (2013, 2017)
- [ ] Hansard voting record scraping
- [ ] EACC asset declaration data
- [ ] Social media monitoring integration
- [ ] Sentiment analysis pipeline

## License

Jalada is licensed under the [GNU Affero General Public License v3.0](LICENSE).

You are free to use, modify, and distribute this software. If you run a modified version as a network service (e.g. your own API), you must make the source code available to users of that service. This ensures the project and any derivatives remain free and open, which matters for election transparency tools.

## Acknowledgements

- [IEBC](https://www.iebc.or.ke) for election data and infrastructure
- [Wikipedia](https://en.wikipedia.org/wiki/13th_Parliament_of_Kenya) for verified parliamentary records
- Kenyan media outlets (Capital FM, The Standard, Nation, Citizen, The Star, KTN, Tuko) for RSS news feeds
- The Kenyan civic tech community
