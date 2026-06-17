## 3. API Contracts

### 3.1 MessagePart

Use a tagged struct with pointer payloads, not `interface{}`:

```text
PartType:
  text
  thinking
  image
  file
  tool_call
  tool_result
```

`MessagePart` expresses content. Lifecycle belongs to `Event`.

Important payloads:

- `TextPart{Text, Phase}`
- `ThinkingPart{Text, Signature, Redacted}`
- `ImagePart{URL, Data, MIMEType, Detail}`
- `FilePart{Name, URL, Data, MIMEType}`
- `ToolCallPart{ID, Name, Arguments, ArgumentsText, Status, ProviderExecuted}`
- `ToolResultPart{ToolCallID, Name, Content, Error, Metadata}`

### 3.2 Message

`Message` is a persistable canonical content snapshot:

```text
Message:
  ID
  RunID
  TurnID
  ParentID
  Role
  Parts
  Model
  Usage
  FinishReason
  Error
  Metadata
  CreatedAt
  UpdatedAt
  CompletedAt
```

Roles:

```text
system
developer
user
assistant
tool
```

Keep `developer`; do not collapse it into `system`.

### 3.3 Event

Events are unified envelopes with typed payload pointers. Do not use `Payload any`.

Core event types:

```text
run.started
run.completed
message.started
message.delta
message.completed
thinking.started
thinking.delta
thinking.completed
tool.started
tool.args.delta
tool.args.completed
tool.result.delta
tool.completed
tool.failed
approval.requested
approval.resolved
error
```

Delivery:

```text
best_effort:
  *.delta

must_deliver:
  *.started
  *.completed
  *.failed
  approval.*
  error
  run.completed
```

Event rules:

- `started` creates a node.
- `delta` modifies an existing node.
- `completed` writes a final snapshot when possible.
- `failed` writes error state.
- `seq` is monotonic in a run/session stream.

### 3.4 RunRequest

`RunRequest` is one run's canonical input snapshot:

```text
RunRequest:
  ID
  SessionID
  ParentRunID
  Input []Message
  Model ModelRef
  Tools []ToolSpec
  Instructions []Instruction
  Options RunOptions
  Metadata
```

`Prompt string` is not core input. CLI and IM adapters convert text to a user `Message`.

`RunOptions` uses pointer scalars when unset vs explicit zero/false matters:

```text
Stream
MaxOutputTokens
Temperature *float64
TopP *float64
Thinking
ToolChoice
ParallelToolCalls *bool
ResponseFormat
Cache
TimeoutMS
```

### 3.5 RunResult

`RunResult` is the final run summary, not the event stream:

```text
RunResult:
  RunID
  TurnID
  Status
  Messages
  Output
  Usage
  Error
  FinishReason
  Metadata
  StartedAt
  CompletedAt
```

`Output` is the primary assistant message. `Messages` are produced by this run, not full session history.

### 3.6 Usage

`Usage` is canonical and provider-normalized:

```text
Usage:
  InputTokens
  OutputTokens
  TotalTokens
  CacheHitTokens
  CacheMissTokens
  CacheHitRate
  ReasoningTokens
```

Provider adapters map vendor-specific token accounting into these fields. Cache fields are optional when a provider cannot report them, but when present `CacheHitTokens + CacheMissTokens` should reconcile with input token accounting.

---

