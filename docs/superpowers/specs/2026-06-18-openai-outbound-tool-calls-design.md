# OpenAI Outbound Tool Calls Design

## Goal

Teach the OpenAI/OpenAI-compatible outbound adapter to map provider tool-call wire formats to Artiworks canonical `api.ToolCall` records, and to map canonical assistant tool-call/tool-result messages back into provider request payloads.

## Scope

Included:

- Chat Completions request mapping:
  - Assistant canonical `tool_call` parts become `message.tool_calls`.
  - Canonical `role=tool` results become `role=tool` messages with `tool_call_id`.
- Chat Completions response mapping:
  - `choices[0].message.tool_calls` becomes canonical `api.ToolCall` values.
  - `finish_reason=tool_calls` becomes `api.FinishReasonToolCalls`.
- Responses request mapping:
  - Assistant canonical `tool_call` parts become `type=function_call` input items.
  - Canonical `role=tool` results become `type=function_call_output` input items.
- Responses response mapping:
  - `output` items with `type=function_call` become canonical `api.ToolCall` values.
- Argument JSON is decoded into `api.JSONObject` when valid and always preserves `ArgumentsText`.
- Non-streaming only.

Excluded:

- Streaming tool-call deltas.
- Provider-specific parallel execution behavior.
- Approval or tool execution logic, which already belongs to runtime wiring.
- Persistent `previous_response_id` threading for Responses API.

## Notes

This follows the provider boundary principle: adapters translate provider payloads into canonical DTOs; the runtime loop only consumes canonical data.

Official OpenAI reference used for wire-shape alignment: <https://platform.openai.com/docs/guides/function-calling>.
