## 2. Package Layout

Target layout:

```text
artiworks/
├── cmd/
│   └── artiworks/
│       └── main.go
│
├── pkg/
│   └── artiworks/
│       ├── api/
│       ├── core/
│       ├── harness/
│       └── config/
│
├── internal/
│   ├── app/
│   │   ├── cli/
│   │   ├── http/
│   │   ├── tui/
│   │   └── wiring/
│   │
│   ├── infra/
│   │   ├── audit/
│   │   ├── control/
│   │   ├── memory/
│   │   ├── observability/
│   │   ├── persistence/
│   │   ├── secrets/
│   │   └── security/
│   │
│   └── adapters/
│       ├── api/
│       │   ├── native/
│       │   └── openai/
│       ├── ai/
│       │   ├── graph/
│       │   │   ├── trpc/
│       │   │   └── eino/
│       │   └── provider/
│       │       ├── anthropic/
│       │       ├── gemini/
│       │       ├── ollama/
│       │       ├── openai/
│       │       └── openaicompat/
│       ├── control/
│       │   ├── app/
│       │   ├── im/
│       │   ├── local/
│       │   └── relay/
│       └── tool/
│           ├── builtin/
│           ├── mcp/
│           └── openapi/
│
├── tools/
│   └── schema/
├── config.toml
├── schema.json
└── README.md
```

Dependency direction:

```text
api <- core <- harness <- internal/app
api <- harness
config <- internal/app
internal/adapters -> api/harness/config
internal/infra -> harness/config
```

Rules:

- `pkg/artiworks/api` uses stdlib only.
- `pkg/artiworks/core` depends on `api`, not on `harness`, `internal`, or provider SDKs.
- `pkg/artiworks/harness` depends on `api` and selected `core` contracts, and defines consumer-side interfaces.
- `pkg/artiworks/config` owns config structs and schema source.
- `internal/app/wiring` is the composition root.

---

