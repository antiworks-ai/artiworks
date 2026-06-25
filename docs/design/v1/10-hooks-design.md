## 10. Hooks Design

Hooks are lifecycle side effects:

```text
Hook = observe + side effect
Middleware = mutate/block/short-circuit canonical flow
Tool = model-callable capability
```

Initial hook events:

```text
on_session_created
on_run_started
on_run_completed
on_run_failed
on_tool_started
on_tool_completed
on_tool_failed
on_memory_proposed
on_memory_written
on_error
```

Hook types:

```text
command
webhook
```

Security:

- Command hooks must use `exec.Command(name, args...)`, not shell string concatenation.
- Webhooks need timeout, retry limits, and payload size limits.
- Hook payloads must be redacted.
- Hook failures do not block main flow unless `critical: true`.
- Matched hook attempts write `hook.executed` or `hook.failed` audit records
  through the app audit store without exposing hook payload content.

---
