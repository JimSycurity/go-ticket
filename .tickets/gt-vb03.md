---
id: gt-vb03
status: closed
deps: [gt-27pw]
links: []
created: 2026-05-28T18:38:06Z
type: task
priority: 1
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#1-project-foundation
parent: gt-ag7g
tags: [mvp, tests, fixtures, ci]
---
# Add MVP fixtures and cross-platform test harness

Add upstream-produced fixture tickets, hand-authored edge fixtures, Go test helpers, and minimal GitHub Actions CI for Windows, macOS, and Linux go test.

## Acceptance Criteria

fixtures cover normal upstream tickets and edge cases; go test exercises fixture loading; GitHub Actions runs go test on all MVP platforms.

