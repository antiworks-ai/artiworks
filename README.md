# artiworks

Go-native terminal AI agent. Inspired by PI and Crush.

## Prerequisites

| Tool | Required | Check |
|------|----------|-------|
| Go | ≥ 1.26 | `go version` |
| git | ✓ | `git --version` |
| Node | * | `node --version` |
| golangci-lint | for lint | `golangci-lint --version` |

\* Node is only needed for `make install-skills` (installing Go agent skills via `npx`).

## Quick Start

```bash
# 1. Clone
git clone https://github.com/antiworks-ai/artiworks.git
cd artiworks

# 2. Install dev tools
make tools         # golangci-lint, govulncheck
make install-skills  # Go agent skills (43 skills from cc-skills-golang)

# 3. Configure
cp config.toml $HOME/.artiworks/config.toml
# Edit api_key with your LLM credentials

# 4. Build & run
make build
./artiworks
```

## Development

```bash
make test          # Run all tests
make lint          # Lint
make lint-fix      # Auto-fix lint issues
make schema        # Regenerate schema.json after config changes
make help          # Show all targets
```

## Configuration

Two config layers, project overrides user:

| Priority | Path |
|----------|------|
| 1 (high) | `.artiworks/config.toml` (per-project) |
| 2 (default) | `$ARTIWORKS_HOME/config.toml` (defaults to `~/.artiworks/config.toml`) |

Config values starting with `$` resolve from environment variables (e.g. `api_key = "$OPENAI_API_KEY"`).

## Project Structure

```
artiworks/
├── harness/
│   ├── api/               # Pure interfaces + types (zero deps)
│   │   ├── config.go      #   AppConfig, ProviderConfig, ...
│   │   ├── constant.go    #   All string constants
│   │   ├── home.go        #   ArtiworksHome(), EnsureDirs()
│   │   └── cmd/schema/    #   JSON Schema generator
│   ├── core/              # Framework implementation (WIP)
│   └── adapters/
│       ├── trpc/           #   trpc-agent-go (GraphAgent)
│       └── eino/           #   Eino (LLM provider)
├── app/                   # CLI + TUI application (WIP)
├── config.toml            # Example configuration
├── schema.json            # Generated JSON Schema
└── Makefile
```

> `harness/api/` is the contract — zero dependencies, imported by every other package.

## License

MIT
