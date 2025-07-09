FROM golang:1.24.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

FROM alpine:latest

WORKDIR /app

# Install required packages for network connectivity and SSL
RUN apk add --no-cache \
    sqlite \
    ca-certificates \
    tzdata \
    curl \
    openssl \
    && update-ca-certificates

# Copy the binary
COPY --from=builder /app/main .

# Create data directory with proper permissions
RUN mkdir -p /app/data && chmod 755 /app/data

# Add a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of app directory
RUN chown -R appuser:appgroup /app

# Set environment variables for better networking
ENV GODEBUG=netdns=go
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV SSL_CERT_DIR=/etc/ssl/certs

# Switch to non-root user
USER appuser

EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["./main"]