# Compatibility Matrix

This matrix tracks `go-ticket` compatibility with the documented `wedow/ticket`
command surface. The upstream reference is the `wedow/ticket` README usage
section at https://github.com/wedow/ticket.

Compatibility target:

- Prefer semantic compatibility for command behavior, file format, status
  changes, relationship updates, and machine-readable fields.
- Require byte-for-byte compatibility only where downstream tooling is likely to
  depend on exact output, such as JSON shape, ticket file frontmatter, and
  command examples promoted as scripting contracts.
- Capture known output differences in this document before marking a command
  fully supported.

Status legend:

- Supported: implemented and covered by tests for expected behavior.
- Partial: implemented for common use but missing upstream options, exact
  output, or edge-case coverage.
- Deferred: intentionally not implemented yet.
- Security-gated: not considered for implementation until security policy and
  use case review are complete.

## Core Commands

| Upstream command | Current status | Compatibility target | Notes |
| --- | --- | --- | --- |
| `create [title] [options]` | Supported | Semantic plus compatible ticket file output | MVP supports description, design, acceptance, type, priority, assignee, external-ref, parent, and tags. |
| `start <id>` | Supported | Semantic | Sets status to `in_progress`. |
| `close <id>` | Supported | Semantic | Sets status to `closed`. |
| `reopen <id>` | Supported | Semantic | Sets status to `open`. |
| `status <id> <status>` | Supported | Semantic | MVP accepts `open`, `in_progress`, and `closed`. |
| `dep <id> <dep-id>` | Supported | Semantic | Adds a dependency once. |
| `dep tree [--full] <id>` | Deferred | Human-readable output should be documented before parity claim | Planned under feature parity. |
| `dep cycle` | Deferred | Human-readable output should be documented before parity claim | Planned under feature parity. |
| `undep <id> <dep-id>` | Supported | Semantic | Removes a dependency. |
| `link <id> <id> [id...]` | Partial | Semantic | MVP links two tickets; upstream supports multiple targets. |
| `unlink <id> <target-id>` | Supported | Semantic | Removes symmetric link. |
| `ls` / `list` | Supported | Semantic; stable shape for human output | MVP also includes closed tickets as an intentional convenience. |
| `ready` | Supported | Semantic | Lists open/in-progress tickets with dependencies resolved. |
| `blocked` | Supported | Semantic | Lists open/in-progress tickets with unresolved dependencies. |
| `closed [--limit=N]` | Deferred | Semantic plus ordering policy | Upstream orders by mtime with default limit 20. |
| `show <id>` | Supported | Byte-for-byte raw file output | Prints raw ticket Markdown content. |
| `add-note <id> [text]` | Supported | Semantic; timestamp format documented by implementation | Accepts argument text or stdin. |
| `super <cmd> [args]` | Security-gated | Semantic after plugin policy exists | Only useful once plugin dispatch exists. |

## Bundled Plugin Commands

| Upstream command | Current status | Compatibility target | Notes |
| --- | --- | --- | --- |
| `edit <id>` | Security-gated | Semantic after editor launch policy exists | Requires editor command resolution and argument handling review. |
| plugin `ls` / `list` | Deferred | To be decided | Core `gtk list` already covers the MVP list workflow. |
| `query [jq-filter]` | Deferred | JSON shape should be stable; filter strategy undecided | Options remain embedded jq-like evaluator, external `jq`, or documented unsupported filters. |
| `migrate-beads` | Deferred | Good-enough import with explicit review report and rollback docs | Does not need perfect migration fidelity for the first parity pass. |

## Plugin Surface

Upstream discovers executables named `tk-<cmd>` or `ticket-<cmd>` on `PATH`.
`go-ticket` will not implement plugin execution until there is a dedicated
security review and a clearer use case for spending implementation effort on
that surface.

Security policy must cover at least:

- PATH lookup order.
- Allowed executable/script extensions by platform.
- PowerShell script handling.
- Environment variables passed to plugins.
- Editor argument handling.
- No implicit shell interpolation.

## Release Readiness Notes

The MVP CI already runs formatting, tests, and a build check on Linux, macOS,
and Windows. Release-readiness CI may still add vet/lint, artifact builds,
checksums, and tag-triggered release publishing.
