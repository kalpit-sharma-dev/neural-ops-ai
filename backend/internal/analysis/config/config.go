package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/spf13/viper"
)

// Config holds analysis service configuration.
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Postgres      PostgresConfig      `mapstructure:"postgres"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Qdrant        QdrantConfig        `mapstructure:"qdrant"`
	LLM           LLMConfig           `mapstructure:"llm"`
	Analysis      AnalysisConfig      `mapstructure:"analysis"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type KafkaConfig struct {
	Brokers        []string          `mapstructure:"brokers"`
	LogTopics      []string          `mapstructure:"log_topics"`
	MetricTopics   []string          `mapstructure:"metric_topics"`
	AnomalyTopic   string            `mapstructure:"anomaly_topic"`
	ConsumerGroups map[string]string `mapstructure:"consumer_groups"`
	Concurrency    int               `mapstructure:"concurrency"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type ElasticsearchConfig struct {
	URL   string `mapstructure:"url"`
	Index string `mapstructure:"index"`
}

type RedisConfig struct {
	URL               string        `mapstructure:"url"`
	ClassificationTTL time.Duration `mapstructure:"classification_ttl"`
	ExplanationTTL    time.Duration `mapstructure:"explanation_ttl"`
}

type QdrantConfig struct {
	URL        string `mapstructure:"url"`
	Collection string `mapstructure:"collection"`
	VectorSize int    `mapstructure:"vector_size"`
}

type LLMConfig struct {
	Provider        string        `mapstructure:"provider"`
	OpenAIBaseURL   string        `mapstructure:"openai_base_url"`
	OpenAIAPIKey    string        `mapstructure:"openai_api_key"`
	OpenAIModel     string        `mapstructure:"openai_model"`
	EmbeddingModel  string        `mapstructure:"embedding_model"`
	AnthropicAPIKey string        `mapstructure:"anthropic_api_key"`
	AnthropicModel  string        `mapstructure:"anthropic_model"`
	OllamaBaseURL   string        `mapstructure:"ollama_base_url"`
	OllamaModel     string        `mapstructure:"ollama_model"`
	ExplainTimeout  time.Duration `mapstructure:"explain_timeout"`
	RCATimeout      time.Duration `mapstructure:"rca_timeout"`
	MaxRetries      int           `mapstructure:"max_retries"`
	InitialBackoff  time.Duration `mapstructure:"initial_backoff"`
}

type AnalysisConfig struct {
	LLMConfidenceThreshold   float64 `mapstructure:"llm_confidence_threshold"`
	EmbedBatchSize           int     `mapstructure:"embed_batch_size"`
	RecurringExceptionWindow int     `mapstructure:"recurring_exception_window"`
	AnomalyZScoreThreshold   float64 `mapstructure:"anomaly_zscore_threshold"`
}

// Load reads analysis configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("analysis")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("ANALYSIS")

	setDefaults(v)
	bindEnvOverrides(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

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
	v.SetDefault("server.port", 8082)
	v.SetDefault("kafka.concurrency", 8)
	v.SetDefault("kafka.log_topics", []string{"raw-logs", "enriched-logs"})
	v.SetDefault("kafka.metric_topics", []string{"raw-metrics"})
	v.SetDefault("kafka.anomaly_topic", "anomalies")
	v.SetDefault("elasticsearch.index", "enriched-logs")
	v.SetDefault("qdrant.collection", "log_embeddings")
	v.SetDefault("qdrant.vector_size", 1536)
	v.SetDefault("analysis.llm_confidence_threshold", 0.8)
	v.SetDefault("analysis.embed_batch_size", 100)
	v.SetDefault("analysis.recurring_exception_window", 100)
	v.SetDefault("analysis.anomaly_zscore_threshold", 3.0)
}

func bindEnvOverrides(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("postgres.dsn", "POSTGRES_DSN")
	_ = v.BindEnv("elasticsearch.url", "ELASTICSEARCH_URL")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("qdrant.url", "QDRANT_URL")
	_ = v.BindEnv("llm.provider", "LLM_PROVIDER")
	_ = v.BindEnv("llm.openai_api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("llm.anthropic_api_key", "ANTHROPIC_API_KEY")

	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		v.Set("kafka.brokers", splitAndTrim(brokers))
	}
}

func splitAndTrim(value string) []string {
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
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers must not be empty")
	}
	if c.Kafka.Concurrency <= 0 {
		return fmt.Errorf("kafka.concurrency must be positive")
	}
	if c.Postgres.DSN == "" {
		return fmt.Errorf("postgres.dsn must not be empty")
	}
	return nil
}

// LLMClientConfig converts config to ai.ClientConfig.
func (c *Config) LLMClientConfig() ai.ClientConfig {
	return ai.ClientConfig{
		Provider:        c.LLM.Provider,
		OpenAIBaseURL:   c.LLM.OpenAIBaseURL,
		OpenAIAPIKey:    c.LLM.OpenAIAPIKey,
		OpenAIModel:     c.LLM.OpenAIModel,
		EmbeddingModel:  c.LLM.EmbeddingModel,
		AnthropicAPIKey: c.LLM.AnthropicAPIKey,
		AnthropicModel:  c.LLM.AnthropicModel,
		OllamaBaseURL:   c.LLM.OllamaBaseURL,
		OllamaModel:     c.LLM.OllamaModel,
		ExplainTimeout:  c.LLM.ExplainTimeout,
		RCATimeout:      c.LLM.RCATimeout,
		MaxRetries:      c.LLM.MaxRetries,
		InitialBackoff:  c.LLM.InitialBackoff,
	}
}

// LogsConsumerGroup returns the logs consumer group name.
func (c *Config) LogsConsumerGroup() string {
	if group, ok := c.Kafka.ConsumerGroups["logs"]; ok && group != "" {
		return group
	}
	return "analysis-errors"
}

// MetricsConsumerGroup returns the metrics consumer group name.
func (c *Config) MetricsConsumerGroup() string {
	if group, ok := c.Kafka.ConsumerGroups["metrics"]; ok && group != "" {
		return group
	}
	return "analysis-metrics"
}
