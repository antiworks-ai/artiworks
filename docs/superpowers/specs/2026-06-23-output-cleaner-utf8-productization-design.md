# Output Cleaner UTF-8 Productization Design

> Status: productization slice for `feature/runtime-skeleton`.

## Goal

Turn the `head_tail` output-cleaning strategy from a byte-slicing MVP into a
UTF-8-safe productized cleaner so model-facing tool output never returns invalid
text when the source content contains multibyte runes.

## Scope

This slice spans:

- `pkg/artiworks/harness` output-cleaner trimming helpers and tests;
- token-economy docs and roadmap evidence.

It keeps:

- the existing byte-budget policy shape;
- the existing warning codes and metadata fields;
- the current cleaning strategies.

It adds:

- rune-boundary-aware head/tail trimming;
- regression tests with multibyte text.

It does not add:

- new cleaning strategies;
- tokenizer integration;
- semantic rewriting or summarization;
- changes to the reference-only strategy.

## Behavior

When `head_tail` cleaning truncates output, the cleaner still respects the byte
budget configured by `MaxBytes`, `HeadBytes`, and `TailBytes`, but it clamps the
actual slice boundaries to valid UTF-8 rune starts so the returned text stays
valid.

The cleaner still records the same metadata and warnings.

## Acceptance Criteria

- `go test ./pkg/artiworks/harness -run TestOutputCleanerHeadTailPreservesUTF8 -count=1` passes;
- `go test ./pkg/artiworks/harness -count=1` passes;
- `go vet ./pkg/artiworks/harness` passes;
- `git diff --check` passes;
- GitNexus change detection stays confined to the output-cleaner path and the
  already-dirty productization worktree.
