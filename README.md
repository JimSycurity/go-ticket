# go-ticket

Cross-platform Go reimplementation of
[`wedow/ticket`](https://github.com/wedow/ticket).

`go-ticket` preserves the git-native `.tickets/` Markdown/YAML workflow while
making the CLI usable on Windows, macOS, and Linux without Bash, coreutils, jq,
grep, or ripgrep for the supported command set.

The primary binary is currently named `gtk` so it can be tested side-by-side
with upstream `tk`.

## Current Features

- Initialize and discover `.tickets/` repositories, including `TICKETS_DIR`
  overrides under the reviewed path policy.
- Create, show, list, start, close, reopen, and set ticket status.
- Add/remove dependencies and links.
- Show ready and blocked tickets.
- Show dependency trees and dependency cycles.
- List recently closed tickets.
- Append notes.
- Emit `list --json` for local editor/tool integrations.
- Emit path-free `query` JSONL. Filtered query parity is intentionally deferred.
- Run reviewed unknown-command plugins and `super` builtin bypass.
- Open tickets with a validated user-configured editor.
- Import Beads JSONL with conflict/reporting behavior.
- Use optional prefix-only `.tickets/settings.json`.
- Build CI on Linux, macOS, and Windows, with tag-triggered release archives and
  checksums.

See [docs/compatibility.md](docs/compatibility.md) for the current parity
matrix and known limitations.

## Install

For local development:

```sh
go build -o /tmp/gtk ./cmd/gtk
/tmp/gtk version
```

For releases, download the archive for your OS and architecture from GitHub
Releases, verify the checksum in `SHA256SUMS`, then place `gtk` or `gtk.exe` on
your PATH.

## Basic Usage

```sh
gtk init
gtk create "Add Windows support"
gtk list
gtk ready
gtk blocked
gtk query
gtk show <id>
gtk start <id>
gtk add-note <id> "Implementation note"
gtk close <id>
```

`gtk list --json` includes absolute paths for editor integrations. Treat those
fields as local-machine metadata because they can expose usernames or repository
locations when copied into logs, tickets, or reports. `gtk query` intentionally
does not include path fields.

## Migration

Existing `.tickets/` repositories should be tested with read-only commands
before write commands:

```sh
gtk list
gtk ready
gtk blocked
gtk query
```

For Beads imports, start with:

```sh
gtk migrate-beads --dry-run
```

Then review the report before running `gtk migrate-beads`. See
[docs/migration.md](docs/migration.md) for rollback guidance.

## Development

```sh
go test ./...
go vet ./...
go build -o /tmp/gtk ./cmd/gtk
/tmp/gtk version
```

The test suite includes upstream `tk` generated fixtures, including a
comprehensive `.tickets/` fixture that covers types, priorities, statuses,
parents, dependencies, cycles, links, tags, notes, and structured sections. It
also includes a hand-authored mock comprehensive fixture with malformed ticket
examples, plus edge fixtures for portability, path safety, process execution,
and migration behavior.

Go embeds VCS metadata in normal local builds when it can see the repository.
For ad hoc smoke binaries, `gtk version` reports the version label, VCS
revision, dirty state, executable path, and executable mtime so `/tmp/gtk` can
prove which file is being smoke tested.

## Documentation

- [Compatibility matrix](docs/compatibility.md)
- [Migration and rollback](docs/migration.md)
- [Settings](docs/settings.md)
- [Plugin and editor security](docs/plugins.md)
- [Shell completion guidance](docs/shell-completion.md)
- [Contributing](docs/contributing.md)

## Attribution

This project is inspired by and targets file-format compatibility with
[`wedow/ticket`](https://github.com/wedow/ticket). `go-ticket` is a compatible
reimplementation, not an upstream-maintained port.
