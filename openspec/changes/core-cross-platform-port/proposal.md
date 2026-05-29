## Why

`wedow/ticket` is a useful git-native ticket workflow, but its Bash/coreutils implementation is awkward for native Windows users and agents. A Go reimplementation can preserve the `.tickets/` file format and CLI workflow while producing a single cross-platform MVP binary that can run side-by-side with upstream `tk`.

## What Changes

- Add a Go CLI that can be built and run on Windows, macOS, and Linux.
- Name the MVP executable `gtk` so users can test it side-by-side with upstream `tk`; defer a `tk` alias/drop-in binary to the parity/release-readiness phase.
- Preserve the `wedow/ticket` Markdown/YAML frontmatter storage model under `.tickets/`.
- Implement the core read/write commands needed for day-to-day ticket use: init, create, show, list, start, close, reopen, status, dependency add/remove, link/unlink, ready, blocked, and add-note.
- Provide `gtk list --json` as the MVP machine-readable integration surface for tools such as vscode-tk and agents.
- Include README attribution to `wedow/ticket` as the inspiration and compatibility target.
- Preserve parent-directory `.tickets/` discovery and `TICKETS_DIR` override behavior.
- Provide compatibility-focused tests that exercise shared `.tickets/` fixtures across platforms.
- Defer plugin override behavior, jq-compatible query filtering, Beads migration, and polished release artifacts to the follow-up parity change.

## Capabilities

### New Capabilities
- `cross-platform-ticket-cli`: Native Go `gtk` CLI that works on Windows, macOS, and Linux while reading and writing compatible ticket Markdown files.
- `core-ticket-workflows`: Core ticket lifecycle, relationship, listing, ready/blocked, and note workflows needed for normal use.

### Modified Capabilities

## Impact

- New Go module and command entrypoint.
- New parser/writer for `.tickets/*.md` frontmatter and Markdown bodies.
- New cross-platform filesystem and process behavior for ticket discovery, editor-less operation, and stdout/stderr output.
- CI will need at least a minimal Go test matrix for Windows, macOS, and Linux before the MVP is called usable.
