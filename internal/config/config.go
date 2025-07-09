package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Aggregator AggregatorConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AggregatorConfig struct {
	Interval        time.Duration
	RequestTimeout  time.Duration
	UserAgent       string
	MaxContentSize  int
	MaxSummarySize  int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "./genje.db?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000&_foreign_keys=ON"),
			MaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Aggregator: AggregatorConfig{
			Interval:        getDuration("AGGREGATION_INTERVAL", 30*time.Minute),
			RequestTimeout:  getDuration("REQUEST_TIMEOUT", 30*time.Second),
			UserAgent:       getEnv("USER_AGENT", "Mozilla/5.0 (compatible; Genje-News-Aggregator/1.0)"),
			MaxContentSize:  getInt("MAX_CONTENT_SIZE", 10000),
			MaxSummarySize:  getInt("MAX_SUMMARY_SIZE", 300),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
} 