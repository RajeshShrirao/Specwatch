# Specwatch AI Feature Guide

## What is This?

Specwatch now has **AI-powered code analysis**! Instead of just checking if your code follows simple rules (like "no console.log"), it can now use AI to understand complex architectural patterns and semantic rules.

## Features

### 1. Multiple AI Providers
You can choose from three different AI services:

| Provider | Best For | Example Model |
|----------|----------|---------------|
| **Anthropic (Claude)** | Fast, cost-effective AI checks | Haiku 4.5 |
| **OpenRouter** | Access to 300+ AI models | Claude via OpenRouter |
| **Google Gemini** | Google's AI models | Gemini 2.0 Flash |

### 2. Dynamic Model Listing
Want to see all available models? Just run:
```bash
specwatch login --provider openrouter --list-models
```

This shows every model you can use, including pricing and context limits.

### 3. Easy Authentication
No more hunting for API keys! Just run:

```bash
# One-time setup
specwatch login --provider anthropic --api-key sk-ant-your-key-here

# Or use environment variables
export ANTHROPIC_API_KEY="sk-ant-your-key-here"
```

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
# Using Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Switch to OpenRouter (alternative)
unset ANTHROPIC_API_KEY
export OPENROUTER_API_KEY="sk-or-v1-..."

# Switch to Gemini
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

| Environment Variable | Provider |
|---------------------|----------|
| `ANTHROPIC_API_KEY` | Anthropic |
| `OPENROUTER_API_KEY` | OpenRouter |
| `GEMINI_API_KEY` | Google Gemini |

## Benefits

1. **Faster Development** - AI handles complex rules automatically
2. **Better Feedback** - Gets explanations, not just "error"
3. **Flexible** - Choose provider based on cost/speed needs
4. **Future-Ready** - Easy to switch models as AI improves
