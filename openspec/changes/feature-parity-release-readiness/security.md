# Security Review: Proposed Feature Parity

Date: 2026-05-29

## Scope

This review covers the proposed feature-parity and release-readiness tasks for
`go-ticket`, especially commands that execute processes, launch editors, ingest
external issue data, read repository-controlled settings, write ticket files, or
publish release artifacts.

Current safety primitives are favorable: ticket discovery rejects symlinked
`.tickets` roots, explicit `TICKETS_DIR` must be absolute and non-symlinked,
ticket writes flow through `ticket.Write`, and `ResolveTicketPath` rejects
unsafe IDs, traversal, symlink ticket files, non-regular targets, and
containment escapes. New parity features must preserve those invariants.

## High: Plugin And Super Execution

Relevant tasks: 2.3a, 2.4, 3.1, 5.2, 5.5.

Plugin parity is arbitrary process execution. A PATH-based `tk-<cmd>` or
`ticket-<cmd>` lookup can turn an unknown command, typo, or malicious PATH entry
into code execution. `super` also matters because it defines the boundary
between builtins and plugins.

Required controls before implementation:

- Define deterministic lookup order and document whether `tk-<cmd>` or
  `ticket-<cmd>` wins.
- Do not search the current directory implicitly.
- Builtins must win by default; `super` must bypass plugins and dispatch only
  builtins.
- Use argv-based process execution only; no shell interpolation.
- Maintain a platform-specific extension allowlist.
- Deny `.ps1` initially, or require explicit opt-in after PowerShell policy is
  reviewed.
- Pass a minimal environment to plugins, limited to ticket/project metadata
  needed for compatibility.
- Add tests for PATH shadowing, builtin shadowing, extension allowlist behavior,
  malicious command names, and no-shell argument preservation.

Policy for implementation:

- Builtin commands win by default. A plugin can run only for a command name that
  is not a builtin.
- `gtk super <cmd> [args...]` dispatches only builtin commands and never runs a
  plugin.
- Plugin command names must be simple command atoms: no path separators, drive
  letters, dots, whitespace, shell metacharacters, or empty names.
- Lookup scans absolute PATH directories in order. Empty, relative, and current
  directory PATH entries are ignored.
- Candidate basename order is `tk-<cmd>` first, then `ticket-<cmd>`.
- Unix candidates must be regular, non-symlink executable files with the exact
  candidate basename.
- Windows candidates initially allow `.exe` only. `.cmd`, `.bat`, and `.ps1`
  remain deferred until their process-launch behavior has dedicated tests and a
  reviewed wrapper policy.
- Plugins receive only a minimal environment: `TICKETS_DIR`,
  `TICKET_PROJECT_DIR`, and the platform-required process environment.
- Plugin execution must use direct argv execution and must not invoke a shell.
- Plugin stdout/stderr are forwarded to the caller; non-zero exit status is
  returned as command failure.

## High: Editor Integration

Relevant tasks: 2.3a, 2.5, 3.1, 5.2.

`edit` is process execution, not just file opening. Risk comes from treating
`EDITOR` or `VISUAL` as shell strings, mishandling quoted args, accepting
repo-controlled editor config, or passing an unsafe path.

Required controls before implementation:

- Resolve the ticket through existing ticket lookup and path checks first.
- Pass the final ticket file path as one argv element.
- Avoid shell parsing and shell fallback.
- Prefer explicit config shape such as `editor.command` plus `editor.args`.
- Do not read editor command configuration from repo-controlled `.tickets`
  settings until separately reviewed.
- Return clear errors when no editor is configured.
- Add tests for spaces in paths, quoted args, missing editor, unsafe editor
  config, and no-shell behavior.

Policy for implementation:

- Editor launch is disabled unless an explicit editor command is configured by
  user environment or user-level configuration. Repository-controlled
  `.tickets` settings must not configure editors.
- `GTK_EDITOR` may name an editor executable path or command name. Values with
  whitespace, shell metacharacters, or inline arguments are rejected.
- `VISUAL` and `EDITOR` may be considered only if they are single command atoms
  under the same validation rules as `GTK_EDITOR`.
- Editor arguments, if supported, must come from a separate argv-shaped setting
  and must not be shell parsed.
- The resolved ticket file path is appended as one argv element after ticket
  path validation.
- No shell fallback is allowed.
- Missing or invalid editor configuration returns a clear error without
  modifying tickets.

## Medium: Beads Migration

Relevant tasks: 2.6, 3.1, 3.3, 3.4, 5.2.

`migrate-beads` ingests external JSONL and performs bulk writes. Treat Beads
data as untrusted input even when it comes from a local repository.

Required controls before implementation:

