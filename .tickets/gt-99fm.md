---
id: gt-99fm
status: open
deps: []
links: []
created: 2026-05-29T14:58:34Z
type: task
priority: 2
assignee: Jim Sykora
external-ref: security-review:action-sha-pinning
parent: gt-z9qe
tags: [release, ci, security, hardening]
---
# Pin GitHub Actions by SHA for trusted releases

Harden the GitHub Actions release and CI workflows by pinning third-party actions to immutable commit SHAs before treating published archives as trusted release artifacts.

## Acceptance Criteria

CI and release workflows pin third-party actions by full commit SHA; comments or documentation preserve the human-readable action/version context; release docs mention the trusted-artifact assumption; validation confirms tag-triggered release still builds archives and SHA256SUMS.

