---
id: gt-mlpu
status: closed
deps: [gt-hhwg]
links: []
created: 2026-05-28T18:46:03Z
type: task
priority: 0
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#2-ticket-storage-core
parent: gt-ag7g
tags: [mvp, security, paths, hitl, cyber-review]
---
# Review ticket-root path security before writes

Perform the focused path-security review for TICKETS_DIR, relative overrides, symlinked ticket roots, external roots, and write-target canonicalization before write-capable gtk commands are enabled.

## Acceptance Criteria

Security-reviewed policy is recorded; write commands have clear allowed and rejected ticket-root cases; storage and mutation tickets can proceed with that policy.

