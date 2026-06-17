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
  type: sqlite
  path: ~/.artiworks/artiworks.db
  event_log:
    enabled: true
    mode: compact
  snapshots:
    enabled: true
    on_run_completed: true
  artifacts:
    type: fs
    path: ~/.artiworks/artifacts

memory:
  enabled: true
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
    mcp:
      type: mcp
      enabled: true
    openapi:
      type: openapi
      enabled: false

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

audit:
  enabled: true
  store: persistence
  include_content: false

control:
  enabled: true
```

The config implementation lives in `pkg/artiworks/config`.

---

