<p align="center">
  <img src="assets/logo.png" width="220" alt="Specwatch logo">
</p>

<h1 align="center">Specwatch</h1>

<p align="center">
  <strong>Keep architecture from drifting while you code.</strong>
</p>

<p align="center">
  Specwatch watches your repo, reads a local <code>spec.md</code>, runs fast static checks first,
  and escalates to AI only when a rule needs semantic judgment.
</p>

<p align="center">
  <img src="https://img.shields.io/github/v/release/rajeshshrirao/specwatch?style=flat-square&color=4F8EF7" alt="Version">
  <img src="https://img.shields.io/github/go-mod/go-version/rajeshshrirao/specwatch?style=flat-square" alt="Go Version">
  <img src="https://img.shields.io/github/license/rajeshshrirao/specwatch?style=flat-square" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/rajeshshrirao/specwatch/ci.yml?style=flat-square" alt="CI">
</p>

---

## What is Specwatch?

Specwatch is a **spec-driven architectural drift detector** that runs in your terminal alongside your editor.

You write a `spec.md` that describes how your project is supposed to be structured — naming conventions, forbidden patterns, import boundaries, file size limits, architectural rules. Specwatch watches your codebase and tells you the moment code drifts from that contract.

Static rules are checked in milliseconds with zero API calls. Rules that require semantic judgment — like "no business logic in components" — are escalated to an LLM, but only when static analysis can't catch it, and only once per 10 file saves to keep costs near zero.

It is a single binary, installs with one command, and reads a plain markdown file. Nothing else required.

---

## Is Specwatch right for you?

**Yes, if you:**

- Use Claude Code, Cursor, or any AI coding tool and want a guardrail that catches when generated code drifts from your intended architecture
- Work on a codebase with rules that aren't enforced anywhere — naming conventions, import boundaries, layering rules — and violations only surface in review
- Want instant terminal feedback on saves rather than waiting for CI
- Prefer a lightweight file-based tool over setting up a complex linting pipeline

**Probably not, if you:**

- Need enforcement across a large monorepo with multiple independent specs — that's not the current focus
- Want deep type-level analysis — Specwatch is not a replacement for `tsc`, ESLint, or Go vet
- Need rules to block merges automatically — use the `check` command in CI for that, but the tool is primarily designed for the inner dev loop

---

## Problems it solves

**AI-assisted drift** — AI coding tools generate code fast. Too fast to manually review every structural decision. Specwatch watches what gets written and catches violations the moment a file is saved, not after a PR review.

**Undocumented architecture rot** — Every codebase has rules that exist only in people's heads: "we don't call the database directly from handlers," "all API errors return `{ error: string }`." Specwatch forces those rules into a file and then actually enforces them.

**Slow feedback loops** — Most teams only discover architectural violations in CI or in code review. Specwatch surfaces them in the terminal within 800ms of saving a file.

**Expensive AI overuse** — Other AI-assisted linters send every file change to an LLM. Specwatch runs static checks first and only escalates to AI when the rule genuinely requires semantic understanding. In practice, most sessions use zero API calls.

---

## Why Specwatch is different

Most linters enforce syntax and style. Specwatch enforces **intent**.

The design philosophy in one sentence: **rules that can be computed should never hit an LLM.**

That means:

- `console.log` is forbidden → regex match, 0ms, no API call
- file exceeds 300 lines → line count, 0ms, no API call
- component imports from the wrong layer → import graph, sub-10ms, no API call
- "no business logic in components" → this needs judgment, LLM fires, but only once per 10 saves

Everything else in this space either does too little (static linters with no architectural awareness) or too much (sends every file to an AI on every save). Specwatch sits exactly in the middle.

```
Your spec.md → static checks (< 10ms) → AI only when needed (1 per 10 saves)
```

---

## What Specwatch is not

- **Not a replacement for ESLint, tsc, or go vet.** Those tools check correctness. Specwatch checks architecture.
- **Not a CI-only tool.** It's designed for the live inner loop. `check` mode exists for CI but the real value is in `watch`.
- **Not magic.** The quality of drift detection is directly proportional to how well you write your `spec.md`. Vague rules produce vague results.
- **Not a security scanner.** Specwatch has no awareness of CVEs, OWASP rules, or dependency vulnerabilities.
- **Not production-ready for monorepos.** Single-repo enforcement is the current focus. Monorepo discovery is on the roadmap.

---

## Quick Start

### Install

```bash
go install github.com/rajeshshrirao/specwatch@latest
```

### Initialize

```bash
cd your-project
specwatch init
```

This generates a starter `spec.md` in the current directory. Edit it to match your project's actual rules.

### Watch

```bash
specwatch watch .
```

The TUI launches. Save any file to see analysis run in real time.

### One-shot check (CI)

```bash
specwatch check . --format json
```

Exits `0` if no violations, `1` if violations found.

---

## The two files that matter

### `spec.md`

The contract for your codebase. Written in plain markdown with a specific key-value syntax per section so static analysis can parse rules deterministically.

```md
## stack
- language: typescript
- framework: next.js@14

## naming
- components: PascalCase
- files: kebab-case

## forbidden
- pattern: "console.log"
  message: use logger from @/lib/logger
- import: "lodash"
  message: use native ES methods

## required
- async functions: try/catch

## limits
- max file lines: 300
- max imports per file: 20

## architecture
- no direct db calls outside src/lib/db
- no business logic in components — belongs in hooks or lib
```

> **Important:** The `forbidden`, `required`, `naming`, and `limits` sections use structured key-value syntax. The `architecture` section takes plain natural language — that's what gets escalated to AI when static heuristics can't catch it.

### `.specwatch.yml`

