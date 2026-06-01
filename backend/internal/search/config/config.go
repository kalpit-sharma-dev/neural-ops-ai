package config

import (
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/spf13/viper"
)

// Config holds search service configuration.
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Qdrant        QdrantConfig        `mapstructure:"qdrant"`
	ClickHouse    ClickHouseConfig    `mapstructure:"clickhouse"`
	LLM           LLMConfig           `mapstructure:"llm"`
	Search        SearchConfig        `mapstructure:"search"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type ElasticsearchConfig struct {
	URL             string `mapstructure:"url"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Index           string `mapstructure:"index"`
	IndexAlias      string `mapstructure:"index_alias"`
	IndexPrefix     string `mapstructure:"index_prefix"`
	InitialIndex    string `mapstructure:"initial_index"`
	ILMPolicy       string `mapstructure:"ilm_policy"`
	RolloverMaxSize string `mapstructure:"rollover_max_size"`
	RolloverMaxAge  string `mapstructure:"rollover_max_age"`
	HotDays         int    `mapstructure:"hot_days"`
	WarmDays        int    `mapstructure:"warm_days"`
	ColdDays        int    `mapstructure:"cold_days"`
	DeleteDays      int    `mapstructure:"delete_days"`
}

type QdrantConfig struct {
	URL        string `mapstructure:"url"`
	Collection string `mapstructure:"collection"`
	VectorSize int    `mapstructure:"vector_size"`
}

type ClickHouseConfig struct {
	DSN              string `mapstructure:"dsn"`
	TransactionTable string `mapstructure:"transaction_table"`
}

type LLMConfig struct {
	Provider        string        `mapstructure:"provider"`
	OpenAIBaseURL   string        `mapstructure:"openai_base_url"`
	OpenAIAPIKey    string        `mapstructure:"openai_api_key"`
	OpenAIModel     string        `mapstructure:"openai_model"`
	EmbeddingModel  string        `mapstructure:"embedding_model"`
	AnthropicAPIKey string        `mapstructure:"anthropic_api_key"`
	OllamaBaseURL   string        `mapstructure:"ollama_base_url"`
	OllamaModel     string        `mapstructure:"ollama_model"`
	MaxRetries      int           `mapstructure:"max_retries"`
	InitialBackoff  time.Duration `mapstructure:"initial_backoff"`
}

type SearchConfig struct {
	DefaultPageSize     int     `mapstructure:"default_page_size"`
	MaxPageSize         int     `mapstructure:"max_page_size"`
	HybridBM25Weight    float64 `mapstructure:"hybrid_bm25_weight"`
	HybridVectorWeight  float64 `mapstructure:"hybrid_vector_weight"`
}

// Load reads search configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("search")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("SEARCH")

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
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8085)
	v.SetDefault("elasticsearch.index", "enriched-logs")
	v.SetDefault("elasticsearch.index_alias", "neuralops-logs")
	v.SetDefault("elasticsearch.index_prefix", "neuralops-logs")
	v.SetDefault("elasticsearch.initial_index", "neuralops-logs-000001")
	v.SetDefault("qdrant.collection", "log_embeddings")
	v.SetDefault("qdrant.vector_size", 1536)
	v.SetDefault("search.default_page_size", 50)
	v.SetDefault("search.max_page_size", 500)
	v.SetDefault("search.hybrid_bm25_weight", 0.5)
	v.SetDefault("search.hybrid_vector_weight", 0.5)
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("elasticsearch.url", "ELASTICSEARCH_URL")
	_ = v.BindEnv("qdrant.url", "QDRANT_URL")
	_ = v.BindEnv("clickhouse.dsn", "CLICKHOUSE_DSN")
	_ = v.BindEnv("llm.openai_api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("llm.provider", "LLM_PROVIDER")
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if c.Elasticsearch.URL == "" {
		return fmt.Errorf("elasticsearch.url is required")
	}
	return nil
}

func (c *Config) LLMClientConfig() ai.ClientConfig {
	return ai.ClientConfig{
		Provider:        c.LLM.Provider,
		OpenAIBaseURL:   c.LLM.OpenAIBaseURL,
		OpenAIAPIKey:    c.LLM.OpenAIAPIKey,
		OpenAIModel:     c.LLM.OpenAIModel,
		EmbeddingModel:  c.LLM.EmbeddingModel,
		AnthropicAPIKey: c.LLM.AnthropicAPIKey,
		OllamaBaseURL:   c.LLM.OllamaBaseURL,
		OllamaModel:     c.LLM.OllamaModel,
		MaxRetries:      c.LLM.MaxRetries,
		InitialBackoff:  c.LLM.InitialBackoff,
	}
}

func (c *ElasticsearchConfig) ResolveIndex() string {
	if c.IndexAlias != "" {
		return c.IndexAlias
	}
	if c.Index != "" {
		return c.Index
	}
	if c.InitialIndex != "" {
		return c.InitialIndex
	}
	return c.IndexPrefix + "-000001"
}
