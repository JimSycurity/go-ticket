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
| `dep tree [--full] <id>` | Supported | Semantic with upstream-style readable tree output | Repeated dependencies are omitted by default and shown with `--full`; missing dependency targets are omitted from tree output. |
| `dep cycle` | Supported | Semantic with stable readable output | Detects cycles among open/in-progress tickets and ignores cycles broken by closed tickets. |
| `undep <id> <dep-id>` | Supported | Semantic | Removes a dependency. |
| `link <id> <id> [id...]` | Partial | Semantic | MVP links two tickets; upstream supports multiple targets. |
| `unlink <id> <target-id>` | Supported | Semantic | Removes symmetric link. |
| `ls` / `list` | Supported | Semantic; stable shape for human output | MVP also includes closed tickets as an intentional convenience. |
| `ready` | Supported | Semantic | Lists open/in-progress tickets with dependencies resolved. |
| `blocked` | Supported | Semantic | Lists open/in-progress tickets with unresolved dependencies. |
| `closed [--limit=N]` | Supported | Semantic plus mtime ordering policy | Lists closed tickets by descending file mtime with default limit 20 and supports `-a`/`-T` filters. |
| `show <id>` | Supported | Byte-for-byte raw file output | Prints raw ticket Markdown content. |
| `add-note <id> [text]` | Supported | Semantic; timestamp format documented by implementation | Accepts argument text or stdin. |
| `super <cmd> [args]` | Supported | Semantic with reviewed plugin policy | Dispatches builtins only and never runs plugins. |

## Bundled Plugin Commands

| Upstream command | Current status | Compatibility target | Notes |
| --- | --- | --- | --- |
| `edit <id>` | Supported | Semantic with reviewed editor launch policy | Uses a validated user-configured editor command, passes the ticket path as one argv element, and rejects inline shell-style editor arguments. |
| plugin `ls` / `list` | Deferred | To be decided | Core `gtk list` already covers the MVP list workflow. |
| `query [jq-filter]` | Partial | Native JSONL output without filtering | `gtk query` emits one compact JSON object per ticket using ticket frontmatter fields and no path fields. jq-style filters are intentionally deferred and return a clear unsupported-filter error. This is not full query feature parity; filtered query support remains future parity work. |
| `migrate-beads` | Supported | Good-enough import with explicit review report and conflict handling | Reads `.beads/issues.jsonl` under the project root, bounds input size, skips existing ticket IDs, reports malformed/partial records for review, and writes only through the ticket writer. |

## Plugin Surface

Upstream discovers executables named `tk-<cmd>` or `ticket-<cmd>` on `PATH`.
`go-ticket` implements a narrowed plugin dispatch policy for unknown commands.
The policy favors explicit review over full upstream shell parity.

Current plugin/editor execution policy:

- Builtins win by default; `super` dispatches builtins and never runs plugins.
- Plugin command names must be simple command atoms with no path separators,
  dots, drive letters, whitespace, or shell metacharacters.
- Plugin lookup scans absolute PATH directories in order, ignoring empty,
  relative, and current-directory entries.
- Candidate order is `tk-<cmd>` before `ticket-<cmd>`.
- Unix plugin candidates must be executable regular non-symlink files.
- Windows plugin execution initially allows `.exe` only; `.cmd`, `.bat`, and
  `.ps1` remain deferred future parity until wrapper behavior is reviewed and
  tested.
- Plugins receive only minimal ticket/project environment variables.
- Process execution must use argv directly and never implicit shell
  interpolation.
- `edit` must use a validated editor command and pass the ticket path as one
  argv element. Repository-controlled `.tickets` settings must not configure
  editors.

## Settings Surface

`go-ticket` supports optional `.tickets/settings.json` with a prefix-only
settings model. Unknown keys fail closed. Settings cannot configure plugins,
editors, external roots, PATH behavior, or process execution policy.

## Release Readiness Notes

The MVP CI already runs formatting, tests, and a build check on Linux, macOS,
and Windows. Release-readiness CI may still add vet/lint, artifact builds,
checksums, and tag-triggered release publishing.
