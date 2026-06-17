## 13. Runtime / Harness Orchestration

`Harness.Run` is the canonical runtime entry:

```text
1. validate RunRequest
2. resolve model alias -> ModelRef
3. resolve provider + capabilities
4. create/load session
5. build MiddlewareContext
6. before_run middleware
7. retrieve memory
8. list/filter tools
9. assemble PromptPlan
10. apply token budget and cache plan
11. emit run.started
12. provider stream
13. reducer apply events
14. persist events/state/messages
15. execute tools if needed
16. loop back to provider if tool results exist
17. emit run.completed
18. after_run hooks/memory extraction
19. return RunResult
```

Agent loop:

```text
model step
  -> assistant message
  -> tool calls?
      yes -> execute tools -> append tool results -> next model step
      no  -> finish run
```

Hard limits:

```yaml
harness:
  max_steps: 12
  max_tool_calls: 32
  timeout: 10m
```

Concurrency rule:

```text
concurrent producers -> sequencer -> reducer.Apply -> persistence -> sinks
```

Reducer application is serial.

---

