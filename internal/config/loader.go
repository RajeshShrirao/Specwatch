package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles configuration loading from various sources
type Loader struct {
	configPath string
	env        Environment
}

// LoaderOption is a functional option for configuring the loader
type LoaderOption func(*Loader)

// WithConfigPath sets the configuration file path
func WithConfigPath(path string) LoaderOption {
	return func(l *Loader) {
		l.configPath = path
	}
}

// WithEnvironment sets the environment
func WithEnvironment(env Environment) LoaderOption {
	return func(l *Loader) {
		l.env = env
	}
}

// NewLoader creates a new configuration loader
func NewLoader(opts ...LoaderOption) *Loader {
	loader := &Loader{
		env: EnvironmentDevelopment,
	}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads configuration from file and environment
func (l *Loader) Load() (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()
	cfg.Env = l.env

	// Load from file if specified
	if l.configPath != "" {
		if err := l.loadFromFile(cfg, l.configPath); err != nil {
			return nil, err
		}
	}

	// Load from environment variables
	if err := l.loadFromEnv(cfg); err != nil {
		return nil, err
	}

	// Apply environment-specific overrides
	l.applyEnvironmentOverrides(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file
func (l *Loader) loadFromFile(cfg *Config, path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try with .yaml extension
		yamlPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".yaml"
		if _, err := os.Stat(yamlPath); err == nil {
			path = yamlPath
		} else {
			// File doesn't exist, return with defaults
			return nil
		}
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Set defaults
	v.SetDefault("app.name", cfg.App.Name)
	v.SetDefault("app.log_level", cfg.App.LogLevel)
	v.SetDefault("llm.enabled", cfg.LLM.Enabled)
	v.SetDefault("llm.provider", cfg.LLM.Provider)
	v.SetDefault("llm.model", cfg.LLM.Model)
	v.SetDefault("llm.timeout", cfg.LLM.Timeout)
	v.SetDefault("llm.max_retries", cfg.LLM.MaxRetries)
	v.SetDefault("llm.temperature", cfg.LLM.Temperature)
	v.SetDefault("llm.max_tokens", cfg.LLM.MaxTokens)
	v.SetDefault("watch.debounce", cfg.Watch.Debounce)
	v.SetDefault("watch.extensions", cfg.Watch.Extensions)
	v.SetDefault("watch.ignore_patterns", cfg.Watch.IgnorePatterns)
	v.SetDefault("watch.max_file_size_mb", cfg.Watch.MaxFileSizeMB)
	v.SetDefault("analyzer.enabled", cfg.Analyzer.Enabled)
	v.SetDefault("analyzer.parallelism", cfg.Analyzer.Parallelism)
	v.SetDefault("analyzer.skip_categories", cfg.Analyzer.SkipCategories)
	v.SetDefault("cache.enabled", cfg.Cache.Enabled)
	v.SetDefault("cache.max_size_mb", cfg.Cache.MaxSizeMB)
	v.SetDefault("cache.ttl_minutes", cfg.Cache.TTLMinutes)
	v.SetDefault("cache.eviction_policy", cfg.Cache.EvictionPolicy)
	v.SetDefault("auth.token_file", cfg.Auth.TokenFile)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into config
	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (l *Loader) loadFromEnv(cfg *Config) error {
	// App config
	if v := os.Getenv("SPECWATCH_LOG_LEVEL"); v != "" {
		cfg.App.LogLevel = v
	}

	// LLM config
	if v := os.Getenv("SPECWATCH_LLM_ENABLED"); v != "" {
		cfg.LLM.Enabled = strings.ToLower(v) == "true"
	}
	if v := os.Getenv("SPECWATCH_LLM_PROVIDER"); v != "" {
		cfg.LLM.Provider = v
	}
	if v := os.Getenv("SPECWATCH_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.Auth.AnthropicAPIKey = v
	}
	if v := os.Getenv("OPENROUTER_API_KEY"); v != "" {
		cfg.Auth.OpenRouterAPIKey = v
	}
	if v := os.Getenv("GEMINI_API_KEY"); v != "" {
		cfg.Auth.GeminiAPIKey = v
	}

	// Watch config
	if v := os.Getenv("SPECWATCH_WATCH_DEBOUNCE"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &cfg.Watch.Debounce); err != nil {
			return fmt.Errorf("invalid SPECWATCH_WATCH_DEBOUNCE value: %w", err)
		}
	}

	// Analyzer config
	if v := os.Getenv("SPECWATCH_ANALYZER_PARALLELISM"); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &cfg.Analyzer.Parallelism); err != nil {
			return fmt.Errorf("invalid SPECWATCH_ANALYZER_PARALLELISM value: %w", err)
		}
	}

	// Cache config
	if v := os.Getenv("SPECWATCH_CACHE_ENABLED"); v != "" {
		cfg.Cache.Enabled = strings.ToLower(v) == "true"
	}

	return nil
}

// applyEnvironmentOverrides applies environment-specific configuration overrides
func (l *Loader) applyEnvironmentOverrides(cfg *Config) {
	switch l.env {
	case EnvironmentDevelopment:
		// Development defaults
		cfg.App.LogLevel = "debug"
		cfg.Analyzer.Parallelism = 2

	case EnvironmentStaging:
		// Staging defaults
		cfg.App.LogLevel = "info"
		cfg.Analyzer.Parallelism = 4
		cfg.Cache.MaxSizeMB = 200

	case EnvironmentProduction:
		// Production defaults
		cfg.App.LogLevel = "warn"
		cfg.Analyzer.Parallelism = 8
		cfg.Cache.MaxSizeMB = 500
	}
}

// Load loads configuration with default options
func Load() (*Config, error) {
	return NewLoader().Load()
}

// LoadFromPath loads configuration from a specific path
func LoadFromPath(path string) (*Config, error) {
	return NewLoader(WithConfigPath(path)).Load()
}

// LoadForEnvironment loads configuration for a specific environment
func LoadForEnvironment(env Environment) (*Config, error) {
	return NewLoader(WithEnvironment(env)).Load()
}

// FindConfigFile searches for config file in common locations
func FindConfigFile(baseDir string) string {
	searchPaths := []string{
		filepath.Join(baseDir, ".specwatch.yml"),
		filepath.Join(baseDir, ".specwatch.yaml"),
		filepath.Join(baseDir, "specwatch.yml"),
		filepath.Join(baseDir, "specwatch.yaml"),
		filepath.Join(baseDir, ".config", "specwatch.yml"),
		filepath.Join(baseDir, ".config", "specwatch.yaml"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
