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

// GeminiClient implements LLMClient for Google Gemini API
type GeminiClient struct {
	client  *http.Client
	apiKey  string
	model   string
	baseURL string
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(apiKey, model string) (*GeminiClient, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	if model == "" {
		model = DefaultModels[ProviderGemini]
	}

	return &GeminiClient{
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://generativelanguage.googleapis.com",
	}, nil
}

// GeminiRequest represents a request to Gemini
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in a request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents a response from Gemini
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// GeminiCandidate represents a candidate in the response
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// Generate generates a response from Gemini
func (g *GeminiClient) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Combine system prompt and user prompt
	fullPrompt := systemPrompt + "\n\n" + userPrompt

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: fullPrompt},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var response GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from model")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}

// GenerateStream generates a response with streaming (not fully implemented)
func (g *GeminiClient) GenerateStream(ctx context.Context, systemPrompt, userPrompt string, onChunk func(string)) error {
	result, err := g.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return err
	}
	onChunk(result)
	return nil
}

// GetModel returns the current model
func (g *GeminiClient) GetModel() string {
	return g.model
}

// GetProvider returns the provider type
func (g *GeminiClient) GetProvider() ProviderType {
	return ProviderGemini
}

// Close closes the client
func (g *GeminiClient) Close() error {
	return nil
}

// GeminiModelInfo represents model info from Gemini API
type GeminiModelInfo struct {
	ID string `json:"id"`
}

// GeminiModelsResponse represents the response from listing models
type GeminiModelsResponse struct {
	Models []GeminiModelInfo `json:"models"`
}

// GeminiModelLister implements ModelLister for Gemini
type GeminiModelLister struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewGeminiModelLister creates a new Gemini model lister
func NewGeminiModelLister(apiKey string) (*GeminiModelLister, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, ErrNoAPIKey
		}
	}

	return &GeminiModelLister{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com",
	}, nil
}

// ListModels returns a list of available Gemini models
func (g *GeminiModelLister) ListModels(ctx context.Context) ([]ModelInfo, error) {
	url := fmt.Sprintf("%s/v1beta/models?key=%s", g.baseURL, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var response GeminiModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]ModelInfo, 0, len(response.Models))
	for _, m := range response.Models {
		models = append(models, ModelInfo{
			ID:       m.ID,
			Name:     m.ID,
			Provider: string(ProviderGemini),
		})
	}

	return models, nil
}
