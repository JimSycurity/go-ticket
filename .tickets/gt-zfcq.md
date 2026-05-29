---
id: gt-zfcq
status: open
deps: []
links: []
created: 2026-05-29T14:59:33Z
type: task
priority: 2
assignee: Jim Sykora
external-ref: security-review:toctou-open-handle-hardening
parent: gt-z9qe
tags: [security, paths, io, hardening]
---
# Harden ticket file IO against TOCTOU races

Review and harden ticket, settings, and migration file IO where current Lstat-then-open checks leave a small local time-of-check/time-of-use window if another process can mutate the repository concurrently.

## Acceptance Criteria

Read paths validate opened file handles where practical; settings and migration source reads avoid symlink swaps between validation and open; platform-specific helpers document Unix and Windows behavior; tests cover symlink swap or equivalent race-resistant behavior where feasible; residual platform limits are documented.

