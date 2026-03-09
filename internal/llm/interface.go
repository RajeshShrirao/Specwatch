package llm

import (
	"context"
	"errors"
)

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderAnthropic  ProviderType = "anthropic"
	ProviderOpenRouter ProviderType = "openrouter"
	ProviderGemini     ProviderType = "gemini"
)

// PricingInfo contains pricing details for a model
type PricingInfo struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Request    string `json:"request"`
	Image      string `json:"image"`
}

// ModelInfo contains information about an available model
type ModelInfo struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Provider    string      `json:"provider"`
	ContextLen  int         `json:"context_length"`
	Pricing     PricingInfo `json:"pricing,omitempty"`
	Description string      `json:"description,omitempty"`
}

// LLMClient is the interface for interacting with LLM providers
type LLMClient interface {
	// Generate generates a response from the LLM
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)

	// GenerateStream generates a response with streaming support
	GenerateStream(ctx context.Context, systemPrompt, userPrompt string, onChunk func(string)) error

	// GetModel returns the current model ID
	GetModel() string

	// GetProvider returns the provider type
	GetProvider() ProviderType

	// Close closes the client and releases resources
	Close() error
}

// ModelLister is the interface for listing available models
type ModelLister interface {
	// ListModels returns a list of available models
	ListModels(ctx context.Context) ([]ModelInfo, error)
}

// Config contains configuration for an LLM provider
type Config struct {
	Provider ProviderType `json:"provider" yaml:"provider"`
	Model    string       `json:"model" yaml:"model"`
	APIKey   string       `json:"-" yaml:"api_key"` // Not stored in config file
}

// ErrNoAPIKey is returned when no API key is provided
var ErrNoAPIKey = errors.New("no API key provided for the selected provider")

// ErrInvalidProvider is returned when an invalid provider is specified
var ErrInvalidProvider = errors.New("invalid provider type")

// ErrModelNotFound is returned when the specified model is not found
var ErrModelNotFound = errors.New("model not found")

// ProviderNames returns human-readable names for providers
var ProviderNames = map[ProviderType]string{
	ProviderAnthropic:  "Anthropic (Claude)",
	ProviderOpenRouter: "OpenRouter",
	ProviderGemini:     "Google Gemini",
}

// DefaultModels returns the default model for each provider
var DefaultModels = map[ProviderType]string{
	ProviderAnthropic:  "claude-haiku-4-5-20251002",
	ProviderOpenRouter: "anthropic/claude-4.5-haiku-20250929",
	ProviderGemini:     "gemini-2.0-flash",
}
