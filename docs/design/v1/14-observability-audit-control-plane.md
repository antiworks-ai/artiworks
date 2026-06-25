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

Local audit storage supports the current in-memory default and a file-backed persistence mode. When `audit.store` is `persistence` or `file`, audit records are appended as JSONL under `persistence.path/audit/records.jsonl`; the store scans existing records on startup and continues sequence assignment from the highest stored sequence. When `audit.store` is omitted, the app keeps the compatibility default in-memory audit store.

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
Hook dispatch now writes `hook.executed` and `hook.failed` audit records for
matched hook attempts, with redacted metadata only.

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

The local text TUI surface is read-only: `artiworks tui` renders the redacted control snapshot, a projected timeline, and the raw redacted event tail. The timeline is derived only from `control.EventSummary` IDs, event types, statuses, sequence ranges, and terminal/frozen state; it does not expose prompt text, provider deltas, tool arguments, memory content, headers, metadata, or secrets. JSON output remains the raw `control.Snapshot` for automation.

The local control-plane run command MVP exposes in-process run inspection, creation, and cancellation:

```text
GET  /control/v1/runs
GET  /control/v1/runs/{run_id}
POST /control/v1/runs
POST /control/v1/runs/{run_id}/cancel
```

Create and cancel require explicit `actor` and `source`, pass through `PermissionAuthorizer` using `control.run_create` or `control.run_cancel`, and write `run.requested` or `run.canceled` audit records after the command manager accepts the command. Create starts the run asynchronously inside the current process; cancel signals the stored run context and final status is visible through the run command record or snapshot/event tail.

The local control-plane approval MVP exposes read and resolve endpoints for existing approval records:

```text
GET  /control/v1/approvals
GET  /control/v1/approvals/{approval_id}
POST /control/v1/approvals/{approval_id}/resolve
```

Resolution accepts only `approved` and `rejected`, passes through `PermissionAuthorizer`, and records `tool.approved` or `tool.denied` audit records. Durable run resume, remote relay authentication, and subscription flows stay behind later explicit control-plane contracts.

---
