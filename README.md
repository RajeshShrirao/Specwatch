# ╭─ specwatch ─────────────────────────────────────────── ● watching ─╮

<p align="center">
  <img src="https://img.shields.io/github/v/release/rajeshshrirao/specwatch?include_prereleases&style=flat&label=version" alt="Version">
  <img src="https://img.shields.io/github/go-mod/go-version/rajeshshrirao/specwatch?style=flat" alt="Go Version">
  <img src="https://img.shields.io/github/license/rajeshshrirao/specwatch?style=flat" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/rajeshshrirao/specwatch/ci.yml?branch=main" alt="CI">
</p>

<p align="center">
  A blazing fast, structured spec-driven static analysis tool for modern web development.
</p>

</p>

---

## ✨ Why Specwatch?

Traditional linters and formatters are great for enforcing syntax and style, but they don't capture your project's **architectural soul**. Specwatch fills that gap by letting you define **custom rules** in a simple `spec.md` file—and then enforces them automatically.

### The Philosophy

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   Your Vision           Static Analysis        LLM Fallback│
│       │                      │                    │        │
│       ▼                      ▼                    ▼        │
│   ┌────────┐          ┌──────────────┐       ┌────────┐    │
│   │ spec.md│    ──►   │   Fast Checks│  ──► │ Semantic│    │
│   └────────┘          └──────────────┘       │  Rules │    │
│                                              └────────┘    │
│                                                             │
│   Rules that can be computed should NEVER hit an LLM       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Target: Max 1 LLM API call per 10 file saves.**

---

## 🚀 Features

| Feature | Description |
|---------|-------------|
| ⚡ **Blazing Fast** | Sub-10ms static checks using regex and AST analysis |
| 🎯 **Spec-Driven** | Centralized `spec.md` defines your project's architectural rules |
| 🖥️ **Beautiful TUI** | Real-time activity feed, violation tracking, and live statistics |
| 🧠 **Smart LLM Usage** | LLMs used sparingly and surgically for complex semantic rules |
| 🔄 **File Watching** | Automatically re-analyzes on file changes with debouncing |
| 📊 **CI-Ready** | One-shot `check` mode with JSON/Text output for pipelines |
| 🔒 **Zero Config** | No complex setup—just a `spec.md` and you're ready |
| 🌈 **Rich TUI** | Built with Bubble Tea for a delightful terminal experience |

---

## 📦 Installation

### From Source

```bash
git clone https://github.com/rajeshshrirao/specwatch.git
cd specwatch
go install
```

### Using Go Install

```bash
go install github.com/rajeshshrirao/specwatch@latest
```

### Verify Installation

```bash
specwatch --version
```

---

## 🏃‍♂️ Quick Start

### 1. Initialize a new spec

```bash
specwatch init
```

This creates a `spec.md` file in your current directory with sensible defaults.

### 2. Start watching for changes

```bash
specwatch watch ./src
```

The TUI will launch showing real-time analysis of file changes.

### 3. Run a one-shot check (great for CI)

```bash
specwatch check ./src
```

---

## 📖 Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| `specwatch init` | Initialize a new `spec.md` | `specwatch init` |
| `specwatch watch [path]` | Watch directory for changes | `specwatch watch ./src` |
| `specwatch check [path]` | Run once and exit (CI mode) | `specwatch check ./src` |

### Check Options

```bash
specwatch check ./src --format json    # JSON output
specwatch check ./src --format text    # Plain text output
```

---

## 📝 The `spec.md` Format

The `spec.md` file is the heart of Specwatch. It defines all your project's rules in a declarative Markdown format.

### Supported Rule Types

| Section | Purpose |
|---------|---------|
| `stack` | Define technology stack (language, framework, styling, runtime) |
| `structure` | Define directory structure and file locations |
| `naming` | Enforce naming conventions for files, functions, components |
| `forbidden` | Block specific patterns, imports, or code patterns |
| `required` | Enforce mandatory patterns (try/catch, return types, etc.) |
| `architecture` | Define import boundaries and architectural constraints |
| `limits` | Set code metrics thresholds (max lines, max imports, etc.) |

### Example `spec.md`

