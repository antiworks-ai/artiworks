## 14. Observability, Audit, Control Plane

### 14.1 Observability

Observability is system health and performance:

```yaml
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
    provider: otel
    exporter: otlp
    endpoint: ""
  profiling:
    enabled: false
    addr: "127.0.0.1:6060"
```

`observability.enabled=false` disables logging/metrics/tracing/profiling. Audit is separate.

Do not log prompt, full response, API keys, provider raw payload, tool args, file contents, memory contents, or headers by default.

### 14.2 Audit

Audit is security/accountability:

```yaml
audit:
  enabled: true
  store: persistence
  include_content: false
```

Audit events:

```text
auth.accepted / auth.denied
run.requested / run.completed / run.failed / run.canceled
provider.selected
middleware.blocked
tool.approval_requested
tool.approved / tool.denied
tool.executed / tool.failed
memory.proposed / memory.written / memory.forgotten
hook.executed / hook.failed
config.loaded / config.reload_failed
```

Audit records decision metadata, not large content or secrets.

### 14.3 Control Plane / Runtime Presence

Future App/IM integration must attach to a control plane, not to CLI internals.

```text
Observability = logs/metrics/traces/profiles
Audit = security decision history
Control Plane = runtime state, commands, approvals, attach/detach
Surface Adapter = CLI/App/IM/HTTP entry adapters
```

Control config:

```yaml
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

IM/App can read runtime snapshots, subscribe to projected events, create/cancel runs, and resolve approvals.

---

