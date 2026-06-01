package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds correlation engine configuration.
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Kafka       KafkaConfig       `mapstructure:"kafka"`
	Postgres    PostgresConfig    `mapstructure:"postgres"`
	ClickHouse  ClickHouseConfig  `mapstructure:"clickhouse"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Correlation CorrelationConfig `mapstructure:"correlation"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type KafkaConfig struct {
	Brokers       []string `mapstructure:"brokers"`
	EventTopics   []string `mapstructure:"event_topics"`
	TraceTopics   []string `mapstructure:"trace_topics"`
	LogTopics     []string `mapstructure:"log_topics"`
	ConsumerGroup string   `mapstructure:"consumer_group"`
	Concurrency   int      `mapstructure:"concurrency"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type ClickHouseConfig struct {
	DSN              string `mapstructure:"dsn"`
	TransactionTable string `mapstructure:"transaction_table"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type CorrelationConfig struct {
	DeploymentWatchWindow time.Duration `mapstructure:"deployment_watch_window"`
	DeploymentEvalWindow  time.Duration `mapstructure:"deployment_eval_window"`
	TemporalWindow        time.Duration `mapstructure:"temporal_window"`
	GraphFlushInterval    time.Duration `mapstructure:"graph_flush_interval"`
}

// Load reads correlation configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("correlation")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("CORRELATION")

	setDefaults(v)
	bindEnv(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8083)
	v.SetDefault("kafka.concurrency", 8)
	v.SetDefault("kafka.consumer_group", "correlation-engine")
	v.SetDefault("kafka.event_topics", []string{"raw-events"})
	v.SetDefault("kafka.trace_topics", []string{"raw-traces"})
	v.SetDefault("kafka.log_topics", []string{"enriched-logs"})
	v.SetDefault("correlation.deployment_watch_window", "15m")
	v.SetDefault("correlation.deployment_eval_window", "10m")
	v.SetDefault("correlation.temporal_window", "5m")
	v.SetDefault("correlation.graph_flush_interval", "1m")
	v.SetDefault("clickhouse.transaction_table", "transaction_journeys")
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("postgres.dsn", "POSTGRES_DSN")
	_ = v.BindEnv("clickhouse.dsn", "CLICKHOUSE_DSN")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		v.Set("kafka.brokers", splitComma(brokers))
	}
}

func splitComma(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// Validate validates configuration.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if c.Postgres.DSN == "" {
		return fmt.Errorf("postgres.dsn is required")
	}
	return nil
}
