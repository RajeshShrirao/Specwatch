# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Specwatch is a Go-based static analysis tool that enforces architectural rules defined in `spec.md` files. It combines fast static checks with a premium terminal interface using Bubble Tea.

**Key Technologies:**
- Go 1.22+ (primary language)
- Bubble Tea (TUI framework)
- Cobra (CLI framework)
- Goldmark (Markdown parsing)
- fsnotify (file watching)
- Anthropic SDK (AI-powered analysis)

## Architecture

```
specwatch/
├── cmd/           # CLI commands (Cobra)
├── internal/
│   ├── analyzer/   # Core analysis engine
│   ├── auth/       # Authentication manager
│   ├── llm/        # LLM providers (Anthropic, OpenRouter, Gemini)
│   ├── spec/       # spec.md parser
│   ├── tui/        # Terminal interface (Bubble Tea)
│   └── watcher/    # File system monitoring
└── spec.md        # Architectural rules
```

## Common Development Commands

### Building & Running
```bash
# Build the binary
go build -o specwatch .

# Run directly
go run .

# Install for system use
go install .
```

### Testing
```bash
# Run all tests
go test -v -race -cover ./...

# Run tests for specific package
go test -v ./internal/analyzer

# Generate coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting
```bash
# Run golangci-lint
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Development Workflow
```bash
# Watch mode (development)
go run . watch ./src

# Check mode (CI)
go run . check ./src

# Initialize spec.md
go run . init
```

## Code Structure Guidelines

### Package Organization
- `cmd/`: CLI commands (Cobra)
- `internal/analyzer/`: Core analysis logic
- `internal/auth/`: Authentication manager
- `internal/llm/`: LLM providers (Anthropic, OpenRouter, Gemini)
- `internal/spec/`: spec.md parsing
- `internal/tui/`: Terminal interface
- `internal/watcher/`: File watching

### Testing Requirements
- All new packages need corresponding test files
- Use table-driven tests for complex logic
- Mock external dependencies in tests

### Performance Requirements
- Static checks must run in sub-10ms
- Use regex for simple pattern matching
- Reserve AI calls for complex semantic rules

## Development Standards

### Go Standards
- Use `gofmt` for code formatting
- Follow standard Go naming conventions
- Use descriptive variable names
- Handle errors explicitly

### TUI Development
- Maintain three-panel layout
- Use color tokens from `internal/tui/styles.go`
- Ensure responsive across terminal sizes
- Test with different terminal dimensions

### Architecture Rules
- No business logic in UI components
- Use interfaces for external dependencies
- Keep functions under 50 lines
- Limit imports per file to 20

## Configuration Files

- `spec.md`: Project-specific architectural rules
- `.specwatch.yml`: Runtime configuration
- `.github/workflows/ci.yml`: CI/CD pipeline

## Common Patterns

### Error Handling
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### File Watching
- Use debouncing (800ms default)
- Filter by extensions
- Skip hidden directories

### TUI Updates
- Use Bubble Tea messages for state updates
- Keep TUI responsive during analysis
- Show progress indicators for long operations

## CI/CD Integration

### Pipeline Steps
1. Checkout code
2. Set up Go 1.22
3. Run tests with race detection
4. Run golangci-lint
5. Build binary
6. Upload coverage

### Environment Variables
- `ANTHROPIC_API_KEY`: Anthropic API key for AI-powered checks
- `OPENROUTER_API_KEY`: OpenRouter API key (alternative to Anthropic)
- `GEMINI_API_KEY`: Google Gemini API key
- `CI`: Set to true in CI environment

## Documentation

- Update README.md for new features
- Document architectural decisions in spec.md
- Use Go doc comments for public APIs
- Include examples in README for complex features