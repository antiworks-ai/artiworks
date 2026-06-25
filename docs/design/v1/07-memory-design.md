## 7. Memory Design

Memory is long-term, retrievable, injectable knowledge. It is not session history and not provider prompt cache.

Kinds:

```text
fact
preference
summary
document
```

Memory scope:

```text
global
tenant
project
user
session-derived
```

Memory pipeline:

```text
before_run:
  build MemoryQuery
  retrieve MemoryHit[]
  middleware filter/rerank/redact
  inject as Instruction

after_run:
  extractor proposes MemoryItem[]
  policy/approval
  writer persists
```

Default write mode should be `propose`, not automatic.
`propose` is non-persistent and does not require approval; `write` and
`forget` require permission authorization before they mutate the memory store.

Memory injection must become `Instruction`, not a fake user message.

---
