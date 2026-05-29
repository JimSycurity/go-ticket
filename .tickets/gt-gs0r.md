---
id: gt-gs0r
status: closed
deps: [gt-81ev]
links: []
created: 2026-05-28T18:38:42Z
type: task
priority: 1
assignee: Jim Sykora
external-ref: openspec/changes/core-cross-platform-port/tasks.md#3-mvp-commands
parent: gt-ag7g
tags: [mvp, commands, json, vscode]
---
# Implement create show list and JSON inventory

Implement gtk create, show, list, ls, filters, and list --json with compact raw fields plus relative and absolute paths and no Markdown bodies.

## Acceptance Criteria

create fails safely without .tickets; show prints raw Markdown; list includes all tickets by default; list --json is compact and tool-friendly.

