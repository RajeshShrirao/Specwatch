# Specwatch AI Feature Guide

## What is This?

Specwatch now has **AI-powered code analysis**! Instead of just checking if your code follows simple rules (like "no console.log"), it can now use AI to understand complex architectural patterns and semantic rules.

## Features

### 1. Multiple AI Providers

| Provider | Default Model | Description |
|----------|--------------|-------------|
| **Anthropic (Claude)** | claude-haiku-4-5-20251002 | Fast, cost-effective |
| **OpenRouter** | anthropic/claude-4.5-haiku-20250929 | 300+ AI models |
| **Google Gemini** | gemini-2.0-flash | Google's fastest model |

### 2. Dynamic Model Listing
Want to see all available models? Just run:
```bash
specwatch login --provider openrouter --list-models
```

### 3. Easy Authentication
No more hunting for API keys! Just run:

```bash
# One-time setup
specwatch login --provider anthropic --api-key sk-ant-your-key-here

# Or use environment variables
export ANTHROPIC_API_KEY="sk-ant-your-key-here"
```

## ⚠️ Important: AI Budget Constraint

**AI analysis fires at most once per 10 file saves**, and only when architecture rules require semantic judgment. This prevents excessive API costs and keeps development fast.

## User Flows

### Flow 1: First-Time Setup

```
1. Get an API key from your preferred provider:
   - Anthropic: https://console.anthropic.com/
   - OpenRouter: https://openrouter.ai/settings
   - Google Gemini: https://aistudio.google.com/app/apikey

2. Configure Specwatch:
   specwatch login --provider anthropic --api-key YOUR_KEY

3. Start using AI analysis:
   specwatch watch ./src
```

### Flow 2: Choosing a Provider

```
1. Try OpenRouter to explore models:
   specwatch login --provider openrouter --list-models
   
2. Pick a model (e.g., claude-4.5-haiku)
   
3. Configure it:
   specwatch login --provider openrouter --api-key YOUR_KEY
```

### Flow 3: Switching Providers

```
# Using Anthropic (default model: claude-haiku-4-5-20251002)
export ANTHROPIC_API_KEY="sk-ant-..."

# Switch to OpenRouter (default model: anthropic/claude-4.5-haiku-20250929)
unset ANTHROPIC_API_KEY
export OPENROUTER_API_KEY="sk-or-v1-..."

# Switch to Gemini (default model: gemini-2.0-flash)
unset OPENROUTER_API_KEY  
export GEMINI_API_KEY="AIza..."
```

## How AI Analysis Works

Before (Static Analysis):
- Checks file names match patterns
- Counts import statements
- Validates code structure

With AI (New!):
- Understands architectural intent
- Detects business logic in UI components
- Identifies design pattern violations
- Explains WHY something is wrong

**Budget**: AI analysis runs at most once per 10 file saves to prevent excessive API costs.

Example in `spec.md`:
```markdown
## architecture
- UI components must not contain business logic
- API calls must go through service layer
```

The AI will analyze your code and explain violations in plain English.

## Configuration

Edit `.specwatch.yml`:
```yaml
llm:
  enabled: true
  provider: anthropic
  model: claude-haiku-4-5-20251002
```

Or use the default model for each provider:
```yaml
llm:
  provider: anthropic    # Uses claude-haiku-4-5-20251002
  # or
  provider: openrouter  # Uses anthropic/claude-4.5-haiku-20250929
  # or  
  provider: gemini      # Uses gemini-2.0-flash
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `specwatch login --provider X` | Configure provider X |
| `specwatch login --provider X --list-models` | See all models for X |
| `specwatch login --provider X --api-key KEY` | Set API key inline |

| Environment Variable | Default Model | Provider |
|---------------------|--------------|----------|
| `ANTHROPIC_API_KEY` | claude-haiku-4-5-20251002 | Anthropic |
| `OPENROUTER_API_KEY` | anthropic/claude-4.5-haiku-20250929 | OpenRouter |
| `GEMINI_API_KEY` | gemini-2.0-flash | Google Gemini |

## Benefits

1. **Faster Development** - AI handles complex rules automatically
2. **Better Feedback** - Gets explanations, not just "error"
3. **Flexible** - Choose provider based on cost/speed needs
4. **Future-Ready** - Easy to switch models as AI improves
5. **Budget-Aware** - Runs max once per 10 saves to control costs
