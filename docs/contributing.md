# Contributing

`go-ticket` is a compatible reimplementation of `wedow/ticket`. Compatibility
work should preserve the `.tickets/` file format and clearly document any
intentional behavior differences.

## Before Changing Commands

- Update `docs/compatibility.md` when support status changes.
- Add or update command tests for user-visible behavior.
- Prefer semantic compatibility unless a downstream machine-readable contract
  requires exact output.
- Keep filtered `query` documented as partial until a reviewed filter strategy
  exists.

## Write-Surface Rules

- Ticket writes must go through `ticket.Write` or a reviewed helper with the
  same containment checks.
- Ticket IDs must pass `ValidateID`; upstream-compatible dotted suffixes such
  as `project-abcd.1` are allowed only between normal ID atoms.
- Target paths must pass `ResolveTicketPath`.
- New bulk operations must report partial failures and avoid overwriting by
  default.
- Tests should cover traversal, separators, Windows reserved names, symlinks,
  non-regular files, and external-root attempts when a feature touches paths.

## Process Execution Rules

- Do not add shell interpolation.
- Do not execute repo-controlled command paths.
- Keep plugin, editor, and future external tool execution aligned with
  `docs/plugins.md` and the active OpenSpec security review.

## Settings Rules

`.tickets/settings.json` is prefix-only for now. Do not add editor, plugin,
PATH, external root, or process policy settings without a separate security
review.

## Validation

Run:

```sh
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
openspec validate --specs --strict
```
