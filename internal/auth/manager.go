package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/rajeshshrirao/specwatch/internal/llm"
)

// Manager handles authentication for LLM providers
type Manager struct {
	factory *llm.Factory
}

// NewManager creates a new auth manager
func NewManager() *Manager {
	return &Manager{
		factory: llm.NewFactory(),
	}
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	AnthropicAPIKey  string `json:"anthropic,omitempty"`
	OpenRouterAPIKey string `json:"openrouter,omitempty"`
	GeminiAPIKey     string `json:"gemini,omitempty"`
}

// GetAPIKey returns the API key for a provider
func (m *Manager) GetAPIKey(provider llm.ProviderType) string {
	switch provider {
	case llm.ProviderAnthropic:
		if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
			return key
		}
		return ""
	case llm.ProviderOpenRouter:
		if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
			return key
		}
		return ""
	case llm.ProviderGemini:
		if key := os.Getenv("GEMINI_API_KEY"); key != "" {
			return key
		}
		return ""
	default:
		return ""
	}
}

// HasAPIKey returns true if the provider has an API key configured
func (m *Manager) HasAPIKey(provider llm.ProviderType) bool {
	return m.GetAPIKey(provider) != ""
}

// GetProviderForConfig determines which provider to use based on available API keys
func (m *Manager) GetProviderForConfig() (llm.ProviderType, string) {
	// Check providers in order of preference
	providers := []llm.ProviderType{
		llm.ProviderAnthropic,
		llm.ProviderOpenRouter,
		llm.ProviderGemini,
	}

	for _, p := range providers {
		if m.HasAPIKey(p) {
			return p, m.GetAPIKey(p)
		}
	}

	return "", ""
}

// CreateClient creates an LLM client for the specified provider
func (m *Manager) CreateClient(provider llm.ProviderType, model string) (llm.LLMClient, error) {
	apiKey := m.GetAPIKey(provider)
	if apiKey == "" {
		envVar := llm.GetAPIKeyEnvVar(provider)
		return nil, fmt.Errorf("%w: %s environment variable not set", llm.ErrNoAPIKey, envVar)
	}

	config := llm.Config{
		Provider: provider,
		Model:    model,
		APIKey:   apiKey,
	}

	return m.factory.CreateClient(config)
}

// ListModels lists available models for a provider
func (m *Manager) ListModels(ctx context.Context, provider llm.ProviderType) ([]llm.ModelInfo, error) {
	apiKey := m.GetAPIKey(provider)
	if apiKey == "" {
		envVar := llm.GetAPIKeyEnvVar(provider)
		return nil, fmt.Errorf("%w: %s environment variable not set", llm.ErrNoAPIKey, envVar)
	}

	return m.factory.ListModelsForProvider(ctx, provider, apiKey)
}

// ValidateAPIKey validates that an API key works for the provider
func (m *Manager) ValidateAPIKey(ctx context.Context, provider llm.ProviderType, apiKey string) error {
	// Try to create a model lister and list models to validate
	lister, err := m.factory.CreateModelLister(provider, apiKey)
	if err != nil {
		return err
	}

	_, err = lister.ListModels(ctx)
	return err
}
