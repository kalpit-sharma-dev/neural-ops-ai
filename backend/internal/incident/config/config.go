package config

import (
	"os"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/spf13/viper"
)

// Config holds incident engine configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	Redis    RedisConfig    `mapstructure:"redis"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Incident IncidentConfig `mapstructure:"incident"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type KafkaConfig struct {
	Brokers       []string `mapstructure:"brokers"`
	LogTopics     []string `mapstructure:"log_topics"`
	AnomalyTopics []string `mapstructure:"anomaly_topics"`
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

type LLMConfig struct {
	Provider        string        `mapstructure:"provider"`
	OpenAIBaseURL   string        `mapstructure:"openai_base_url"`
	OpenAIAPIKey    string        `mapstructure:"openai_api_key"`
	OpenAIModel     string        `mapstructure:"openai_model"`
	AnthropicAPIKey string        `mapstructure:"anthropic_api_key"`
	OllamaBaseURL   string        `mapstructure:"ollama_base_url"`
	OllamaModel     string        `mapstructure:"ollama_model"`
	RCATimeout      time.Duration `mapstructure:"rca_timeout"`
	MaxRetries      int           `mapstructure:"max_retries"`
	InitialBackoff  time.Duration `mapstructure:"initial_backoff"`
}

type IncidentConfig struct {
	DedupWindow     time.Duration `mapstructure:"dedup_window"`
	DefaultTenant   string        `mapstructure:"default_tenant"`
	PaymentServices []string      `mapstructure:"payment_services"`
}

// Load reads incident configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("incident")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("INCIDENT")

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
	v.SetDefault("server.port", 8084)
	v.SetDefault("kafka.concurrency", 8)
	v.SetDefault("kafka.consumer_group", "incident-engine")
	v.SetDefault("kafka.log_topics", []string{"enriched-logs"})
	v.SetDefault("kafka.anomaly_topics", []string{"anomalies"})
	v.SetDefault("incident.dedup_window", "5m")
	v.SetDefault("incident.default_tenant", "default")
	v.SetDefault("clickhouse.transaction_table", "transaction_journeys")
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("postgres.dsn", "POSTGRES_DSN")
	_ = v.BindEnv("clickhouse.dsn", "CLICKHOUSE_DSN")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("llm.openai_api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("llm.anthropic_api_key", "ANTHROPIC_API_KEY")
	_ = v.BindEnv("llm.provider", "LLM_PROVIDER")
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

// LLMClientConfig converts to ai client config.
func (c *Config) LLMClientConfig() ai.ClientConfig {
	return ai.ClientConfig{
		Provider:        c.LLM.Provider,
		OpenAIBaseURL:   c.LLM.OpenAIBaseURL,
		OpenAIAPIKey:    c.LLM.OpenAIAPIKey,
		OpenAIModel:     c.LLM.OpenAIModel,
		AnthropicAPIKey: c.LLM.AnthropicAPIKey,
		OllamaBaseURL:   c.LLM.OllamaBaseURL,
		OllamaModel:     c.LLM.OllamaModel,
		RCATimeout:      c.LLM.RCATimeout,
		MaxRetries:      c.LLM.MaxRetries,
		InitialBackoff:  c.LLM.InitialBackoff,
	}
}
