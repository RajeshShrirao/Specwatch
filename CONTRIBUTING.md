# Contributing to Specwatch

First off, thank you for considering contributing to Specwatch! It's people like you that make Specwatch such a great tool for the community.

## 🌈 Our Philosophy

Specwatch is built on the idea that architectural rules should be:
1. **Fast**: Static checks must run in sub-10ms.
2. **Deterministic**: Rules in `spec.md` should produce consistent results.
3. **Beautiful**: The TUI should provide a premium developer experience.

## 🛠️ Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (1.22 or later)
- [Git](https://git-scm.com/downloads)

### Setup

1. Fork the repository on GitHub.
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/specwatch.git
   cd specwatch
   ```
3. Create a branch for your changes:
   ```bash
   git checkout -b feature/amazing-feature
   ```

## 🏗️ Project Structure

- `cmd/`: CLI command definitions (Cobra).
- `internal/analyzer/`: Core analysis engine and rule checkers.
- `internal/spec/`: Parser for the `spec.md` format.
- `internal/tui/`: Bubble Tea / Lipgloss based terminal interface.
- `internal/watcher/`: File system monitoring logic.

## 📜 Development Guidelines

### Coding Style
- Follow standard Go idioms and `gofmt`.
- Use descriptive variable and function names.
- Ensure all new features are documented in the README.

### TUI Changes
- When modifying the TUI, maintain the **three-panel layout**.
- Use the color tokens defined in `internal/tui/styles.go`.
- Ensure responsiveness across different terminal sizes.

## 🧪 Testing

Before submitting a PR, ensure your changes don't break existing functionality:
```bash
go test ./...
```

## 🚀 Submitting a Pull Request

1. Push your changes to your fork.
2. Open a Pull Request against the `main` branch.
3. Provide a clear description of the changes and why they are needed.
4. Link any related issues.

## 📄 License

By contributing to Specwatch, you agree that your contributions will be licensed under its [MIT License](LICENSE).
