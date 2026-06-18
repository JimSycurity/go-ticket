---
id: gt-d10e
status: open
deps: [gt-4b5d, gt-6d36, gt-c8df, gt-422a, gt-b0b6]
links: []
created: 2026-06-18T16:30:02Z
type: feature
priority: 3
assignee: Jim Sykora
external-ref: github:MrLesk/Backlog.md
parent: gt-6fc6
tags: [tui, kanban, agents, cli, exploration]
---

# Explore optional TUI views for terminal-first workflows

Explore whether go-ticket should add an optional terminal UI for agentic coding fans, including kanban-style status views, ready/blocked dashboards, ticket search/browse, and general ticket workflow navigation. Keep web UI out of go-ticket scope and leave richer editor visuals to vscode-tk.

## Acceptance Criteria

Exploration documents target users, non-goals, and whether TUI belongs in go-ticket core, an optional subcommand, or a separate companion; any prototype is read-only first and built on stable list/search/query contracts; kanban/editing behavior is compared with vscode-tk responsibilities; accessibility, terminal portability, and dependency impact are reviewed; no TUI implementation starts until CLI reference, search/query, and config foundations are stable or explicitly accepted as dependencies; security review covers terminal escape/output handling and any editor/process integration.
