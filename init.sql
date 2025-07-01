-- Initialize the Genje database
-- This script is run when the PostgreSQL container starts

-- Create database if it doesn't exist (handled by docker-compose environment)

-- Set timezone
SET TIME ZONE 'UTC';

-- Create extensions that might be useful
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Note: Tables will be created by sqlx migrations when the API service starts 