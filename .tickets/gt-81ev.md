---
id: gt-81ev
status: open
deps: [gt-hhwg, gt-mlpu]
links: []
created: 2026-05-28T18:38:28Z
type: task
priority: 1
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#2-ticket-storage-core
parent: gt-ag7g
tags: [mvp, storage, yaml, ids]
---
# Build ticket storage parser writer and ID core

After the path-security checkpoint, implement gtk init, compatible Markdown/YAML parsing, stable LF atomic writes, upstream-default ID generation, partial ID resolution, and tolerant read/strict write validation.

## Acceptance Criteria

gtk init is explicit; parser preserves unknown fields by value; writes are LF and same-directory atomic; IDs avoid collisions; ambiguous partial IDs fail safely.

