## Why

After the MVP proves a native Go `tk` can operate on compatible `.tickets/` repositories, the project needs confidence that it can replace `wedow/ticket` for real workflows. That requires broader feature parity, compatibility tests, polished docs, and repeatable versioned releases.

## What Changes

- Expand the Go implementation toward solid feature parity with `wedow/ticket`.
- Add compatibility coverage for plugins, `query`, `migrate-beads`, dependency trees, cycle detection, closed-ticket history, and editor integration.
- Build a release pipeline that produces signed or checksummed binaries for Windows, macOS, and Linux.
- Add a lightweight optional settings file under `.tickets/` for prefix and behavior customization while preserving upstream-compatible defaults when settings are absent.
- Add a test matrix and fixture strategy that compares behavior against upstream-compatible `.tickets/` repositories.
- Add shell completion generation and installation documentation as release polish.
- Improve documentation for installation, migration, compatibility scope, contribution, and release usage.
- Add explicit attribution to `wedow/ticket` and document the compatible-reimplementation relationship.

## Capabilities

### New Capabilities
- `feature-parity-compatibility`: Complete enough CLI/file-format compatibility for existing `wedow/ticket` users to switch with confidence.
- `release-readiness`: Automated builds, tests, versioned artifacts, documentation, and migration guidance for public releases.

### Modified Capabilities

## Impact

- Adds additional CLI commands and plugin/process execution behavior.
- Adds optional dependencies or embedded libraries for jq-like query behavior if chosen.
- Adds GitHub Actions or equivalent CI/CD workflows.
- Adds release documentation, compatibility matrix, migration docs, and generated artifact checksums.
- May require security review for plugin execution, editor launching, Beads migration input parsing, and PATH lookup behavior.
