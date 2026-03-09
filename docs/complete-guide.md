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

This is your **architectural blueprint**. Example:

```markdown
# My Project Spec

## naming
- Components must use PascalCase
- Hooks must start with "use"

## limits
- Max 200 lines per file
- Max 10 imports per file

## forbidden
- No console.log statements
- No TODO comments

## architecture
- UI components must not contain business logic
- API calls must go through service layer
```

### How Analysis Works

```
Your Code → Specwatch → Violations Report
                ↓
           spec.md rules
                ↓
        [Static] + [AI] checks
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
# Set up Anthropic (Claude)
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

### Supported AI Providers

| Provider | Best For | Model |
|----------|----------|-------|
| **Anthropic** | Fast, cheap | Haiku 4.5 |
| **OpenRouter** | Model variety (300+) | Claude variants |
| **Google Gemini** | Google's models | Gemini 2.0 Flash |

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

## Rule Categories

### 1. Naming Rules
```markdown
## naming
- Files must use kebab-case
- Components must use PascalCase
- Hooks must start with "use"
```

### 2. Forbidden Patterns
```markdown
## forbidden
- No console.log statements
- No debugger statements
- No TODO comments without assignee
```

### 3. Limits
```markdown
## limits
- Max 200 lines per file
- Max 10 imports per file
- Max 3 levels of nesting
```

### 4. Architecture
```markdown
## architecture
- UI components must not contain business logic
- API calls must go through service layer
- No direct database access in handlers
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
│  └── watcher/      - File system monitoring      │
├─────────────────────────────────────────────────┤
│  UI (Bubble Tea)                                 │
│  ├── Activity feed - Recent checks               │
│  ├── Violations    - Issues list                │
│  └── Stats         - Error/warning counts        │
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

## Example Workflow

```bash
# 1. Initialize project
specwatch init

# 2. Edit spec.md with your rules
vim spec.md

# 3. Configure AI (optional)
specwatch login --provider anthropic --api-key YOUR_KEY

# 4. Start watching
specwatch watch ./src

# 5. Fix violations as they appear
```

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
