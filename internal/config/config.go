package config

import (
	"fmt"
	"time"

	specerr "github.com/rajeshshrirao/specwatch/internal/errors"
	"github.com/rajeshshrirao/specwatch/internal/llm"
)

// Environment represents the runtime environment
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
)

// Config is the main application configuration
type Config struct {
	// App settings
	App AppConfig `yaml:"app" json:"app"`

	// LLM settings
	LLM LLMConfig `yaml:"llm" json:"llm"`

	// Watch settings
	Watch WatchConfig `yaml:"watch" json:"watch"`

	// Analyzer settings
	Analyzer AnalyzerConfig `yaml:"analyzer" json:"analyzer"`

	// Cache settings
	Cache CacheConfig `yaml:"cache" json:"cache"`

	// Auth settings
	Auth AuthConfig `yaml:"auth" json:"auth"`

	// Environment
	Env Environment `yaml:"-" json:"environment"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string `yaml:"name" json:"name"`
	Version     string `yaml:"version" json:"version"`
	Description string `yaml:"description" json:"description"`
	LogLevel    string `yaml:"log_level" json:"log_level"`
}

// LLMConfig contains LLM provider settings
type LLMConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Provider string `yaml:"provider" json:"provider"`
	Model    string `yaml:"model" json:"model"`
	// APIKey is loaded from environment variable, not from config file
	APIKey      string        `yaml:"-" json:"-"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	MaxRetries  int           `yaml:"max_retries" json:"max_retries"`
	Temperature float64       `yaml:"temperature" json:"temperature"`
	MaxTokens   int           `yaml:"max_tokens" json:"max_tokens"`
}

// WatchConfig contains file watcher settings
type WatchConfig struct {
	Debounce       time.Duration `yaml:"debounce" json:"debounce"`
	Extensions     []string      `yaml:"extensions" json:"extensions"`
	IgnorePatterns []string      `yaml:"ignore_patterns" json:"ignore_patterns"`
	MaxFileSizeMB  int           `yaml:"max_file_size_mb" json:"max_file_size_mb"`
}

// AnalyzerConfig contains analyzer settings
type AnalyzerConfig struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	Parallelism    int      `yaml:"parallelism" json:"parallelism"`
	SkipCategories []string `yaml:"skip_categories" json:"skip_categories"`
}

