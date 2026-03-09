# Specwatch AI Integration Plan

## Overview

This plan outlines implementing Claude Haiku 4.5 model support with multiple LLM providers (Anthropic, OpenRouter, Gemini) and improved authentication experience in Specwatch.

---

## Research Findings

### 1. Anthropic (Direct)

**Haiku 4.5 Model:**
- Model ID: `claude-haiku-4-5-2025xxxxx` (pattern like `claude-sonnet-4-5-20250929`)
- SDK: [`github.com/anthropics/anthropic-sdk-go`](go.mod:5)
- Authentication: API key via `ANTHROPIC_API_KEY` env var or `option.WithAPIKey()`

```go
client := anthropic.NewClient(
    option.WithAPIKey("your-api-key"),
)
message, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
    Model: anthropic.ModelClaudeHaiku4_5_20251002,
    Messages: []anthropic.MessageParam{
        anthropic.NewUserMessage(anthropic.NewTextBlock("Your prompt")),
    },
    MaxTokens: 1024,
})
```

---

### 2. OpenRouter

**Features:**
- Unified API for 300+ models including Claude Haiku
- Model ID: `anthropic/claude-4.5-haiku-20250929`
- Authentication: Bearer token (`OPENROUTER_API_KEY`)

**Endpoints:**
- POST `https://openrouter.ai/api/v1/responses`
- Headers: `Authorization: Bearer YOUR_KEY`

```go
// Using HTTP directly (recommended for Go)
resp, err := http.Post("https://openrouter.ai/api/v1/responses", "application/json", body)
// Headers: Authorization: Bearer YOUR_OPENROUTER_API_KEY
```

**Key Benefits:**
- Access to Claude models without direct Anthropic API
- Credit limits per key
- Automatic fallbacks

---

### 3. Google Gemini

**SDK:**
- Package: `google.golang.org/genai`
- Authentication: API key via `GEMINI_API_KEY` or `x-goog-api-key` header

```go
client, err := genai.NewClient(ctx, &genai.ClientConfig{
    APIKey:  "YOUR_API_KEY",
    Backend: genai.BackendGeminiAPI,
})
result, err := client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text("prompt"), nil)
```

**Models:**
- `gemini-2.0-flash` (fast)
- `gemini-2.0-pro` (capable)
- `gemini-1.5-flash` / `gemini-1.5-pro`

---

## Architecture

```mermaid
flowchart TB
    subgraph Config
        C[.specwatch.yml]
    end
    
    subgraph Auth
        AM[Auth Manager]
        K[Keychain Storage]
    end
    
    subgraph Providers
        AN[Anthropic]
        OR[OpenRouter]
        GM[Gemini]
    end
    
    subgraph Core
        LLM[LLM Interface]
    end
    
    subgraph TUI
        L[Login Flow]
        S[Settings Panel]
    end
    
    C --> AM
    K --> AM
    AM --> LLM
    LLM --> AN
    LLM --> OR
    LLM --> GM
    L --> AM
    S --> AM
```

---

## Implementation Tasks

### Phase 1: Dependencies

1. **Add Go SDKs**
   - Anthropic: `github.com/anthropics/anthropic-sdk-go`
   - Gemini: `google.golang.org/genai`
   - Keychain: `github.com/keybase/go-keychain` (macOS)

### Phase 2: Core Interface

2. **Create LLM Interface** (`internal/llm/interface.go`)
   ```go
   type LLMClient interface {
       Generate(ctx context.Context, prompt string) (string, error)
       GetModel() string
       GetProvider() string
   }
   
   type ModelInfo struct {
       ID          string
       Name        string
       Provider    string
       ContextLen  int
       Pricing     PricingInfo
   }
   
   type ModelLister interface {
       ListModels(ctx context.Context) ([]ModelInfo, error)
   }
   ```

3. **Create Provider Implementations**
   - `internal/llm/anthropic.go` - Direct Anthropic API + model listing
   - `internal/llm/openrouter.go` - OpenRouter unified API + model listing
   - `internal/llm/gemini.go` - Google Gemini API + model listing

### Phase 3: Authentication

4. **Auth Manager** (`internal/auth/manager.go`)
   - Provider-specific API keys
   - Secure credential storage (Keychain)
   - Config priority: Keychain > Env > Config file

