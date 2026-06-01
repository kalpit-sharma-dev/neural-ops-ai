package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds ingestion service configuration.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Parser     ParserConfig     `mapstructure:"parser"`
	Connectors ConnectorsConfig `mapstructure:"connectors"`
	Ingestion  IngestionConfig  `mapstructure:"ingestion"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type KafkaConfig struct {
	Brokers       []string      `mapstructure:"brokers"`
	TopicPrefix   string        `mapstructure:"topic_prefix"`
	BatchSize     int           `mapstructure:"batch_size"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	DLQTopic      string        `mapstructure:"dlq_topic"`
}

type ClickHouseConfig struct {
	DSN           string        `mapstructure:"dsn"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	MaxBatchSize  int           `mapstructure:"max_batch_size"`
	MetricsTable  string        `mapstructure:"metrics_table"`
}

type RateLimitConfig struct {
	PerTenant map[string]float64 `mapstructure:"per_tenant"`
}

type ParserConfig struct {
	StackTraceLanguages []string `mapstructure:"stack_trace_languages"`
	TraceIDPatterns     []string `mapstructure:"trace_id_patterns"`
	TxnIDPatterns       []string `mapstructure:"txn_id_patterns"`
}

type FileTailConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Paths   []string `mapstructure:"paths"`
}

type SyslogConfig struct {
	Enabled bool `mapstructure:"enabled"`
	UDPPort int  `mapstructure:"udp_port"`
	TCPPort int  `mapstructure:"tcp_port"`
}

type FluentConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type ConnectorsConfig struct {
	FileTail FileTailConfig `mapstructure:"file_tail"`
	Syslog   SyslogConfig   `mapstructure:"syslog"`
	Fluent   FluentConfig   `mapstructure:"fluent"`
}

type IngestionConfig struct {
	IngestorID     string        `mapstructure:"ingestor_id"`
	Datacenter     string        `mapstructure:"datacenter"`
	GeoIPEnabled   bool          `mapstructure:"geoip_enabled"`
	MaxBatchSize   int           `mapstructure:"max_batch_size"`
	TimestampSkew  time.Duration `mapstructure:"timestamp_skew"`
}

// Load reads ingestion configuration from file and environment variables.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("ingestion")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	v.SetEnvPrefix("INGESTION")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	bindEnvOverrides(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8081)
	v.SetDefault("server.environment", "development")
	v.SetDefault("kafka.brokers", []string{"kafka:9092"})
	v.SetDefault("kafka.batch_size", 500)
	v.SetDefault("kafka.flush_interval", "1s")
	v.SetDefault("kafka.dlq_topic", "dlq-logs")
	v.SetDefault("clickhouse.dsn", "clickhouse://default:@clickhouse:9000/neuralops")
	v.SetDefault("clickhouse.flush_interval", "1s")
	v.SetDefault("clickhouse.max_batch_size", 50000)
	v.SetDefault("clickhouse.metrics_table", "metrics")
	v.SetDefault("rate_limit.per_tenant.startup", 10000)
	v.SetDefault("rate_limit.per_tenant.business", 100000)
	v.SetDefault("rate_limit.per_tenant.enterprise", 0)
	v.SetDefault("ingestion.ingestor_id", "ingestion-1")
	v.SetDefault("ingestion.datacenter", "local")
	v.SetDefault("ingestion.max_batch_size", 10000)
	v.SetDefault("ingestion.timestamp_skew", "1h")
}

func bindEnvOverrides(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("server.environment", "ENVIRONMENT")
	_ = v.BindEnv("clickhouse.dsn", "CLICKHOUSE_DSN")

	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		v.Set("kafka.brokers", splitAndTrim(brokers))
	}
}

// Validate ensures configuration values are usable.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers must not be empty")
	}
	if c.Kafka.BatchSize <= 0 {
		return fmt.Errorf("kafka.batch_size must be positive")
	}
	if c.ClickHouse.MaxBatchSize <= 0 {
		return fmt.Errorf("clickhouse.max_batch_size must be positive")
	}
	if c.Ingestion.MaxBatchSize <= 0 || c.Ingestion.MaxBatchSize > 10000 {
		return fmt.Errorf("ingestion.max_batch_size must be between 1 and 10000")
	}
	if c.Ingestion.TimestampSkew <= 0 {
		return fmt.Errorf("ingestion.timestamp_skew must be positive")
	}
	return nil
}

// Topic returns a Kafka topic name with optional prefix.
func (c *Config) Topic(name string) string {
	if c.Kafka.TopicPrefix == "" {
		return name
	}
	return c.Kafka.TopicPrefix + name
}

// RateForPlan returns the per-second rate limit for a tenant plan.
func (c *Config) RateForPlan(plan string) float64 {
	if rate, ok := c.RateLimit.PerTenant[strings.ToLower(plan)]; ok {
		return rate
	}
	return c.RateLimit.PerTenant["startup"]
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
