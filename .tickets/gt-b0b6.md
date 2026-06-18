---
id: gt-b0b6
status: open
deps: [gt-4b5d]
links: []
created: 2026-06-18T16:29:31Z
type: feature
priority: 1
assignee: Jim Sykora
external-ref: github:MrLesk/Backlog.md
parent: gt-6fc6
tags: [config, agents, dod, workflow]
---

# Expand project settings for safe agent workflow defaults and DoD

Extend .tickets/settings.json beyond prefix-only configuration to support safe project defaults for agent workflows, definition-of-done prompts/defaults, ticket templates, default type/priority/assignee/tag behavior, and documentation/reference generation without allowing repository-controlled process execution.

## Acceptance Criteria

A reviewed settings schema covers safe defaults and definition-of-done fields; unknown or unsafe keys still fail closed or warn loudly according to the security review; settings cannot configure editors, hooks, PATH behavior, external roots, or shell/process execution; create/help/reference paths surface configured DoD/defaults where appropriate; docs explain repo-controlled versus user-controlled configuration boundaries; tests cover valid settings, unknown keys, invalid schema, DoD/default application, and backward compatibility with prefix-only settings; post-implementation security review validates the expanded config boundary.
