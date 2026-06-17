## 6. State, Reducer, Session, Persistence

### 6.1 State / Reducer

`core.State` is a runtime projection. `api.Event` is the fact stream. Reducer is deterministic and side-effect free:

```text
api.Event -> core.Reducer -> core.State + Patch
```

State shape:

```text
Runs map[RunID]*RunNode
Turns map[TurnID]*TurnNode
Messages map[MessageID]*MessageNode
Tools map[ToolCallID]*ToolNode
RootRuns []RunID
LastSeq int64
```

Reducer validates event order/references, mutates state, and returns a patch for renderers/session materialization.

### 6.2 Session

`Session` is a durable conversation boundary, not `State`, not `EventLog`, and not `Memory`.

```text
Session:
  ID
  Title
  Status
  RootRunIDs
  HeadRunID
  Metadata
  CreatedAt
  UpdatedAt
  ArchivedAt
```

### 6.3 Persistence

Use hybrid persistence:

```text
EventLog       # replay/audit-friendly fact stream
Snapshot       # fast restore of core.State
Materialized   # query/list/load views for sessions/runs/messages/tools
ArtifactStore  # binary payload storage
```

Default:

- Must-deliver events are persisted.
- Delta events can be `compact` or `audit`.
- Save a snapshot on `run.completed`.
- Store large files/images in `ArtifactStore`; keep references in `MessagePart`.

---

