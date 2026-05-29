## Context

`wedow/ticket` stores tickets as Markdown files in `.tickets/`, with YAML frontmatter carrying IDs, statuses, dependencies, links, parent references, priority, assignee, tags, and external references. The upstream CLI is small and productive, but it relies on POSIX shell behavior, coreutils, and optional Unix-oriented tools. The MVP is a compatible Go implementation focused on native Windows viability without changing the file format.

## Goals / Non-Goals

**Goals:**
- Produce a `gtk` CLI implemented in Go that runs natively on Windows, macOS, and Linux.
- Preserve ticket file compatibility with upstream `wedow/ticket` for the MVP command set.
- Keep the storage model simple: `.tickets/<id>.md` remains the source of truth.
- Build a testable core package for parsing, writing, filtering, and graph analysis.
- Establish enough cross-platform CI to catch path, newline, and shell-assumption regressions.
- Add minimal GitHub Actions CI for Windows, macOS, and Linux `go test`.

**Non-Goals:**
- Full upstream feature parity in the MVP.
- Drop-in replacement of the `tk` executable name during MVP.
- Replacing upstream `wedow/ticket` or requiring a migration for existing `.tickets/` users.
- `closed --limit=N`; closed-ticket history belongs in the feature-parity phase.
- `dep tree` and `dep cycle`; dependency graph reporting belongs in the feature-parity phase.
- Shell completions; command completion belongs in release-readiness.
- Database-backed indexing, daemon behavior, or non-file storage.
- UI/IDE features; the VS Code extension can continue using the ticket file format and CLI behavior.

## Decisions

- Implement as Go with a thin `cmd/gtk` entrypoint and reusable internal packages.
  - Rationale: Go produces straightforward static binaries for Windows and Unix-like systems, keeps startup fast, and avoids requiring Python, Bash, or PowerShell on target machines.
  - Alternative considered: Python would be faster to draft but would add packaging friction for `tk.exe`.

- Use a small command dispatcher with Go's standard `flag` package for MVP.
  - Rationale: The MVP command surface is small, this keeps dependencies low, and the decision is reversible if completions or richer CLI UX later justify Cobra or another framework.
  - Alternative considered: Adopt Cobra or urfave/cli immediately; rejected because those frameworks solve problems the MVP does not yet have.

- Preserve `.tickets/` Markdown/YAML as the compatibility contract.
  - Rationale: Existing repos should continue to work without migration, and agents can keep reading/editing Markdown directly.
  - Alternative considered: A normalized JSON or SQLite index would speed queries but would violate the simplicity that made `ticket` attractive.

- Use structured parsing/writing for frontmatter rather than shell-like text substitutions.
  - Rationale: Native Windows support needs predictable handling of paths, newlines, lists, and quoting.
  - Alternative considered: Regex-only parsing would be smaller but fragile for tags, dependencies, and future fields.

- Preserve unknown frontmatter fields by value while allowing stable normalized YAML output.
  - Rationale: Existing repos may carry custom metadata, but exact byte-preservation would complicate the writer and make cross-platform mutation behavior brittle.
  - Alternative considered: Preserve frontmatter byte-for-byte; rejected for MVP because stable semantic preservation is the better cost/benefit point.

- Use same-directory atomic writes for ticket mutations.
  - Rationale: Writing a temporary file beside the ticket and renaming it over the original reduces corruption risk and keeps replacement on the same filesystem, including on Windows.
  - Alternative considered: Rewrite ticket files in place; rejected because interrupted writes can leave damaged tickets.

- Do not add file locking in MVP.
  - Rationale: Atomic writes cover the main corruption risk, while cross-platform locking adds complexity before concurrent mutation is proven to matter.
  - Alternative considered: Add lock files or OS-level locks immediately; rejected because the repo-backed workflow is usually single-actor and can revisit locking later with real requirements.

- Normalize written ticket files to LF newlines on all platforms.
  - Rationale: Tickets are git-backed Markdown files and LF avoids cross-platform diff churn while preserving compatibility with Unix agents and upstream `tk`.
  - Alternative considered: Use native platform newlines; rejected because Windows-created files would create unnecessary churn.

- Make `add-note` timestamp formatting fixture-driven.
  - Rationale: If upstream-produced fixtures show a clear note timestamp format, matching it improves compatibility; otherwise a documented stable UTC ISO-8601 format is enough for MVP.
  - Alternative considered: Pick a new format immediately; rejected because fixture evidence is cheap to gather.

- Keep built-in MVP commands editor-less.
  - Rationale: `edit`, plugin overrides, and shell-sensitive command discovery are higher-risk compatibility work and belong in the parity change.
  - Alternative considered: Implement plugin behavior immediately; rejected to keep MVP scope tight.

- Do not execute plugins in the MVP.
  - Rationale: Plugin behavior requires PATH lookup and arbitrary process execution, which needs a dedicated compatibility and security pass.
  - Alternative considered: Implement upstream-style plugin dispatch immediately; rejected because it increases risk without helping the core Windows file-format proof.

