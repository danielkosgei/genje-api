# Multi-stage build for Genje News Aggregator API
FROM rust:1.88-slim AS builder

# Install system dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Create app directory
WORKDIR /app

# Copy manifests
COPY Cargo.toml Cargo.lock ./

# Copy source code
COPY src ./src
COPY migrations ./migrations

# Build the application
RUN cargo build --release

# Runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libssl3 \
    libpq5 \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create a non-privileged user
RUN useradd --create-home --shell /bin/bash genje

# Set working directory
WORKDIR /app

# Copy the built binary
COPY --from=builder /app/target/release/genje /app/
COPY --from=builder /app/migrations /app/migrations

# Copy configuration
COPY config.toml /app/

# Change ownership
RUN chown -R genje:genje /app

# Switch to non-privileged user
USER genje

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./genje"] 