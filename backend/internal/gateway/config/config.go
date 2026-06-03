package config

import (
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/spf13/viper"
)

// Config holds API gateway configuration.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Services  ServicesConfig  `mapstructure:"services"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Tenant    TenantConfig    `mapstructure:"tenant"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	AuthRateLimit AuthRateLimitConfig `mapstructure:"auth_rate_limit"`
	TenantQuota TenantQuotaConfig `mapstructure:"tenant_quota"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	LLM       LLMConfig       `mapstructure:"llm"`
	Dashboard DashboardConfig `mapstructure:"dashboard"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	DemoMode  bool            `mapstructure:"demo_mode"`
}

type PrometheusConfig struct {
	URL string `mapstructure:"url"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Environment string `mapstructure:"environment"`
}

type ServicesConfig struct {
	Ingestion   string `mapstructure:"ingestion"`
	Analysis    string `mapstructure:"analysis"`
	Correlation string `mapstructure:"correlation"`
	Incident    string `mapstructure:"incident"`
	Search      string `mapstructure:"search"`
	Alerting    string `mapstructure:"alerting"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type AuthConfig struct {
	Disabled         bool                       `mapstructure:"disabled"`
	JWTPrivateKeyPEM string                     `mapstructure:"jwt_private_key_pem"`
	JWTPublicKeyPEM  string                     `mapstructure:"jwt_public_key_pem"`
	JWTIssuer        string                     `mapstructure:"jwt_issuer"`
	JWTAudience      string                     `mapstructure:"jwt_audience"`
	AccessTokenTTL   time.Duration              `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration              `mapstructure:"refresh_token_ttl"`
	FrontendURL      string                     `mapstructure:"frontend_url"`
	AllowDevLogin    bool                       `mapstructure:"allow_dev_login"`
	APIKeys          map[string]APIKeyPrincipal `mapstructure:"api_keys"`
	OIDC             OIDCConfig                 `mapstructure:"oidc"`
	SAML             SAMLConfig                 `mapstructure:"saml"`
	MTLS             MTLSConfig                 `mapstructure:"mtls"`
}

type APIKeyPrincipal struct {
	TenantID string `mapstructure:"tenant_id"`
	UserID   string `mapstructure:"user_id"`
	Role     string `mapstructure:"role"`
}

type OIDCConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	Issuer          string   `mapstructure:"issuer"`
	IssuerInternal  string   `mapstructure:"issuer_internal"`
	BrowserIssuer   string   `mapstructure:"browser_issuer"`
	ClientID        string   `mapstructure:"client_id"`
	ClientSecret    string   `mapstructure:"client_secret"`
	RedirectURL     string   `mapstructure:"redirect_url"`
	Scopes          []string `mapstructure:"scopes"`
	JWKSURL         string   `mapstructure:"jwks_url"`
	Audience        string   `mapstructure:"audience"`
	DefaultTenantID string   `mapstructure:"default_tenant_id"`
}

type SAMLConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	EntityID        string `mapstructure:"entity_id"`
	MetadataURL     string `mapstructure:"metadata_url"`
	SSOURL          string `mapstructure:"sso_url"`
	ACSURL          string `mapstructure:"acs_url"`
	CertificatePEM  string `mapstructure:"certificate_pem"`
	KeyPEM          string `mapstructure:"key_pem"`
	CertificateFile string `mapstructure:"certificate_file"`
	KeyFile         string `mapstructure:"key_file"`
	DefaultTenant   string `mapstructure:"default_tenant"`
	DefaultRole     string `mapstructure:"default_role"`
}

type MTLSConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	CAFile     string `mapstructure:"ca_file"`
	ServerName string `mapstructure:"server_name"`
}

type TenantConfig struct {
	DefaultTenant  string   `mapstructure:"default_tenant"`
	AllowedTenants []string `mapstructure:"allowed_tenants"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

type AuthRateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
}

// TenantQuotaConfig bounds per-tenant daily ingest volume for multi-tenant
// fairness. Zero quota values mean "unlimited"; empty Paths uses the built-in
// ingest path set.
type TenantQuotaConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	DailyRequestQuota int64    `mapstructure:"daily_request_quota"`
	DailyBytesQuota   int64    `mapstructure:"daily_bytes_quota"`
	Paths             []string `mapstructure:"paths"`
}

type ClickHouseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
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

type DashboardConfig struct {
	FetchTimeout time.Duration `mapstructure:"fetch_timeout"`
}

type WebSocketConfig struct {
	PingInterval time.Duration `mapstructure:"ping_interval"`
}