- Include `gtk list` and `gtk ls` as MVP convenience commands.
  - Rationale: A general inventory command is useful for humans, agents, and vscode-tk integration even if it is not strict upstream parity in the currently installed `tk` help.
  - Alternative considered: Only implement upstream-documented commands; rejected because the convenience is low-risk and clearly distinguishable as go-ticket MVP behavior.

- Provide MVP-specific help text.
  - Rationale: `gtk` is side-by-side and not a full parity binary yet, so help should only describe supported MVP commands.
  - Alternative considered: Mirror upstream help with unsupported commands; rejected because it would mislead users.

- Make `gtk list` a full inventory by default.
  - Rationale: `ready` and `blocked` are work-selection commands; `list` should support audits, browsing, and vscode-tk-style inventory without hiding closed tickets unless filters request it.
  - Alternative considered: Show only open and in-progress tickets by default; rejected because it would make list less useful as a general inventory.

- Preserve and display parent metadata without parent/child traversal commands in MVP.
  - Rationale: Parent fields are part of the file format and useful to vscode-tk, but hierarchy traversal commands are not required to prove the cross-platform core.
  - Alternative considered: Add parent/child traversal commands immediately; rejected to keep MVP focused.

- Make `gtk show <id>` print the raw ticket Markdown file.
  - Rationale: Raw output is simplest, preserves human formatting, and avoids inventing a normalized display that might hide compatibility issues.
  - Alternative considered: Render a parsed normalized view; rejected for MVP because list output already provides normalized summary.

- Add `gtk list --json` as the MVP machine-readable integration surface.
  - Rationale: vscode-tk and agents should not need to scrape human-oriented list output, but full `query`/jq compatibility is post-MVP.
  - Alternative considered: Defer all machine-readable output; rejected because JSON list output is low-risk and useful immediately.

- Keep `gtk list --json` compact and body-free.
  - Rationale: Tools can use the emitted file path or `gtk show` when they need full Markdown; summary JSON should remain fast and predictable.
  - Alternative considered: Include raw Markdown bodies in list JSON; rejected because it can create large, noisy output.

- Include both relative and absolute paths in `gtk list --json`.
  - Rationale: Relative paths are stable for git-backed scripts, while absolute paths are useful for local editor integrations such as vscode-tk.
  - Alternative considered: Emit only one path form; rejected because consumers have different needs.

- Keep `gtk list --json` to raw ticket fields and paths in MVP.
  - Rationale: Derived graph fields such as children, blocks, and blocked_by expand the API contract; consumers can derive them from `deps`, `links`, and `parent` for now.
  - Alternative considered: Emit derived graph fields immediately; rejected to keep MVP JSON stable and compact.

- Skip malformed ticket files with warnings during broad reads, but fail targeted mutations on malformed targets.
  - Rationale: One bad ticket should not make list/ready/blocked unusable, while mutation must not overwrite files that could not be parsed safely.
  - Alternative considered: Reject the whole ticket directory when any file is malformed; rejected because it is too brittle for real repositories.

- Default ticket ID generation to upstream-compatible behavior when no settings exist.
  - Rationale: MVP should not require configuration, and existing behavior should remain familiar. A future lightweight `.tickets/` settings file can configure prefixes and other behavior after the core is stable.
  - Alternative considered: Require or introduce a settings file in MVP; rejected because it adds migration and compatibility surface too early.

- Do not implement settings parsing in MVP.
  - Rationale: The future settings contract is captured in the parity change, but the implementation should stay clean until the settings shape is designed.
  - Alternative considered: Add an unused parser scaffold; rejected because unused configuration code invites premature coupling.

- Treat missing or malformed dependency targets as blockers.
  - Rationale: A typo or stale dependency reference should not make work appear ready.
  - Alternative considered: Ignore unresolved dependency references; rejected because it hides data-quality problems.

- Treat status values as strict on write and tolerant on read.
  - Rationale: `gtk` should not create unsupported statuses in MVP, but existing repositories may contain unknown statuses that still need to be listed and inspected.
  - Alternative considered: Reject unknown statuses during reads; rejected because it would make broad operations too brittle.

- Treat ticket type values as strict on write and tolerant on read.
  - Rationale: MVP-created tickets should use the known type vocabulary `bug`, `feature`, `task`, `epic`, and `chore`, while existing repos with custom types should remain readable.
  - Alternative considered: Allow arbitrary types on write; rejected because it makes typos easy and weakens filter behavior.

- Treat priority values as strict on write and tolerant on read.
  - Rationale: MVP-created tickets should use the known `0` through `4` priority range, while existing repos with unusual priorities should remain readable.
  - Alternative considered: Allow arbitrary priority values on write; rejected because it weakens sorting/filtering consistency.

