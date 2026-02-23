FROM golang:1.24.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /jalada ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata curl

RUN adduser -D -g '' appuser

COPY --from=builder /jalada /jalada
COPY --from=builder /app/internal/database/migrations /migrations

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/jalada"]
