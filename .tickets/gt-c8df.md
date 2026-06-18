---
id: gt-c8df
status: open
deps: [gt-4b5d]
links: []
created: 2026-06-18T16:29:15Z
type: feature
priority: 1
assignee: Jim Sykora
external-ref: github:MrLesk/Backlog.md
parent: gt-6fc6
tags: [search, agents, cli, query]
---

# Add native ticket search for humans and agents

Add a native gtk search command for quickly finding tickets by title, body text, notes, ID, tags, status, assignee, type, parent, and external reference without depending on grep, ripgrep, jq, Bash, or platform-specific shell tooling.

## Acceptance Criteria

Search works cross-platform with stable human output and an optional machine-readable output mode; search defaults avoid leaking absolute paths; filters can narrow status/type/assignee/tag where useful; malformed tickets are reported without aborting broad search; docs and agent reference explain when to use search versus list/query; tests cover text search, metadata search, malformed tickets, and path-disclosure expectations; security review validates input handling and output disclosure.
