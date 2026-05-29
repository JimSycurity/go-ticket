# Plugin and Editor Security

`go-ticket` supports a narrowed plugin model for unknown commands. This is not
full shell parity with upstream `wedow/ticket`.

## Plugin Policy

- Builtin commands win by default.
- `gtk super <cmd> [args...]` dispatches builtins only and never runs plugins.
- Plugin command names must be simple command atoms with no path separators,
  dots, drive letters, whitespace, or shell metacharacters.
- Lookup scans absolute PATH directories in order, ignoring empty, relative, and
  current-directory entries.
- Candidate order is `tk-<cmd>` before `ticket-<cmd>`.
- Unix candidates must be executable regular non-symlink files.
- Windows candidates currently allow `.exe` only.
- `.cmd`, `.bat`, and `.ps1` plugins are deferred future parity until wrapper
  behavior is reviewed and tested.
- Plugins receive minimal ticket/project environment variables.
- Process execution uses argv directly and does not invoke a shell.

## Editor Policy

`gtk edit <id>` launches a user-configured editor after resolving the ticket
through the normal path checks.

Editor command sources are checked in this order:

1. `GTK_EDITOR`
2. `VISUAL`
3. `EDITOR`

The editor value must be a command name or absolute executable path. Inline
arguments, whitespace, shell metacharacters, and relative paths are rejected.
The ticket path is passed as one argv element.

Repository-controlled `.tickets/settings.json` cannot configure editors,
plugins, PATH behavior, external roots, or process execution policy.
