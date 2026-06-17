# Artiworks Design v1

> Status: v1 architecture baseline  
> Source: architecture discussion snapshot, 2026-06-17  
> Future docs site: this file is structured so VitePress can render included sections later.

Artiworks is a Go-native terminal AI agent runtime. The design is canonical-first: Artiworks defines its own run, event, message, tool, memory, model, configuration, security, and runtime contracts, then adapts external protocols around them.

## Reading Guide

This document is the entry point. Each section below is split into a small file under `docs/design/v1/` so the design can later move into VitePress without another restructuring pass.

## Sections

1. [Design Principles](./design/v1/01-design-principles.md)
2. [Package Layout](./design/v1/02-package-layout.md)
3. [API Contracts](./design/v1/03-api-contracts.md)
4. [Native API and OpenAI-compatible API](./design/v1/04-native-and-openai-compatible-api.md)
5. [Model and Provider Registry](./design/v1/05-model-and-provider-registry.md)
6. [State, Reducer, Session, Persistence](./design/v1/06-state-reducer-session-persistence.md)
7. [Memory Design](./design/v1/07-memory-design.md)
8. [Tool Design](./design/v1/08-tool-design.md)
9. [Middleware Pipeline](./design/v1/09-middleware-pipeline.md)
10. [Hooks Design](./design/v1/10-hooks-design.md)
11. [Prompt / Instruction Assembly](./design/v1/11-prompt-instruction-assembly.md)
12. [Token Economy / Cache-Aware Context](./design/v1/12-token-economy-cache-aware-context.md)
13. [Runtime / Harness Orchestration](./design/v1/13-runtime-harness-orchestration.md)
14. [Observability, Audit, Control Plane](./design/v1/14-observability-audit-control-plane.md)
15. [Security, Permissions, Approval, Secrets](./design/v1/15-security-permissions-approval-secrets.md)
16. [Config Design](./design/v1/16-config-design.md)
17. [Schema Stability and Versioning](./design/v1/17-schema-stability-versioning.md)
18. [Implementation Roadmap](./design/v1/18-implementation-roadmap.md)
19. [Open Decisions](./design/v1/19-open-decisions.md)

## VitePress Includes

The following include directives are intentionally kept in comments. GitHub renders the linked table above cleanly; VitePress can later render the full composed document by enabling these includes.

<!--@include: ./design/v1/01-design-principles.md-->

<!--@include: ./design/v1/02-package-layout.md-->

<!--@include: ./design/v1/03-api-contracts.md-->

<!--@include: ./design/v1/04-native-and-openai-compatible-api.md-->

<!--@include: ./design/v1/05-model-and-provider-registry.md-->

<!--@include: ./design/v1/06-state-reducer-session-persistence.md-->

<!--@include: ./design/v1/07-memory-design.md-->

<!--@include: ./design/v1/08-tool-design.md-->

<!--@include: ./design/v1/09-middleware-pipeline.md-->

<!--@include: ./design/v1/10-hooks-design.md-->

<!--@include: ./design/v1/11-prompt-instruction-assembly.md-->

<!--@include: ./design/v1/12-token-economy-cache-aware-context.md-->

<!--@include: ./design/v1/13-runtime-harness-orchestration.md-->

<!--@include: ./design/v1/14-observability-audit-control-plane.md-->

<!--@include: ./design/v1/15-security-permissions-approval-secrets.md-->

<!--@include: ./design/v1/16-config-design.md-->

<!--@include: ./design/v1/17-schema-stability-versioning.md-->

<!--@include: ./design/v1/18-implementation-roadmap.md-->

<!--@include: ./design/v1/19-open-decisions.md-->