Optional runtime config. Lives next to `spec.md`. Silently ignored if absent.

```yaml
llm:
  enabled: true
  provider: anthropic        # anthropic | openrouter | gemini
  model: claude-haiku-4-5-20251002

watch:
  debounce: 800              # ms to wait after save before analyzing
  extensions: [go, ts, tsx, js, jsx]
```

---

## AI checks

AI is intentionally constrained. This is by design, not a limitation.

| Condition | What happens |
|---|---|
| Static rule triggered | LLM never called |
| Architecture rule, AI disabled | Static heuristics only |
| Architecture rule, AI enabled, file clean | LLM fires |
| 10th save not yet reached in watch mode | LLM skipped, queued |
| `ANTHROPIC_API_KEY` missing | One-time warning, tool continues without AI |

Set up a provider:

```bash
export ANTHROPIC_API_KEY="your-key"
```

Inspect available models:

```bash
specwatch login --provider openrouter --list-models
```

Supported providers: `anthropic`, `openrouter`, `gemini`

---

## Commands

| Command | What it does |
|---|---|
| `specwatch init` | Generate a starter `spec.md` |
| `specwatch watch [path]` | Live watch mode with TUI |
| `specwatch check [path]` | One-shot CI check |
| `specwatch login` | Configure and validate provider auth |
| `specwatch version` | Print version |

```bash
# Examples
specwatch watch . --debounce 1200
specwatch watch ./src --ext go,ts,tsx
specwatch watch . --skip limits
specwatch check . --format json
```

---

## Terminal UI

Watch mode is a live architecture console, not a log viewer.

- Animated startup and shutdown sequences
- Three-panel layout: activity feed, violations list, live stats
- Violations sorted by severity then recency
- Detail pane on `Enter` showing excerpt and suggested fix
- Compact fallback layout for small terminals
- `8ms` analysis latency visible in footer at all times

**Keyboard shortcuts:**

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Expand violation detail |
| `c` | Clear all violations |
| `f` | Filter by severity |
| `q` | Quit |

---

## FAQ

**Do I need an API key to use Specwatch?**
No. Static analysis works entirely offline. An API key is only needed if you set `llm.enabled: true` in `.specwatch.yml` and want architecture rules checked semantically. Most users will never need it.

**How much does the AI cost to run?**
In watch mode, AI fires at most once per 10 file saves, only for architecture-section rules, and only when static analysis found nothing. A typical hour-long session might trigger 3–5 AI calls. At Haiku pricing that's a fraction of a cent.

**Will it slow down my editor?**
No. Specwatch is a separate process. It watches the filesystem and runs in its own terminal. Your editor is never touched.

**Can I use it without a `spec.md`?**
No. The spec is the entire point. `specwatch init` generates one in under a second.

**Does it work with any language?**
The static checks are pattern-based so they work on any text file. The named analyzers (naming conventions, import detection, try/catch) currently understand TypeScript, JavaScript, and Go best. Support for other languages improves with community contributions.

**Can I run it in CI?**
Yes. Use `specwatch check . --format json`. It exits `1` if violations exist, `0` if clean. Drop it into any pipeline.

**What if I disagree with a violation?**
Edit `spec.md`. Specwatch only knows what you tell it. You own the rules entirely.

---

## Development

```bash
git clone https://github.com/rajeshshrirao/specwatch.git
cd specwatch

go build -o specwatch .
go test -v -race -cover ./...
go run . watch .
go run . check .
```

The codebase is deliberately simple. Key directories:

- `internal/spec/` — `spec.md` parser and rule types
- `internal/analyzer/` — static checks and LLM escalation engine
- `internal/watcher/` — fsnotify wrapper with debouncer
- `internal/tui/` — Bubble Tea UI model and styles
- `cmd/` — Cobra CLI commands

---

## Roadmap

**v0.2**
- [ ] Upward directory traversal for `spec.md` discovery (monorepo support)
- [ ] Per-project config namespace so concurrent repos don't bleed settings
- [ ] Tests for `fix drift → event → violation cleared` loop

**v0.3**
- [ ] VS Code extension — inline violation markers without leaving the editor
- [ ] `specwatch explain` — natural language explanation of any spec rule
- [ ] Git hook integration — block commits with unresolved drift

**Longer term**
- [ ] Language server protocol support
- [ ] Team-shared spec registries
- [ ] Spec generation from existing codebase (reverse engineering rules from patterns)

Have a feature idea? [Open a discussion](https://github.com/rajeshshrirao/specwatch/discussions).

---

## Contributing

Contributions are welcome. The bar is high on purpose — this is a tool people run in their inner loop and slowness or noise is a first-class bug.

Before opening a PR:

1. Check existing issues and discussions to avoid duplication
2. For non-trivial changes, open an issue first to align on approach
3. Run `go test -v -race -cover ./...` — all tests must pass
4. Fast checks must stay fast — static analysis should never block on IO
5. UI changes should not reduce clarity — the TUI is minimal by design

```bash
git checkout -b feature/your-feature
# make changes
go test ./...
git commit -m "feat: describe what changed and why"
git push origin feature/your-feature
# open PR
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full details.

---

## Community

- **Issues** — [github.com/rajeshshrirao/specwatch/issues](https://github.com/rajeshshrirao/specwatch/issues)
- **Discussions** — [github.com/rajeshshrirao/specwatch/discussions](https://github.com/rajeshshrirao/specwatch/discussions)
- **X / Twitter** — [@RajeshShrirao](https://x.com/RajeshShrirao)

If Specwatch is useful to you, a star on GitHub helps others find it.

---

<p align="center">
  Built by <a href="https://github.com/rajeshshrirao">Rajesh Shrirao</a>
</p>