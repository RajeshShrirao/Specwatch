package llm

import (
	"context"
	"fmt"
)

// Factory creates LLM clients based on configuration
type Factory struct{}

// NewFactory creates a new LLM factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateClient creates an LLM client based on the config
func (f *Factory) CreateClient(config Config) (LLMClient, error) {
	switch config.Provider {
	case ProviderAnthropic:
		return NewAnthropicClient(config.APIKey, config.Model)
	case ProviderOpenRouter:
		return NewOpenRouterClient(config.APIKey, config.Model)
	case ProviderGemini:
		return NewGeminiClient(config.APIKey, config.Model)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidProvider, config.Provider)
	}
}

// CreateModelLister creates a model lister for the specified provider
func (f *Factory) CreateModelLister(provider ProviderType, apiKey string) (ModelLister, error) {
	switch provider {
	case ProviderAnthropic:
		return NewAnthropicModelLister(apiKey)
	case ProviderOpenRouter:
		return NewOpenRouterModelLister(apiKey)
	case ProviderGemini:
		return NewGeminiModelLister(apiKey)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidProvider, provider)
	}
}

// ListModelsForProvider lists available models for a given provider
func (f *Factory) ListModelsForProvider(ctx context.Context, provider ProviderType, apiKey string) ([]ModelInfo, error) {
	lister, err := f.CreateModelLister(provider, apiKey)
	if err != nil {
		return nil, err
	}
	return lister.ListModels(ctx)
}

// GetAPIKeyEnvVar returns the environment variable name for the provider's API key
func GetAPIKeyEnvVar(provider ProviderType) string {
	switch provider {
	case ProviderAnthropic:
		return "ANTHROPIC_API_KEY"
	case ProviderOpenRouter:
		return "OPENROUTER_API_KEY"
	case ProviderGemini:
		return "GEMINI_API_KEY"
	default:
		return ""
	}
}
