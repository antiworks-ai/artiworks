## 16. Config Design

Target top-level config:

```yaml
version: 1

server:
  addr: ":8080"
  api:
    native:
      enabled: true
      prefix: /api/v1
      streaming:
        transport: sse
        resume: true
        heartbeat_interval: 15s
    openai:
      enabled: true
      prefix: /v1
      mode: agent
      compatibility: best_effort
      endpoints:
        models: true
        chat_completions: true
        responses: true
  auth:
    enabled: false
    type: none
    tenant_header: X-Artiworks-Tenant
    project_header: X-Artiworks-Project
  idempotency:
    enabled: true
    header: Idempotency-Key
    ttl: 24h

models:
  default: default-chat
  aliases:
    default-chat:
      provider: openai
      name: gpt-4.1

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

harness:
  max_steps: 12
  max_tool_calls: 32
  timeout: 10m
  assembly:
    history:
      strategy: recent
      max_messages: 40
      summarize: true
    memory:
      enabled: true
      inject_as: developer_instruction
      max_tokens: 2000
    on_unsupported:
      thinking: downgrade
      image_input: error
      structured_output: downgrade
  token:
    enabled: true
    soft_compact_ratio: 0.5
    compact_ratio: 0.8
    compact_force_ratio: 0.9
    cleaners:
      tool_output:
        default: failure_focus
        max_bytes: 65536
      shell:
        strategy: failure_focus
        max_lines: 200
      grep:
        strategy: dedup
        max_lines: 300
  cache:
    enabled: true
    strategy: stable_prefix
    stable_prefix:
      include_tools: true
      include_project_memory: true
      deterministic_serialization: true
      forbid_volatile_fields: true

middleware:
  enabled: true
  starlark:
    enabled: true
    paths:
      - ~/.artiworks/middleware
  chains:
    before_run:
      - policy.default
      - redact.secrets
    before_emit_event:
      - redact.thinking

persistence:
  # Current productized local backends: file, memory.
  # SQLite remains a future backend behind the same persistence contract.
  type: file
  path: ~/.artiworks/persistence
  event_log:
    enabled: true
  snapshots:
    enabled: true
    on_run_completed: true

memory:
  enabled: true
  # Supported local stores: memory, persistence, file.
  # `persistence` and `file` write current-state JSON under
  # persistence.path/memory/items.json.
  store: persistence
  retrieval:
    top_k: 8
    min_score: 0.35
    inject_as: developer_instruction
    max_tokens: 2000
  write:
    enabled: true
    mode: propose

tools:
  enabled: true
  providers:
    builtin:
      type: builtin
      enabled: true
    local:
      type: local
      enabled: true
      allowed_commands:
        - git
        - rg
      allowed_roots:
        - ~/.artiworks
      timeout: 5s
      max_output_bytes: 65536
    mcp:
      type: mcp
      enabled: true
      command: npx
      args:
        - -y
        - @modelcontextprotocol/server-filesystem
      env:
        - MCP_LOG_LEVEL=error
    openapi:
      type: openapi
      enabled: true
      base_url: https://api.example.test
      spec_path: ./openapi.json
      timeout: 5s

hooks: {}

permissions:
  mode: ask
  approval:
    enabled: true
    timeout: 5m

secrets:
  providers:
    env:
      enabled: true
    file:
      enabled: false
      allowed_roots:
        - ~/.artiworks/secrets

observability:
  enabled: true
  logging:
    enabled: true
    level: info
    format: json
    include_content: false
    redact: true
  metrics:
    enabled: true
    path: /metrics
  tracing:
    enabled: false
  profiling:
    enabled: false
    addr: 127.0.0.1:6060

audit:
  enabled: true
  # Supported local stores: memory, persistence, file.
  # `persistence` and `file` append JSONL to
  # <persistence.path>/audit/records.jsonl.
  # Empty store keeps the compatibility default in-memory audit store.
  store: persistence
  include_content: false

control:
  enabled: true
  presence:
    enabled: true
    heartbeat_interval: 10s
  local:
    enabled: true
    transport: unix
    socket_path: ~/.artiworks/run/artiworks.sock
  relay:
    enabled: false
    endpoint: ""
    token_env: ARTIWORKS_RELAY_TOKEN
  expose:
    process: true
    active_runs: true
    event_tail: true
    content: redacted
```

The config implementation lives in `pkg/artiworks/config`.

When `secrets.providers.file.allowed_roots` is empty, `file:` refs keep the
compatibility default behavior. Once roots are present, only files that resolve
inside those roots are accepted. `artiworks status` uses the same local secret
resolution path as runtime wiring: it does not contact providers, but it fails
if configured provider credentials cannot be resolved locally.

When `memory.store` is omitted, Artiworks keeps the compatibility default
in-memory store. File-backed memory requires `persistence.path`; `propose` mode
returns proposed memories without persisting them or requesting approval.
`write` and `forget` modes require permission authorization before persistence.

Tool adapters currently ship with concrete `builtin`, `local`, `mcp`, and
`openapi` providers. `local` exposes allowlisted `shell.exec`, `fs.read`, and
`fs.write` tools; `mcp` discovers tools through a stdio MCP server; `openapi`
registers one tool per operation ID from a local OpenAPI v3 JSON spec.

When `control.local.transport = "unix"` and `socket_path` is set, `artiworks
serve` listens on that Unix socket unless `--addr` overrides it. Relay, IM, and
other external control surfaces stay behind explicit future contracts; enabling
`control.relay` currently fails fast instead of silently doing nothing.

---
