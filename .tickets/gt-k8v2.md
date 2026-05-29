---
id: gt-k8v2
status: closed
deps: []
links: []
created: 2026-05-29T01:29:01Z
type: task
priority: 1
assignee: Jim Sykora
external-ref: openspec/changes/feature-parity-release-readiness/tasks.md#1-compatibility-baseline
parent: gt-z9qe
tags: [parity, compatibility, fixtures, tests]
---
# Build feature parity compatibility baseline

Create the fixture and golden-output baseline for the feature parity phase so later command work can be judged against explicit upstream-compatible behavior instead of vibes.

## Acceptance Criteria

- Representative upstream-compatible `.tickets/` fixture repositories exist for parity-sensitive commands.
- Golden command-output tests cover the matrix entries that need exact or stable output behavior.
- The compatibility matrix is updated when fixture evidence changes command status or known differences.
- Remaining undecided command policies are linked to follow-up tickets rather than blocking fixture coverage.
