# Jalada

Neutral, comprehensive Kenya 2027 election tracker API.

Jalada provides structured, sourced data on politicians, political parties, election timelines, manifestos, promises, court cases, voting records, and real-time election news — designed for transparency and accountability.

## Quick Start

```bash
# Start PostgreSQL and the API
docker compose up -d

# Or run locally
cp .env.example .env
make run
```

## API

All endpoints are prefixed with `/api/v1/`. See the full endpoint list in the documentation.

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Router:** chi
- **Migrations:** golang-migrate
