# Ticket Settings

`go-ticket` supports an optional `.tickets/settings.json` file.

The first supported setting is `prefix`:

```json
{
  "prefix": "gt"
}
```

When present, new ticket IDs use the configured prefix instead of deriving one
from the project directory name.

Security boundaries:

- settings must be a regular, non-symlink file under the active `.tickets`
  directory;
- the opened settings handle is revalidated against the checked path to catch
  local symlink swaps before reading;
- the file is limited to 4 KiB;
- unknown keys fail closed;
- `prefix` must be a simple ticket ID atom without hyphens;
- settings cannot configure editors, plugins, PATH behavior, external roots, or
  process execution policy.

If `.tickets/settings.json` is absent, `go-ticket` uses the upstream-compatible
default project-name prefix behavior.