```markdown
# My Project Spec

## stack
- language: typescript
- framework: next.js@14
- styling: tailwind
- runtime: node@20

## structure
- components: src/components/**
- api routes: src/app/api/**
- utilities: src/lib/**
- types: src/types/**
- tests: **/*.test.ts, **/*.spec.ts

## naming
- components: PascalCase
- functions: camelCase
- files: kebab-case
- constants: SCREAMING_SNAKE_CASE
- interfaces: PascalCase prefixed with I

## forbidden
- pattern: "console.log"
  message: Use logger utility from @/lib/logger
- pattern: "\\bany\\b"
  message: No any types — use unknown or explicit type
- pattern: "style={{"
  message: No inline styles — use tailwind classes
- import: "lodash"
  message: Use native ES methods instead
- import: "moment"
  message: Use date-fns instead

## required
- async functions: try/catch
- api routes: return type { data, error }
- components: must have displayName
- new files in src/components: must have matching *.test.ts

## architecture
- no direct db calls outside src/lib/db
- no business logic in components — belongs in hooks or lib
- server components by default — client components need explicit justification

## limits
- max file lines: 300
- max function lines: 50
- max imports per file: 20
- max component props: 8
```

---

## 🖥️ TUI Controls

When running `specwatch watch`, use these keyboard shortcuts:

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down in violation list |
| `k` / `↑` | Move up in violation list |
| `Enter` | Expand violation detail |
| `c` | Clear all violations |
| `f` | Filter by severity level |
| `q` / `Esc` | Quit |

### TUI Layout

```
┌─────────────────────────────────────────────────────────────┐
│  specwatch                                    ✓ Watching   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─ Recent Violations ─────────────────────────────────┐   │
│  │  ✗ src/components/Button.tsx:23                     │   │
│  │    → Inline styles detected                          │   │
│  │  ✗ src/utils/api.ts:45                              │   │
│  │    → Missing try/catch in async function            │   │
│  │  ⚠ src/hooks/useAuth.ts:12                          │   │
│  │    → Function exceeds 50 lines (62)                 │   │
│  └────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─ Stats ─────────────────────────────────────────────┐    │
│  │  Files: 156  │  Violations: 3  │  Time: 4.2ms     │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
│  [j/k] Navigate  [c] Clear  [f] Filter  [q] Quit         │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔧 CI/CD Integration

### GitHub Actions

```yaml
name: Spec Check
on: [push, pull_request]

jobs:
  speccheck:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Install specwatch
        run: go install github.com/rajeshshrirao/specwatch@latest
      
      - name: Run specwatch
        run: specwatch check ./src --format json >> $GITHUB_STEP_SUMMARY
```

### GitLab CI

```yaml
speccheck:
  image: golang:1.22
  script:
    - go install github.com/rajeshshrirao/specwatch@latest
    - specwatch check ./src --format json
  allow_failure: false  # Set to true to warn only
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | No violations found |
| `1` | Violations found |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        specwatch                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌───────────┐    ┌──────────────────┐    │
│  │  CLI     │───►│  Engine   │───►│  Analyzers       │    │
│  │ (Cobra)  │    │           │    │                  │    │
│  └──────────┘    └───────────┘    │  - Forbidden    │    │
│         │              │          │  - Naming       │    │
│         │              ▼          │  - Limits        │    │
│         │        ┌─────────┐      │  - Required      │    │
│         │        │ Parser  │      │  - Architecture  │    │
│         │        │(Markdown│      └──────────────────┘    │
│         │        └─────────┘                  │            │
│         │               │                     ▼            │
│         │               ▼            ┌────────────────┐    │
│         │        ┌────────────┐       │   Reporters    │    │
│         │        │  spec.md   │       │                │    │
│         │        └────────────┘       │  - Console     │    │
│         │                             │  - JSON         │    │
│         │                             │  - TUI          │    │
│         │                             └────────────────┘    │
│         │                                              │
│         ▼                                              ▼
│  ┌─────────────┐                           ┌─────────────┐
│  │   Watcher   │                           │  Reporter   │
│  │ (fsnotify)  │                           │             │
│  └─────────────┘                           └─────────────┘
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Core Components

- **CLI** - Command-line interface using Cobra
- **Parser** - Parses `spec.md` into structured rules
- **Engine** - Orchestrates analysis and coordinates analyzers
- **Analyzers** - Individual rule checkers (Forbidden, Naming, Limits, etc.)
- **Watcher** - File system watcher with debouncing
- **Reporters** - Output formatters (Console, JSON, TUI)

---

## 🤝 Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) first.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

Built with these amazing open-source projects:

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [fsnotify](https://github.com/fsnotify/fsnotify) - File system notifications
- [Goldmark](https://github.com/yuin/goldmark) - Markdown parser

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/rajeshshrirao">rajeshshrirao</a>
</p>
