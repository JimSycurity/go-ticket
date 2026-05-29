---
id: gt-7aan
status: closed
deps: [gt-gs0r, gt-wv7t]
links: []
created: 2026-05-28T18:39:07Z
type: task
priority: 2
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#3-mvp-commands
parent: gt-ag7g
tags: [mvp, notes, tests, commands]
---
# Implement notes unsupported commands and command tests

Implement add-note using argument or stdin with fixture-driven timestamp format, unsupported-command errors without plugin execution, and command-output tests for core workflows.

## Acceptance Criteria

add-note appends safely; unknown commands never execute plugins; command-output tests cover core workflows in temporary repositories.

