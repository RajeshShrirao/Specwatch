package plugin

import (
	"context"
)

// Plugin represents an LLM provider plugin
type Plugin interface {
	// ID returns the unique plugin identifier
	ID() string

	// Name returns the human-readable name
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a brief description of the plugin
	Description() string

	// SupportedModels returns list of supported model IDs
	SupportedModels() []string

	// CreateClient creates an LLM client instance
	// apiKey: the API key for authentication
	// model: the model ID to use (empty = default)
	CreateClient(apiKey, model string) (LLMClient, error)

	// Configure applies additional configuration
	Configure(config map[string]interface{}) error
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
	GetProvider() string

	// Close closes the client and releases resources
	Close() error
}

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
}

// BasePlugin provides common functionality for plugins
type BasePlugin struct{}

// Configure is a no-op by default
func (bp *BasePlugin) Configure(config map[string]interface{}) error {
	return nil
}
