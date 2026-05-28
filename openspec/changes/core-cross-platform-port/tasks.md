## 1. Project Foundation

- [ ] 1.1 Initialize the Go module, `cmd/gtk` command entrypoint, package layout, formatter, and baseline README usage notes.
- [ ] 1.2 Add a minimal cross-platform test harness that can run on Linux, macOS, and Windows in CI.
- [ ] 1.3 Add golden `.tickets/` fixtures produced by upstream `tk` for normal compatibility workflows.
- [ ] 1.4 Add hand-authored edge fixtures for malformed, ambiguous, CRLF, and Windows-specific cases.

## 2. Ticket Storage Core

- [ ] 2.1 Implement ticket directory discovery with `TICKETS_DIR` override and parent walking.
- [ ] 2.1a Perform a focused path-security review for `TICKETS_DIR`, symlinked ticket roots, external roots, and write-target canonicalization before enabling write-capable commands.
- [ ] 2.2 Implement explicit `gtk init` that creates `.tickets/` only in the current directory and reports existing ticket roots safely.
- [ ] 2.3 Implement Markdown/YAML frontmatter parsing for known fields while tolerating unknown fields.
- [ ] 2.4 Implement stable ticket writing for create and mutation paths using same-directory atomic replacement.
- [ ] 2.5 Implement ticket ID generation and partial ID resolution with ambiguity errors.

## 3. MVP Commands

- [ ] 3.1 Implement `create` with title, description, design, acceptance, type, priority, assignee, external-ref, parent, and tags options, failing safely when `.tickets/` is missing.
- [ ] 3.2 Implement `show`, `ls`, and `list` with status, assignee, and type filters plus `list --json`.
- [ ] 3.3 Implement `start`, `close`, `reopen`, and `status` lifecycle commands.
- [ ] 3.4 Implement `dep`, `undep`, `link`, and `unlink` relationship commands.
- [ ] 3.5 Implement `ready` and `blocked` dependency classification.
- [ ] 3.6 Implement `add-note` using argument text or stdin, matching upstream note timestamps when fixture evidence confirms the format.
- [ ] 3.7 Return clear unsupported-command errors without plugin execution for unknown commands.

## 4. Cross-Platform Verification

- [ ] 4.1 Add tests for Windows path handling, CRLF/LF tolerance, and case-insensitive ambiguity behavior.
- [ ] 4.2 Add command-output tests for core workflows using temporary repositories.
- [ ] 4.3 Configure CI to run Go tests on Windows, macOS, and Linux.
- [ ] 4.4 Document MVP scope, omitted parity features, compatibility expectations, and attribution to `wedow/ticket`.