// Load reads gateway configuration.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("gateway")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")
	v.AutomaticEnv()
	v.SetEnvPrefix("GATEWAY")

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
	v.SetDefault("server.port", 8080)
	v.SetDefault("services.ingestion", "http://ingestion:8081")
	v.SetDefault("services.analysis", "http://analysis:8082")
	v.SetDefault("services.correlation", "http://correlation:8083")
	v.SetDefault("services.incident", "http://incident:8084")
	v.SetDefault("services.search", "http://search:8085")
	v.SetDefault("services.alerting", "http://alerting:8086")
	v.SetDefault("auth.disabled", true)
	v.SetDefault("auth.jwt_issuer", "neuralops")
	v.SetDefault("auth.jwt_audience", "neuralops-api")
	v.SetDefault("auth.access_token_ttl", "1h")
	v.SetDefault("auth.refresh_token_ttl", "168h")
	v.SetDefault("auth.frontend_url", "http://localhost:5173")
	v.SetDefault("auth.allow_dev_login", true)
	v.SetDefault("tenant.default_tenant", "00000000-0000-0000-0000-000000000002")
	v.SetDefault("rate_limit.requests_per_minute", 600)
	v.SetDefault("rate_limit.burst", 100)
	v.SetDefault("auth_rate_limit.enabled", true)
	v.SetDefault("auth_rate_limit.requests_per_minute", 20)
	v.SetDefault("tenant_quota.enabled", false)
	v.SetDefault("tenant_quota.daily_request_quota", 0)
	v.SetDefault("tenant_quota.daily_bytes_quota", 0)
	v.SetDefault("dashboard.fetch_timeout", "3s")
	v.SetDefault("websocket.ping_interval", "30s")
	v.SetDefault("prometheus.url", "http://prometheus:9090")
	v.SetDefault("demo_mode", false)
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("server.port", "HTTP_PORT")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("postgres.dsn", "POSTGRES_DSN")
	_ = v.BindEnv("clickhouse.dsn", "CLICKHOUSE_DSN")
	_ = v.BindEnv("auth.disabled", "AUTH_DISABLED")
	_ = v.BindEnv("auth.jwt_private_key_pem", "JWT_PRIVATE_KEY_PEM")
	_ = v.BindEnv("auth.jwt_public_key_pem", "JWT_PUBLIC_KEY_PEM")
	_ = v.BindEnv("auth.frontend_url", "FRONTEND_URL")
	_ = v.BindEnv("auth.allow_dev_login", "AUTH_ALLOW_DEV_LOGIN")
	_ = v.BindEnv("auth.oidc.enabled", "OIDC_ENABLED")
	_ = v.BindEnv("auth.oidc.issuer", "OIDC_ISSUER")
	_ = v.BindEnv("auth.oidc.issuer_internal", "OIDC_ISSUER_INTERNAL")
	_ = v.BindEnv("auth.oidc.browser_issuer", "OIDC_BROWSER_ISSUER")
	_ = v.BindEnv("auth.oidc.client_id", "OIDC_CLIENT_ID")
	_ = v.BindEnv("auth.oidc.client_secret", "OIDC_CLIENT_SECRET")
	_ = v.BindEnv("auth.oidc.redirect_url", "OIDC_REDIRECT_URL")
	_ = v.BindEnv("auth.saml.enabled", "SAML_ENABLED")
	_ = v.BindEnv("auth.saml.entity_id", "SAML_ENTITY_ID")
	_ = v.BindEnv("auth.saml.sso_url", "SAML_SSO_URL")
	_ = v.BindEnv("auth.saml.acs_url", "SAML_ACS_URL")
	_ = v.BindEnv("auth.saml.metadata_url", "SAML_METADATA_URL")
	_ = v.BindEnv("auth.saml.certificate_file", "SAML_CERT_FILE")
	_ = v.BindEnv("auth.saml.key_file", "SAML_KEY_FILE")
	_ = v.BindEnv("auth.mtls.enabled", "INTERNAL_MTLS_ENABLED")
	_ = v.BindEnv("auth.mtls.cert_file", "INTERNAL_MTLS_CERT_FILE")
	_ = v.BindEnv("auth.mtls.key_file", "INTERNAL_MTLS_KEY_FILE")
	_ = v.BindEnv("auth.mtls.ca_file", "INTERNAL_MTLS_CA_FILE")
	_ = v.BindEnv("auth.mtls.server_name", "INTERNAL_MTLS_SERVER_NAME")
	_ = v.BindEnv("llm.openai_api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("llm.provider", "LLM_PROVIDER")
	_ = v.BindEnv("services.ingestion", "INGESTION_URL")
	_ = v.BindEnv("services.analysis", "ANALYSIS_URL")
	_ = v.BindEnv("services.correlation", "CORRELATION_URL")
	_ = v.BindEnv("services.incident", "INCIDENT_URL")
	_ = v.BindEnv("services.search", "SEARCH_URL")
	_ = v.BindEnv("services.alerting", "ALERTING_URL")
	_ = v.BindEnv("demo_mode", "DEMO_MODE")
	_ = v.BindEnv("tenant_quota.enabled", "TENANT_QUOTA_ENABLED")
	_ = v.BindEnv("tenant_quota.daily_request_quota", "TENANT_QUOTA_DAILY_REQUESTS")
	_ = v.BindEnv("tenant_quota.daily_bytes_quota", "TENANT_QUOTA_DAILY_BYTES")
	_ = v.BindEnv("prometheus.url", "PROMETHEUS_URL")
	_ = v.BindEnv("server.environment", "ENVIRONMENT")
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if c.Services.Ingestion == "" {
		return fmt.Errorf("services.ingestion is required")
	}
	return c.ValidateProduction()
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
