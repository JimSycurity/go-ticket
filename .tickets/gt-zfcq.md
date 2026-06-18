---
id: gt-zfcq
status: closed
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

## Notes

**2026-06-16T12:54:00Z**

Hardened regular-file reads by validating opened handles against checked paths, routed ticket/settings/migration reads through the helper, added symlink-swap coverage, documented the Unix/Windows identity behavior and residual post-open mutation limit, and passed go test, go vet, diff check, and OpenSpec spec validation.