- Include basic `wedow/ticket` attribution in MVP README docs.
  - Rationale: Attribution is low-cost, clarifies the compatibility target, and aligns with the project intent from the start.
  - Alternative considered: Defer all attribution docs to release-readiness; rejected because readers should know the relationship immediately.

- Treat upstream behavior as a compatibility target, not copied implementation.
  - Rationale: The repo can stay MIT-licensed with attribution while avoiding a line-by-line shell port.

- Name the MVP binary `gtk` for side-by-side testing.
  - Rationale: Users can keep upstream `tk` installed while testing the Go port against the same `.tickets/` directory.
  - Alternative considered: Ship the MVP as `tk`; rejected because it implies drop-in readiness before parity has been proven.

- Require explicit `gtk init` before writes in a repo without `.tickets/`.
  - Rationale: Explicit initialization avoids accidentally creating ticket roots in the wrong directory, especially when agents run commands from nested or temporary paths.
  - Alternative considered: Auto-create `.tickets/` from `gtk create`; rejected because the convenience is not worth the ambiguity.

- Resolve `.tickets/` itself as a valid current working directory.
  - Rationale: Users and agents sometimes operate from inside the ticket directory; discovery should still identify the active ticket directory and its parent project root.
  - Alternative considered: Only search for child `.tickets/` directories from cwd; rejected because it breaks a common filesystem edge case.

- Treat ticket-root path handling as security-sensitive.
  - Rationale: `gtk` mutates files, so environment overrides, symlinks, relative paths, and external roots can redirect writes in surprising ways.
  - Decision: MVP discovery uses conservative ancestor `.tickets/` lookup plus canonicalized explicit overrides only. `TICKETS_DIR` must resolve to an existing directory before use. Symlink and external-root behavior require a focused security review before implementation.
  - Alternative considered: Freely accept relative `TICKETS_DIR` paths and symlinked directories; rejected until path-security implications are reviewed.

- Use a conservative MVP path-security policy before enabling write commands.
  - Rationale: Discovery and ticket path resolution are shared by every future mutation command, so the safe boundary should be enforced once rather than reimplemented per command.
  - Decision: Ancestor-discovered `.tickets/` entries must be real directories; a `.tickets` path that exists but is not a directory is an error rather than a reason to continue walking upward. Explicit `TICKETS_DIR` must be absolute, existing, and non-symlinked; it may point outside the current working tree because it is an explicit environment override and the resolved root is marked as environment-selected. Ticket writes must resolve IDs through one containment-checked path helper that rejects path separators, traversal, drive-letter syntax, Windows reserved device names, symlinked ticket files, and non-regular existing targets.
  - Alternative considered: Permit symlinked ticket roots and rely on canonicalization alone; rejected because it can hide where future writes will land.

- Use both upstream-produced fixtures and hand-authored edge fixtures.
  - Rationale: Real files produced by upstream `tk` provide compatibility confidence, while hand-authored fixtures cover malformed, ambiguous, CRLF, and Windows-specific edge cases that are awkward to generate from upstream commands.
  - Alternative considered: Only write synthetic fixtures; rejected because it could drift from actual upstream output.

- Include minimal cross-platform CI in MVP.
  - Rationale: Windows viability is the core project reason, so Windows/macOS/Linux `go test` coverage must exist before the MVP is considered usable.
  - Alternative considered: Defer CI to release-readiness; rejected because local-only testing would not prove native Windows behavior.

## Risks / Trade-offs

- Upstream output compatibility may be underspecified -> capture golden fixtures and command-output tests for the MVP commands.
- YAML round-tripping may reorder or normalize fields -> define stable field order for writes and preserve unknown fields where practical.
- Partial ID matching can become surprising on case-insensitive Windows filesystems -> match by normalized ticket ID and fail on ambiguity.
- Go dependencies can overcomplicate a small tool -> keep dependencies limited to CLI parsing, YAML, and optional test helpers.
- Native Windows path handling may diverge from POSIX examples -> add Windows CI early and avoid shell-dependent tests.

## Migration Plan

The MVP does not require migration. Users can point `go-ticket` at an existing `.tickets/` directory, run read-only commands, then test write commands on a branch. Rollback is deleting the Go binary and continuing with upstream `tk`; ticket files remain Markdown.

## Open Questions

- Should the binary be named only `tk`, or should releases also include `go-ticket` as an alias?
  - Resolved for MVP: the executable SHALL be `gtk` for side-by-side testing. A `tk` alias/drop-in binary and optional `go-ticket` alias belong in the parity/release-readiness phase.
- Should MVP command output be byte-for-byte compatible where possible, or only semantically compatible?
  - byte for byte compatible in areas where it may impact vscode-tk compatibility.  Otherwise semantically compatible.
- Should `query` be omitted from MVP entirely, or included as JSON output without jq filtering?
  - I didn't even know tk had a query option, so ommit it.
