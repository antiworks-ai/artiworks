## 11. Prompt / Instruction Assembly

Assembly is provider-independent planning:

```text
RunRequest
 + Session history
 + Memory hits
 + Tool specs
 + Model capabilities
 + Runtime config
 -> PromptPlan
 -> ProviderAdapter
```

`ProviderAdapter` translates `PromptPlan` to vendor payloads. It does not decide memory injection, history trimming, tool filtering, or capability downgrade.

`PromptPlan` carries both content and prompt-economy decisions:

```text
PromptPlan:
  StablePrefix []Instruction
  VolatileTail []Message
  Tools []ToolSpec
  CachePlan
  Warnings []AssemblyWarning
```

Instruction order:

```text
1. system instructions
2. developer instructions
3. policy/safety instructions
4. memory/context instructions
5. tool-use instructions
6. session/history messages
7. current user messages
```

Assembly records warnings:

```text
thinking_disabled
image_dropped
tool_streaming_disabled
history_truncated
memory_truncated
tool_output_cleaned
tool_output_truncated
artifact_referenced_only
context_budget_exceeded
structured_output_downgraded
tools_unsupported
```

---

