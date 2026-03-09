package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// OpenRouterClient implements LLMClient for OpenRouter
type OpenRouterClient struct {
	client  *http.Client
	apiKey  string
	model   string
	baseURL string
}

// NewOpenRouterClient creates a new OpenRouter client
func NewOpenRouterClient(apiKey, model string) (*OpenRouterClient, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	if model == "" {
		model = DefaultModels[ProviderOpenRouter]
	}

	return &OpenRouterClient{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://openrouter.ai",
	}, nil
}

// OpenRouterResponse represents the response from OpenRouter
type OpenRouterResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// Message represents a message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Generate generates a response from OpenRouter
func (o *OpenRouterClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens": 4096,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://specwatch.dev")
	req.Header.Set("X-Title", "Specwatch")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var response OpenRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("empty response from model")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateStream generates a response with streaming (not fully implemented)
func (o *OpenRouterClient) GenerateStream(ctx context.Context, systemPrompt, userPrompt string, onChunk func(string)) error {
	// For simplicity, we'll use non-streaming for now
	result, err := o.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}
	onChunk(result)
	return nil
}

// GetModel returns the current model
func (o *OpenRouterClient) GetModel() string {
	return o.model
}

// GetProvider returns the provider type
func (o *OpenRouterClient) GetProvider() ProviderType {
	return ProviderOpenRouter
}

// Close closes the client
func (o *OpenRouterClient) Close() error {
	return nil
}

// OpenRouterModelsResponse represents the response from listing models
type OpenRouterModelsResponse struct {
	Data []OpenRouterModel `json:"data"`
}

// OpenRouterModel represents a model from OpenRouter
type OpenRouterModel struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ContextLen  int         `json:"context_length"`
	Pricing     PricingInfo `json:"pricing"`
}

// OpenRouterModelLister implements ModelLister for OpenRouter
type OpenRouterModelLister struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewOpenRouterModelLister creates a new OpenRouter model lister
func NewOpenRouterModelLister(apiKey string) (*OpenRouterModelLister, error) {
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
		if apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	return &OpenRouterModelLister{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai",
	}, nil
}

// ListModels returns a list of available OpenRouter models
func (o *OpenRouterModelLister) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var response OpenRouterModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, 0, len(response.Data))
	for _, m := range response.Data {
		models = append(models, ModelInfo{
			ID:          m.ID,
			Name:        m.Name,
			Provider:    string(ProviderOpenRouter),
			ContextLen:  m.ContextLen,
			Pricing:     m.Pricing,
			Description: m.Description,
		})
	}

	return models, nil
}
