---
id: gt-wv7t
status: open
deps: [gt-81ev]
links: []
created: 2026-05-28T18:38:54Z
type: task
priority: 1
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#3-mvp-commands
parent: gt-ag7g
tags: [mvp, commands, deps, relationships]
---
# Implement lifecycle relationships ready and blocked

Implement status lifecycle commands, dep/undep, link/unlink, parent metadata preservation/display, ready, and blocked with conservative unresolved dependency handling.

## Acceptance Criteria

status writes only valid MVP statuses; unknown read statuses are surfaced; missing malformed or unknown-status dependencies block readiness; relationship mutations are compatible.