- Bound total input size and per-line size.
- Parse with a strict schema and collect per-record diagnostics.
- Validate every generated ticket ID and relationship ID.
- Write only through `ticket.Write` and `ResolveTicketPath`.
- Never overwrite existing tickets by default.
- Produce an import report that separates imported, skipped, conflicted, and
  manually reviewed records.
- Support dry-run/report-first behavior.
- Produce rollback guidance or a manifest of generated files.
- Add malformed, partial, duplicate, conflicting, oversized, and relationship
  edge-case fixtures.

## Medium: Optional Tickets Settings

Relevant tasks: 2.7, 3.1, 5.2, 5.5.

Settings under `.tickets/` are repository-controlled. Prefix settings are
reasonable, but this file must not quietly become a way to inject command paths,
plugin paths, editor config, external roots, or process policy toggles.

Required controls before implementation:

- Implement prefix-only settings first.
- Validate prefixes as ticket ID atoms, not free-form strings.
- Read settings only from a regular, non-symlink file under the active
  `.tickets` directory.
- Enforce a small file-size limit.
- Fail closed or warn loudly for unknown keys.
- Explicitly forbid process, path, plugin, and editor settings until a later
  security review.

## Medium: Write Surface Invariants

Relevant tasks: 2.6, 2.7, 3.1, 3.2, 3.3, 5.5.

The current write surface is intentionally centralized. Future commands must not
bypass it with direct path construction or ad hoc writes.

Required controls before implementation:

- All ticket file writes must use `ticket.Write` or a separately reviewed helper
  with equivalent containment checks.
- All ticket IDs must pass `ValidateID`.
- All target paths must pass `ResolveTicketPath`.
- Existing symlink and non-regular file rejection must remain covered by tests.
- Bulk operations must be atomic per file and report partial failures.
- Regression tests must cover traversal, path separators, Windows reserved
  device names, symlink ticket files, non-regular files, and external-root
  attempts.

## Medium: Filtered Query Strategy

Relevant tasks: 2.3, 3.1, 5.2.

The current `query` implementation intentionally provides JSONL output without
filters. This is partial query parity, not full feature parity. External `jq`
would reintroduce PATH process execution for a feature that is currently
deferred.

Required controls before filtered query parity:

- Keep filtered query documented as deferred until there is a concrete need.
- Prefer an embedded evaluator with bounded complexity over PATH-resolved `jq`.
- If external `jq` is ever supported, require explicit opt-in and reuse the
  plugin/process execution policy.
- Keep `gtk query` path-free unless a future option explicitly requests paths.

Task 2.3 revisit result:

- Current `gtk query` satisfies the security baseline by emitting JSONL without
  `abs_path`, `rel_path`, or `path` fields.
- Filter arguments return a clear deferred-feature error and do not execute
  external `jq`.
- Full filtered query remains future parity work.

## Task 3.1 Pre-Implementation Review Disposition

The pre-implementation review has been completed for the proposed parity
surface. The review does not approve unrestricted process execution. It defines
the gates above that must be satisfied by plugin execution, editor launching,
PATH lookup, Beads parsing, and ticket writes before individual implementation
tasks are considered complete.

## Low: JSON Path Disclosure

Relevant tasks: 3.1, 5.2.

`list --json` emits absolute paths for editor integrations. That is useful but
can leak local usernames and repository layout into logs or pasted reports.
`query` currently avoids absolute paths.

Required controls:

- Document `abs_path` as local-machine data.
- Keep `query` output path-free.
- Consider a future `--no-abs-path` or opt-in absolute path mode if leakage
  becomes an issue.

## Low: Release Pipeline Supply Chain

Relevant tasks: 4.1, 4.2, 4.3, 4.4, 5.2.

CI currently uses least-privilege read permissions for tests. Release workflows
will need additional permissions and should not broaden trust unnecessarily.

Required controls before trusted public artifacts:

- Do not use `pull_request_target` for build or release logic.
- Scope release permissions to the minimum needed, usually `contents: write`
  only in the tag-release job.
- Build version metadata from the tag and commit.
- Publish SHA256 checksums for every archive.
- Keep dependency caches disabled or dependency-keyed.
- Consider SHA-pinning GitHub Actions before making strong trusted-artifact
  claims.

## Recommended Gate Order

1. Complete task 2.3a as a written plugin/editor execution policy.
2. Perform task 3.1 against that policy before implementing 2.4, 2.5, or 2.6.
3. Implement `edit` only after process execution rules are settled.
4. Implement `migrate-beads` only after import, conflict, and rollback tests are
   specified.
5. Keep filtered `query` deferred and documented as future parity.
6. Keep `.tickets` settings prefix-only for the first settings slice.
