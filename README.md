# artiworks

Go-native terminal AI agent runtime. artiworks is canonical-first: it defines its own run, event, message, tool, memory, model, and error contracts, then adapts external protocols around them.

## Prerequisites

| Tool | Required | Check |
|------|----------|-------|
| Go | >= 1.26 | `go version` |
| git | yes | `git --version` |
| Node | optional | `node --version` |
| golangci-lint | for lint | `golangci-lint --version` |

Node is only needed for development tasks that install agent skills through `npx`.

## Quick Start

```bash
git clone https://github.com/antiworks-ai/artiworks.git
cd artiworks

make tools
make install-skills

cp config.toml "$HOME/.artiworks/config.toml"
make build
./artiworks
```

## Development

```bash
make test
make lint
make lint-fix
make schema
make help
```

## Configuration

Two config layers are planned. Project config overrides user config.

| Priority | Path |
|----------|------|
| 1 | `.artiworks/config.toml` |
| 2 | `$ARTIWORKS_HOME/config.toml` |

Provider names are instances. `type` selects the adapter. `api` selects the outbound provider protocol.

```yaml
providers:
  openai:
    type: openai
    api: auto
    api_key_env: OPENAI_API_KEY

  deepseek:
    type: openai-compatible
    api: chat_completions
    base_url: https://api.deepseek.com/v1
    api_key_env: DEEPSEEK_API_KEY

models:
  default: default-chat
  aliases:
    default-chat:
      provider: openai
      name: gpt-4.1
    deepseek-chat:
      provider: deepseek
      name: deepseek-chat

harness:
  token:
    enabled: true
    soft_compact_ratio: 0.5
    compact_ratio: 0.8
    compact_force_ratio: 0.9
  cache:
    enabled: true
    strategy: stable_prefix
```

OpenAI-compatible inbound APIs and outbound provider protocols are independent. artiworks must support both `/v1/chat/completions` and `/v1/responses` inbound, even when an outbound provider only supports Chat Completions-compatible requests.

Token cleaning and cache-hit optimization are harness responsibilities. The runtime should preserve raw tool output in events/artifacts, pass cleaned or referenced content to the model, and protect a stable prompt prefix whenever possible.

## Project Structure

```text
artiworks/
├── cmd/
│   └── artiworks/          # CLI entrypoint
├── pkg/
│   └── artiworks/
│       ├── api/            # Public contracts and canonical DTOs
│       ├── core/           # Pure state/reducer/snapshot/replay logic
│       ├── harness/        # Agent runtime shell around LLM calls
│       └── config/         # Config structs and schema source
├── internal/
│   ├── app/                # CLI/HTTP/TUI process wiring
│   ├── infra/              # Persistence, secrets, audit, observability, control
│   └── adapters/
│       ├── api/            # Native and OpenAI-compatible inbound adapters
│       ├── ai/
│       │   ├── graph/      # trpc/eino graph adapters
│       │   └── provider/   # OpenAI, OpenAI-compatible, Anthropic, Gemini, Ollama
│       ├── control/        # Local/App/IM/relay control adapters
│       └── tool/           # Builtin, MCP, OpenAPI tool adapters
├── tools/
│   └── schema/             # JSON Schema generator
├── config.toml
├── schema.json             # Config schema for now
└── Makefile
```

Package meanings are strict:

- `pkg/artiworks/api` is the public contract layer, not the HTTP layer.
- `pkg/artiworks/core` is pure reducer/state logic.
- `pkg/artiworks/harness` is the Agent runtime shell: middleware, memory, tools, permissions, approval, prompt assembly, token budgeting, cache-aware context planning, session coordination, event routing, and agent loop.
- `pkg/artiworks/config` owns configuration models and schema source.
- `internal/` contains concrete app wiring, infrastructure, and adapters.

## Design Notes

The v1 design entry point is [docs/artiworks-design-v1.md](./docs/artiworks-design-v1.md). The design is split into smaller files under [docs/design/v1](./docs/design/v1/) for a future VitePress site.

## License

MIT
