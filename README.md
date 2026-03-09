# ╭─ specwatch ─────────────────────────── ● watching ─╮

A fast, structured spec-driven static analysis tool for modern web development.

`specwatch` ensures your codebase strictly adheres to your architectural rules, naming conventions, and structural patterns—without reaching for an LLM until it absolutely has to.

## 🚀 The Philosophy

```text
structured spec → parsed rule set → static analysis first
                                  → LLM only for semantic rules
                                  → never for rules that can be computed
```

**LLM budget target: max 1 API call per 10 file saves.**

## ✨ Features

- **⚡ Blazing Fast**: Static checks run in sub-10ms.
- **🛠️ Structured Spec**: Centralized `spec.md` defines your project's soul.
- **📟 Beautiful TUI**: Real-time activity feed, violation tracking, and stats.
- **🤖 Smart LLM Usage**: Architecture rules use LLMs sparingly and surgically.
- **🏗️ CI Ready**: One-shot check mode for your pipelines.

## 📦 Installation

```bash
go install github.com/rajeshshrirao/specwatch@latest
```

## 🛠️ Getting Started

1. **Initialize spec**:
   ```bash
   specwatch init
   ```
2. **Start watching**:
   ```bash
   specwatch watch ./src
   ```

## 📄 The `spec.md` Format

Define rules for `stack`, `structure`, `naming`, `forbidden`, `required`, `architecture`, and `limits`.

```markdown
# specwatch

## naming
- components: PascalCase
- functions: camelCase
- files: kebab-case

## forbidden
- pattern: "console.log"
  message: use logger utility from @/lib/logger
```

## ⌨️ TUI Shortcuts

| Key | Action |
| --- | --- |
| `j`/`k` or `↑`/`↓` | Navigate violations |
| `enter` | Expand violation detail |
| `c` | Clear all violations |
| `f` | Filter by severity |
| `q` | Quit |

---

Built with Go, Bubbletea, and ❤️ by [rajeshshrirao](https://github.com/rajeshshrirao)
