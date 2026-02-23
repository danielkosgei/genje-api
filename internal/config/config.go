package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Aggregation AggregationConfig
	Log         LogConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL string
}

type AggregationConfig struct {
	Interval       time.Duration
	RequestTimeout time.Duration
	UserAgent      string
}

type LogConfig struct {
	Level  string
	JSON   bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://jalada:jalada@localhost:5432/jalada?sslmode=disable"),
		},
		Aggregation: AggregationConfig{
			Interval:       parseDuration(getEnv("AGGREGATION_INTERVAL", "15m")),
			RequestTimeout: parseDuration(getEnv("REQUEST_TIMEOUT", "30s")),
			UserAgent:      getEnv("USER_AGENT", "Jalada/1.0"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
			JSON:  getEnv("LOG_JSON", "false") == "true",
		},
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
