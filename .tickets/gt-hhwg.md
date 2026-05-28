---
id: gt-hhwg
status: closed
deps: [gt-27pw]
links: []
created: 2026-05-28T18:38:18Z
type: task
priority: 0
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#2-ticket-storage-core
parent: gt-ag7g
tags: [mvp, paths, discovery]
---
# Implement ticket-root discovery up to security checkpoint

Implement conservative read-only ticket-root discovery up to the focused path-security checkpoint. The cyber/security review for TICKETS_DIR, symlinked roots, external roots, and canonical write targets is tracked separately in gt-mlpu.

## Acceptance Criteria

ancestor .tickets discovery works; cwd=.tickets works; absolute TICKETS_DIR override resolves to an existing directory; invalid TICKETS_DIR fails closed; symlinked ancestor .tickets is rejected pending the path-security review.
