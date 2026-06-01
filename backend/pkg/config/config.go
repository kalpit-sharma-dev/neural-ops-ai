package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds shared service configuration loaded from environment variables.
type Config struct {
	ServiceName string
	Environment string
	HTTPPort    int

	PostgresDSN       string
	ClickHouseDSN     string
	ElasticsearchURL  string
	QdrantURL         string
	RedisURL          string
	KafkaBrokers      string
	KafkaConsumerGroup string
}

// Load reads configuration from environment variables with sensible defaults for local development.
func Load(serviceName string) (*Config, error) {
	port, err := strconv.Atoi(getEnv("HTTP_PORT", defaultPort(serviceName)))
	if err != nil {
		return nil, fmt.Errorf("invalid HTTP_PORT: %w", err)
	}

	return &Config{
		ServiceName:        serviceName,
		Environment:        getEnv("ENVIRONMENT", "development"),
		HTTPPort:           port,
		PostgresDSN:        getEnv("POSTGRES_DSN", "postgres://neuralops:neuralops@postgres:5432/neuralops?sslmode=disable"),
		ClickHouseDSN:      getEnv("CLICKHOUSE_DSN", "clickhouse://default:@clickhouse:9000/neuralops"),
		ElasticsearchURL:   getEnv("ELASTICSEARCH_URL", "http://elasticsearch:9200"),
		QdrantURL:          getEnv("QDRANT_URL", "http://qdrant:6333"),
		RedisURL:           getEnv("REDIS_URL", "redis://redis:6379/0"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "kafka:9092"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", serviceName),
	}, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Getenv returns an environment variable or fallback (exported for auxiliary binaries).
func Getenv(key, fallback string) string {
	return getEnv(key, fallback)
}

// GetenvInt parses an integer environment variable.
func GetenvInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func defaultPort(serviceName string) string {
	ports := map[string]string{
		"gateway":     "8080",
		"ingestion":   "8081",
		"analysis":    "8082",
		"correlation": "8083",
		"incident":    "8084",
		"search":      "8085",
	"alerting":    "8086",
		"collector":   "8090",
	}
	if port, ok := ports[serviceName]; ok {
		return port
	}
	return "8080"
}
