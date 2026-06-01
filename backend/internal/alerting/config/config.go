package config

import (
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/spf13/viper"
)

// Config holds alerting service configuration.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Incident  IncidentConfig  `mapstructure:"incident"`
	LLM       LLMConfig       `mapstructure:"llm"`
	Alerting  AlertingConfig  `mapstructure:"alerting"`
	Oncall    OncallConfig    `mapstructure:"oncall"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type IncidentConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type LLMConfig struct {
	Provider        string        `mapstructure:"provider"`
	OpenAIBaseURL   string        `mapstructure:"openai_base_url"`
	OpenAIAPIKey    string        `mapstructure:"openai_api_key"`
	OpenAIModel     string        `mapstructure:"openai_model"`
	AnthropicAPIKey string        `mapstructure:"anthropic_api_key"`
	OllamaBaseURL   string        `mapstructure:"ollama_base_url"`
	OllamaModel     string        `mapstructure:"ollama_model"`
	MaxRetries      int           `mapstructure:"max_retries"`
	InitialBackoff  time.Duration `mapstructure:"initial_backoff"`
}

type AlertingConfig struct {
	DefaultTenant           string        `mapstructure:"default_tenant"`
	GroupWindow             time.Duration `mapstructure:"group_window"`
	EscalationCheckInterval time.Duration `mapstructure:"escalation_check_interval"`
	P1EscalateL2            time.Duration `mapstructure:"p1_escalate_l2"`
	P1EscalateManager       time.Duration `mapstructure:"p1_escalate_manager"`
	DemoMode                bool          `mapstructure:"demo_mode"`
}

type OncallConfig struct {
	SchedulesFile string `mapstructure:"schedules_file"`
}

// Load reads alerting configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("alerting")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("ALERTING")

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
	v.SetDefault("server.port", 8086)
	v.SetDefault("incident.base_url", "http://incident:8084")
	v.SetDefault("alerting.default_tenant", "default")
	v.SetDefault("alerting.group_window", "5m")
	v.SetDefault("alerting.escalation_check_interval", "1m")
	v.SetDefault("alerting.p1_escalate_l2", "5m")
	v.SetDefault("alerting.p1_escalate_manager", "15m")
	v.SetDefault("alerting.demo_mode", false)
	v.SetDefault("oncall.schedules_file", "configs/oncall_schedules.json")
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("postgres.dsn", "POSTGRES_DSN")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("llm.openai_api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("llm.provider", "LLM_PROVIDER")
	_ = v.BindEnv("incident.base_url", "INCIDENT_URL")
	_ = v.BindEnv("alerting.demo_mode", "DEMO_MODE")
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if c.Postgres.DSN == "" {
		return fmt.Errorf("postgres.dsn is required")
	}
	return nil
}

func (c *Config) LLMClientConfig() ai.ClientConfig {
	return ai.ClientConfig{
		Provider:        c.LLM.Provider,
		OpenAIBaseURL:   c.LLM.OpenAIBaseURL,
		OpenAIAPIKey:    c.LLM.OpenAIAPIKey,
		OpenAIModel:     c.LLM.OpenAIModel,
		AnthropicAPIKey: c.LLM.AnthropicAPIKey,
		OllamaBaseURL:   c.LLM.OllamaBaseURL,
		OllamaModel:     c.LLM.OllamaModel,
		MaxRetries:      c.LLM.MaxRetries,
		InitialBackoff:  c.LLM.InitialBackoff,
	}
}
