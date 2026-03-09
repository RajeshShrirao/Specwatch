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

// AnthropicClient implements LLMClient for Anthropic's Claude API
type AnthropicClient struct {
	client  *http.Client
	apiKey  string
	model   string
	baseURL string
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(apiKey, model string) (*AnthropicClient, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	if model == "" {
		model = DefaultModels[ProviderAnthropic]
	}

	return &AnthropicClient{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com",
	}, nil
}

// AnthropicRequest represents a request to Anthropic
type AnthropicRequest struct {
	Model     string                   `json:"model"`
	MaxTokens int                      `json:"max_tokens"`
	System    []AnthropicSystemContent `json:"system"`
	Messages  []AnthropicMessage       `json:"messages"`
}

// AnthropicSystemContent represents system content
type AnthropicSystemContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicMessage represents a message
type AnthropicMessage struct {
	Role    string             `json:"role"`
	Content []AnthropicContent `json:"content"`
}

// AnthropicContent represents content in a message
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicResponse represents a response from Anthropic
type AnthropicResponse struct {
	ID      string                  `json:"id"`
	Type    string                  `json:"type"`
	Role    string                  `json:"role"`
	Content []AnthropicContentBlock `json:"content"`
	Model   string                  `json:"model"`
	Usage   AnthropicUsage          `json:"usage"`
}

// AnthropicContentBlock represents a content block in response
type AnthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents token usage
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Generate generates a response from Claude
func (a *AnthropicClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := AnthropicRequest{
		Model:     a.model,
		MaxTokens: 4096,
		System: []AnthropicSystemContent{
			{Type: "text", Text: systemPrompt},
		},
		Messages: []AnthropicMessage{
			{
				Role: "user",
				Content: []AnthropicContent{
					{Type: "text", Text: userPrompt},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var response AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", errors.New("empty response from model")
	}

	return response.Content[0].Text, nil
}

// GenerateStream generates a response with streaming (not fully implemented)
func (a *AnthropicClient) GenerateStream(ctx context.Context, systemPrompt, userPrompt string, onChunk func(string)) error {
	result, err := a.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}
	onChunk(result)
	return nil
}

// GetModel returns the current model
func (a *AnthropicClient) GetModel() string {
	return a.model
}

// GetProvider returns the provider type
func (a *AnthropicClient) GetProvider() ProviderType {
	return ProviderAnthropic
}

// Close closes the client
func (a *AnthropicClient) Close() error {
	return nil
}

// AnthropicModelInfo represents model info from Anthropic API
type AnthropicModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// AnthropicModelsResponse represents the response from listing models
type AnthropicModelsResponse struct {
	Data []AnthropicModelInfo `json:"data"`
}

// AnthropicModelLister implements ModelLister for Anthropic
type AnthropicModelLister struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewAnthropicModelLister creates a new Anthropic model lister
func NewAnthropicModelLister(apiKey string) (*AnthropicModelLister, error) {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	return &AnthropicModelLister{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com",
	}, nil
}

// ListModels returns a list of available Anthropic models
func (a *AnthropicModelLister) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", a.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var response AnthropicModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, 0, len(response.Data))
	for _, m := range response.Data {
		models = append(models, ModelInfo{
			ID:          m.ID,
			Name:        m.DisplayName,
			Provider:    string(ProviderAnthropic),
			Description: m.Description,
		})
	}

	return models, nil
}