// CacheConfig contains caching settings
type CacheConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	MaxSizeMB      int    `yaml:"max_size_mb" json:"max_size_mb"`
	TTLMinutes     int    `yaml:"ttl_minutes" json:"ttl_minutes"`
	EvictionPolicy string `yaml:"eviction_policy" json:"eviction_policy"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	// AnthropicAPIKey is loaded from environment variable
	AnthropicAPIKey string `yaml:"-" json:"-"`
	// OpenRouterAPIKey is loaded from environment variable
	OpenRouterAPIKey string `yaml:"-" json:"-"`
	// GeminiAPIKey is loaded from environment variable
	GeminiAPIKey string `yaml:"-" json:"-"`
	// TokenFile is the path to the stored token
	TokenFile string `yaml:"token_file" json:"token_file"`
	// TokenExpiry is the token expiration time
	TokenExpiry time.Time `yaml:"-" json:"token_expiry"`
}

// Validate validates the configuration and returns errors for any invalid fields
func (c *Config) Validate() error {
	var errs []error

	// Validate LLM config
	if c.LLM.Enabled {
		if c.LLM.Provider == "" {
			c.LLM.Provider = string(llm.ProviderAnthropic)
		}
		if c.LLM.Model == "" {
			c.LLM.Model = llm.DefaultModels[llm.ProviderType(c.LLM.Provider)]
		}
		if c.LLM.Timeout == 0 {
			c.LLM.Timeout = 30 * time.Second
		}
		if c.LLM.MaxRetries == 0 {
			c.LLM.MaxRetries = 3
		}
		if c.LLM.Temperature < 0 || c.LLM.Temperature > 1 {
			errs = append(errs, specerr.New(specerr.ErrCodeInvalidInput, "llm.temperature must be between 0 and 1"))
		}
		if c.LLM.MaxTokens <= 0 {
			errs = append(errs, specerr.New(specerr.ErrCodeInvalidInput, "llm.max_tokens must be greater than 0"))
		}
	}

	// Validate Watch config
	if c.Watch.Debounce == 0 {
		c.Watch.Debounce = 800 * time.Millisecond
	}
	if c.Watch.MaxFileSizeMB == 0 {
		c.Watch.MaxFileSizeMB = 10
	}
	if len(c.Watch.Extensions) == 0 {
		c.Watch.Extensions = []string{".go", ".ts", ".tsx", ".js", ".jsx"}
	}

	// Validate Analyzer config
	if c.Analyzer.Parallelism <= 0 {
		c.Analyzer.Parallelism = 4
	}

	// Validate Cache config
	if c.Cache.MaxSizeMB == 0 {
		c.Cache.MaxSizeMB = 100
	}
	if c.Cache.TTLMinutes == 0 {
		c.Cache.TTLMinutes = 30
	}
	if c.Cache.EvictionPolicy == "" {
		c.Cache.EvictionPolicy = "lru"
	}

	// Validate App config
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}

	if len(errs) > 0 {
		return specerr.New(specerr.ErrCodeInvalidInput, "configuration validation failed").
			WithUnderlying(fmt.Errorf("%v", errs))
	}

	return nil
}

// GetProvider returns the LLM provider type
func (c *LLMConfig) GetProvider() llm.ProviderType {
	return llm.ProviderType(c.Provider)
}

// GetAPIKey returns the appropriate API key based on provider
func (c *AuthConfig) GetAPIKey(provider llm.ProviderType) string {
	switch provider {
	case llm.ProviderAnthropic:
		return c.AnthropicAPIKey
	case llm.ProviderOpenRouter:
		return c.OpenRouterAPIKey
	case llm.ProviderGemini:
		return c.GeminiAPIKey
	default:
		return ""
	}
}

// IsTokenValid returns true if the token is not expired
func (c *AuthConfig) IsTokenValid() bool {
	return c.TokenExpiry.IsZero() || time.Now().Before(c.TokenExpiry)
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:     "specwatch",
			Version:  "dev",
			LogLevel: "info",
		},
		LLM: LLMConfig{
			Enabled:     false,
			Provider:    string(llm.ProviderAnthropic),
			Model:       llm.DefaultModels[llm.ProviderAnthropic],
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Watch: WatchConfig{
			Debounce:       800 * time.Millisecond,
			Extensions:     []string{".go", ".ts", ".tsx", ".js", ".jsx"},
			IgnorePatterns: []string{".git", "node_modules", "vendor"},
			MaxFileSizeMB:  10,
		},
		Analyzer: AnalyzerConfig{
			Enabled:        true,
			Parallelism:    4,
			SkipCategories: []string{},
		},
		Cache: CacheConfig{
			Enabled:        true,
			MaxSizeMB:      100,
			TTLMinutes:     30,
			EvictionPolicy: "lru",
		},
		Auth: AuthConfig{
			TokenFile: "~/.specwatch/token.json",
		},
		Env: EnvironmentDevelopment,
	}
}

// Merge merges another config into this one, overwriting zero values
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}

	// Merge App config
	if other.App.Name != "" {
		c.App.Name = other.App.Name
	}
	if other.App.Version != "" {
		c.App.Version = other.App.Version
	}
	if other.App.LogLevel != "" {
		c.App.LogLevel = other.App.LogLevel
	}

	// Merge LLM config
	if other.LLM.Provider != "" {
		c.LLM.Provider = other.LLM.Provider
	}
	if other.LLM.Model != "" {
		c.LLM.Model = other.LLM.Model
	}
	if other.LLM.Timeout != 0 {
		c.LLM.Timeout = other.LLM.Timeout
	}
	if other.LLM.MaxRetries != 0 {
		c.LLM.MaxRetries = other.LLM.MaxRetries
	}
	if other.LLM.Temperature != 0 {
		c.LLM.Temperature = other.LLM.Temperature
	}
	if other.LLM.MaxTokens != 0 {
		c.LLM.MaxTokens = other.LLM.MaxTokens
	}

	// Merge Watch config
	if other.Watch.Debounce != 0 {
		c.Watch.Debounce = other.Watch.Debounce
	}
	if len(other.Watch.Extensions) > 0 {
		c.Watch.Extensions = other.Watch.Extensions
	}
	if len(other.Watch.IgnorePatterns) > 0 {
		c.Watch.IgnorePatterns = other.Watch.IgnorePatterns
	}
	if other.Watch.MaxFileSizeMB != 0 {
		c.Watch.MaxFileSizeMB = other.Watch.MaxFileSizeMB
	}

	// Merge Analyzer config
	if other.Analyzer.Parallelism != 0 {
		c.Analyzer.Parallelism = other.Analyzer.Parallelism
	}
	if len(other.Analyzer.SkipCategories) > 0 {
		c.Analyzer.SkipCategories = other.Analyzer.SkipCategories
	}

	// Merge Cache config
	if other.Cache.MaxSizeMB != 0 {
		c.Cache.MaxSizeMB = other.Cache.MaxSizeMB
	}
	if other.Cache.TTLMinutes != 0 {
		c.Cache.TTLMinutes = other.Cache.TTLMinutes
	}
	if other.Cache.EvictionPolicy != "" {
		c.Cache.EvictionPolicy = other.Cache.EvictionPolicy
	}

	// Merge Auth config
	if other.Auth.TokenFile != "" {
		c.Auth.TokenFile = other.Auth.TokenFile
	}
}