5. **Config Structure**
   ```yaml
   llm:
     provider: anthropic  # or "openrouter", "gemini"
     model: claude-haiku-4-5-20251002
     
   auth:
     anthropic:
       api_key: ${ANTHROPIC_API_KEY}
     openrouter:
       api_key: ${OPENROUTER_API_KEY}
     gemini:
       api_key: ${GEMINI_API_KEY}
   ```

### Phase 4: TUI Integration

6. **Login Flow** (`internal/tui/login.go`)
   - Provider selection (Anthropic, OpenRouter, Gemini)
   - API key input with validation
   - **Model selection dropdown** - dynamically fetches available models
   - Account status display

7. **Settings Panel** (`internal/tui/settings.go`)
   - Provider dropdown
   - **Model dropdown** - dynamically populated from provider API
   - Auth status indicators
   - Refresh models button

### Phase 5: Integration

8. **Connect to Analyzer** (`internal/analyzer/engine.go`)
   - Add LLM client to engine
   - Trigger AI checks for architecture rules

---

## Configuration Examples

### Using Anthropic Direct
```yaml
llm:
  provider: anthropic
  model: claude-haiku-4-5-20251002
  
auth:
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
```

### Using OpenRouter
```yaml
llm:
  provider: openrouter
  model: anthropic/claude-4.5-haiku-20250929
  
auth:
  openrouter:
    api_key: ${OPENROUTER_API_KEY}
```

### Using Gemini
```yaml
llm:
  provider: gemini
  model: gemini-2.0-flash
  
auth:
  gemini:
    api_key: ${GEMINI_API_KEY}
```

---

## Environment Variables

| Provider | Env Variable |
|----------|-------------|
| Anthropic | `ANTHROPIC_API_KEY` |
| OpenRouter | `OPENROUTER_API_KEY` |
| Gemini | `GEMINI_API_KEY` |

---

## File Changes Summary

### New Files
- `internal/llm/interface.go` - LLM interface
- `internal/llm/anthropic.go` - Anthropic provider
- `internal/llm/openrouter.go` - OpenRouter provider  
- `internal/llm/gemini.go` - Gemini provider
- `internal/llm/factory.go` - Provider factory
- `internal/auth/manager.go` - Auth manager
- `internal/tui/login.go` - Login TUI

### Modified Files
- `go.mod` / `go.sum` - Add dependencies
- `.specwatch.yml` - Update config schema
- `internal/analyzer/engine.go` - Integrate LLM
- `cmd/root.go` - Add login command

---

## Model Listing APIs

### OpenRouter
- **Endpoint**: `GET https://openrouter.ai/api/v1/models`
- **Auth**: Bearer token
- **Response**: Array of model objects with id, name, pricing, context_length

```go
// List models from OpenRouter
func (o *OpenRouterClient) ListModels(ctx context.Context) ([]ModelInfo, error) {
    req, _ := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
    req.Header.Set("Authorization", "Bearer "+o.apiKey)
    // Parse response and return model list
}
```

### Gemini  
- **Endpoint**: `GET https://generativelanguage.googleapis.com/v1beta/openai/models`
- **Auth**: Bearer token (`GEMINI_API_KEY`)
- **Response**: Array of model objects with id

```go
// List models from Gemini
func (g *GeminiClient) ListModels(ctx context.Context) ([]ModelInfo, error) {
    req, _ := http.NewRequest("GET", 
        "https://generativelanguage.googleapis.com/v1beta/openai/models", nil)
    req.Header.Set("Authorization", "Bearer "+g.apiKey)
    // Parse response and return model list
}
```

### Anthropic
- Use SDK: `client.Models.ListAutoPaging()`
- Or API: `GET https://api.anthropic.com/v1/models`

---

## Model Mapping

| Provider | Haiku 4.5 Model ID |
|----------|-------------------|
| Anthropic | `claude-haiku-4-5-20251002` |
| OpenRouter | `anthropic/claude-4.5-haiku-20250929` |
| Gemini | N/A (use `gemini-2.0-flash`) |

---

## Next Steps

Ready to implement. Switch to Code mode to begin:
1. Add dependencies to `go.mod`
2. Create LLM interface and providers
3. Implement auth manager
4. Build login TUI
5. Integrate with analyzer
