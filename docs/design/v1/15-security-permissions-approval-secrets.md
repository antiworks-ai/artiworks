## 15. Security, Permissions, Approval, Secrets

### 15.1 Security Layers

```text
Authentication = who you are
Authorization  = whether you can act
Permission     = action/resource policy decision
Approval       = human confirmation when decision is ask
```

All sensitive actions produce `PermissionRequest` and `PermissionDecision`.

Actions requiring permission:

```text
tool.execute
tool.approve
tool.deny
memory.write
memory.forget
control.run_create
control.run_cancel
control.status_read
control.attach
control.config_reload
hook.command_execute
hook.webhook_send
file.read
file.write
shell.exec
network.request
provider.use
model.use
```

### 15.2 Approval Flow

Approval is first-class and cross-surface:

```text
permission decision = ask
 -> create ApprovalRequest
 -> emit approval.requested
 -> persist pending approval
 -> CLI/App/IM approve or deny
 -> emit approval.resolved
 -> harness continues or fails
```

### 15.3 Hard Security Rules

These are mandatory:

- Middleware cannot bypass permission gates.
- Starlark cannot read API keys or provider raw headers.
- Tool executor cannot skip approval on its own.
- Remote control command must include actor and source.
- Approval and denial must write audit records.
- Secrets can only be resolved through `SecretProvider`.
- Secrets must not enter `RunRequest`, `Event`, `RunResult`, `MessagePart`, `Metadata`, `MiddlewareContext`, memory, prompt, logs, traces, metrics, or audit payloads.
- Command hooks must not pass secrets through shell string interpolation.
- Provider raw headers must not enter events or debug output.

### 15.4 Secrets

Config stores secret references, not values:

```yaml
providers:
  openai:
    type: openai
    credentials:
      api_key:
        ref: env:OPENAI_API_KEY
```

Supported refs:

```text
env:OPENAI_API_KEY
file:/path/to/token
keychain:artiworks/openai
vault:path/to/secret#field
```

Current local secret resolution supports `env` and `file`. `keychain` and
`vault` remain reserved reference forms for future secret-provider backends.

For product use, `file` refs are constrained by
`secrets.providers.file.allowed_roots` when configured. Empty roots preserve
the compatibility default, but configured roots must contain the resolved file
path after canonicalization and symlink resolution.

---
