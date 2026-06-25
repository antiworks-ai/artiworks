## 17. Schema Stability and Versioning

Separate:

```text
HTTP API version:
  /api/v1
  /v1 for OpenAI-compatible

Canonical schema version:
  api.Event / RunRequest / RunResult

Config schema version:
  version: 1
```

Allowed non-breaking changes:

- Add optional field.
- Add enum value.
- Add event type.
- Add metadata key.
- Add capability flag.
- Add optional config field.

Breaking changes:

- Rename field.
- Remove field.
- Change field type.
- Change event ordering guarantee.
- Change existing event type semantics.
- Make optional field required.
- Move data from one payload to another.

`Metadata` is for small, non-sensitive, non-core fields only. If a metadata key becomes core behavior, promote it to a real field.

Generated schemas are split by contract:

```text
config.schema.json
api.schema.json
events.schema.json
```

Root `schema.json` remains as a compatibility alias for `config.schema.json`.
Config documents that omit `version` are loaded as the current config version;
documents declaring a future version are rejected before runtime wiring.

---
