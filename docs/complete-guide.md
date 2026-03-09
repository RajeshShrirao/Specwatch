# Specwatch - Complete Guide

## What is Specwatch?

Specwatch is a **smart code analysis tool** that enforces architectural rules in your project. Think of it as a guardrail that keeps your codebase clean and consistent.

### Two Types of Analysis

| Type | How It Works | Speed |
|------|-------------|-------|
| **Static** | Checks patterns, names, file sizes | < 10ms (instant) |
| **AI-Powered** | Understands code meaning & intent | Uses AI models |

---

## Quick Start

### 1. Installation

```bash
# Install
go install github.com/rajeshshrirao/specwatch@latest

# Verify
specwatch --version
```

### 2. Initialize Your Project

```bash
# Creates spec.md template
specwatch init
```

This creates a `spec.md` file where you define your rules.

---

## Core Concepts

### The spec.md File

This is your **architectural blueprint**. You MUST use the exact syntax below or rules won't be detected.

#### Required Syntax (Copy This!)

```markdown
# My Project Spec

## stack
- language: typescript
- framework: react
- styling: tailwind

## structure
- pages: src/pages/*
- components: src/components/*
- hooks: src/hooks/*

## naming
- components: PascalCase
- files: kebab-case
- hooks: use*

## forbidden
- pattern: "console.log" message: Use the logger from lib/logger instead
- import: "lib/db" message: Use the service layer instead

## required
- async: try-catch

## limits
- max file lines: 200
- max imports per file: 10

## architecture
- No business logic in UI components
- API calls must go through service layer
```

⚠️ **Important**: The format above is REQUIRED. Do NOT use freeform text. Each rule type has a specific syntax.

---

## Rule Syntax Reference

### `## stack`
```
- language: typescript
- framework: react
- runtime: node
```

### `## structure`
```
- components: src/components/*
- hooks: src/hooks/*
```

### `## naming`
```
- components: PascalCase
- files: kebab-case
- functions: camelCase
- constants: UPPER_SNAKE_CASE
```

### `## forbidden` (MUST use this exact format)
```
- pattern: "console.log" message: Use logger instead
- pattern: "debugger" message: Remove debugger statement
- import: "lib/db" message: Use service layer
```

### `## required`
```
- async: try-catch
- exports: default
```

### `## limits` (MUST use this exact format)
```
- max file lines: 200
- max imports per file: 10
- max function lines: 50
```

### `## architecture` (Freeform - uses AI)
Freeform text rules are analyzed by AI:
```
- No business logic in UI components
- API calls must go through service layer
```

---

## Commands

### `specwatch init`
Creates a fresh `spec.md` template in your project.

```bash
specwatch init
```

### `specwatch watch`
Live monitoring with terminal UI.

```bash
# Watch current directory
specwatch watch

# Watch specific folder
specwatch watch ./src

# Watch with options
specwatch watch ./src --ext ts,tsx --debounce 1000
```

**Flags:**
- `--ext` - File extensions to watch (default: ts, tsx, js, jsx)
- `--debounce` - Delay before analysis (ms)
- `--skip` - Skip rule categories

### `specwatch check`
One-time check (good for CI/CD).

```bash
# Check directory
specwatch check ./src

# JSON output for CI
specwatch check ./src --format json
```

### `specwatch login`
Configure AI provider for intelligent analysis.

```bash
# Set up Anthropic (Claude Haiku 4.5)
specwatch login --provider anthropic --api-key sk-ant-...

# Set up OpenRouter
specwatch login --provider openrouter --api-key sk-or-v1-...

# Set up Google Gemini
specwatch login --provider gemini --api-key AIza...

# List available models
specwatch login --provider openrouter --list-models
```

---

## AI Integration

### Why AI Analysis?

Static rules catch simple issues:
- ✅ "No console.log"
- ✅ "Max 200 lines"
- ✅ "Components use PascalCase"

But AI catches complex issues:
- ✅ "Business logic in UI component"
- ✅ "Missing error handling in async code"
- ✅ "Security vulnerability in data flow"

### ⚠️ Important: AI Budget Constraint

**AI analysis fires at most once per 10 file saves**, and only when architecture rules require semantic judgment. This prevents excessive API costs.

### Supported AI Providers

| Provider | Default Model | Notes |
|----------|--------------|-------|
| **Anthropic** | claude-haiku-4-5-20251002 | Direct API |
| **OpenRouter** | anthropic/claude-4.5-haiku-20250929 | 300+ models |
| **Google Gemini** | gemini-2.0-flash | Google's fastest model |

### Environment Setup

```bash
# Option 1: Environment variables
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENROUTER_API_KEY="sk-or-v1-..."
export GEMINI_API_KEY="AIza..."

# Option 2: Login command (saves to session)
specwatch login --provider anthropic --api-key YOUR_KEY
```

### Configuration

In `.specwatch.yml`:

```yaml
llm:
  enabled: true
  provider: anthropic
  model: claude-haiku-4-5-20251002

watch:
  debounce: 800
  extensions: [go, ts, tsx, js, jsx]
```

---

## Use Cases

### CI/CD Integration
```bash
# In your pipeline
specwatch check ./src --format json
```

### Pre-commit Hook
```bash
# .git/hooks/pre-commit
#!/bin/bash
specwatch check ./src
```

### Editor Integration
```bash
# Run on save (with VS Code tasks or similar)
specwatch watch ./src --debounce 500
```

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│                  Specwatch                        │
├─────────────────────────────────────────────────┤
│  CLI (Cobra)                                    │
│  ├── watch   - Live monitoring                   │
│  ├── check   - CI/CD mode                        │
│  ├── init    - Create spec.md                    │
│  └── login   - AI configuration                  │
├─────────────────────────────────────────────────┤
│  Core Engine                                     │
│  ├── spec/parser   - Parse spec.md              │
│  ├── analyzer/     - Run static checks           │
│  ├── llm/          - AI-powered analysis         │
│  └── watcher/      - File system monitoring     │
├─────────────────────────────────────────────────┤
│  UI (Bubble Tea)                                 │
│  ├── Activity feed - Recent checks              │
│  ├── Violations    - Issues list                │
│  └── Stats         - Error/warning counts       │
└─────────────────────────────────────────────────┘
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Anthropic/Claude API key |
| `OPENROUTER_API_KEY` | OpenRouter API key |
| `GEMINI_API_KEY` | Google Gemini API key |
| `CI` | Set to "true" in CI environments |

---

## Troubleshooting

### "No API key configured"
```bash
# Set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Or use login command
specwatch login --provider anthropic --api-key YOUR_KEY
```

### "No spec.md found"
```bash
# Create one
specwatch init
```

### "No violations detected"
Check your spec.md syntax! Freeform text won't work. Use:
```
- pattern: "console.log" message: Use logger
```
NOT:
```
- No console.log statements
```

### "Permission denied"
```bash
# Make specwatch executable
chmod +x specwatch
# or
go install
```

---

## Links

- GitHub: https://github.com/rajeshshrirao/specwatch
- Issues: https://github.com/rajeshshrirao/specwatch/issues
